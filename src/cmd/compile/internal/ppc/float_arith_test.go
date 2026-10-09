// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestFloatArithmeticQEMU(t *testing.T) {
	for _, width := range []int{32, 64} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			var src strings.Builder
			src.WriteString("package main\n")
			functions := []string{"add", "sub", "mul", "div", "neg"}
			for i, op := range []string{"x+y", "x-y", "x*y", "x/y", "-x"} {
				fmt.Fprintf(&src, "//go:noinline\n//go:nosplit\nfunc %s(x,y float%d) float%d {return %s}\n", functions[i], width, width, op)
			}
			dir, dump := compileFloatFixture(t, src.String(), functions, "")
			var body strings.Builder
			tiny := math.SmallestNonzeroFloat64
			if width == 32 {
				tiny = math.SmallestNonzeroFloat32
			}
			for _, pair := range [][2]float64{{0, math.Copysign(0, -1)}, {math.Copysign(0, -1), 2}, {1.25, 2.5}, {-123.5, 0.25}, {1.5, math.Inf(1)}, {tiny, 2}, {math.Ldexp(1, 24), 1}} {
				x, y := pair[0], pair[1]
				want := []float64{x + y, x - y, x * y, x / y, -x}
				if width == 32 {
					a, b := float32(x), float32(y)
					want = []float64{float64(a + b), float64(a - b), float64(a * b), float64(a / b), float64(-a)}
				}
				for i, fn := range functions {
					if math.IsNaN(want[i]) {
						continue
					} // NaN payload is not promised; comparison tests cover NaNs.
					if width == 32 {
						body.WriteString(setWord(24, math.Float32bits(float32(x))) + setWord(28, math.Float32bits(float32(y))) + "bl fn_" + fn + "\n" + checkWord(32, math.Float32bits(float32(want[i]))))
					} else {
						body.WriteString(setDouble(24, math.Float64bits(x)) + setDouble(32, math.Float64bits(y)) + "bl fn_" + fn + "\n" + checkDouble(40, math.Float64bits(want[i])))
					}
				}
			}
			runFloatELF(t, dir, dump, functions, body.String(), nil)
		})
	}
}

func TestFloatComparisonsQEMU(t *testing.T) {
	for _, width := range []int{32, 64} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			var src strings.Builder
			src.WriteString("package main\n")
			var functions []string
			for i, op := range []string{"==", "!=", "<", "<="} {
				for _, form := range []string{"bool", "branch", "inverted"} {
					fn := fmt.Sprintf("cmp%d_%s", i, form)
					functions = append(functions, fn)
					fmt.Fprintf(&src, "//go:noinline\n//go:nosplit\nfunc %s(x,y float%d", fn, width)
					if form == "bool" {
						fmt.Fprintf(&src, ") bool {return x%s y}\n", op)
					} else if form == "branch" {
						fmt.Fprintf(&src, ",p *uint32) {if x%s y {*p=17;return};*p=23}\n", op)
					} else {
						fmt.Fprintf(&src, ",p *uint32) {if !(x%s y) {*p=23;return};*p=17}\n", op)
					}
				}
			}
			dir, dump := compileFloatFixture(t, src.String(), functions, "")
			var body strings.Builder
			for _, pair := range [][2]float64{{1, 2}, {2, 1}, {1, 1}, {0, math.Copysign(0, -1)}, {math.NaN(), 1}, {1, math.NaN()}, {math.NaN(), math.NaN()}, {math.Inf(1), math.Inf(1)}, {math.Inf(-1), math.Inf(1)}, {math.SmallestNonzeroFloat32, 0}} {
				x, y := pair[0], pair[1]
				want := []bool{x == y, x != y, x < y, x <= y}
				for i, fn := range functions {
					off := 32
					if width == 32 {
						body.WriteString(setWord(24, math.Float32bits(float32(x))) + setWord(28, math.Float32bits(float32(y))))
					} else {
						body.WriteString(setDouble(24, math.Float64bits(x)) + setDouble(32, math.Float64bits(y)))
						off = 40
					}
					if strings.HasSuffix(fn, "bool") {
						// Bool occupies one byte; sentinel bytes prove no over-wide result store.
						body.WriteString(setWord(off, 0x00abcdef) + "bl fn_" + fn + "\n")
						b := uint32(0x00abcdef)
						if want[i/3] {
							b |= 1 << 24
						}
						body.WriteString(checkWord(off, b))
					} else {
						fmt.Fprintf(&body, "addi 3,1,256\nstw 3,%d(1)\nbl fn_%s\n", off, fn)
						b := uint32(23)
						if want[i/3] {
							b = 17
						}
						body.WriteString(checkWord(256, b))
					}
				}
			}
			t.Run("ordered-return-and-both-branch-layouts", func(t *testing.T) { runFloatELF(t, dir, dump, functions, body.String(), nil) })
			t.Run("negative-unordered-le", func(t *testing.T) {
				runFloatELF(t, dir, dump, functions, body.String(), func(fn string, w uint32) uint32 {
					if fn == "cmp3_bool" && w>>26 == 16 && (w>>16)&31 == 3 {
						return 0x60000000
					} // remove unordered rejection
					return w
				})
			})
		})
	}
}
