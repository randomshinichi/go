// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"encoding/binary"
	"testing"
)

func TestIndirectCallAndTailFrames(t *testing.T) {
	for _, tc := range []struct {
		name     string
		locals   int64
		indirect bool
		frame    int64
	}{
		{"frameless-tail", 0, false, 0},
		{"local-tail", 4, false, 32},
		{"indirect-nonleaf-tail", 0, true, 32},
		{"indirect-local-tail", 40, true, 64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			target := ctxt.Lookup("tail_target")
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: tc.locals}}
			ret := &obj.Prog{As: obj.ARET, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: target, Offset: 4}}
			text.Link = ret
			if tc.indirect {
				mtctr := &obj.Prog{As: AMTCTR, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R12}}
				call := &obj.Prog{As: ABCTRL}
				text.Link, mtctr.Link, call.Link = mtctr, call, ret
			}
			s := ctxt.Lookup("fixture")
			s.Set(obj.AttrNoSplit, true)
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			if text.To.Offset != tc.frame || s.Leaf() == tc.indirect {
				t.Fatalf("frame=%d leaf=%v", text.To.Offset, s.Leaf())
			}
			var delta int32
			for p := text.Link; p != nil; p = p.Link {
				if p.As == AB {
					if delta != 0 || p.To.Sym != target || p.To.Offset != 4 {
						t.Fatalf("tail branch: SP delta=%d target=%v", delta, p.To)
					}
				} else {
					delta += p.Spadj
				}
			}
			assemble(ctxt, s, nil)
			n := len(s.P) / 4
			word := func(i int) uint32 { return binary.BigEndian.Uint32(s.P[4*i:]) }
			if word(n-1) != 0x48000000 {
				t.Fatalf("tail branch links or returns: %08x", word(n-1))
			}
			if tc.indirect {
				want := []uint32{0x7d8903a6, 0x4e800421, 0x38210000 | uint32(tc.frame), 0x83e10008, 0x7fe803a6, 0x48000000}
				for i, w := range want {
					if word(n-len(want)+i) != w {
						t.Errorf("word %d = %08x want %08x", i, word(n-len(want)+i), w)
					}
				}
			} else if tc.frame != 0 && word(n-2) != 0x38210000|uint32(tc.frame) {
				t.Fatal("local tail lost SP restoration")
			}
			if len(s.R) != 1 || s.R[0] != (obj.Reloc{Off: int32(4 * (n - 1)), Siz: 4, Type: objabi.R_CALLPOWER, Sym: target, Add: 4}) {
				t.Fatalf("tail relocations=%+v", s.R)
			}
		})
	}
}

func TestRejectUnalignedTailAddend(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	diagnosed := false
	ctxt.DiagFunc = func(string, ...interface{}) { diagnosed = true }
	p := &obj.Prog{As: AB, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee"), Offset: 2}}
	if _, ok := encodeProg(ctxt, p, 0); ok || !diagnosed {
		t.Fatal("unaligned tail addend accepted")
	}
}
