// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package liveness

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"internal/buildcfg"
	"testing"
)

// simdStruct returns a named one-field struct of the shape CalcStructSize
// recognizes as SIMD: a struct named v128 in package simd/archsimd with
// goexperiment.SIMD enabled (types/size.go:489). simdify, which sets the flag
// the readers of IsSIMD test, is unexported, so the only way to a SIMD struct
// from outside cmd/compile/internal/types is to name it the way the type
// checker would and let CalcSize mark it.
func simdStruct(t *testing.T) *types.Type {
	old := buildcfg.Experiment.SIMD
	buildcfg.Experiment.SIMD = true
	t.Cleanup(func() { buildcfg.Experiment.SIMD = old })

	pkg := types.NewPkg("simd/archsimd", "archsimd")
	named := types.NewNamed(simdObj{pkg.Lookup("v128")})
	named.SetUnderlying(types.NewStruct([]*types.Field{{Type: types.TypeVec128}}))
	types.CalcSize(named)
	return named
}

// simdObj is the Object types.NewNamed wants; NewNamed only reads the symbol.
type simdObj struct{ sym *types.Sym }

func (o simdObj) Pos() src.XPos     { return src.NoXPos }
func (o simdObj) Sym() *types.Sym   { return o.sym }
func (o simdObj) Type() *types.Type { return nil }

// TestFullWriteAgreesWithIsFat asserts that ssa.FullWrite, the scheduler's
// answer to "does one store to this type cover all of it", is the negation of
// isfat over every type shape a result can have.
//
// The comparison is against the real isfat in this package, not against a
// second copy of it: a reimplementation would agree with ssa.FullWrite
// wherever the two share a mistake, which is exactly the defect this test
// exists to catch (goppc-10mz: FullWrite called every array and struct not
// fully written, while isfat calls a store to [1]*T or struct{p *T} a kill).
// The scheduler orders a result's addresses after the first value that kills
// the result (ssa/killsVar), so a row that disagrees here is a schedule that
// orders an address after the OpVarDef of a result that something else already
// killed -- the edge X12's counterexample cannot satisfy.
//
// The fat column is what isfat answers, asserted as well: it records the
// authority's answer for each shape, so a change to either side shows up as
// the row that changed rather than as a silent disagreement.
func TestFullWriteAgreesWithIsFat(t *testing.T) {
	ptr := types.NewPtr(types.Types[types.TUINT8])
	str := types.Types[types.TSTRING]
	for _, tc := range []struct {
		name string
		typ  *types.Type
		fat  bool
	}{
		{"nil", nil, false},
		{"byte", types.Types[types.TUINT8], false},
		{"ptr", ptr, false},
		{"map", types.NewMap(str, ptr), false},
		{"slice", types.NewSlice(ptr), true},
		{"string", str, true},
		{"interface", types.Types[types.TINTER], true},

		{"array1ptr", types.NewArray(ptr, 1), false},
		{"array2ptr", types.NewArray(ptr, 2), true},
		{"array1string", types.NewArray(str, 1), true},
		{"array1struct1ptr", types.NewArray(types.NewStruct([]*types.Field{{Type: ptr}}), 1), false},

		{"struct1ptr", types.NewStruct([]*types.Field{{Type: ptr}}), false},
		{"struct1string", types.NewStruct([]*types.Field{{Type: str}}), true},
		{"struct2ptr", types.NewStruct([]*types.Field{{Type: ptr}, {Type: ptr}}), true},
		{"struct1array1ptr", types.NewStruct([]*types.Field{{Type: types.NewArray(ptr, 1)}}), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isfat(tc.typ); got != tc.fat {
				t.Errorf("isfat(%v) = %v, want %v", tc.typ, got, tc.fat)
			}
			if got := ssa.FullWrite(tc.typ); got != !tc.fat {
				t.Errorf("ssa.FullWrite(%v) = %v, want !isfat = %v", tc.typ, got, !tc.fat)
			}
		})
	}

	t.Run("simdstruct", func(t *testing.T) {
		simd := simdStruct(t)
		if !simd.IsSIMD() {
			t.Fatalf("test premise failed: %v is not SIMD, so the row does not test isfat's SIMD case", simd)
		}
		if isfat(simd) {
			t.Errorf("isfat(%v) = true, want false: a SIMD struct is not fat", simd)
		}
		if !ssa.FullWrite(simd) {
			t.Errorf("ssa.FullWrite(%v) = false, want true: one store covers a SIMD struct", simd)
		}
	})
}
