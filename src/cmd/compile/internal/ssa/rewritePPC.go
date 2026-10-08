// Code generated from _gen/PPC.rules using 'go generate'; DO NOT EDIT.

package ssa

func rewriteValuePPC(v *Value) bool {
	switch v.Op {
	case OpAdd32:
		return rewriteValuePPC_OpAdd32(v)
	case OpAdd32carry:
		return rewriteValuePPC_OpAdd32carry(v)
	case OpAdd32carrywithcarry:
		return rewriteValuePPC_OpAdd32carrywithcarry(v)
	case OpAdd32withcarry:
		return rewriteValuePPC_OpAdd32withcarry(v)
	case OpAddr:
		return rewriteValuePPC_OpAddr(v)
	case OpAnd32:
		return rewriteValuePPC_OpAnd32(v)
	case OpConst32:
		return rewriteValuePPC_OpConst32(v)
	case OpConstBool:
		return rewriteValuePPC_OpConstBool(v)
	case OpConstNil:
		return rewriteValuePPC_OpConstNil(v)
	case OpCopy:
		return rewriteValuePPC_OpCopy(v)
	case OpLoad:
		return rewriteValuePPC_OpLoad(v)
	case OpLocalAddr:
		return rewriteValuePPC_OpLocalAddr(v)
	case OpLsh32x32:
		return rewriteValuePPC_OpLsh32x32(v)
	case OpLsh32x64:
		return rewriteValuePPC_OpLsh32x64(v)
	case OpMul32:
		return rewriteValuePPC_OpMul32(v)
	case OpMul32uhilo:
		return rewriteValuePPC_OpMul32uhilo(v)
	case OpNeg32:
		return rewriteValuePPC_OpNeg32(v)
	case OpOffPtr:
		return rewriteValuePPC_OpOffPtr(v)
	case OpOr32:
		return rewriteValuePPC_OpOr32(v)
	case OpRsh32Ux32:
		return rewriteValuePPC_OpRsh32Ux32(v)
	case OpRsh32Ux64:
		return rewriteValuePPC_OpRsh32Ux64(v)
	case OpRsh32x32:
		return rewriteValuePPC_OpRsh32x32(v)
	case OpRsh32x64:
		return rewriteValuePPC_OpRsh32x64(v)
	case OpStore:
		return rewriteValuePPC_OpStore(v)
	case OpSub32:
		return rewriteValuePPC_OpSub32(v)
	case OpSub32carry:
		return rewriteValuePPC_OpSub32carry(v)
	case OpSub32withcarry:
		return rewriteValuePPC_OpSub32withcarry(v)
	case OpXor32:
		return rewriteValuePPC_OpXor32(v)
	}
	return false
}
func rewriteValuePPC_OpAdd32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add32 x y)
	// result: (ADD x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADD)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAdd32carry(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add32carry x y)
	// result: (ADDC x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADDC)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAdd32carrywithcarry(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add32carrywithcarry x y c)
	// result: (ADDEC x y c)
	for {
		x := v_0
		y := v_1
		c := v_2
		v.reset(OpPPCADDEC)
		v.AddArg3(x, y, c)
		return true
	}
}
func rewriteValuePPC_OpAdd32withcarry(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add32withcarry x y c)
	// result: (ADDE x y c)
	for {
		x := v_0
		y := v_1
		c := v_2
		v.reset(OpPPCADDE)
		v.AddArg3(x, y, c)
		return true
	}
}
func rewriteValuePPC_OpAddr(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Addr {sym} base)
	// result: (MOVWaddr {sym} base)
	for {
		sym := auxToSym(v.Aux)
		base := v_0
		v.reset(OpPPCMOVWaddr)
		v.Aux = symToAux(sym)
		v.AddArg(base)
		return true
	}
}
func rewriteValuePPC_OpAnd32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (And32 x y)
	// result: (AND x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpConst32(v *Value) bool {
	// match: (Const32 [c])
	// result: (MOVWconst [c])
	for {
		c := auxIntToInt32(v.AuxInt)
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(c)
		return true
	}
}
func rewriteValuePPC_OpConstBool(v *Value) bool {
	// match: (ConstBool [t])
	// result: (MOVWconst [b2i32(t)])
	for {
		t := auxIntToBool(v.AuxInt)
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(b2i32(t))
		return true
	}
}
func rewriteValuePPC_OpConstNil(v *Value) bool {
	// match: (ConstNil)
	// result: (MOVWconst [0])
	for {
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
}
func rewriteValuePPC_OpCopy(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Copy x)
	// result: (MOVWreg x)
	for {
		x := v_0
		v.reset(OpPPCMOVWreg)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpLoad(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Load <t> ptr mem)
	// cond: t.IsBoolean()
	// result: (MOVBUload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.IsBoolean()) {
			break
		}
		v.reset(OpPPCMOVBUload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is8BitInt(t) && t.IsSigned()
	// result: (MOVBload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is8BitInt(t) && t.IsSigned()) {
			break
		}
		v.reset(OpPPCMOVBload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is8BitInt(t) && !t.IsSigned()
	// result: (MOVBUload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is8BitInt(t) && !t.IsSigned()) {
			break
		}
		v.reset(OpPPCMOVBUload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is16BitInt(t) && t.IsSigned()
	// result: (MOVHload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is16BitInt(t) && t.IsSigned()) {
			break
		}
		v.reset(OpPPCMOVHload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is16BitInt(t) && !t.IsSigned()
	// result: (MOVHUload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is16BitInt(t) && !t.IsSigned()) {
			break
		}
		v.reset(OpPPCMOVHUload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: (is32BitInt(t) || isPtr(t))
	// result: (MOVWload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is32BitInt(t) || isPtr(t)) {
			break
		}
		v.reset(OpPPCMOVWload)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValuePPC_OpLocalAddr(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (LocalAddr <t> {sym} base mem)
	// cond: t.Elem().HasPointers()
	// result: (MOVWaddr {sym} (SPanchored base mem))
	for {
		t := v.Type
		sym := auxToSym(v.Aux)
		base := v_0
		mem := v_1
		if !(t.Elem().HasPointers()) {
			break
		}
		v.reset(OpPPCMOVWaddr)
		v.Aux = symToAux(sym)
		v0 := b.NewValue0(v.Pos, OpSPanchored, typ.Uintptr)
		v0.AddArg2(base, mem)
		v.AddArg(v0)
		return true
	}
	// match: (LocalAddr <t> {sym} base _)
	// cond: !t.Elem().HasPointers()
	// result: (MOVWaddr {sym} base)
	for {
		t := v.Type
		sym := auxToSym(v.Aux)
		base := v_0
		if !(!t.Elem().HasPointers()) {
			break
		}
		v.reset(OpPPCMOVWaddr)
		v.Aux = symToAux(sym)
		v.AddArg(base)
		return true
	}
	return false
}
func rewriteValuePPC_OpLsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Lsh32x32 x y)
	// result: (SLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSLW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpLsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Lsh32x64 x y)
	// result: (SLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSLW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpMul32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mul32 x y)
	// result: (MULLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULLW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpMul32uhilo(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mul32uhilo x y)
	// result: (MULLU x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULLU)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpNeg32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Neg32 x)
	// result: (NEG x)
	for {
		x := v_0
		v.reset(OpPPCNEG)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpOffPtr(v *Value) bool {
	v_0 := v.Args[0]
	// match: (OffPtr [off] ptr)
	// result: (MOVWaddr [int32(off)] ptr)
	for {
		off := auxIntToInt64(v.AuxInt)
		ptr := v_0
		v.reset(OpPPCMOVWaddr)
		v.AuxInt = int32ToAuxInt(int32(off))
		v.AddArg(ptr)
		return true
	}
}
func rewriteValuePPC_OpOr32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or32 x y)
	// result: (OR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCOR)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpRsh32Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Rsh32Ux32 x y)
	// result: (SRW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpRsh32Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Rsh32Ux64 x y)
	// result: (SRW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpRsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Rsh32x32 x y)
	// result: (SRAW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpRsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Rsh32x64 x y)
	// result: (SRAW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpStore(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 1
	// result: (MOVBstore ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 1) {
			break
		}
		v.reset(OpPPCMOVBstore)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 2
	// result: (MOVHstore ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 2) {
			break
		}
		v.reset(OpPPCMOVHstore)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 4 && !t.IsFloat()
	// result: (MOVWstore ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 4 && !t.IsFloat()) {
			break
		}
		v.reset(OpPPCMOVWstore)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValuePPC_OpSub32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub32 x y)
	// result: (SUB x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUB)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpSub32carry(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub32carry x y)
	// result: (SUBC x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUBC)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpSub32withcarry(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub32withcarry x y c)
	// result: (SUBE x y c)
	for {
		x := v_0
		y := v_1
		c := v_2
		v.reset(OpPPCSUBE)
		v.AddArg3(x, y, c)
		return true
	}
}
func rewriteValuePPC_OpXor32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Xor32 x y)
	// result: (XOR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCXOR)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteBlockPPC(b *Block) bool {
	return false
}
