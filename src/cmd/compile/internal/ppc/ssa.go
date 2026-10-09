// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc"
	"internal/abi"
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

	case ssa.OpPPCDIVWU, ssa.OpPPCROTLW:
		reg3(s, v.Op.Asm(), v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
	case ssa.OpPPCDIVW:
		// divw has undefined results for signed overflow. Negating the
		// numerator for every divisor -1 gives Go's wrapping result.
		p := s.Prog(ppc.ACMPWI)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[1].Reg()}
		p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: -1}
		minusOne := localBranch(s, 12, 2)
		reg3(s, ppc.ADIVW, v.Reg(), v.Args[0].Reg(), v.Args[1].Reg())
		over := s.Prog(ppc.AB)
		over.To.Type = obj.TYPE_BRANCH
		p = s.Prog(ppc.ANEG)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
		minusOne.To.SetTarget(p)
		p = s.Prog(ppc.AOR)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
		p.Reg = v.Reg()
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
		over.To.SetTarget(p)

	case ssa.OpPPCLoweredWB:
		p := s.Prog(obj.ACALL)
		p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ir.Syms.GCWriteBarrier[v.AuxInt-1]}

	case ssa.OpPPCLoweredPanicBounds, ssa.OpPPCLoweredPanicExtend:
		code, signed := ssa.BoundsKind(v.AuxInt).Code()
		x := int(v.Args[0].Reg() - ppc.REG_R3)
		y := int(v.Args[1].Reg() - ppc.REG_R3)
		sym := ir.Syms.PanicBounds
		if v.Op == ssa.OpPPCLoweredPanicExtend {
			x = x<<2 | y // high and low save-array slots
			y = int(v.Args[2].Reg() - ppc.REG_R3)
			sym = ir.Syms.PanicExtend
		}
		p := s.Prog(obj.APCDATA)
		p.From.SetConst(abi.PCDATA_PanicBounds)
		p.To.SetConst(int64(abi.BoundsEncode(code, signed, true, true, x, y)))
		p = s.Prog(obj.ACALL)
		p.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: sym}

	case ssa.OpPPCLoweredZero, ssa.OpPPCLoweredMove:
		emitBulkMemory(s, v)

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
	case ssa.OpCopy, ssa.OpPPCMOVWreg:
		if v.Type.IsMemory() || v.Reg() == v.Args[0].Reg() {
			return
		}
		// or rd, rs, rs is the register move, including regalloc copies.
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
		if v.Type.Size() == 1 && v.Type.IsSigned() {
			reg2(s, ppc.AEXTSB, v.Reg(), v.Reg())
		}
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
		if v.Op == ssa.OpPPCMOVBload {
			reg2(s, ppc.AEXTSB, v.Reg(), v.Reg())
		}

	case ssa.OpPPCMOVBstore, ssa.OpPPCMOVHstore, ssa.OpPPCMOVWstore:
		p := s.Prog(v.Op.Asm())
		p.To = obj.Addr{Type: obj.TYPE_MEM, Reg: v.Args[0].Reg()}
		ssagen.AddAux(&p.To, v)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[1].Reg()}

	case ssa.OpPPCMOVWconst:
		emitMOVWconst(s, v)

	case ssa.OpPPCCMPW, ssa.OpPPCCMPWU:
		p := s.Prog(v.Op.Asm())
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[1].Reg()}
	case ssa.OpPPCCMPWconst, ssa.OpPPCCMPWUconst:
		p := s.Prog(v.Op.Asm())
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		p.To = obj.Addr{Type: obj.TYPE_CONST, Offset: v.AuxInt}
	case ssa.OpPPCLoweredGetClosurePtr:
		// Darwin ppc32 closure context is passed in R11.
		p := s.Prog(ppc.AOR)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R11}
		p.Reg = ppc.REG_R11
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R11}
	case ssa.OpPPCLoweredGetCallerSP:
		p := s.Prog(ppc.AADDI)
		// NAME_PARAM includes the allocated frame in its resolved offset.
		// Subtract the linkage area to obtain the caller's actual SP.
		p.From = obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_PARAM, Offset: -base.Ctxt.Arch.FixedFrameSize}
		p.Reg = ppc.REGSP
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	case ssa.OpPPCLoweredGetCallerPC:
		p := s.Prog(obj.AGETCALLERPC)
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	case ssa.OpPPCLoweredNilCheck:
		// A faulting byte load tests the address without requiring alignment.
		// R31 is reserved to the assembler and never allocated to an SSA value.
		p := s.Prog(ppc.ALBZ)
		p.From = obj.Addr{Type: obj.TYPE_MEM, Reg: v.Args[0].Reg()}
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R31}
	case ssa.OpPPCInvertFlags, ssa.OpPPCFlagEQ, ssa.OpPPCFlagLT, ssa.OpPPCFlagGT:
		return
	case ssa.OpPPCEqual, ssa.OpPPCNotEqual, ssa.OpPPCLessThan, ssa.OpPPCLessEqual,
		ssa.OpPPCGreaterThan, ssa.OpPPCGreaterEqual:
		emitPredicate(s, v)
	case ssa.OpPPCCALLstatic:
		s.Call(v)
	case ssa.OpPPCCALLtail:
		s.TailCall(v)
	case ssa.OpPPCCALLclosure, ssa.OpPPCCALLinter:
		if v.Args[0].Reg() != ppc.REG_R12 {
			v.Fatalf("ppc32: indirect call target must be in R12: %v", v.LongString())
		}
		if v.Op == ssa.OpPPCCALLclosure && v.Args[1].Reg() != ppc.REG_R11 {
			v.Fatalf("ppc32: closure context must be in R11: %v", v.LongString())
		}
		p := s.Prog(ppc.AMTCTR)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Args[0].Reg()}
		// Keep Call's stack-map/argument bookkeeping, but branch through CTR
		// with link. obj/ppc must treat this as a non-leaf call too.
		p = s.Call(v)
		p.As = ppc.ABCTRL
		p.To = obj.Addr{}

	default:
		v.Fatalf("unhandled op %v", v.Op)
	}
}

