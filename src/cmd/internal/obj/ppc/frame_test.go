// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"encoding/binary"
	"testing"
)

func TestGoStackFrames(t *testing.T) {
	for _, tc := range []struct {
		name          string
		locals, frame int64
		call          bool
	}{
		{"frameless-leaf", 0, 0, false},
		{"leaf-local", 4, 32, false},
		{"minimum-nonleaf", 0, 32, true},
		{"rounded-nonleaf", 40, 64, true},
		{"large-local", 4096, 4128, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: tc.locals, Val: int32(4)}}
			load := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_PARAM, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
			ret := &obj.Prog{As: ABLR}
			text.Link, load.Link = load, ret
			if tc.call {
				call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}, Link: ret}
				load.Link = call
			}
			s := ctxt.Lookup("fixture")
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			if s.Func().Args != 4 || s.Func().Locals != int32(tc.locals) || s.Leaf() == tc.call {
				t.Fatalf("wrong frame metadata: args=%d locals=%d leaf=%v", s.Func().Args, s.Func().Locals, s.Leaf())
			}
			if text.To.Offset != tc.frame || tc.frame%16 != 0 {
				t.Fatalf("allocated frame %d, want aligned %d", text.To.Offset, tc.frame)
			}
			if off, err := resolveOffset(ctxt, load, &load.From, tc.frame); err != nil || off != tc.frame+24 {
				t.Fatalf("incoming parameter offset %d, %v: want frame+24", off, err)
			}
			var spadj int32
			for p := text.Link; p != nil; p = p.Link {
				if p.As == ABLR {
					if spadj != 0 {
						t.Fatalf("return with SP delta %d", spadj)
					}
				} else {
					spadj += p.Spadj
				}
			}
			assemble(ctxt, s, nil)
			word := func(i int) uint32 { return binary.BigEndian.Uint32(s.P[4*i:]) }
			if tc.call {
				// Independently assembled shape: mflr0; stw0,8(SP); li0,0;
				// stwu SP,-N(SP). LR store precedes allocation, not follows it.
				want := []uint32{0x7c0802a6, 0x90010008, 0x38000000, 0x94210000 | uint32(uint16(-tc.frame))}
				for i, w := range want {
					if word(i) != w {
						t.Errorf("prologue word %d = %08x, want %08x", i, word(i), w)
					}
				}
				n := len(s.P) / 4
				wantRet := []uint32{0x38210000 | uint32(tc.frame), 0x83e10008, 0x7fe803a6, 0x4e800020}
				for i, w := range wantRet {
					if word(n-4+i) != w {
						t.Errorf("epilogue word %d = %08x, want %08x", i, word(n-4+i), w)
					}
				}
			} else if tc.frame == 0 && (len(s.P) != 8 || word(0) != 0x80610018 || word(1) != 0x4e800020) {
				t.Fatalf("frameless leaf has unexpected prologue: %x", s.P)
			}
		})
	}
}

func TestLargeFramePrologue(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	diagnosed := false
	ctxt.DiagFunc = func(string, ...interface{}) { diagnosed = true }
	s := ctxt.Lookup("large")
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: 32768}}
	text.Link = &obj.Prog{As: obj.ARET}
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	if diagnosed {
		t.Fatal("large frame rejected")
	}
	var foundIndexedAllocation bool
	for p := text.Link; p != nil; p = p.Link {
		if p.As == ASTWUX {
			foundIndexedAllocation = true
		}
	}
	if !foundIndexedAllocation {
		t.Fatal("large frame did not use indexed stack allocation")
	}
}
