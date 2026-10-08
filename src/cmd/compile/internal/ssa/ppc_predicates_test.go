// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"testing"
)

func TestPPCConstantPredicates(t *testing.T) {
	c := testConfigPPC32(t)
	predicates := []struct {
		op   Op
		want [3]int64 // EQ, LT, GT
	}{
		{OpPPCEqual, [3]int64{1, 0, 0}},
		{OpPPCNotEqual, [3]int64{0, 1, 1}},
		{OpPPCLessThan, [3]int64{0, 1, 0}},
		{OpPPCLessEqual, [3]int64{1, 1, 0}},
		{OpPPCGreaterThan, [3]int64{0, 0, 1}},
		{OpPPCGreaterEqual, [3]int64{1, 0, 1}},
	}
	for _, p := range predicates {
		for i, flag := range []Op{OpPPCFlagEQ, OpPPCFlagLT, OpPPCFlagGT} {
			t.Run(p.op.String()+"/"+flag.String(), func(t *testing.T) {
				fun := c.Fun("entry", Bloc("entry",
					Valu("mem", OpInitMem, types.TypeMem, 0, nil),
					Valu("flag", flag, types.TypeFlags, 0, nil),
					Valu("bool", p.op, c.config.Types.Bool, 0, nil, "flag"),
					Exit("mem")))
				v := fun.values["bool"]
				if !rewriteValuePPC(v) || v.Op != OpPPCMOVWconst || v.AuxInt != p.want[i] {
					t.Fatalf("predicate = %s, want MOVWconst [%d]", v.LongString(), p.want[i])
				}
			})
		}
	}
}

func TestPPCInvertPredicates(t *testing.T) {
	c := testConfigPPC32(t)
	for _, p := range []struct{ op, want Op }{
		{OpPPCEqual, OpPPCEqual}, {OpPPCNotEqual, OpPPCNotEqual},
		{OpPPCLessThan, OpPPCGreaterThan}, {OpPPCLessEqual, OpPPCGreaterEqual},
		{OpPPCGreaterThan, OpPPCLessThan}, {OpPPCGreaterEqual, OpPPCLessEqual},
	} {
		t.Run(p.op.String(), func(t *testing.T) {
			fun := c.Fun("entry", Bloc("entry",
				Valu("mem", OpInitMem, types.TypeMem, 0, nil),
				Valu("flag", OpPPCFlagEQ, types.TypeFlags, 0, nil),
				Valu("invert", OpPPCInvertFlags, types.TypeFlags, 0, nil, "flag"),
				Valu("bool", p.op, c.config.Types.Bool, 0, nil, "invert"),
				Exit("mem")))
			v := fun.values["bool"]
			if !rewriteValuePPC(v) || v.Op != p.want || v.Args[0] != fun.values["flag"] {
				t.Fatalf("predicate = %s, want %s of original flags", v.LongString(), p.want)
			}
		})
	}
}

func TestPPCWrappedPredicates(t *testing.T) {
	c := testConfigPPC32(t)
	for _, p := range []struct {
		op   Op
		want BlockKind
	}{
		{OpPPCEqual, BlockPPCEQ}, {OpPPCNotEqual, BlockPPCNE},
		{OpPPCLessThan, BlockPPCLT}, {OpPPCLessEqual, BlockPPCLE},
		{OpPPCGreaterThan, BlockPPCGT}, {OpPPCGreaterEqual, BlockPPCGE},
	} {
		t.Run(p.op.String(), func(t *testing.T) {
			fun := c.Fun("entry",
				Bloc("entry",
					Valu("mem", OpInitMem, types.TypeMem, 0, nil),
					Valu("flag", OpPPCFlagEQ, types.TypeFlags, 0, nil),
					Valu("bool", p.op, c.config.Types.Bool, 0, nil, "flag"),
					Valu("wrapped", OpPPCCMPWconst, types.TypeFlags, 0, nil, "bool"),
					If("wrapped", "yes", "no")),
				Bloc("yes", Exit("mem")),
				Bloc("no", Exit("mem")))
			b := fun.blocks["entry"]
			b.Kind = BlockPPCNE
			if !rewriteBlockPPC(b) || b.Kind != p.want || b.Controls[0] != fun.values["flag"] {
				t.Fatalf("wrapped predicate block = %s, want %s of original flags", b.LongString(), p.want)
			}
		})
	}
}