// localBranch targets CR0; its caller must set a real instruction target.
func localBranch(s *ssagen.State, bo, bi int16) *obj.Prog {
	p := s.Prog(ppc.ABC)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(bo)}
	p.Reg = bi
	p.To.Type = obj.TYPE_BRANCH
	return p
}

// emitBulkMemory preserves pointer-store atomicity when alignment permits word
// accesses. Byte-aligned aggregates use byte accesses, never unaligned words.
// Move's SSA contract excludes partial overlap; identical pointers are harmless.
func emitBulkMemory(s *ssagen.State, v *ssa.Value) {
	size := v.AuxInt
	step := int64(1)
	load, store := ppc.ALBZ, ppc.ASTB
	if v.Aux.(*types.Type).Alignment() >= 4 {
		step, load, store = 4, ppc.ALWZ, ppc.ASTW
	}
	moving := v.Op == ssa.OpPPCLoweredMove
	access := func(off int64, ld, st obj.As) *obj.Prog {
		r := int16(ppc.REG_R0)
		var first *obj.Prog
		if moving {
			r = ppc.REG_R31
			first = s.Prog(ld)
			first.From = obj.Addr{Type: obj.TYPE_MEM, Reg: ppc.REG_R4, Offset: off}
			first.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
		}
		p := s.Prog(st)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: r}
		p.To = obj.Addr{Type: obj.TYPE_MEM, Reg: ppc.REG_R3, Offset: off}
		if first == nil {
			first = p
		}
		return first
	}
	advance := func(r int16) {
		p := s.Prog(ppc.AADDI)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: step}
		p.Reg = r
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
	}
	n := size / step
	if size > 32 {
		emitConst32(s, ppc.REG_R5, int32(n))
		start := access(0, load, store)
		advance(ppc.REG_R3)
		if moving {
			advance(ppc.REG_R4)
		}
		p := s.Prog(ppc.AADDI)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: -1}
		p.Reg = ppc.REG_R5
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R5}
		p = s.Prog(ppc.ACMPWI)
		p.From = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R5}
		p.To = obj.Addr{Type: obj.TYPE_CONST}
		localBranch(s, 4, 2).To.SetTarget(start)
		for off := int64(0); off < size%step; off++ {
			access(off, ppc.ALBZ, ppc.ASTB)
		}
		return
	}
	for off := int64(0); off < n*step; off += step {
		access(off, load, store)
	}
	for off := n * step; off < size; off++ {
		access(off, ppc.ALBZ, ppc.ASTB)
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
// pair. ori combines unsigned low bits: unlike addi, it needs no high-half bias.
func emitMOVWconst(s *ssagen.State, v *ssa.Value) {
	emitConst32(s, v.Reg(), int32(v.AuxInt))
}

func emitConst32(s *ssagen.State, r int16, c int32) {
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
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(int16(uint32(c) >> 16))}
	p = s.Prog(ppc.AORI)
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
	p.Reg = r
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(uint32(c) & 0xffff)}
}

