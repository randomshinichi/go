// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math_test

import (
	. "math"
	"math/big"
	"testing"
)

// The ppc970 floor has no hardware round-to-integral instruction (FRIN/FRIZ/FRIP/FRIM are
// Power ISA v2.02, absent on the 970), so Floor, Ceil, Trunc and Round run the portable
// Go code there. This test checks all of them against exact big.Float arithmetic, bit for
// bit (so the sign of zero is checked), over the edge values where round-to-integral
// implementations usually break: zeros, infinities, NaN, the smallest denormal, values
// around 0.5, and values at, below and above 2^52, 2^53 and 2^63.

// roundRef returns the exactly rounded integral value of x for the given rule, as a float64.
func roundRef(x float64, rule string) float64 {
	if IsNaN(x) || IsInf(x, 0) || x == 0 {
		return x
	}
	f := new(big.Float).SetPrec(2200).SetFloat64(x)
	z, _ := f.Int(nil) // truncated toward zero
	zf := new(big.Float).SetPrec(2200).SetInt(z)
	exact := f.Cmp(zf) == 0
	switch rule {
	case "trunc":
	case "floor":
		if !exact && x < 0 {
			z.Sub(z, big.NewInt(1))
		}
	case "ceil":
		if !exact && x > 0 {
			z.Add(z, big.NewInt(1))
		}
	case "round": // half away from zero
		diff := new(big.Float).SetPrec(2200).Sub(f, zf)
		diff.Abs(diff)
		if diff.Cmp(big.NewFloat(0.5)) >= 0 {
			if x < 0 {
				z.Sub(z, big.NewInt(1))
			} else {
				z.Add(z, big.NewInt(1))
			}
		}
	case "roundeven":
		diff := new(big.Float).SetPrec(2200).Sub(f, zf)
		diff.Abs(diff)
		c := diff.Cmp(big.NewFloat(0.5))
		if c > 0 || (c == 0 && z.Bit(0) == 1) {
			if x < 0 {
				z.Sub(z, big.NewInt(1))
			} else {
				z.Add(z, big.NewInt(1))
			}
		}
	}
	r, _ := new(big.Float).SetPrec(2200).SetInt(z).Float64()
	if r == 0 && Signbit(x) {
		return Copysign(0, -1)
	}
	return r
}

func floorEdgeValues() []float64 {
	base := []float64{
		0, 0.5, 0.49999999999999994, 1, 1.5, 2.5, 3.5, 4503599627370495.5, // 2^52 - 0.5
		4503599627370496, 4503599627370497, 4503599627370497.5, 9007199254740991, 9007199254740992,
		9223372036854775807, 9223372036854775808, 1.8446744073709552e19, MaxFloat64, SmallestNonzeroFloat64,
		Float64frombits(0x000fffffffffffff), 1e-300, 1e300, Pi, E, 1 << 31, 1<<31 + 0.5, 1 << 32, 1<<32 + 0.5,
	}
	var vals []float64
	for _, b := range base {
		for _, s := range []float64{1, -1} {
			x := s * b
			vals = append(vals, x, Nextafter(x, Inf(1)), Nextafter(x, Inf(-1)))
		}
	}
	vals = append(vals, Inf(1), Inf(-1), NaN(), Copysign(0, -1))
	// deterministic spread over the whole exponent range
	var s uint64 = 0x9e3779b97f4a7c15
	for i := 0; i < 20000; i++ {
		s = s*6364136223846793005 + 1442695040888963407
		vals = append(vals, Float64frombits(s))
	}
	return vals
}

func TestRoundToIntegralEdges(t *testing.T) {
	type fn struct {
		name, rule string
		f          func(float64) float64
	}
	fns := []fn{
		{"Trunc", "trunc", Trunc}, {"Floor", "floor", Floor}, {"Ceil", "ceil", Ceil},
		{"Round", "round", Round}, {"RoundToEven", "roundeven", RoundToEven},
	}
	for _, x := range floorEdgeValues() {
		for _, c := range fns {
			got, want := c.f(x), roundRef(x, c.rule)
			if IsNaN(want) {
				if !IsNaN(got) {
					t.Fatalf("%s(%v) = %v, want NaN", c.name, x, got)
				}
				continue
			}
			if Float64bits(got) != Float64bits(want) {
				t.Fatalf("%s(%v [%#x]) = %v [%#x], want %v [%#x]", c.name, x, Float64bits(x), got, Float64bits(got), want, Float64bits(want))
			}
		}
	}
}
