// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"math"
	"strings"
	"testing"
)

const floatConversionFixture = `package main
//go:noinline
//go:nosplit
func signed64(x int32) float64 {return float64(x)}
//go:noinline
//go:nosplit
func signed32(x int32) float32 {return float32(x)}
//go:noinline
//go:nosplit
func unsigned64(x uint32) float64 {return float64(x)}
//go:noinline
//go:nosplit
func unsigned32(x uint32) float32 {return float32(x)}
//go:noinline
//go:nosplit
func toSigned(x float64) int32 {return int32(x)}
//go:noinline
//go:nosplit
func toUnsigned(x float64) uint32 {return uint32(x)}
//go:noinline
//go:nosplit
func singleToSigned(x float32) int32 {return int32(x)}
//go:noinline
//go:nosplit
func singleToUnsigned(x float32) uint32 {return uint32(x)}
`

func TestFloatConversionsQEMU(t *testing.T) {
	functions := []string{"signed64", "signed32", "unsigned64", "unsigned32", "toSigned", "toUnsigned", "singleToSigned", "singleToUnsigned"}
	dir, dump := compileFloatFixture(t, floatConversionFixture, functions, "")
	var body strings.Builder
	for _, x := range []uint32{0, 1, 16777215, 16777216, 16777217, 2147483647, 2147483648, 2147483649, 4294967294, 4294967295} {
		body.WriteString(setWord(24, x) + "bl fn_unsigned64\n" + checkDouble(28, math.Float64bits(float64(x))))
		body.WriteString(setWord(24, x) + "bl fn_unsigned32\n" + checkWord(28, math.Float32bits(float32(x))))
		body.WriteString(setWord(24, x) + "bl fn_signed64\n" + checkDouble(28, math.Float64bits(float64(int32(x)))))
		body.WriteString(setWord(24, x) + "bl fn_signed32\n" + checkWord(28, math.Float32bits(float32(int32(x)))))
	}
	for _, x := range []float64{0, math.Copysign(0, -1), math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64, 0.9, -0.9, 1.9, -1.9, 16777215.9, 16777216.9, 16777217.9, -16777217.9, 2147483646.9, 2147483647.9, -2147483648, -2147483648.9} {
		body.WriteString(setDouble(24, math.Float64bits(x)) + "bl fn_toSigned\n" + checkWord(32, uint32(int32(x))))
		f := float32(x)
		if float64(f) >= -2147483648 && float64(f) < 2147483648 {
			body.WriteString(setWord(24, math.Float32bits(f)) + "bl fn_singleToSigned\n" + checkWord(28, uint32(int32(f))))
		}
	}
	for _, x := range []float64{0, math.Copysign(0, -1), math.SmallestNonzeroFloat64, 0.9, 1.9, 16777215.9, 16777216.9, 16777217.9, 2147483647.9, 2147483648, 2147483648.9, 2147483649.9, 4294967294.9, 4294967295, 4294967295.9} {
		body.WriteString(setDouble(24, math.Float64bits(x)) + "bl fn_toUnsigned\n" + checkWord(32, uint32(x)))
		f := float32(x)
		if float64(f) < 4294967296 {
			body.WriteString(setWord(24, math.Float32bits(f)) + "bl fn_singleToUnsigned\n" + checkWord(28, uint32(f)))
		}
	}
	t.Run("bias-truncation-high-bit-boundaries", func(t *testing.T) { runFloatELF(t, dir, dump, functions, body.String(), nil) })
	t.Run("negative-unsigned-high-bit", func(t *testing.T) {
		runFloatELF(t, dir, dump, functions, body.String(), func(fn string, w uint32) uint32 {
			if fn == "toUnsigned" && w>>26 == 27 {
				return 0x60000000
			} // xoris restoring high bit -> nop
			return w
		})
	})
}
