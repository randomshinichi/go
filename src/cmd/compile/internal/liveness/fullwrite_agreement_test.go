// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package liveness

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
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
// isfat over the 16 hand-picked shapes in its table, and it pins isfat's own
// answer for each of them in the fat column.
//
// The table is a sample, not an enumeration: an edit to isfat that leaves
// these 16 answers alone walks straight past it (goppc-tvtr: a one-field rule
// that ignores zero-length array fields changes isfat for struct{ _ [0]func();
// p *T } and this table does not notice).
// TestFullWriteAgreesWithIsFatExhaustive is the enumeration -- the same
// comparison over a generated cross product of type shapes.
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

// wrapping is one of the five ways the generated comparison wraps a base type.
// They are the shapes isfat's two recursive cases look through (array1 and
// structX) and the shapes it answers without looking (array0, array2 and
// structXptr); array0 is the one that is fat while contributing no field to a
// struct that contains it, which is the shape goppc-tvtr's edit C moved.
type wrapping struct {
	name string
	wrap func(*types.Type) *types.Type
}

// wrappings returns the five wrappings, with ptr as the pointer the two-field
// struct carries as its second field.
func wrappings(ptr *types.Type) []wrapping {
	return []wrapping{
		{"array0", func(x *types.Type) *types.Type { return types.NewArray(x, 0) }},
		{"array1", func(x *types.Type) *types.Type { return types.NewArray(x, 1) }},
		{"array2", func(x *types.Type) *types.Type { return types.NewArray(x, 2) }},
		{"structX", func(x *types.Type) *types.Type {
			return types.NewStruct([]*types.Field{{Type: x}})
		}},
		{"structXptr", func(x *types.Type) *types.Type {
			return types.NewStruct([]*types.Field{{Type: x}, {Type: ptr}})
		}},
	}
}

// typeShape is one generated type together with a name describing how it was
// built. The name is built from the generator's own labels rather than from
// the type, because the base set includes pseudo-types whose printed form is
// not a type expression at all.
type typeShape struct {
	name string
	typ  *types.Type
}

// exhaustiveShapes builds the type shape set the exhaustive agreement check
// runs over: every non-nil predeclared type in types.Types, a pointer, a
// slice, a map, a channel, a func, struct{} and the SIMD struct, each wrapped
// two levels deep through the five wrappings -- 31 shapes per base.
//
// nil is covered unwrapped and only unwrapped: a nil element or field type is
// not something a wrapper can be built around, because types.NewArray calls
// elem.HasShape() and types.NewStruct calls f.Type.NotInHeap() on it.
//
// Two levels, not one, is what reaches the shapes the pin is weakest on: a
// struct{ _ [0]func(); p *uint8 } is structXptr applied to array0 applied to
// the func base, and one store to that struct is not a full write while
// isfat's field count makes it fat.
func exhaustiveShapes(t *testing.T) []typeShape {
	ptr := types.NewPtr(types.Types[types.TUINT8])

	nilShape := typeShape{"nil", nil}
	bases := []typeShape{nilShape}
	for k := types.Kind(0); k < types.NTYPE; k++ {
		if typ := types.Types[k]; typ != nil {
			bases = append(bases, typeShape{fmt.Sprintf("Types[%v]", k), typ})
		}
	}
	bases = append(bases,
		typeShape{"ptr", ptr},
		typeShape{"slice", types.NewSlice(ptr)},
		typeShape{"map", types.NewMap(types.Types[types.TSTRING], ptr)},
		typeShape{"chan", types.NewChan(types.Types[types.TUINT8], types.Cboth)},
		typeShape{"func", types.NewSignature(nil, nil, nil)},
		typeShape{"struct{}", types.NewStruct(nil)},
		typeShape{"simdstruct", simdStruct(t)},
	)

	ws := wrappings(ptr)
	shapes := []typeShape{nilShape} // nil, unwrapped
	for _, base := range bases[1:] {
		before := len(shapes)
		shapes = append(shapes, base)
		one := make([]typeShape, 0, len(ws))
		for _, w := range ws {
			one = append(one, typeShape{base.name + "/" + w.name, w.wrap(base.typ)})
		}
		shapes = append(shapes, one...)
		for _, o := range one {
			for _, w := range ws {
				shapes = append(shapes, typeShape{o.name + "/" + w.name, w.wrap(o.typ)})
			}
		}
		if got, want := len(shapes)-before, 1+len(ws)+len(ws)*len(ws); got != want {
			t.Fatalf("base %s generated %d shapes, want %d (itself, one level of wrapping and two)", base.name, got, want)
		}
	}

	return shapes
}

// TestFullWriteAgreesWithIsFatExhaustive is the exhaustive half of the pin on
// ssa.FullWrite == !isfat: it compares the two predicates over every shape of
// exhaustiveShapes instead of over the 16 hand-picked rows of
// TestFullWriteAgreesWithIsFat, so an edit to isfat that leaves those 16
// answers alone is still caught, and the failure names the shape it changed.
//
// This is the check that has to run against an edited isfat for the pin to
// mean anything: the table alone passes edits A, B and C of goppc-tvtr,
// including the one that changes the answer for a pointer-carrying
// struct{ _ [0]func(); p *T }.
func TestFullWriteAgreesWithIsFatExhaustive(t *testing.T) {
	shapes := exhaustiveShapes(t)
	t.Logf("generated %d type shapes", len(shapes))
	for _, s := range shapes {
		t.Run(s.name, func(t *testing.T) {
			fat := isfat(s.typ)
			if got := ssa.FullWrite(s.typ); got != !fat {
				t.Errorf("ssa.FullWrite(%v) = %v, want !isfat(%v) = %v: the scheduler's full write test disagrees with liveness's isfat for this shape", s.typ, got, s.typ, !fat)
			}
		})
	}
}
