// Code generated from _gen/PPC.rules using 'go generate'; DO NOT EDIT.

package ssa

import "cmd/compile/internal/types"

func rewriteValuePPC(v *Value) bool {
	switch v.Op {
	case OpAdd16:
		return rewriteValuePPC_OpAdd16(v)
	case OpAdd32:
		return rewriteValuePPC_OpAdd32(v)
	case OpAdd32F:
		return rewriteValuePPC_OpAdd32F(v)
	case OpAdd32carry:
		return rewriteValuePPC_OpAdd32carry(v)
	case OpAdd32carrywithcarry:
		return rewriteValuePPC_OpAdd32carrywithcarry(v)
	case OpAdd32withcarry:
		return rewriteValuePPC_OpAdd32withcarry(v)
	case OpAdd64F:
		return rewriteValuePPC_OpAdd64F(v)
	case OpAdd8:
		return rewriteValuePPC_OpAdd8(v)
	case OpAddPtr:
		return rewriteValuePPC_OpAddPtr(v)
	case OpAddr:
		return rewriteValuePPC_OpAddr(v)
	case OpAnd16:
		return rewriteValuePPC_OpAnd16(v)
	case OpAnd32:
		return rewriteValuePPC_OpAnd32(v)
	case OpAnd8:
		return rewriteValuePPC_OpAnd8(v)
	case OpAndB:
		return rewriteValuePPC_OpAndB(v)
	case OpAvg32u:
		return rewriteValuePPC_OpAvg32u(v)
	case OpClosureCall:
		v.Op = OpPPCCALLclosure
		return true
	case OpCom16:
		return rewriteValuePPC_OpCom16(v)
	case OpCom32:
		return rewriteValuePPC_OpCom32(v)
	case OpCom8:
		return rewriteValuePPC_OpCom8(v)
	case OpConst16:
		return rewriteValuePPC_OpConst16(v)
	case OpConst32:
		return rewriteValuePPC_OpConst32(v)
	case OpConst32F:
		return rewriteValuePPC_OpConst32F(v)
	case OpConst64F:
		return rewriteValuePPC_OpConst64F(v)
	case OpConst8:
		return rewriteValuePPC_OpConst8(v)
	case OpConstBool:
		return rewriteValuePPC_OpConstBool(v)
	case OpConstNil:
		return rewriteValuePPC_OpConstNil(v)
	case OpCopy:
		return rewriteValuePPC_OpCopy(v)
	case OpCvt32Fto32:
		return rewriteValuePPC_OpCvt32Fto32(v)
	case OpCvt32Fto32U:
		return rewriteValuePPC_OpCvt32Fto32U(v)
	case OpCvt32Fto64F:
		return rewriteValuePPC_OpCvt32Fto64F(v)
	case OpCvt32Uto32F:
		return rewriteValuePPC_OpCvt32Uto32F(v)
	case OpCvt32Uto64F:
		return rewriteValuePPC_OpCvt32Uto64F(v)
	case OpCvt32to32F:
		return rewriteValuePPC_OpCvt32to32F(v)
	case OpCvt32to64F:
		return rewriteValuePPC_OpCvt32to64F(v)
	case OpCvt64Fto32:
		return rewriteValuePPC_OpCvt64Fto32(v)
	case OpCvt64Fto32F:
		return rewriteValuePPC_OpCvt64Fto32F(v)
	case OpCvt64Fto32U:
		return rewriteValuePPC_OpCvt64Fto32U(v)
	case OpCvtBoolToUint8:
		return rewriteValuePPC_OpCvtBoolToUint8(v)
	case OpDiv16:
		return rewriteValuePPC_OpDiv16(v)
	case OpDiv16u:
		return rewriteValuePPC_OpDiv16u(v)
	case OpDiv32:
		return rewriteValuePPC_OpDiv32(v)
	case OpDiv32F:
		return rewriteValuePPC_OpDiv32F(v)
	case OpDiv32u:
		return rewriteValuePPC_OpDiv32u(v)
	case OpDiv64F:
		return rewriteValuePPC_OpDiv64F(v)
	case OpDiv8:
		return rewriteValuePPC_OpDiv8(v)
	case OpDiv8u:
		return rewriteValuePPC_OpDiv8u(v)
	case OpEq16:
		return rewriteValuePPC_OpEq16(v)
	case OpEq32:
		return rewriteValuePPC_OpEq32(v)
	case OpEq32F:
		return rewriteValuePPC_OpEq32F(v)
	case OpEq64F:
		return rewriteValuePPC_OpEq64F(v)
	case OpEq8:
		return rewriteValuePPC_OpEq8(v)
	case OpEqB:
		return rewriteValuePPC_OpEqB(v)
	case OpEqPtr:
		return rewriteValuePPC_OpEqPtr(v)
	case OpGetCallerPC:
		v.Op = OpPPCLoweredGetCallerPC
		return true
	case OpGetCallerSP:
		v.Op = OpPPCLoweredGetCallerSP
		return true
	case OpGetClosurePtr:
		v.Op = OpPPCLoweredGetClosurePtr
		return true
	case OpHmul32:
		return rewriteValuePPC_OpHmul32(v)
	case OpHmul32u:
		return rewriteValuePPC_OpHmul32u(v)
	case OpInterCall:
		v.Op = OpPPCCALLinter
		return true
	case OpIsInBounds:
		return rewriteValuePPC_OpIsInBounds(v)
	case OpIsNonNil:
		return rewriteValuePPC_OpIsNonNil(v)
	case OpIsSliceInBounds:
		return rewriteValuePPC_OpIsSliceInBounds(v)
	case OpLeq16:
		return rewriteValuePPC_OpLeq16(v)
	case OpLeq16U:
		return rewriteValuePPC_OpLeq16U(v)
	case OpLeq32:
		return rewriteValuePPC_OpLeq32(v)
	case OpLeq32F:
		return rewriteValuePPC_OpLeq32F(v)
	case OpLeq32U:
		return rewriteValuePPC_OpLeq32U(v)
	case OpLeq64F:
		return rewriteValuePPC_OpLeq64F(v)
	case OpLeq8:
		return rewriteValuePPC_OpLeq8(v)
	case OpLeq8U:
		return rewriteValuePPC_OpLeq8U(v)
	case OpLess16:
		return rewriteValuePPC_OpLess16(v)
	case OpLess16U:
		return rewriteValuePPC_OpLess16U(v)
	case OpLess32:
		return rewriteValuePPC_OpLess32(v)
	case OpLess32F:
		return rewriteValuePPC_OpLess32F(v)
	case OpLess32U:
		return rewriteValuePPC_OpLess32U(v)
	case OpLess64F:
		return rewriteValuePPC_OpLess64F(v)
	case OpLess8:
		return rewriteValuePPC_OpLess8(v)
	case OpLess8U:
		return rewriteValuePPC_OpLess8U(v)
	case OpLoad:
		return rewriteValuePPC_OpLoad(v)
	case OpLocalAddr:
		return rewriteValuePPC_OpLocalAddr(v)
	case OpLsh16x16:
		return rewriteValuePPC_OpLsh16x16(v)
	case OpLsh16x32:
		return rewriteValuePPC_OpLsh16x32(v)
	case OpLsh16x64:
		return rewriteValuePPC_OpLsh16x64(v)
	case OpLsh16x8:
		return rewriteValuePPC_OpLsh16x8(v)
	case OpLsh32x16:
		return rewriteValuePPC_OpLsh32x16(v)
	case OpLsh32x32:
		return rewriteValuePPC_OpLsh32x32(v)
	case OpLsh32x64:
		return rewriteValuePPC_OpLsh32x64(v)
	case OpLsh32x8:
		return rewriteValuePPC_OpLsh32x8(v)
	case OpLsh8x16:
		return rewriteValuePPC_OpLsh8x16(v)
	case OpLsh8x32:
		return rewriteValuePPC_OpLsh8x32(v)
	case OpLsh8x64:
		return rewriteValuePPC_OpLsh8x64(v)
	case OpLsh8x8:
		return rewriteValuePPC_OpLsh8x8(v)
	case OpMod16:
		return rewriteValuePPC_OpMod16(v)
	case OpMod16u:
		return rewriteValuePPC_OpMod16u(v)
	case OpMod32:
		return rewriteValuePPC_OpMod32(v)
	case OpMod32u:
		return rewriteValuePPC_OpMod32u(v)
	case OpMod8:
		return rewriteValuePPC_OpMod8(v)
	case OpMod8u:
		return rewriteValuePPC_OpMod8u(v)
	case OpMove:
		return rewriteValuePPC_OpMove(v)
	case OpMul16:
		return rewriteValuePPC_OpMul16(v)
	case OpMul32:
		return rewriteValuePPC_OpMul32(v)
	case OpMul32F:
		return rewriteValuePPC_OpMul32F(v)
	case OpMul32uhilo:
		return rewriteValuePPC_OpMul32uhilo(v)
	case OpMul64F:
		return rewriteValuePPC_OpMul64F(v)
	case OpMul8:
		return rewriteValuePPC_OpMul8(v)
	case OpNeg16:
		return rewriteValuePPC_OpNeg16(v)
	case OpNeg32:
		return rewriteValuePPC_OpNeg32(v)
	case OpNeg32F:
		return rewriteValuePPC_OpNeg32F(v)
	case OpNeg64F:
		return rewriteValuePPC_OpNeg64F(v)
	case OpNeg8:
		return rewriteValuePPC_OpNeg8(v)
	case OpNeq16:
		return rewriteValuePPC_OpNeq16(v)
	case OpNeq32:
		return rewriteValuePPC_OpNeq32(v)
	case OpNeq32F:
		return rewriteValuePPC_OpNeq32F(v)
	case OpNeq64F:
		return rewriteValuePPC_OpNeq64F(v)
	case OpNeq8:
		return rewriteValuePPC_OpNeq8(v)
	case OpNeqB:
		return rewriteValuePPC_OpNeqB(v)
	case OpNeqPtr:
		return rewriteValuePPC_OpNeqPtr(v)
	case OpNilCheck:
		v.Op = OpPPCLoweredNilCheck
		return true
	case OpNot:
		return rewriteValuePPC_OpNot(v)
	case OpOffPtr:
		return rewriteValuePPC_OpOffPtr(v)
	case OpOr16:
		return rewriteValuePPC_OpOr16(v)
	case OpOr32:
		return rewriteValuePPC_OpOr32(v)
	case OpOr8:
		return rewriteValuePPC_OpOr8(v)
	case OpOrB:
		return rewriteValuePPC_OpOrB(v)
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
	case OpPanicBounds:
		v.Op = OpPPCLoweredPanicBounds
		return true
	case OpPanicExtend:
		v.Op = OpPPCLoweredPanicExtend
		return true
	case OpRotateLeft16:
		return rewriteValuePPC_OpRotateLeft16(v)
	case OpRotateLeft32:
		return rewriteValuePPC_OpRotateLeft32(v)
	case OpRotateLeft8:
		return rewriteValuePPC_OpRotateLeft8(v)
	case OpRound32F:
		return rewriteValuePPC_OpRound32F(v)
	case OpRound64F:
		return rewriteValuePPC_OpRound64F(v)
	case OpRsh16Ux16:
		return rewriteValuePPC_OpRsh16Ux16(v)
	case OpRsh16Ux32:
		return rewriteValuePPC_OpRsh16Ux32(v)
	case OpRsh16Ux64:
		return rewriteValuePPC_OpRsh16Ux64(v)
	case OpRsh16Ux8:
		return rewriteValuePPC_OpRsh16Ux8(v)
	case OpRsh16x16:
		return rewriteValuePPC_OpRsh16x16(v)
	case OpRsh16x32:
		return rewriteValuePPC_OpRsh16x32(v)
	case OpRsh16x64:
		return rewriteValuePPC_OpRsh16x64(v)
	case OpRsh16x8:
		return rewriteValuePPC_OpRsh16x8(v)
	case OpRsh32Ux16:
		return rewriteValuePPC_OpRsh32Ux16(v)
	case OpRsh32Ux32:
		return rewriteValuePPC_OpRsh32Ux32(v)
	case OpRsh32Ux64:
		return rewriteValuePPC_OpRsh32Ux64(v)
	case OpRsh32Ux8:
		return rewriteValuePPC_OpRsh32Ux8(v)
	case OpRsh32x16:
		return rewriteValuePPC_OpRsh32x16(v)
	case OpRsh32x32:
		return rewriteValuePPC_OpRsh32x32(v)
	case OpRsh32x64:
		return rewriteValuePPC_OpRsh32x64(v)
	case OpRsh32x8:
		return rewriteValuePPC_OpRsh32x8(v)
	case OpRsh8Ux16:
		return rewriteValuePPC_OpRsh8Ux16(v)
	case OpRsh8Ux32:
		return rewriteValuePPC_OpRsh8Ux32(v)
	case OpRsh8Ux64:
		return rewriteValuePPC_OpRsh8Ux64(v)
	case OpRsh8Ux8:
		return rewriteValuePPC_OpRsh8Ux8(v)
	case OpRsh8x16:
		return rewriteValuePPC_OpRsh8x16(v)
	case OpRsh8x32:
		return rewriteValuePPC_OpRsh8x32(v)
	case OpRsh8x64:
		return rewriteValuePPC_OpRsh8x64(v)
	case OpRsh8x8:
		return rewriteValuePPC_OpRsh8x8(v)
	case OpSignExt16to32:
		return rewriteValuePPC_OpSignExt16to32(v)
	case OpSignExt8to16:
		return rewriteValuePPC_OpSignExt8to16(v)
	case OpSignExt8to32:
		return rewriteValuePPC_OpSignExt8to32(v)
	case OpSignmask:
		return rewriteValuePPC_OpSignmask(v)
	case OpSlicemask:
		return rewriteValuePPC_OpSlicemask(v)
	case OpStaticCall:
		v.Op = OpPPCCALLstatic
		return true
	case OpStore:
		return rewriteValuePPC_OpStore(v)
	case OpSub16:
		return rewriteValuePPC_OpSub16(v)
	case OpSub32:
		return rewriteValuePPC_OpSub32(v)
	case OpSub32F:
		return rewriteValuePPC_OpSub32F(v)
	case OpSub32carry:
		return rewriteValuePPC_OpSub32carry(v)
	case OpSub32withcarry:
		return rewriteValuePPC_OpSub32withcarry(v)
	case OpSub64F:
		return rewriteValuePPC_OpSub64F(v)
	case OpSub8:
		return rewriteValuePPC_OpSub8(v)
	case OpTailCall:
		v.Op = OpPPCCALLtail
		return true
	case OpTrunc16to8:
		return rewriteValuePPC_OpTrunc16to8(v)
	case OpTrunc32to16:
		return rewriteValuePPC_OpTrunc32to16(v)
	case OpTrunc32to8:
		return rewriteValuePPC_OpTrunc32to8(v)
	case OpWB:
		v.Op = OpPPCLoweredWB
		return true
	case OpXor16:
		return rewriteValuePPC_OpXor16(v)
	case OpXor32:
		return rewriteValuePPC_OpXor32(v)
	case OpXor8:
		return rewriteValuePPC_OpXor8(v)
	case OpZero:
		return rewriteValuePPC_OpZero(v)
	case OpZeroExt16to32:
		return rewriteValuePPC_OpZeroExt16to32(v)
	case OpZeroExt8to16:
		return rewriteValuePPC_OpZeroExt8to16(v)
	case OpZeroExt8to32:
		return rewriteValuePPC_OpZeroExt8to32(v)
	case OpZeromask:
		return rewriteValuePPC_OpZeromask(v)
	}
	return false
}
func rewriteValuePPC_OpAdd16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add16 x y)
	// result: (ADD x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADD)
		v.AddArg2(x, y)
		return true
	}
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
func rewriteValuePPC_OpAdd32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Add32F x y)
	// result: (FRSP (FADDS <typ.Float32> x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCFADDS, typ.Float32)
		v0.AddArg2(x, y)
		v.AddArg(v0)
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
func rewriteValuePPC_OpAdd64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add64F x y)
	// result: (FADD x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFADD)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAdd8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Add8 x y)
	// result: (ADD x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADD)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAddPtr(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (AddPtr x y)
	// result: (ADD x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADD)
		v.AddArg2(x, y)
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
func rewriteValuePPC_OpAnd16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (And16 x y)
	// result: (AND x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v.AddArg2(x, y)
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
func rewriteValuePPC_OpAnd8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (And8 x y)
	// result: (AND x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAndB(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (AndB x y)
	// result: (AND x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpAvg32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Avg32u x y)
	// result: (ADD (SRW <typ.UInt32> (SUB <typ.UInt32> x y) (MOVWconst [1])) y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCADD)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSUB, typ.UInt32)
		v1.AddArg2(x, y)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v2.AuxInt = int32ToAuxInt(1)
		v0.AddArg2(v1, v2)
		v.AddArg2(v0, y)
		return true
	}
}
func rewriteValuePPC_OpCom16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com16 x)
	// result: (XOR x (MOVWconst [-1]))
	for {
		x := v_0
		v.reset(OpPPCXOR)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpCom32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com32 x)
	// result: (XOR x (MOVWconst [-1]))
	for {
		x := v_0
		v.reset(OpPPCXOR)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpCom8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com8 x)
	// result: (XOR x (MOVWconst [-1]))
	for {
		x := v_0
		v.reset(OpPPCXOR)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpConst16(v *Value) bool {
	// match: (Const16 [c])
	// result: (MOVWconst [int32(c)])
	for {
		c := auxIntToInt16(v.AuxInt)
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(int32(c))
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
func rewriteValuePPC_OpConst32F(v *Value) bool {
	// match: (Const32F [c])
	// result: (FMOVSconst [float64(c)])
	for {
		c := auxIntToFloat32(v.AuxInt)
		v.reset(OpPPCFMOVSconst)
		v.AuxInt = float64ToAuxInt(float64(c))
		return true
	}
}
func rewriteValuePPC_OpConst64F(v *Value) bool {
	// match: (Const64F [c])
	// result: (FMOVDconst [c])
	for {
		c := auxIntToFloat64(v.AuxInt)
		v.reset(OpPPCFMOVDconst)
		v.AuxInt = float64ToAuxInt(c)
		return true
	}
}
func rewriteValuePPC_OpConst8(v *Value) bool {
	// match: (Const8 [c])
	// result: (MOVWconst [int32(c)])
	for {
		c := auxIntToInt8(v.AuxInt)
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(int32(c))
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
	// match: (Copy <t> x)
	// cond: !t.IsFloat() && !t.IsMemory()
	// result: (MOVWreg x)
	for {
		t := v.Type
		x := v_0
		if !(!t.IsFloat() && !t.IsMemory()) {
			break
		}
		v.reset(OpPPCMOVWreg)
		v.AddArg(x)
		return true
	}
	// match: (Copy <t> x)
	// cond: t.IsFloat()
	// result: (FMR x)
	for {
		t := v.Type
		x := v_0
		if !(t.IsFloat()) {
			break
		}
		v.reset(OpPPCFMR)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpCvt32Fto32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32Fto32 x)
	// result: (LoweredCvt64Fto32 x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt64Fto32)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt32Fto32U(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32Fto32U x)
	// result: (LoweredCvt64Fto32U x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt64Fto32U)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt32Fto64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32Fto64F x)
	// result: (FMR x)
	for {
		x := v_0
		v.reset(OpPPCFMR)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt32Uto32F(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Cvt32Uto32F x)
	// result: (FRSP (LoweredCvt32Uto64F <typ.Float64> x))
	for {
		x := v_0
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCLoweredCvt32Uto64F, typ.Float64)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpCvt32Uto64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32Uto64F x)
	// result: (LoweredCvt32Uto64F x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt32Uto64F)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt32to32F(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Cvt32to32F x)
	// result: (FRSP (LoweredCvt32to64F <typ.Float64> x))
	for {
		x := v_0
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCLoweredCvt32to64F, typ.Float64)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpCvt32to64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32to64F x)
	// result: (LoweredCvt32to64F x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt32to64F)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt64Fto32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt64Fto32 x)
	// result: (LoweredCvt64Fto32 x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt64Fto32)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt64Fto32F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt64Fto32F x)
	// result: (FRSP x)
	for {
		x := v_0
		v.reset(OpPPCFRSP)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvt64Fto32U(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt64Fto32U x)
	// result: (LoweredCvt64Fto32U x)
	for {
		x := v_0
		v.reset(OpPPCLoweredCvt64Fto32U)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpCvtBoolToUint8(v *Value) bool {
	v_0 := v.Args[0]
	// match: (CvtBoolToUint8 x)
	// result: (Copy x)
	for {
		x := v_0
		v.reset(OpCopy)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpDiv16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div16 [false] x y)
	// result: (DIVW (SignExt16to32 x) (SignExt16to32 y))
	for {
		if auxIntToBool(v.AuxInt) != false {
			break
		}
		x := v_0
		y := v_1
		v.reset(OpPPCDIVW)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValuePPC_OpDiv16u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div16u x y)
	// result: (DIVWU (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCDIVWU)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpDiv32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Div32 x y)
	// result: (DIVW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCDIVW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpDiv32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div32F x y)
	// result: (FRSP (FDIVS <typ.Float32> x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCFDIVS, typ.Float32)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpDiv32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Div32u x y)
	// result: (DIVWU x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCDIVWU)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpDiv64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Div64F x y)
	// result: (FDIV x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFDIV)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpDiv8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div8 x y)
	// result: (DIVW (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCDIVW)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpDiv8u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div8u x y)
	// result: (DIVWU (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCDIVWU)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpEq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Eq16 x y)
	// result: (Equal (CMPW (ZeroExt16to32 x) (ZeroExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
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
func rewriteValuePPC_OpEq32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Eq32F x y)
	// result: (Equal (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpEq64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Eq64F x y)
	// result: (Equal (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpEq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Eq8 x y)
	// result: (Equal (CMPW (ZeroExt8to32 x) (ZeroExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpEqB(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (EqB x y)
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
func rewriteValuePPC_OpEqPtr(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (EqPtr x y)
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
func rewriteValuePPC_OpHmul32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Hmul32 x y)
	// result: (MULHW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULHW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpHmul32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Hmul32u x y)
	// result: (MULHWU x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULHWU)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpIsInBounds(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (IsInBounds x y)
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
func rewriteValuePPC_OpIsNonNil(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	// match: (IsNonNil x)
	// result: (NotEqual (CMPWconst [0] x))
	for {
		x := v_0
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWconst, types.TypeFlags)
		v0.AuxInt = int32ToAuxInt(0)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpIsSliceInBounds(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (IsSliceInBounds x y)
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
func rewriteValuePPC_OpLeq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq16 x y)
	// result: (LessEqual (CMPW (SignExt16to32 x) (SignExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLeq16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq16U x y)
	// result: (LessEqual (CMPWU (ZeroExt16to32 x) (ZeroExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
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
func rewriteValuePPC_OpLeq32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Leq32F x y)
	// result: (FLessEqual (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
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
func rewriteValuePPC_OpLeq64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Leq64F x y)
	// result: (FLessEqual (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLeq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq8 x y)
	// result: (LessEqual (CMPW (SignExt8to32 x) (SignExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLeq8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq8U x y)
	// result: (LessEqual (CMPWU (ZeroExt8to32 x) (ZeroExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less16 x y)
	// result: (LessThan (CMPW (SignExt16to32 x) (SignExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less16U x y)
	// result: (LessThan (CMPWU (ZeroExt16to32 x) (ZeroExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
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
func rewriteValuePPC_OpLess32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Less32F x y)
	// result: (LessThan (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
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
func rewriteValuePPC_OpLess64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Less64F x y)
	// result: (LessThan (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less8 x y)
	// result: (LessThan (CMPW (SignExt8to32 x) (SignExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpLess8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less8U x y)
	// result: (LessThan (CMPWU (ZeroExt8to32 x) (ZeroExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCLessThan)
		v0 := b.NewValue0(v.Pos, OpPPCCMPWU, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
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
	// match: (Load <t> ptr mem)
	// cond: is32BitFloat(t)
	// result: (FMOVSload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is32BitFloat(t)) {
			break
		}
		v.reset(OpPPCFMOVSload)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitFloat(t)
	// result: (FMOVDload ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitFloat(t)) {
			break
		}
		v.reset(OpPPCFMOVDload)
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
func rewriteValuePPC_OpLsh16x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x16 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt16to32 x) (ZeroExt16to32 y)) (MOVWconst [65535])) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] (ZeroExt16to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v3.AddArg(y)
		v1.AddArg2(v2, v3)
		v4 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v4.AuxInt = int32ToAuxInt(65535)
		v0.AddArg2(v1, v4)
		v5 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v6 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v7 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v7.AuxInt = int32ToAuxInt(16)
		v7.AddArg(v3)
		v6.AddArg(v7)
		v5.AddArg(v6)
		v.AddArg2(v0, v5)
		return true
	}
}
func rewriteValuePPC_OpLsh16x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x32 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt16to32 x) y) (MOVWconst [65535])) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(x)
		v1.AddArg2(v2, y)
		v3 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v3.AuxInt = int32ToAuxInt(65535)
		v0.AddArg2(v1, v3)
		v4 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v6.AuxInt = int32ToAuxInt(16)
		v6.AddArg(y)
		v5.AddArg(v6)
		v4.AddArg(v5)
		v.AddArg2(v0, v4)
		return true
	}
}
func rewriteValuePPC_OpLsh16x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SLW x (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSLW)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh16x64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh16x64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt16to32 x) (MOVWconst <typ.UInt32> [int32(c)])) (MOVWconst <typ.UInt32> [65535]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpPPCAND)
		v.Type = typ.UInt32
		v0 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.UInt32)
		v2.AuxInt = int32ToAuxInt(int32(c))
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.UInt32)
		v3.AuxInt = int32ToAuxInt(65535)
		v.AddArg2(v0, v3)
		return true
	}
	// match: (Lsh16x64 _ (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpLsh16x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x8 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt16to32 x) (ZeroExt8to32 y)) (MOVWconst [65535])) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] (ZeroExt8to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v3.AddArg(y)
		v1.AddArg2(v2, v3)
		v4 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v4.AuxInt = int32ToAuxInt(65535)
		v0.AddArg2(v1, v4)
		v5 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v6 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v7 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v7.AuxInt = int32ToAuxInt(16)
		v7.AddArg(v3)
		v6.AddArg(v7)
		v5.AddArg(v6)
		v.AddArg2(v0, v5)
		return true
	}
}
func rewriteValuePPC_OpLsh32x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x16 x y)
	// result: (Lsh32x32 x (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpLsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x32 x y)
	// result: (AND (SLW <typ.UInt32> x y) (NEG <typ.UInt32> (LessThan (CMPWUconst [32] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v0.AddArg2(x, y)
		v1 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v3.AuxInt = int32ToAuxInt(32)
		v3.AddArg(y)
		v2.AddArg(v3)
		v1.AddArg(v2)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpLsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SLW x (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSLW)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh32x64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpLsh32x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x8 x y)
	// result: (Lsh32x32 x (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpLsh8x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x16 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt8to32 x) (ZeroExt16to32 y)) (MOVWconst [255])) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] (ZeroExt16to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v3.AddArg(y)
		v1.AddArg2(v2, v3)
		v4 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v4.AuxInt = int32ToAuxInt(255)
		v0.AddArg2(v1, v4)
		v5 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v6 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v7 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v7.AuxInt = int32ToAuxInt(8)
		v7.AddArg(v3)
		v6.AddArg(v7)
		v5.AddArg(v6)
		v.AddArg2(v0, v5)
		return true
	}
}
func rewriteValuePPC_OpLsh8x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x32 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt8to32 x) y) (MOVWconst [255])) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(x)
		v1.AddArg2(v2, y)
		v3 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v3.AuxInt = int32ToAuxInt(255)
		v0.AddArg2(v1, v3)
		v4 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v6.AuxInt = int32ToAuxInt(8)
		v6.AddArg(y)
		v5.AddArg(v6)
		v4.AddArg(v5)
		v.AddArg2(v0, v4)
		return true
	}
}
func rewriteValuePPC_OpLsh8x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SLW x (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSLW)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh8x64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh8x64 x (Const64 [c]))
	// cond: uint64(c) < 8
	// result: (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt8to32 x) (MOVWconst <typ.UInt32> [int32(c)])) (MOVWconst <typ.UInt32> [255]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 8) {
			break
		}
		v.reset(OpPPCAND)
		v.Type = typ.UInt32
		v0 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.UInt32)
		v2.AuxInt = int32ToAuxInt(int32(c))
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.UInt32)
		v3.AuxInt = int32ToAuxInt(255)
		v.AddArg2(v0, v3)
		return true
	}
	// match: (Lsh8x64 _ (Const64 [c]))
	// cond: uint64(c) >= 8
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 8) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpLsh8x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x8 x y)
	// result: (AND (AND <typ.UInt32> (SLW <typ.UInt32> (ZeroExt8to32 x) (ZeroExt8to32 y)) (MOVWconst [255])) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] (ZeroExt8to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCAND, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCSLW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v3.AddArg(y)
		v1.AddArg2(v2, v3)
		v4 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v4.AuxInt = int32ToAuxInt(255)
		v0.AddArg2(v1, v4)
		v5 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v6 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v7 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v7.AuxInt = int32ToAuxInt(8)
		v7.AddArg(v3)
		v6.AddArg(v7)
		v5.AddArg(v6)
		v.AddArg2(v0, v5)
		return true
	}
}
func rewriteValuePPC_OpMod16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod16 x y)
	// result: (Mod32 (SignExt16to32 x) (SignExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpMod32)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpMod16u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod16u x y)
	// result: (Mod32u (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpMod32u)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpMod32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod32 x y)
	// result: (SUB x (MULLW <typ.Int32> y (DIVW <typ.Int32> x y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUB)
		v0 := b.NewValue0(v.Pos, OpPPCMULLW, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpPPCDIVW, typ.Int32)
		v1.AddArg2(x, y)
		v0.AddArg2(y, v1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpMod32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod32u x y)
	// result: (SUB x (MULLW <typ.UInt32> y (DIVWU <typ.UInt32> x y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUB)
		v0 := b.NewValue0(v.Pos, OpPPCMULLW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCDIVWU, typ.UInt32)
		v1.AddArg2(x, y)
		v0.AddArg2(y, v1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpMod8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod8 x y)
	// result: (Mod32 (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpMod32)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpMod8u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod8u x y)
	// result: (Mod32u (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpMod32u)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpMove(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Move [0] _ _ mem)
	// result: (Copy mem)
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		mem := v_2
		v.reset(OpCopy)
		v.AddArg(mem)
		return true
	}
	// match: (Move [s] {t} dst src mem)
	// cond: s > 0
	// result: (LoweredMove [s] {t} dst src mem)
	for {
		s := auxIntToInt64(v.AuxInt)
		t := auxToType(v.Aux)
		dst := v_0
		src := v_1
		mem := v_2
		if !(s > 0) {
			break
		}
		v.reset(OpPPCLoweredMove)
		v.AuxInt = int64ToAuxInt(s)
		v.Aux = typeToAux(t)
		v.AddArg3(dst, src, mem)
		return true
	}
	return false
}
func rewriteValuePPC_OpMul16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mul16 x y)
	// result: (MULLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULLW)
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
func rewriteValuePPC_OpMul32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mul32F x y)
	// result: (FRSP (FMULS <typ.Float32> x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCFMULS, typ.Float32)
		v0.AddArg2(x, y)
		v.AddArg(v0)
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
func rewriteValuePPC_OpMul64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mul64F x y)
	// result: (FMUL x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFMUL)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpMul8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mul8 x y)
	// result: (MULLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCMULLW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpNeg16(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Neg16 x)
	// result: (NEG x)
	for {
		x := v_0
		v.reset(OpPPCNEG)
		v.AddArg(x)
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
func rewriteValuePPC_OpNeg32F(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neg32F x)
	// result: (FRSP (FNEG <typ.Float32> x))
	for {
		x := v_0
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCFNEG, typ.Float32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpNeg64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Neg64F x)
	// result: (FNEG x)
	for {
		x := v_0
		v.reset(OpPPCFNEG)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpNeg8(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Neg8 x)
	// result: (NEG x)
	for {
		x := v_0
		v.reset(OpPPCNEG)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpNeq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neq16 x y)
	// result: (NotEqual (CMPW (ZeroExt16to32 x) (ZeroExt16to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
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
func rewriteValuePPC_OpNeq32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Neq32F x y)
	// result: (NotEqual (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpNeq64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (Neq64F x y)
	// result: (NotEqual (FCMPU x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCFCMPU, types.TypeFlags)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpNeq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neq8 x y)
	// result: (NotEqual (CMPW (ZeroExt8to32 x) (ZeroExt8to32 y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCNotEqual)
		v0 := b.NewValue0(v.Pos, OpPPCCMPW, types.TypeFlags)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v.AddArg(v0)
		return true
	}
}
func rewriteValuePPC_OpNeqB(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (NeqB x y)
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
func rewriteValuePPC_OpNeqPtr(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (NeqPtr x y)
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
func rewriteValuePPC_OpNot(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Not x)
	// result: (XOR x (MOVWconst [1]))
	for {
		x := v_0
		v.reset(OpPPCXOR)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(1)
		v.AddArg2(x, v0)
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
func rewriteValuePPC_OpOr16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or16 x y)
	// result: (OR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCOR)
		v.AddArg2(x, y)
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
func rewriteValuePPC_OpOr8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or8 x y)
	// result: (OR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCOR)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpOrB(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (OrB x y)
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
func rewriteValuePPC_OpRotateLeft16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (RotateLeft16 x y)
	// result: (Or16 <x.Type> (Lsh16x32 <x.Type> x (And32 <y.Type> y (MOVWconst <y.Type> [15]))) (Rsh16Ux32 <x.Type> x (Sub32 <y.Type> (MOVWconst <y.Type> [16]) (And32 <y.Type> y (MOVWconst <y.Type> [15])))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOr16)
		v.Type = x.Type
		v0 := b.NewValue0(v.Pos, OpLsh16x32, x.Type)
		v1 := b.NewValue0(v.Pos, OpAnd32, y.Type)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, y.Type)
		v2.AuxInt = int32ToAuxInt(15)
		v1.AddArg2(y, v2)
		v0.AddArg2(x, v1)
		v3 := b.NewValue0(v.Pos, OpRsh16Ux32, x.Type)
		v4 := b.NewValue0(v.Pos, OpSub32, y.Type)
		v5 := b.NewValue0(v.Pos, OpPPCMOVWconst, y.Type)
		v5.AuxInt = int32ToAuxInt(16)
		v4.AddArg2(v5, v1)
		v3.AddArg2(x, v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRotateLeft32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (RotateLeft32 x y)
	// result: (ROTLW x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCROTLW)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpRotateLeft8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (RotateLeft8 x y)
	// result: (Or8 <x.Type> (Lsh8x32 <x.Type> x (And32 <y.Type> y (MOVWconst <y.Type> [7]))) (Rsh8Ux32 <x.Type> x (Sub32 <y.Type> (MOVWconst <y.Type> [8]) (And32 <y.Type> y (MOVWconst <y.Type> [7])))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOr8)
		v.Type = x.Type
		v0 := b.NewValue0(v.Pos, OpLsh8x32, x.Type)
		v1 := b.NewValue0(v.Pos, OpAnd32, y.Type)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, y.Type)
		v2.AuxInt = int32ToAuxInt(7)
		v1.AddArg2(y, v2)
		v0.AddArg2(x, v1)
		v3 := b.NewValue0(v.Pos, OpRsh8Ux32, x.Type)
		v4 := b.NewValue0(v.Pos, OpSub32, y.Type)
		v5 := b.NewValue0(v.Pos, OpPPCMOVWconst, y.Type)
		v5.AuxInt = int32ToAuxInt(8)
		v4.AddArg2(v5, v1)
		v3.AddArg2(x, v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRound32F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Round32F x)
	// result: (FRSP x)
	for {
		x := v_0
		v.reset(OpPPCFRSP)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpRound64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Round64F x)
	// result: (Copy x)
	for {
		x := v_0
		v.reset(OpCopy)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpRsh16Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux16 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt16to32 x) (ZeroExt16to32 y)) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] (ZeroExt16to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(16)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRsh16Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux32 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt16to32 x) y) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, y)
		v2 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v4.AuxInt = int32ToAuxInt(16)
		v4.AddArg(y)
		v3.AddArg(v4)
		v2.AddArg(v3)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValuePPC_OpRsh16Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SRW (ZeroExt16to32 x) (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSRW)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Rsh16Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh16Ux64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (SRW (ZeroExt16to32 x) (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpPPCSRW)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Rsh16Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh16Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux8 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt16to32 x) (ZeroExt8to32 y)) (NEG <typ.UInt32> (LessThan (CMPWUconst [16] (ZeroExt8to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(16)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRsh16x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x16 x y)
	// result: (SRAW (SignExt16to32 x) (OR <typ.UInt32> (ZeroExt16to32 y) (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [16] (ZeroExt16to32 y))))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(16)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v1.AddArg2(v2, v3)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh16x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x32 x y)
	// result: (SRAW (SignExt16to32 x) (OR <typ.UInt32> y (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [16] y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v4.AuxInt = int32ToAuxInt(16)
		v4.AddArg(y)
		v3.AddArg(v4)
		v2.AddArg(v3)
		v1.AddArg2(y, v2)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh16x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (SRAWI (SignExt16to32 x) [int32(c)])
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(int32(c))
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	// match: (Rsh16x64 x (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (Signmask (SignExt16to32 x))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpSignmask)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh16x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x8 x y)
	// result: (SRAW (SignExt16to32 x) (OR <typ.UInt32> (ZeroExt8to32 y) (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [16] (ZeroExt8to32 y))))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(16)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v1.AddArg2(v2, v3)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh32Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux16 x y)
	// result: (Rsh32Ux32 x (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpRsh32Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux32 x y)
	// result: (AND (SRW <typ.UInt32> x y) (NEG <typ.UInt32> (LessThan (CMPWUconst [32] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v0.AddArg2(x, y)
		v1 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v3.AuxInt = int32ToAuxInt(32)
		v3.AddArg(y)
		v2.AddArg(v3)
		v1.AddArg(v2)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh32Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SRW x (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSRW)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh32Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux8 x y)
	// result: (Rsh32Ux32 x (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpRsh32x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x16 x y)
	// result: (Rsh32x32 x (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpRsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x32 x y)
	// result: (SRAW x (OR <typ.UInt32> y (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [32] y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v3.AuxInt = int32ToAuxInt(32)
		v3.AddArg(y)
		v2.AddArg(v3)
		v1.AddArg(v2)
		v0.AddArg2(y, v1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpRsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Rsh32x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SRAWI x [int32(c)])
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg(x)
		return true
	}
	// match: (Rsh32x64 x (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (SRAWI x [31])
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(31)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh32x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x8 x y)
	// result: (Rsh32x32 x (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpRsh8Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux16 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt8to32 x) (ZeroExt16to32 y)) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] (ZeroExt16to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(8)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRsh8Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux32 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt8to32 x) y) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, y)
		v2 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v4.AuxInt = int32ToAuxInt(8)
		v4.AddArg(y)
		v3.AddArg(v4)
		v2.AddArg(v3)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValuePPC_OpRsh8Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (SRW (ZeroExt8to32 x) (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpPPCSRW)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Rsh8Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh8Ux64 x (Const64 [c]))
	// cond: uint64(c) < 8
	// result: (SRW (ZeroExt8to32 x) (MOVWconst [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 8) {
			break
		}
		v.reset(OpPPCSRW)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Rsh8Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 8
	// result: (MOVWconst [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 8) {
			break
		}
		v.reset(OpPPCMOVWconst)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh8Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux8 x y)
	// result: (AND (SRW <typ.UInt32> (ZeroExt8to32 x) (ZeroExt8to32 y)) (NEG <typ.UInt32> (LessThan (CMPWUconst [8] (ZeroExt8to32 y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCLessThan, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(8)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValuePPC_OpRsh8x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x16 x y)
	// result: (SRAW (SignExt8to32 x) (OR <typ.UInt32> (ZeroExt16to32 y) (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [8] (ZeroExt16to32 y))))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v2.AddArg(y)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(8)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v1.AddArg2(v2, v3)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh8x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x32 x y)
	// result: (SRAW (SignExt8to32 x) (OR <typ.UInt32> y (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [8] y)))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v4.AuxInt = int32ToAuxInt(8)
		v4.AddArg(y)
		v3.AddArg(v4)
		v2.AddArg(v3)
		v1.AddArg2(y, v2)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpRsh8x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x64 x (Const64 [c]))
	// cond: uint64(c) < 8
	// result: (SRAWI (SignExt8to32 x) [int32(c)])
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 8) {
			break
		}
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(int32(c))
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	// match: (Rsh8x64 x (Const64 [c]))
	// cond: uint64(c) >= 8
	// result: (Signmask (SignExt8to32 x))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 8) {
			break
		}
		v.reset(OpSignmask)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValuePPC_OpRsh8x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x8 x y)
	// result: (SRAW (SignExt8to32 x) (OR <typ.UInt32> (ZeroExt8to32 y) (NEG <typ.UInt32> (GreaterEqual (CMPWUconst [8] (ZeroExt8to32 y))))))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSRAW)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpPPCOR, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v2.AddArg(y)
		v3 := b.NewValue0(v.Pos, OpPPCNEG, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpPPCGreaterEqual, typ.Bool)
		v5 := b.NewValue0(v.Pos, OpPPCCMPWUconst, types.TypeFlags)
		v5.AuxInt = int32ToAuxInt(8)
		v5.AddArg(v2)
		v4.AddArg(v5)
		v3.AddArg(v4)
		v1.AddArg2(v2, v3)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValuePPC_OpSignExt16to32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (SignExt16to32 x)
	// result: (EXTSH x)
	for {
		x := v_0
		v.reset(OpPPCEXTSH)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpSignExt8to16(v *Value) bool {
	v_0 := v.Args[0]
	// match: (SignExt8to16 x)
	// result: (EXTSB x)
	for {
		x := v_0
		v.reset(OpPPCEXTSB)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpSignExt8to32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (SignExt8to32 x)
	// result: (EXTSB x)
	for {
		x := v_0
		v.reset(OpPPCEXTSB)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpSignmask(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Signmask x)
	// result: (SRAWI x [31])
	for {
		x := v_0
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(31)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpSlicemask(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Slicemask x)
	// result: (SRAWI (NEG <typ.Int32> x) [31])
	for {
		x := v_0
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(31)
		v0 := b.NewValue0(v.Pos, OpPPCNEG, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
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
	// match: (Store {t} ptr val mem)
	// cond: is32BitFloat(t)
	// result: (FMOVSstore ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(is32BitFloat(t)) {
			break
		}
		v.reset(OpPPCFMOVSstore)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: is64BitFloat(t)
	// result: (FMOVDstore ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(is64BitFloat(t)) {
			break
		}
		v.reset(OpPPCFMOVDstore)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValuePPC_OpSub16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub16 x y)
	// result: (SUB x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUB)
		v.AddArg2(x, y)
		return true
	}
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
func rewriteValuePPC_OpSub32F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Sub32F x y)
	// result: (FRSP (FSUBS <typ.Float32> x y))
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFRSP)
		v0 := b.NewValue0(v.Pos, OpPPCFSUBS, typ.Float32)
		v0.AddArg2(x, y)
		v.AddArg(v0)
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
func rewriteValuePPC_OpSub64F(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub64F x y)
	// result: (FSUB x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCFSUB)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpSub8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Sub8 x y)
	// result: (SUB x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCSUB)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpTrunc16to8(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Trunc16to8 x)
	// result: (Copy x)
	for {
		x := v_0
		v.reset(OpCopy)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpTrunc32to16(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Trunc32to16 x)
	// result: (Copy x)
	for {
		x := v_0
		v.reset(OpCopy)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpTrunc32to8(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Trunc32to8 x)
	// result: (Copy x)
	for {
		x := v_0
		v.reset(OpCopy)
		v.AddArg(x)
		return true
	}
}
func rewriteValuePPC_OpXor16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Xor16 x y)
	// result: (XOR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCXOR)
		v.AddArg2(x, y)
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
func rewriteValuePPC_OpXor8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Xor8 x y)
	// result: (XOR x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpPPCXOR)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValuePPC_OpZero(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Zero [0] _ mem)
	// result: (Copy mem)
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		mem := v_1
		v.reset(OpCopy)
		v.AddArg(mem)
		return true
	}
	// match: (Zero [s] {t} ptr mem)
	// cond: s > 0
	// result: (LoweredZero [s] {t} ptr mem)
	for {
		s := auxIntToInt64(v.AuxInt)
		t := auxToType(v.Aux)
		ptr := v_0
		mem := v_1
		if !(s > 0) {
			break
		}
		v.reset(OpPPCLoweredZero)
		v.AuxInt = int64ToAuxInt(s)
		v.Aux = typeToAux(t)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValuePPC_OpZeroExt16to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt16to32 x)
	// result: (AND x (MOVWconst [65535]))
	for {
		x := v_0
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(65535)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpZeroExt8to16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt8to16 x)
	// result: (AND x (MOVWconst [255]))
	for {
		x := v_0
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(255)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpZeroExt8to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt8to32 x)
	// result: (AND x (MOVWconst [255]))
	for {
		x := v_0
		v.reset(OpPPCAND)
		v0 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v0.AuxInt = int32ToAuxInt(255)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValuePPC_OpZeromask(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Zeromask x)
	// result: (SRAWI (SUB <typ.Int32> (SRW <typ.UInt32> x (MOVWconst [1])) x) [31])
	for {
		x := v_0
		v.reset(OpPPCSRAWI)
		v.AuxInt = int32ToAuxInt(31)
		v0 := b.NewValue0(v.Pos, OpPPCSUB, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpPPCSRW, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpPPCMOVWconst, typ.Int32)
		v2.AuxInt = int32ToAuxInt(1)
		v1.AddArg2(x, v2)
		v0.AddArg2(v1, x)
		v.AddArg(v0)
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
		// match: (If (FLessEqual cc) yes no)
		// result: (FLE cc yes no)
		for b.Controls[0].Op == OpPPCFLessEqual {
			v_0 := b.Controls[0]
			cc := v_0.Args[0]
			b.resetWithControl(BlockPPCFLE, cc)
			return true
		}
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
		// match: (NE (CMPWconst [0] (FLessEqual cc)) yes no)
		// result: (FLE cc yes no)
		for b.Controls[0].Op == OpPPCCMPWconst {
			v_0 := b.Controls[0]
			if auxIntToInt32(v_0.AuxInt) != 0 {
				break
			}
			v_0_0 := v_0.Args[0]
			if v_0_0.Op != OpPPCFLessEqual {
				break
			}
			cc := v_0_0.Args[0]
			b.resetWithControl(BlockPPCFLE, cc)
			return true
		}
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
