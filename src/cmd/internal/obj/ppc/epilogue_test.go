// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/src"
	"encoding/binary"
	"fmt"
	"testing"
)

// epilogueFixture runs the real assembler and PCSP encoder, not a table derived
// by the assertion from the frame-release instruction it is checking.
func epilogueFixture(t *testing.T, locals int64, tail, broken bool) *obj.LSym {
	t.Helper()
	arch := Linkppc
	if broken {
		arch.Preprocess = func(ctxt *obj.Link, s *obj.LSym, alloc obj.ProgAlloc) {
			preprocess(ctxt, s, alloc)
			// Reintroduce the old small-frame shape, including its SP metadata:
			// ADDI SP,frame; LWZ R31,8(SP); MTLR R31; return.
			for p := s.Func().Text.Link; p != nil; p = p.Link {
				if p.As == ALWZ && p.To.Reg == REG_R31 && p.From.Reg == REGSP {
					mtlr, pop := p.Link, p.Link.Link
					if mtlr.As != AMTLR || pop.As != AADDI || pop.To.Reg != REGSP {
						t.Fatal("negative control did not find LR-before-SP epilogue")
					}
					p.As, p.From, p.Reg, p.To, p.Spadj = pop.As, pop.From, pop.Reg, pop.To, pop.Spadj
					mtlr.As, mtlr.From, mtlr.Reg, mtlr.To = ALWZ, obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: 8}, 0, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}
					pop.As, pop.From, pop.Reg, pop.To, pop.Spadj = AMTLR, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}, 0, obj.Addr{}, 0
					return
				}
			}
			t.Fatal("negative control did not mutate an epilogue")
		}
	}
	ctxt := obj.Linknew(&arch)
	ctxt.Pkgpath = "epiloguetest"
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("epiloguetest.fixture")
	pos := ctxt.PosTable.XPos(src.MakePos(src.NewFileBase("epilogue.s", "/epilogue.s"), 1, 1))
	ctxt.InitTextSym(s, obj.NOSPLIT, pos)
	text := &obj.Prog{As: obj.ATEXT, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: s}, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals}}
	ret := &obj.Prog{As: obj.ARET}
	if tail {
		ret.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("tail")}
	}
	text.Link = &obj.Prog{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}, Link: ret}
	text.Pos = pos
	s.Func().Text = text
	obj.Flushplist(ctxt, &obj.Plist{Firstpc: text}, func() *obj.Prog { return new(obj.Prog) })
	return s
}

// epilogueStateErrors executes the emitted register writes and reads the actual
// PCSP table at every instruction boundary after allocation. LR must name the
// entry caller whenever the table says the frame is gone, regardless of unsafe
// markers (SIGPROF can sample even PCs excluded from async preemption).
func epilogueStateErrors(s *obj.LSym) []string {
	const entrySP, entryLR = uint32(0x100000), uint32(0x12345678)
	var regs [32]uint32
	regs[1] = entrySP
	lr := entryLR
	mem := map[uint32]uint32{}
	allocated := false
	var errors []string
	it := obj.PCIter{PCScale: 4}
	it.Init(s.Func().Pcln.Pcsp.P)
	for pc := uint32(0); pc < uint32(len(s.P)); pc += 4 {
		for !it.Done && pc >= it.NextPC {
			it.Next()
		}
		if it.Done {
			return append(errors, fmt.Sprintf("PCSP missing at +%#x", pc))
		}
		if allocated {
			actual := int32(entrySP - regs[1])
			if it.Value != actual {
				errors = append(errors, fmt.Sprintf("+%#x: PCSP=%d actual=%d", pc, it.Value, actual))
			}
			if it.Value == 0 && lr != entryLR {
				errors = append(errors, fmt.Sprintf("+%#x: frame gone with LR=%#x, want entry LR=%#x", pc, lr, entryLR))
			}
		}
		w := binary.BigEndian.Uint32(s.P[pc:])
		rt, ra, rb := (w>>21)&31, (w>>16)&31, (w>>11)&31
		d := uint32(int32(int16(w)))
		switch w >> 26 {
		case 14, 15: // addi/addis (RA=0 means literal zero)
			base := uint32(0)
			if ra != 0 {
				base = regs[ra]
			}
			if w>>26 == 15 {
				d <<= 16
			}
			regs[rt] = base + d
		case 32: // lwz
			regs[rt] = mem[regs[ra]+d]
		case 36: // stw
			mem[regs[ra]+d] = regs[rt]
		case 37: // stwu
			addr := regs[ra] + d
			mem[addr], regs[ra] = regs[rt], addr
		case 18: // bl overwrites LR, non-link tail b does not
			if w&1 != 0 {
				lr = pc + 4
			}
		case 31:
			switch (w >> 1) & 1023 {
			case 339: // mflr
				regs[rt] = lr
			case 467: // mtlr
				lr = regs[rt]
			case 266: // add
				regs[rt] = regs[ra] + regs[rb]
			case 183: // stwux
				addr := regs[ra] + regs[rb]
				mem[addr], regs[ra] = regs[rt], addr
			default:
				return append(errors, fmt.Sprintf("unmodelled word %08x at +%#x", w, pc))
			}
		case 19: // blr
		default:
			return append(errors, fmt.Sprintf("unmodelled word %08x at +%#x", w, pc))
		}
		allocated = allocated || regs[1] != entrySP
	}
	return errors
}

func TestEpilogueLRBeforeSP(t *testing.T) {
	for _, locals := range []int64{0, 40, 32728, 32744, 65544} {
		for _, tail := range []bool{false, true} {
			t.Run(fmt.Sprintf("locals-%d/tail-%v", locals, tail), func(t *testing.T) {
				if errors := epilogueStateErrors(epilogueFixture(t, locals, tail, false)); len(errors) != 0 {
					t.Fatalf("epilogue state: %v", errors)
				}
			})
		}
	}
}

func TestEpilogueRejectsPopBeforeLR(t *testing.T) {
	errors := epilogueStateErrors(epilogueFixture(t, 40, false, true))
	if len(errors) != 2 {
		t.Fatalf("old ordering: got %v, want two frame-gone/stale-LR PCs", errors)
	}
	t.Logf("broken-shape control rejected: %v", errors)
}
