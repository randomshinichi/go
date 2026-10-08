// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc"
)

// This file turns the lowered ppc32 ops into obj.Progs. The operand convention
// is the one obj/ppc's assembler reads back (see obj/ppc/link.go):
//
//	op rd,ra,rb   To=rd, From.Reg=ra, Reg=rb
//	op rd,ra      To=rd, From.Reg=ra
//	op rd,ra,si   To=rd, Reg=ra, From={TYPE_CONST, Offset=si}
//	op ra,rs,ui   To=ra, Reg=rs, From={TYPE_CONST, Offset=ui}
//
// Only the ops registered in _gen/PPCOps.go are handled; anything else reaching
// this switch is a bug in the rule set, and the default arm says so.

func ssaGenValue(s *ssagen.State, v *ssa.Value) {
	switch v.Op {
	case ssa.OpPPCADD:
		// add rd, ra, rb
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AADD)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCADDC:
		// addc rd, ra, rb : rd = ra + rb, XER[CA] = carry out.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AADDC)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCADDE:
		// adde rd, ra, rb : rd = ra + rb + XER[CA]. v.Args[2] is the Flags value
		// that addc (or an earlier adde) left in XER[CA]; it has no register.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AADDE)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCADDEC:
		// As ADDE, and the carry out is this value's second result.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AADDE)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCSUB:
		// subf rd, ra, rb computes rb - ra, so arg0 and arg1 swap places.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASUBF)
		p.To.Reg = r
		p.From.Reg = r2
		p.Reg = r1

	case ssa.OpPPCSUBC:
		// subfc rd, ra, rb = rb - ra, with XER[CA] = no-borrow.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASUBFC)
		p.To.Reg = r
		p.From.Reg = r2
		p.Reg = r1

	case ssa.OpPPCSUBE:
		// subfe rd, ra, rb = rb - ra - (1 - XER[CA]), which is x - y - borrow
		// when the preceding subfc computed x - y.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASUBFE)
		p.To.Reg = r
		p.From.Reg = r2
		p.Reg = r1

	case ssa.OpPPCSUBEC:
		// As SUBE, and the borrow out is this value's second result.
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASUBFE)
		p.To.Reg = r
		p.From.Reg = r2
		p.Reg = r1

	case ssa.OpPPCMULLW:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AMULLW)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCMULHW:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AMULHW)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCMULHWU:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AMULHWU)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCMULLU:
		// No single G3 instruction produces both halves of a 32x32 product, so
		// this op is two instructions: mulhwu for the high half, mullw for the
		// low one. v.Reg0() is the high half, v.Reg1() the low half.
		r1, r2 := v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AMULLW)
		p.To.Reg = v.Reg1()
		p.From.Reg = r1
		p.Reg = r2
		p = s.Prog(ppc.AMULHWU)
		p.To.Reg = v.Reg0()
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCNEG:
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.ANEG)
		p.To.Reg = r
		p.From.Reg = r1

	case ssa.OpPPCAND:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AAND)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCOR:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AOR)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCXOR:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.AXOR)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCSLW:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASLW)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCSRW:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASRW)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCSRAW:
		r, r1, r2 := v.Reg(), v.Args[0].Reg(), v.Args[1].Reg()
		p := s.Prog(ppc.ASRAW)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r2

	case ssa.OpPPCSRAWI:
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.ASRAWI)
		p.To.Reg = r
		p.Reg = r1
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}

	case ssa.OpPPCCNTLZW:
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.ACNTLZW)
		p.To.Reg = r
		p.From.Reg = r1

	case ssa.OpPPCEXTSB:
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.AEXTSB)
		p.To.Reg = r
		p.From.Reg = r1

	case ssa.OpPPCEXTSH:
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.AEXTSH)
		p.To.Reg = r
		p.From.Reg = r1

	case ssa.OpPPCMOVWreg:
		// or rd, rs, rs is the register move.
		r, r1 := v.Reg(), v.Args[0].Reg()
		p := s.Prog(ppc.AOR)
		p.To.Reg = r
		p.From.Reg = r1
		p.Reg = r1

	case ssa.OpPPCMOVWconst:
		emitMOVWconst(s, v)

	default:
		v.Fatalf("unhandled op %v", v.Op)
	}
}

// emitMOVWconst writes a 32-bit constant into a register. addi/addis sign-extend
// their 16-bit immediate, so a constant that does not fit needs the addis/ori
// pair; the (c+0x8000)>>16 bias is what makes the pair exact for negative values
// as well.
func emitMOVWconst(s *ssagen.State, v *ssa.Value) {
	c := int32(v.AuxInt)
	r := v.Reg()
	if c >= -32768 && c <= 32767 {
		p := s.Prog(ppc.AADDI)
		p.To.Reg = r
		p.Reg = ppc.REG_R0
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(c)}
		return
	}
	p := s.Prog(ppc.AADDIS)
	p.To.Reg = r
	p.Reg = ppc.REG_R0
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(uint32(c)+0x8000) >> 16}
	p = s.Prog(ppc.AORI)
	p.To.Reg = r
	p.Reg = r
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(uint32(c) & 0xffff)}
}

func ssaGenBlock(s *ssagen.State, b, next *ssa.Block) {
	switch b.Kind {
	case ssa.BlockPlain:
		// Fall through when the next block in layout order is the successor.
		if b.Succs[0].Block() != next {
			b.Fatalf("ppc32: a jump is needed here, and branch emission is not part of this slice: %s", b.LongString())
		}
	case ssa.BlockExit, ssa.BlockRet, ssa.BlockRetJmp:
		// blr returns through LR. Restoring the frame the prologue built belongs
		// to the frame-writing lane (brief 3.4), so this is the return only.
		s.Prog(ppc.ABLR)
	default:
		b.Fatalf("branch not implemented: %s", b.LongString())
	}
}
