// Code generated from _gen/PPC.rules using 'go generate'; DO NOT EDIT.

package ssa

import "cmd/compile/internal/types"

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
	case OpEq32:
		return rewriteValuePPC_OpEq32(v)
	case OpGetCallerPC:
		v.Op = OpPPCLoweredGetCallerPC
		return true
	case OpGetCallerSP:
		v.Op = OpPPCLoweredGetCallerSP
		return true
	case OpLeq32:
		return rewriteValuePPC_OpLeq32(v)
	case OpLeq32U:
		return rewriteValuePPC_OpLeq32U(v)
	case OpLess32:
		return rewriteValuePPC_OpLess32(v)
	case OpLess32U:
		return rewriteValuePPC_OpLess32U(v)
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
	case OpNeq32:
		return rewriteValuePPC_OpNeq32(v)
	case OpOffPtr:
		return rewriteValuePPC_OpOffPtr(v)
	case OpOr32:
		return rewriteValuePPC_OpOr32(v)
	case OpPPCCMPW:
		return rewriteValuePPC_OpPPCCMPW(v)
	case OpPPCCMPWU:
		return rewriteValuePPC_OpPPCCMPWU(v)
	case OpPPCCMPWUconst:
		return rewriteValuePPC_OpPPCCMPWUconst(v)
	case OpPPCCMPWconst:
		return rewriteValuePPC_OpPPCCMPWconst(v)
	case OpPPCEqual:
		return rewriteValuePPC_OpPPCEqual(v)
	case OpPPCGreaterEqual:
		return rewriteValuePPC_OpPPCGreaterEqual(v)
	case OpPPCGreaterThan:
		return rewriteValuePPC_OpPPCGreaterThan(v)
	case OpPPCLessEqual:
		return rewriteValuePPC_OpPPCLessEqual(v)
	case OpPPCLessThan:
		return rewriteValuePPC_OpPPCLessThan(v)
	case OpPPCNotEqual:
		return rewriteValuePPC_OpPPCNotEqual(v)
	case OpRsh32Ux32:
		return rewriteValuePPC_OpRsh32Ux32(v)
	case OpRsh32Ux64:
		return rewriteValuePPC_OpRsh32Ux64(v)
	case OpRsh32x32:
		return rewriteValuePPC_OpRsh32x32(v)
	case OpRsh32x64:
		return rewriteValuePPC_OpRsh32x64(v)
	case OpStaticCall:
		v.Op = OpPPCCALLstatic
		return true
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
func rewriteValuePPC_OpEq32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Eq32 x y)
	// result: (Equal (CMPW x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLeq32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Leq32 x y)
	// result: (LessEqual (CMPW x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLeq32U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Leq32U x y)
	// result: (LessEqual (CMPWU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Less32 x y)
	// result: (LessThan (CMPW x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess32U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Less32U x y)
	// result: (LessThan (CMPWU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
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
func rewriteValuePPC_OpNeq32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Neq32 x y)
	// result: (NotEqual (CMPW x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
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
func rewriteValuePPC_OpPPCCMPW(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (CMPW (MOVWconst [c]) y)
	// cond: is16Bit(int64(c))
	// result: (InvertFlags (CMPWconst y [int32(c)]))
	for {
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		c := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(is16Bit(int64(c))) {
			break
		}
		v.reset(OpPPCInvertFlags)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWconst, types.TypeFlags)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v0.AddArg(y)
		v.AddArg(v0)
		return true
	}
	// match: (CMPW x (MOVWconst [c]))
	// cond: is16Bit(int64(c))
	// result: (CMPWconst x [int32(c)])
	for {
		x := v_0
		if v_1.Op != OpPPCMOVWconst {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(is16Bit(int64(c))) {
			break
		}
		v.reset(OpPPCCMPWconst)
		v.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg(x)
		return true
	}
	// match: (CMPW x y)
	// cond: canonLessThan(x,y)
	// result: (InvertFlags (CMPW y x))
	for {
		x := v_0
		y := v_1
		if !(canonLessThan(x, y)) {
			break
		}
		v.reset(OpPPCInvertFlags)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v0.AddArg2(y, x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCCMPWU(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (CMPWU (MOVWconst [c]) y)
	// cond: isU16Bit(int64(c))
	// result: (InvertFlags (CMPWUconst y [int32(c)]))
	for {
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		c := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(isU16Bit(int64(c))) {
			break
		}
		v.reset(OpPPCInvertFlags)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v0.AddArg(y)
		v.AddArg(v0)
		return true
	}
	// match: (CMPWU x (MOVWconst [c]))
	// cond: isU16Bit(int64(c))
	// result: (CMPWUconst x [int32(c)])
	for {
		x := v_0
		if v_1.Op != OpPPCMOVWconst {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(isU16Bit(int64(c))) {
			break
		}
		v.reset(OpPPCCMPWUconst)
		v.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg(x)
		return true
	}
	// match: (CMPWU x y)
	// cond: canonLessThan(x,y)
	// result: (InvertFlags (CMPWU y x))
	for {
		x := v_0
		y := v_1
		if !(canonLessThan(x, y)) {
			break
		}
		v.reset(OpPPCInvertFlags)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v0.AddArg2(y, x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCCMPWUconst(v *Value) bool {
	v_0 := v.Args[0]
	// match: (CMPWUconst (MOVWconst [x]) [y])
	// cond: uint32(x)==uint32(y)
	// result: (FlagEQ)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(uint32(x) == uint32(y)) {
			break
		}
		v.reset(OpPPCFlagEQ)
		return true
	}
	// match: (CMPWUconst (MOVWconst [x]) [y])
	// cond: uint32(x)<uint32(y)
	// result: (FlagLT)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(uint32(x) < uint32(y)) {
			break
		}
		v.reset(OpPPCFlagLT)
		return true
	}
	// match: (CMPWUconst (MOVWconst [x]) [y])
	// cond: uint32(x)>uint32(y)
	// result: (FlagGT)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(uint32(x) > uint32(y)) {
			break
		}
		v.reset(OpPPCFlagGT)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCCMPWconst(v *Value) bool {
	v_0 := v.Args[0]
	// match: (CMPWconst (MOVWconst [x]) [y])
	// cond: int32(x)==int32(y)
	// result: (FlagEQ)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(int32(x) == int32(y)) {
			break
		}
		v.reset(OpPPCFlagEQ)
		return true
	}
	// match: (CMPWconst (MOVWconst [x]) [y])
	// cond: int32(x)<int32(y)
	// result: (FlagLT)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(int32(x) < int32(y)) {
			break
		}
		v.reset(OpPPCFlagLT)
		return true
	}
	// match: (CMPWconst (MOVWconst [x]) [y])
	// cond: int32(x)>int32(y)
	// result: (FlagGT)
	for {
		y := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPCMOVWconst {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if !(int32(x) > int32(y)) {
			break
		}
		v.reset(OpPPCFlagGT)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCEqual(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Equal (FlagEQ))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (Equal (FlagLT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Equal (FlagGT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Equal (InvertFlags x))
	// result: (Equal x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCEqual)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCGreaterEqual(v *Value) bool {
	v_0 := v.Args[0]
	// match: (GreaterEqual (FlagEQ))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (GreaterEqual (FlagLT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (GreaterEqual (FlagGT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (GreaterEqual (InvertFlags x))
	// result: (LessEqual x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCLessEqual)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCGreaterThan(v *Value) bool {
	v_0 := v.Args[0]
	// match: (GreaterThan (FlagEQ))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (GreaterThan (FlagLT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (GreaterThan (FlagGT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (GreaterThan (InvertFlags x))
	// result: (LessThan x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCLessThan)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCLessEqual(v *Value) bool {
	v_0 := v.Args[0]
	// match: (LessEqual (FlagEQ))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (LessEqual (FlagLT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (LessEqual (FlagGT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (LessEqual (InvertFlags x))
	// result: (GreaterEqual x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCGreaterEqual)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCLessThan(v *Value) bool {
	v_0 := v.Args[0]
	// match: (LessThan (FlagEQ))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (LessThan (FlagLT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (LessThan (FlagGT))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (LessThan (InvertFlags x))
	// result: (GreaterThan x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCGreaterThan)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpPPCNotEqual(v *Value) bool {
	v_0 := v.Args[0]
	// match: (NotEqual (FlagEQ))
	// result: (MOVWconst [0])
	for {
		if v_0.Op != OpPPCFlagEQ {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (NotEqual (FlagLT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagLT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (NotEqual (FlagGT))
	// result: (MOVWconst [1])
	for {
		if v_0.Op != OpPPCFlagGT {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (NotEqual (InvertFlags x))
	// result: (NotEqual x)
	for {
		if v_0.Op != OpPPCInvertFlags {
			break
		}
		x := v_0.Args[0]
		v.reset(OpPPCNotEqual)
		v.AddArg(x)
		return true
	}
	return false
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
	switch b.Kind {
	case BlockPPCEQ:
		// match: (EQ (FlagEQ) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			return true
		}
		// match: (EQ (FlagLT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (EQ (FlagGT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (EQ (InvertFlags cmp) yes no)
		// result: (EQ cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCEQ, cmp)
			return true
		}
	case BlockPPCGE:
		// match: (GE (FlagEQ) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			return true
		}
		// match: (GE (FlagLT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (GE (FlagGT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (GE (InvertFlags cmp) yes no)
		// result: (LE cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCLE, cmp)
			return true
		}
	case BlockPPCGT:
		// match: (GT (FlagEQ) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (GT (FlagLT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (GT (FlagGT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (GT (InvertFlags cmp) yes no)
		// result: (LT cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCLT, cmp)
			return true
		}
	case BlockIf:
		// match: (If (Equal cc) yes no)
		// result: (EQ cc yes no)
		for b.Controls[0].Op == OpPPCEqual {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCEQ, cc)
			return true
		}
		// match: (If (NotEqual cc) yes no)
		// result: (NE cc yes no)
		for b.Controls[0].Op == OpPPCNotEqual {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCNE, cc)
			return true
		}
		// match: (If (LessThan cc) yes no)
		// result: (LT cc yes no)
		for b.Controls[0].Op == OpPPCLessThan {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCLT, cc)
			return true
		}
		// match: (If (LessEqual cc) yes no)
		// result: (LE cc yes no)
		for b.Controls[0].Op == OpPPCLessEqual {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCLE, cc)
			return true
		}
		// match: (If (GreaterThan cc) yes no)
		// result: (GT cc yes no)
		for b.Controls[0].Op == OpPPCGreaterThan {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCGT, cc)
			return true
		}
		// match: (If (GreaterEqual cc) yes no)
		// result: (GE cc yes no)
		for b.Controls[0].Op == OpPPCGreaterEqual {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCGE, cc)
			return true
		}
		// match: (If cond yes no)
		// result: (NE (CMPWconst [0] cond) yes no)
		for {
			cond := b.Controls[0]
			v0 := b.NewValue0(cond.Pos, OpPPCCMPWconst, types.TypeFlags)
			v0.AuxInt = int32ToAuxInt(0)
			v0.AddArg(cond)
			b.resetWithControl(BlockPPCNE, v0)
			return true
		}
	case BlockPPCLE:
		// match: (LE (FlagEQ) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			return true
		}
		// match: (LE (FlagLT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (LE (FlagGT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (LE (InvertFlags cmp) yes no)
		// result: (GE cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCGE, cmp)
			return true
		}
	case BlockPPCLT:
		// match: (LT (FlagEQ) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (LT (FlagLT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (LT (FlagGT) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (LT (InvertFlags cmp) yes no)
		// result: (GT cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCGT, cmp)
			return true
		}
	case BlockPPCNE:
		// match: (NE (CMPWconst [0] (Equal cc)) yes no)
		// result: (EQ cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCEqual {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCEQ, cc)
			return true
		}
		// match: (NE (CMPWconst [0] (NotEqual cc)) yes no)
		// result: (NE cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCNotEqual {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCNE, cc)
			return true
		}
		// match: (NE (CMPWconst [0] (LessThan cc)) yes no)
		// result: (LT cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCLessThan {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCLT, cc)
			return true
		}
		// match: (NE (CMPWconst [0] (LessEqual cc)) yes no)
		// result: (LE cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCLessEqual {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCLE, cc)
			return true
		}
		// match: (NE (CMPWconst [0] (GreaterThan cc)) yes no)
		// result: (GT cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCGreaterThan {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCGT, cc)
			return true
		}
		// match: (NE (CMPWconst [0] (GreaterEqual cc)) yes no)
		// result: (GE cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCGreaterEqual {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCGE, cc)
			return true
		}
		// match: (NE (FlagEQ) yes no)
		// result: (First no yes)
		for b.Controls[0].Op == OpPPCFlagEQ {
			b.Reset(BlockFirst)
			b.swapSuccessors()
			return true
		}
		// match: (NE (FlagLT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagLT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (NE (FlagGT) yes no)
		// result: (First yes no)
		for b.Controls[0].Op == OpPPCFlagGT {
			b.Reset(BlockFirst)
			return true
		}
		// match: (NE (InvertFlags cmp) yes no)
		// result: (NE cmp yes no)
		for b.Controls[0].Op == OpPPCInvertFlags {
			v_0 := b.Controls[0]
			cmp := v_0.Args[0]
			b.resetWithControl(BlockPPCNE, cmp)
			return true
		}
	}
	return false
}
