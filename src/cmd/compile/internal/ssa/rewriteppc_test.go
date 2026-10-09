// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// These tests guard the ppc32 (GOARCH=ppc) lowering of the generic ops the
// 64-bit decomposition leaves behind, and in particular the XER[CA] model: the
// carry is an SSA value of type Flags produced by ADDC and consumed by ADDE, so
// (a) the carry is a data dependence and (b) an instruction that writes CA
// without producing it cannot be allowed to sit between the pair.

import (
	"fmt"
	"testing"

	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc"
)

// testConfigPPC32 builds a Config for the ppc32 back end. testConfigArch cannot
// be reused: it insists on 64-bit test types, and a 4-byte pointer is the whole
// point of this target.
func testConfigPPC32(tb testing.TB) *Conf {
	ctxt := obj.Linknew(&ppc.Linkppc)
	if ctxt.Arch.PtrSize != 4 {
		tb.Fatalf("ppc LinkArch PtrSize = %d, want 4", ctxt.Arch.PtrSize)
	}
	return &Conf{
		config: NewConfig("ppc", testTypes, ctxt, true, false),
		tb:     tb,
	}
}

// carryChainFunc builds, un-lowered:
//
//	addc = Add32carry a0 b0        ; the low word of a + b
//	ca   = Select1 addc            ; its carry out, in XER[CA]
//	adde = Add32withcarry a1 b1 ca ; the high word, plus that carry
//
// which is exactly what _gen/dec64.rules generates for an Add64.
func carryChainFunc(c *Conf) fun {
	return c.Fun("entry",
		Bloc("entry",
			Valu("mem", OpInitMem, types.TypeMem, 0, nil),
			Valu("a0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("a1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("addc", OpAdd32carry, types.NewTuple(c.config.Types.UInt32, types.TypeFlags), 0, nil, "a0", "b0"),
			Valu("lo", OpSelect0, c.config.Types.UInt32, 0, nil, "addc"),
			Valu("ca", OpSelect1, types.TypeFlags, 0, nil, "addc"),
			Valu("adde", OpAdd32withcarry, c.config.Types.UInt32, 0, nil, "a1", "b1", "ca"),
			// Root the result: the block's exit is reachable from the store's
			// memory, so dead-code elimination cannot drop the carry chain.
			Valu("sp", OpSP, c.config.Types.Uintptr, 0, nil),
			Valu("off", OpOffPtr, types.NewPtr(c.config.Types.UInt32), 0, nil, "sp"),
			Valu("st", OpPPCMOVWstore, types.TypeMem, 0, nil, "off", "adde", "mem"),
			Exit("st")),
	)
}

// TestPPCLowerCarryChain checks that the residue lowers to the carry pair, that
// the carry argument of the high word is the Flags value produced by the low
// word, and that no generic op is left behind.
func TestPPCLowerNarrowRotate(t *testing.T) {
	for _, width := range []int{8, 16} {
		for _, countWidth := range []int{8, 16, 32, 64} {
			t.Run(fmt.Sprintf("%d/count%d", width, countWidth), func(t *testing.T) {
				c := testConfigPPC32(t)
				valueType, rotateOp, store := c.config.Types.UInt8, OpRotateLeft8, OpPPCMOVBstore
				if width == 16 {
					valueType, rotateOp, store = c.config.Types.UInt16, OpRotateLeft16, OpPPCMOVHstore
				}
				countType := c.config.Types.UInt8
				switch countWidth {
				case 16:
					countType = c.config.Types.UInt16
				case 32:
					countType = c.config.Types.UInt32
				}
				values := []any{
					Valu("mem", OpInitMem, types.TypeMem, 0, nil),
					Valu("x", OpArg, valueType, 0, nil),
				}
				if countWidth == 64 {
					values = append(values,
						Valu("khi", OpArg, c.config.Types.UInt32, 0, nil),
						Valu("klo", OpArg, c.config.Types.UInt32, 0, nil),
						Valu("k", OpInt64Make, c.config.Types.UInt64, 0, nil, "khi", "klo"),
					)
				} else {
					values = append(values, Valu("k", OpArg, countType, 0, nil))
				}
				values = append(values,
					Valu("rot", rotateOp, valueType, 0, nil, "x", "k"),
					Valu("sp", OpSP, c.config.Types.Uintptr, 0, nil),
					Valu("off", OpOffPtr, types.NewPtr(valueType), 0, nil, "sp"),
					Valu("st", store, types.TypeMem, 0, nil, "off", "rot", "mem"),
					Exit("st"),
				)
				f := c.Fun("entry", Bloc("entry", values...))
				lower(f.f)
				checkLower(f.f)
				foundLeft, foundRight := false, false
				for _, v := range f.blocks["entry"].Values {
					if v.Op == OpPPCSLW {
						foundLeft = true
					}
					if v.Op == OpPPCSRW {
						foundRight = true
					}
				}
				if !foundLeft || !foundRight {
					t.Fatalf("RotateLeft%d lowered without both word shifts: left=%t right=%t", width, foundLeft, foundRight)
				}
			})
		}
	}
}

func TestPPCLowerCarryChain(t *testing.T) {
	c := testConfigPPC32(t)
	f := carryChainFunc(c)
	lower(f.f)
	checkLower(f.f)

	addc := f.values["addc"]
	if addc.Op != OpPPCADDC {
		t.Fatalf("Add32carry lowered to %v, want OpPPCADDC (addc)", addc.Op)
	}
	if !addc.Type.IsTuple() || !addc.Type.FieldType(0).IsInteger() || !addc.Type.FieldType(1).IsFlags() {
		t.Fatalf("ADDC type is %v, want a (integer, Flags) pair", addc.Type)
	}
	adde := f.values["adde"]
	if adde.Op != OpPPCADDE {
		t.Fatalf("Add32withcarry lowered to %v, want OpPPCADDE (adde)", adde.Op)
	}
	carry := adde.Args[2]
	if carry.Op != OpSelect1 || carry.Args[0] != addc {
		t.Fatalf("adde carry argument is %v, want Select1 of the addc", carry.LongString())
	}
}

// TestPPCLowerSubCarryChain checks the subtraction half: Sub32carry is subfc and
// Sub32withcarry is subfe, and the borrow travels as a Flags value like the
// carry does.
func TestPPCLowerSubCarryChain(t *testing.T) {
	c := testConfigPPC32(t)
	f := c.Fun("entry",
		Bloc("entry",
			Valu("mem", OpInitMem, types.TypeMem, 0, nil),
			Valu("a0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("a1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("subc", OpSub32carry, types.NewTuple(c.config.Types.UInt32, types.TypeFlags), 0, nil, "a0", "b0"),
			Valu("lo", OpSelect0, c.config.Types.UInt32, 0, nil, "subc"),
			Valu("borrow", OpSelect1, types.TypeFlags, 0, nil, "subc"),
			Valu("sube", OpSub32withcarry, c.config.Types.UInt32, 0, nil, "a1", "b1", "borrow"),
			Valu("sp", OpSP, c.config.Types.Uintptr, 0, nil),
			Valu("off", OpOffPtr, types.NewPtr(c.config.Types.UInt32), 0, nil, "sp"),
			Valu("st", OpPPCMOVWstore, types.TypeMem, 0, nil, "off", "sube", "mem"),
			Exit("st")),
	)
	lower(f.f)
	checkLower(f.f)

	if got := f.values["subc"].Op; got != OpPPCSUBC {
		t.Errorf("Sub32carry lowered to %v, want OpPPCSUBC (subfc)", got)
	}
	if got := f.values["sube"].Op; got != OpPPCSUBE {
		t.Errorf("Sub32withcarry lowered to %v, want OpPPCSUBE (subfe)", got)
	}
	if borrow := f.values["sube"].Args[2]; borrow.Op != OpSelect1 || borrow.Args[0] != f.values["subc"] {
		t.Errorf("sube borrow argument is %v, want Select1 of the subc", borrow.LongString())
	}
}

// TestPPCCarryClobberedBetweenAddcAndAdde is the guard on the XER[CA] model.
//
// srawi writes XER[CA] but produces no carry value, so it is marked clobberFlags
// in _gen/PPCOps.go. Placed between the addc and the adde, the flag allocator
// must notice that the carry is no longer in the machine and recompute it --
// which means a second addc, after the clobber. If clobberFlags were missing
// from SRAWI the allocator would leave the single addc where it was and the adde
// would add whatever carry the shift left behind.
func TestPPCCarryClobberedBetweenAddcAndAdde(t *testing.T) {
	c := testConfigPPC32(t)
	f := c.Fun("entry",
		Bloc("entry",
			Valu("mem", OpInitMem, types.TypeMem, 0, nil),
			Valu("a0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b0", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("a1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("b1", OpArg, c.config.Types.UInt32, 0, nil),
			Valu("addc", OpAdd32carry, types.NewTuple(c.config.Types.UInt32, types.TypeFlags), 0, nil, "a0", "b0"),
			Valu("lo", OpSelect0, c.config.Types.UInt32, 0, nil, "addc"),
			Valu("ca", OpSelect1, types.TypeFlags, 0, nil, "addc"),
			// A carry-clobbering instruction with no dependence on the pair.
			Valu("clobber", OpPPCSRAWI, c.config.Types.UInt32, 1, nil, "a0"),
			Valu("adde", OpAdd32withcarry, c.config.Types.UInt32, 0, nil, "a1", "b1", "ca"),
			Valu("sp", OpSP, c.config.Types.Uintptr, 0, nil),
			Valu("off", OpOffPtr, types.NewPtr(c.config.Types.UInt32), 0, nil, "sp"),
			Valu("st", OpPPCMOVWstore, types.TypeMem, 0, nil, "off", "adde", "mem"),
			// The clobbering instruction needs a use of its own, or the dead
			// value pass removes it before the flag allocator ever sees it.
			Valu("st2", OpPPCMOVWstore, types.TypeMem, 0, nil, "off", "clobber", "st"),
			Exit("st2")),
	)
	lower(f.f)
	checkLower(f.f)
	flagalloc(f.f)

	b := f.blocks["entry"]
	var addcs []*Value
	var clobberIndex, addeIndex int
	for i, v := range b.Values {
		switch v.Op {
		case OpPPCADDC:
			addcs = append(addcs, v)
		case OpPPCSRAWI:
			clobberIndex = i
		case OpPPCADDE:
			addeIndex = i
		}
	}
	if len(addcs) != 2 {
		t.Fatalf("block has %d addc instructions, want 2 (the original and the recomputed one)", len(addcs))
	}
	if !(clobberIndex < indexOfValue(b, addcs[1]) && indexOfValue(b, addcs[1]) < addeIndex) {
		t.Fatalf("the recomputed addc is not between the clobber and the adde: clobber at %d, addcs at %d,%d, adde at %d",
			clobberIndex, indexOfValue(b, addcs[0]), indexOfValue(b, addcs[1]), addeIndex)
	}
	adde := f.values["adde"]
	if adde.Args[2].Op != OpSelect1 || adde.Args[2].Args[0] != addcs[1] {
		t.Fatalf("adde reads the carry from %v, want the recomputed addc at %d",
			adde.Args[2].LongString(), indexOfValue(b, addcs[1]))
	}
}

func indexOfValue(b *Block, want *Value) int {
	for i, v := range b.Values {
		if v == want {
			return i
		}
	}
	panic("value not in block")
}
