// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/src"
	"encoding/binary"
	"fmt"
	"internal/abi"
	"testing"
)

// systemstackTailFixture exercises RET (CTR) in a function with dynamic SP
// writes, where source-level constant adjustments must remain undescribed.
func systemstackTailFixture(t *testing.T, locals int64, noframe, broken bool) *obj.LSym {
	t.Helper()
	arch := Linkppc
	if broken {
		arch.Preprocess = func(ctxt *obj.Link, s *obj.LSym, alloc obj.ProgAlloc) {
			preprocess(ctxt, s, alloc)
			for p := s.Func().Text.Link; p != nil; p = p.Link {
				if p.Link != nil && p.Link.As == obj.AJMP && p.Link.To.Reg == REG_CTR {
					// Reproduce the unannotated handwritten pop.
					if p.Spadj >= 0 {
						t.Fatal("negative control did not find the frame pop")
					}
					p.Spadj = 0
					return
				}
			}
			t.Fatal("negative control did not find RET (CTR)")
		}
	}
	ctxt := obj.Linknew(&arch)
	ctxt.Pkgpath = "systemstacktest"
	ctxt.IsAsm = true
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("systemstacktest.fixture")
	pos := ctxt.PosTable.XPos(src.MakePos(src.NewFileBase("systemstack.s", "/systemstack.s"), 1, 1))
	flags := obj.NOSPLIT
	if noframe {
		flags |= obj.NOFRAME
	}
	ctxt.InitTextSym(s, flags, pos)
	text := &obj.Prog{As: obj.ATEXT, Pos: pos, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: s}, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals}}
	call := &obj.Prog{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}}
	push := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: -32}, Reg: REGSP, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}}
	switchSP := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: 24}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}}
	tail := &obj.Prog{As: obj.ARET, To: obj.Addr{Type: obj.TYPE_MEM, Reg: REG_CTR}}
	ret := &obj.Prog{As: obj.ARET}
	text.Link, call.Link, push.Link, switchSP.Link, tail.Link = call, push, switchSP, tail, ret
	s.Func().Text = text
	obj.Flushplist(ctxt, &obj.Plist{Firstpc: text}, func() *obj.Prog { return new(obj.Prog) })
	if push.Spadj != 0 || s.Func().FuncFlag&abi.FuncFlagSPWrite == 0 {
		t.Fatal("RET (CTR) changed dynamic-SP suppression")
	}
	return s
}

func systemstackReturnDeltas(s *obj.LSym) (tail, ret []int32) {
	it := obj.PCIter{PCScale: 4}
	it.Init(s.Func().Pcln.Pcsp.P)
	for pc := uint32(0); pc < uint32(len(s.P)); pc += 4 {
		for !it.Done && pc >= it.NextPC {
			it.Next()
		}
		switch binary.BigEndian.Uint32(s.P[pc:]) {
		case 0x4e800420: // BCCTR, LK clear
			tail = append(tail, it.Value)
		case 0x4e800020: // BLR
			ret = append(ret, it.Value)
		}
	}
	return
}

func TestSystemstackTailPCSP(t *testing.T) {
	for _, tc := range []struct {
		locals  int64
		noframe bool
	}{{0, false}, {40, false}, {32744, false}, {0, true}} {
		t.Run(fmt.Sprintf("locals-%d/noframe-%v", tc.locals, tc.noframe), func(t *testing.T) {
			tail, ret := systemstackReturnDeltas(systemstackTailFixture(t, tc.locals, tc.noframe, false))
			if len(tail) != 1 || tail[0] != 0 || len(ret) != 1 || ret[0] != 0 {
				t.Fatalf("restored-SP branches: BCCTR deltas=%v BLR deltas=%v, want [0] for each", tail, ret)
			}
		})
	}
}

func TestSystemstackTailPCSPRejectsMissingSpadj(t *testing.T) {
	tail, _ := systemstackReturnDeltas(systemstackTailFixture(t, 0, false, true))
	if len(tail) != 1 || tail[0] != 32 {
		t.Fatalf("broken pop: BCCTR deltas=%v, want stale delta [32]", tail)
	}
}
