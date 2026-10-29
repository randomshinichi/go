// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build math_big_pure_go || ppc

package big

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// TestArithDeclPure compares the linkname entry points with the generic bodies,
// including the length checks that the assembly implementations do not provide.
func TestArithDeclPure(t *testing.T) {
	max := ^Word(0)
	vectors := [][]Word{nil, {}, {0}, {1}, {max}, {max, max, max}, {0, 0, 0}, {0, max, 1, max}}
	rng := rand.New(rand.NewSource(67401))
	for n := 0; n < 32; n++ {
		v := make([]Word, n)
		for i := range v {
			v[i] = Word(rng.Uint64())
		}
		vectors = append(vectors, v)
	}

	for i, x := range vectors {
		t.Run(fmt.Sprintf("vector%d", i), func(t *testing.T) {
			y := make([]Word, len(x))
			for i := range y {
				y[i] = max
			}
			for _, op := range []struct {
				name      string
				got, want func(z, x, y []Word) Word
			}{
				{"addVV", addVV, addVV_g},
				{"subVV", subVV, subVV_g},
			} {
				t.Run(op.name, func(t *testing.T) {
					checkArithDeclPure(t, x, func(z []Word) Word { return op.got(z, x, y) }, func(z []Word) Word { return op.want(z, x, y) }, nil)
					short := make([]Word, len(x)+1)
					checkArithDeclPure(t, short, func(z []Word) Word { return op.got(z, x, y) }, func(z []Word) Word { return op.want(z, x, y) }, op.name+" len")
					checkArithDeclPure(t, x, func(z []Word) Word { return op.got(z, x, short) }, func(z []Word) Word { return op.want(z, x, short) }, op.name+" len")
				})
			}

			for _, s := range []uint{0, 1, _W - 1} {
				t.Run(fmt.Sprintf("shlVU/shift%d", s), func(t *testing.T) {
					checkArithDeclPure(t, y, func(z []Word) Word { return shlVU(z, x, s) }, func(z []Word) Word { return lshVU_g(z, x, s) }, nil)
					if s != 0 {
						short := make([]Word, len(x)+1)
						checkArithDeclPure(t, short, func(z []Word) Word { return shlVU(z, x, s) }, func(z []Word) Word { return lshVU_g(z, x, s) }, "lshVU len")
					}
				})
			}

			for _, m := range []Word{0, 1, max} {
				t.Run(fmt.Sprintf("mulAddVWW/m%d", m), func(t *testing.T) {
					checkArithDeclPure(t, y, func(z []Word) Word { return mulAddVWW(z, x, m, max) }, func(z []Word) Word { return mulAddVWW_g(z, x, m, max) }, nil)
					short := make([]Word, len(x)+1)
					checkArithDeclPure(t, short, func(z []Word) Word { return mulAddVWW(z, x, m, max) }, func(z []Word) Word { return mulAddVWW_g(z, x, m, max) }, "mulAddVWW len")
				})
				t.Run(fmt.Sprintf("addMulVVW/m%d", m), func(t *testing.T) {
					checkArithDeclPure(t, y, func(z []Word) Word { return addMulVVW(z, x, m) }, func(z []Word) Word { return addMulVVWW_g(z, z, x, m, 0) }, nil)
					short := make([]Word, len(x)+1)
					checkArithDeclPure(t, short, func(z []Word) Word { return addMulVVW(z, x, m) }, func(z []Word) Word { return addMulVVWW_g(z, z, x, m, 0) }, "addMulVVW len")
				})
			}
		})
	}

	// Like the assembly-path compatibility wrapper, a zero shift copies even
	// when lengths differ. lshVU_g itself checks lengths before checking s.
	for _, z := range [][]Word{nil, {max}, {max, max, max}} {
		x := []Word{1, 2}
		checkArithDeclPure(t, z, func(z []Word) Word { return shlVU(z, x, 0) }, func(z []Word) Word { copy(z, x); return 0 }, nil)
	}
}

func checkArithDeclPure(t *testing.T, initial []Word, got, want func([]Word) Word, wantPanic any) {
	t.Helper()
	type result struct {
		z          []Word
		carry      Word
		panicValue any
	}
	run := func(f func([]Word) Word) (r result) {
		r.z = append([]Word(nil), initial...)
		defer func() { r.panicValue = recover() }()
		r.carry = f(r.z)
		return
	}
	g, w := run(got), run(want)
	if !reflect.DeepEqual(w.panicValue, wantPanic) {
		t.Fatalf("reference panic = %v, want %v", w.panicValue, wantPanic)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("entry point = %+v, generic = %+v", g, w)
	}
}