// emitPredicate materializes a CR0 comparison when it survives block lowering,
// for example a returned bool. None of these instructions changes CR0 or XER.
func emitPredicate(s *ssagen.State, v *ssa.Value) {
	bo, bi := int16(4), int64(2) // branch if Equal is false
	switch v.Op {
	case ssa.OpPPCNotEqual:
		bo = 12
	case ssa.OpPPCLessThan:
		bi = 0
	case ssa.OpPPCLessEqual:
		bo, bi = 12, 1
	case ssa.OpPPCGreaterThan:
		bi = 1
	case ssa.OpPPCGreaterEqual:
		bo, bi = 12, 0
	}
	p := s.Prog(ppc.AADDI)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
	p.Reg = ppc.REG_R0
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	branch := s.Prog(ppc.ABC)
	branch.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(bo)}
	branch.Reg = int16(bi)
	branch.To.Type = obj.TYPE_BRANCH
	p = s.Prog(ppc.AADDI)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 1}
	p.Reg = ppc.REG_R0
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	// A real instruction gives the local branch a target even at block end.
	p = s.Prog(ppc.AOR)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	p.Reg = v.Reg()
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: v.Reg()}
	branch.To.SetTarget(p)
}

func ssaGenBlock(s *ssagen.State, b, next *ssa.Block) {
	branch := func(as ppc.Inst, bo, bi int32, target *ssa.Block) {
		p := s.Prog(as)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: int64(bo)}
		p.Reg = int16(bi)
		p.To.Type = obj.TYPE_BRANCH
		s.Branches = append(s.Branches, ssagen.Branch{P: p, B: target})
	}
	unconditional := func(target *ssa.Block) {
		p := s.Prog(ppc.AB)
		p.To.Type = obj.TYPE_BRANCH
		s.Branches = append(s.Branches, ssagen.Branch{P: p, B: target})
	}
	cond := func(bo, bi int32, yes, no *ssa.Block) {
		if yes == next {
			branch(ppc.ABC, bo^8, bi, no)
			return
		}
		branch(ppc.ABC, bo, bi, yes)
		if no != next {
			unconditional(no)
		}
	}
	multiCond := func(yes, no *ssa.Block, trueBits, falseBits []int32) {
		if yes == next {
			for _, bi := range falseBits {
				branch(ppc.ABC, 12, bi, no)
			}
			return
		}
		for _, bi := range trueBits {
			branch(ppc.ABC, 12, bi, yes)
		}
		if no != next {
			unconditional(no)
		}
	}
	var yes, no *ssa.Block
	if len(b.Succs) > 0 {
		yes = b.Succs[0].Block()
	}
	if len(b.Succs) > 1 {
		no = b.Succs[1].Block()
	}
	kind := b.Kind
	if b.Controls[0] != nil && b.Controls[0].Op == ssa.OpPPCInvertFlags {
		switch kind {
		case ssa.BlockPPCLT:
			kind = ssa.BlockPPCGT
		case ssa.BlockPPCGT:
			kind = ssa.BlockPPCLT
		case ssa.BlockPPCLE:
			kind = ssa.BlockPPCGE
		case ssa.BlockPPCGE:
			kind = ssa.BlockPPCLE
		}
	}
	switch kind {
	case ssa.BlockPlain:
		if yes != next {
			unconditional(yes)
		}
	case ssa.BlockExit:
	case ssa.BlockRet:
		// blr returns through LR. Frame restoration belongs to the prologue lane.
		s.Prog(ppc.ABLR)
	case ssa.BlockRetJmp:
		// CALLtail emitted RET symbol. obj/ppc restores SP/LR before the
		// non-linking branch; do not emit a second return here.
	case ssa.BlockPPCEQ:
		cond(12, 2, yes, no)
	case ssa.BlockPPCNE:
		cond(4, 2, yes, no)
	case ssa.BlockPPCLT, ssa.BlockPPCFLT:
		cond(12, 0, yes, no)
	case ssa.BlockPPCGE:
		cond(4, 0, yes, no)
	case ssa.BlockPPCFGE:
		// Ordered >= is GT or EQ; unordered (SO) must remain false.
		multiCond(yes, no, []int32{1, 2}, []int32{0, 3})
	case ssa.BlockPPCGT, ssa.BlockPPCFGT:
		cond(12, 1, yes, no)
	case ssa.BlockPPCLE:
		cond(4, 1, yes, no)
	case ssa.BlockPPCFLE:
		// Ordered <= is LT or EQ; unordered (SO) must remain false.
		multiCond(yes, no, []int32{0, 2}, []int32{1, 3})
	default:
		b.Fatalf("branch not implemented: %s", b.LongString())
	}
}
