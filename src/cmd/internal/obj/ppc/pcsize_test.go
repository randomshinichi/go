// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/src"
	"encoding/binary"
	"fmt"
	"testing"
)

// Expansions of stack operands must advance PCs, not just the output buffer.
// Otherwise PCSP drops the frame early and relocations patch unrelated words.
func TestExpandedStackOperandPCs(t *testing.T) {
	for _, locals := range []int64{40, 32728, 32744, 44264, 65512} {
		t.Run(fmt.Sprintf("locals-%d", locals), func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.Pkgpath = "fixture"
			ctxt.Headtype = objabi.Hlinux
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			s := ctxt.Lookup("fixture.expanded")
			pos := ctxt.PosTable.XPos(src.MakePos(src.NewFileBase("fixture.s", "fixture.s"), 1, 0))
			ctxt.InitTextSym(s, obj.NOSPLIT, pos)
			text := &obj.Prog{As: obj.ATEXT, Pos: pos, From: obj.Addr{Sym: s}, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals, Val: int32(0)}}
			s.Func().Text = text
			reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
			ret := &obj.Prog{As: obj.ARET}
			branch := &obj.Prog{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}}
			branch.To.SetTarget(ret)
			addr := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_AUTO, Offset: -8}, Reg: REGSP, To: reg(REG_R3)}
			load := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_PARAM, Reg: REGSP}, To: reg(REG_R4)}
			store := &obj.Prog{As: ASTW, From: reg(REG_R4), To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_AUTO, Reg: REGSP, Offset: -4}}
			call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("fixture.callee")}}
			marker := &obj.Prog{As: obj.APCDATA, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 1}, To: obj.Addr{Type: obj.TYPE_CONST, Offset: 7}}
			text.Link, branch.Link, addr.Link, load.Link, store.Link, call.Link, marker.Link = branch, addr, load, store, call, marker, ret
			obj.Flushplist(ctxt, &obj.Plist{Firstpc: text}, nil)

			// Re-encode each Prog to count its actual output independently of Pc.
			var actual, callPC, branchPC, targetPC int64
			for p := text.Link; p != nil; p = p.Link {
				if p == call {
					callPC = actual
				}
				if p == branch {
					branchPC = actual
				}
				if p == ret {
					targetPC = actual
				}
				if p.Pc != actual {
					t.Errorf("%v: Pc=%#x, actual byte offset=%#x", p.As, p.Pc, actual)
				}
				w, ok := encodeProg(ctxt, p, text.To.Offset)
				if !ok {
					t.Fatal("encode failed")
				}
				actual += int64(4 * len(w))
			}
			if actual != s.Size {
				t.Fatalf("counted %d bytes, symbol size %d", actual, s.Size)
			}
			if len(s.R) != 1 || int64(s.R[0].Off) != callPC {
				t.Errorf("call relocation=%v, actual call PC=%#x", s.R, callPC)
			}
			w := binary.BigEndian.Uint32(s.P[branchPC:])
			if got := branchPC + int64(int16(w&0xfffc)); got != targetPC {
				t.Errorf("branch targets %#x, want real epilogue %#x", got, targetPC)
			}
			pcdata := obj.NewPCIter(4)
			pcdata.Init(s.Func().Pcln.Pcdata[1].P)
			for !pcdata.Done && pcdata.Value != 7 {
				pcdata.Next()
			}
			if pcdata.Done || int64(pcdata.PC) != callPC+4 {
				t.Errorf("PCDATA marker starts at %#x, want %#x", pcdata.PC, callPC+4)
			}

			// Scan real words for the SP release, then decode the emitted PCSP.
			var release uint32
			for i := 0; i < len(s.P); i += 4 {
				w := binary.BigEndian.Uint32(s.P[i:])
				if w == 0x7c21fa14 || w>>16 == 0x3821 {
					release = uint32(i)
				}
			}
			if release == 0 {
				t.Fatal("missing SP release")
			}
			sp := obj.NewPCIter(4)
			sp.Init(s.Func().Pcln.Pcsp.P)
			for !sp.Done {
				for pc := sp.PC; pc < sp.NextPC; pc += 4 {
					if pc >= uint32(callPC) && pc <= release && sp.Value != int32(text.To.Offset) {
						t.Errorf("pcsp(%#x)=%d, live frame=%d (release %#x)", pc, sp.Value, text.To.Offset, release)
					}
					if pc == release+4 && sp.Value != 0 {
						t.Errorf("pcsp after release=%d, want 0", sp.Value)
					}
				}
				sp.Next()
			}
		})
	}
}

// A branch that looked short under the old one-word estimate needs relaxation
// once each out-of-range memory operand is counted as four words.
func TestExpandedStackOperandBranchRelaxation(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: 40000}}
	branch := &obj.Prog{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	text.Link = branch
	prev := branch
	for i := 0; i < 2100; i++ {
		p := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_AUTO, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
		prev.Link = p
		prev = p
	}
	ret := &obj.Prog{As: ABLR}
	prev.Link = ret
	branch.To.SetTarget(ret)
	s := ctxt.Lookup("expanded-conditional")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	if s.Size != 16*2100+16 {
		t.Fatalf("relaxed size=%d, want %d", s.Size, 16*2100+16)
	}
	if got := binary.BigEndian.Uint32(s.P); got != 0x40820008 {
		t.Errorf("conditional=%08x, want inverted skip 40820008", got)
	}
	long := binary.BigEndian.Uint32(s.P[4:])
	if long>>26 != 18 || int64(long&0x03fffffc)+4 != s.Size-4 {
		t.Errorf("long branch=%08x does not reach final BLR", long)
	}
}
