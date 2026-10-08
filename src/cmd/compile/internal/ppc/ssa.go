// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/types"
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
	// Three-register forms: add, addc, adde, subf, subfc, subfe, mullw, mulhw,
	// mulhwu, and, or, xor, slw, srw, sraw.
	case ssa.OpPPCADD:
		reg3(s, ppc.AADD, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCADDC:
		// Two outputs (value, carry), so the value half is v.Reg0(); the carry
		// half lives in XER[CA] and has no register. adde reads CA back.
		reg3(s, ppc.AADDC, v.Reg0(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCADDE:
		reg3(s, ppc.AADDE, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCADDEC:
		reg3(s, ppc.AADDE, v.Reg0(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCSUB:
		// subf rd, ra, rb computes rb - ra, so arg0 and arg1 change places.
		reg3(s, ppc.ASUBF, v.Reg(), v.Args[1].Reg(), v.Args[0].Reg())
	case ssa.OpPPCSUBC:
		reg3(s, ppc.ASUBFC, v.Reg0(), v.Args[1].Reg(), v.Args[0].Reg())
	case ssa.OpPPCSUBE:
		// subfe rd, ra, rb = rb - ra - (1 - XER[CA]), which is x - y - borrow
		// when the preceding subfc computed x - y and left CA = no-borrow.
		reg3(s, ppc.ASUBFE, v.Reg(), v.Args[1].Reg(), v.Args[0].Reg())
	case ssa.OpPPCSUBEC:
		reg3(s, ppc.ASUBFE, v.Reg0(), v.Args[1].Reg(), v.Args[0].Reg())
	case ssa.OpPPCMULLW:
		reg3(s, ppc.AMULLW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCMULHW:
		reg3(s, ppc.AMULHW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCMULHWU:
		reg3(s, ppc.AMULHWU, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCAND:
		reg3(s, ppc.AAND, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCOR:
		reg3(s, ppc.AOR, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCXOR:
		reg3(s, ppc.AXOR, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCSLW:
		reg3(s, ppc.ASLW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCSRW:
		reg3(s, ppc.ASRW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCSRAW:
		reg3(s, ppc.ASRAW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())

	case ssa.OpPPCMULLU:
		// No single G3 instruction produces both halves of a 32x32 product, so
		// this op is two instructions: mullw for the low half into v.Reg1(),
		// mulhwu for the high half into v.Reg0().
		reg3(s, ppc.AMULLW, v.Reg1(), v.Args[0].Reg(), v.Args[1].Reg())
		reg3(s, ppc.AMULHWU, v.Reg0(), v.Args[0].Reg(), v.Args[1].Reg())

	// Two-register forms.
	case ssa.OpPPCNEG:
		reg2(s, ppc.ANEG, v.Reg(), v.Args[0].Reg())
	case ssa.OpPPCCNTLZW:
		reg2(s, ppc.ACNTLZW, v.Reg(), v.Args[0].Reg())
	case ssa.OpPPCEXTSB:
		reg2(s, ppc.AEXTSB, v.Reg(), v.Args[0].Reg())
	case ssa.OpPPCEXTSH:
		reg2(s, ppc.AEXTSH, v.Reg(), v.Args[0].Reg())
	case ssa.OpPPCMOVWreg:
		// or rd, rs, rs is the register move.
		reg3(s, ppc.AOR, v.Reg(), v.Args[0].Reg(), v.Args[0].Reg())

	// Shift by an immediate amount.
	case ssa.OpPPCSRAWI:
		p := s.Prog(ppc.ASRAWI)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
		p.Reg = v.Args[0].Reg()
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}

	// Spill and reload, inserted by the register allocator.
	case ssa.OpLoadReg:
		if v.Type.IsFlags() {
			v.Fatalf("ppc32: the carry lives in XER[CA] and cannot be spilled: %v", v.LongString())
		}
		p := s.Prog(loadByType(v.Type))
		ssagen.AddrAuto(&p.From, v.Args[0])
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	case ssa.OpStoreReg:
		if v.Type.IsFlags() {
			v.Fatalf("ppc32: the carry lives in XER[CA] and cannot be spilled: %v", v.LongString())
		}
		p := s.Prog(storeByType(v.Type))
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		ssagen.AddrAuto(&p.To, v)

	// Address formation and memory.
	case ssa.OpPPCMOVWaddr:
		// addi rd, base, off puts the offset in the immediate field. A symbol
		// needs the relocation path that brief 3.4 owns; obj/ppc's assembler
		// refuses one rather than emitting a wrong address.
		p := s.Prog(ppc.AADDI)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
		p.Reg = v.Args[0].Reg()
		p.From.Type = obj.TYPE_ADDR
		ssagen.AddAux(&p.From, v)
		if p.From.Sym == nil && p.From.Name == 0 {
			// No name to resolve, so this is an ordinary addi immediate; obj's
			// operand checker requires TYPE_ADDR to carry a name or a symbol.
			p.From.Type = obj.TYPE_CONST
		}

	case ssa.OpPPCMOVBload, ssa.OpPPCMOVBUload, ssa.OpPPCMOVHload,
		ssa.OpPPCMOVHUload, ssa.OpPPCMOVWload:
		p := s.Prog(v.Op.Asm())
		p.From = obj.Addr{Type: obj.TYPE_MEM, Reg: v.Args[0].Reg()}
		ssagen.AddAux(&p.From, v)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}

	case ssa.OpPPCMOVBstore, ssa.OpPPCMOVHstore, ssa.OpPPCMOVWstore:
		p := s.Prog(v.Op.Asm())
		p.To = obj.Addr{Type: obj.TYPE_MEM, Reg: v.Args[0].Reg()}
		ssagen.AddAux(&p.To, v)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[1].Reg()}

	case ssa.OpPPCMOVWconst:
		emitMOVWconst(s, v)

	default:
		v.Fatalf("unhandled op %v", v.Op)
	}
}

// reg2 and reg3 emit a two- and a three-register form in the operand convention
// obj/ppc's assembler reads back (obj/ppc/link.go): To=rd, From=ra, Reg=rb.
func reg2(s *ssagen.State, as obj.As, rd, ra int16) {
	p := s.Prog(as)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: rd}
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: ra}
}

func reg3(s *ssagen.State, as obj.As, rd, ra, rb int16) {
	p := s.Prog(as)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: rd}
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: ra}
	p.Reg = rb
}

// loadByType and storeByType name the instruction that moves a value of type t
// between a register and a spill slot. A 64-bit value is already two 32-bit
// halves by the time these are called.
func loadByType(t *types.Type) obj.As {
	switch t.Size() {
	case 1:
		if t.IsSigned() {
			return ppc.ALHA // lha sign-extends
		}
		return ppc.ALBZ
	case 2:
		if t.IsSigned() {
			return ppc.ALHA
		}
		return ppc.ALHZ
	case 4:
		return ppc.ALWZ
	}
	base.Fatalf("ppc32: loadByType: unsupported type %v", t)
	return 0
}

func storeByType(t *types.Type) obj.As {
	switch t.Size() {
	case 1:
		return ppc.ASTB
	case 2:
		return ppc.ASTH
	case 4:
		return ppc.ASTW
	}
	base.Fatalf("ppc32: storeByType: unsupported type %v", t)
	return 0
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
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
		p.Reg = ppc.REG_R0
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(c)}
		return
	}
	p := s.Prog(ppc.AADDIS)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
	p.Reg = ppc.REG_R0
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(uint32(c)+0x8000) >> 16}
	p = s.Prog(ppc.AORI)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
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
