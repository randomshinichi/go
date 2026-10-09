// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm || 386 || mips || mipsle || ppc

package runtime_test

import (
	"math"
	bigfloat "math/big"
	"math/bits"
	"testing"
)

//go:noinline
func wideUnsignedSingle(x uint64) float32 { return float32(x) }

//go:noinline
func wideUnsignedDouble(x uint64) float64 { return float64(x) }

//go:noinline
func wideSignedSingle(x int64) float32 { return float32(x) }

//go:noinline
func wideSignedDouble(x int64) float64 { return float64(x) }

// Integer-only IEEE round-to-nearest-even; also cross-checked against math/big.
func wideIntegerBits(x uint64, fraction, bias int) uint64 {
	if x == 0 {
		return 0
	}
	exp := bits.Len64(x) - 1
	shift := exp - fraction
	var q uint64
	if shift > 0 {
		q = x >> shift
		rem := x & ((uint64(1) << shift) - 1)
		half := uint64(1) << (shift - 1)
		if rem > half || rem == half && q&1 != 0 {
			q++
		}
	} else {
		q = x << (-shift)
	}
	if q == uint64(1)<<(fraction+1) {
		q >>= 1
		exp++
	}
	return uint64(exp+bias)<<fraction | q&((uint64(1)<<fraction)-1)
}

func TestWideFloatIntegerOracle(t *testing.T) {
	count := 0
	check := func(x uint64) {
		t.Helper()
		want32 := uint32(wideIntegerBits(x, 23, 127))
		want64 := wideIntegerBits(x, 52, 1023)
		oracle := new(bigfloat.Float).SetPrec(24).SetMode(bigfloat.ToNearestEven).SetUint64(x)
		big32, _ := oracle.Float32()
		big64, _ := new(bigfloat.Float).SetPrec(53).SetMode(bigfloat.ToNearestEven).SetUint64(x).Float64()
		if math.Float32bits(big32) != want32 || math.Float64bits(big64) != want64 {
			t.Fatalf("independent oracles disagree for %016x", x)
		}
		got32 := math.Float32bits(wideUnsignedSingle(x))
		got64 := math.Float64bits(wideUnsignedDouble(x))
		if got32 != want32 || got64 != want64 {
			t.Fatalf("unsigned x=%016x single=%08x want=%08x double=%016x want=%016x", x, got32, want32, got64, want64)
		}
		mag, sign32, sign64 := x, uint32(0), uint64(0)
		if int64(x) < 0 {
			mag = -x
			sign32 = 1 << 31
			sign64 = 1 << 63
		}
		want32 = uint32(wideIntegerBits(mag, 23, 127)) | sign32
		want64 = wideIntegerBits(mag, 52, 1023) | sign64
		got32 = math.Float32bits(wideSignedSingle(int64(x)))
		got64 = math.Float64bits(wideSignedDouble(int64(x)))
		if got32 != want32 || got64 != want64 {
			t.Fatalf("signed x=%016x single=%08x want=%08x double=%016x want=%016x", x, got32, want32, got64, want64)
		}
		count++
	}
	// Include the old helper's guard/sticky defect and both directions of
	// binary64 double-rounding around a binary32 midpoint.
	for _, x := range []uint64{1<<46 + 1, 1<<46 + (1 << 22) + 1, 1<<63 + 1, 1<<63 + (1 << 39) + 1, 1<<63 + 3*(1<<39) - 1, ^uint64(0)} {
		check(x)
	}
	for e := 0; e < 64; e++ {
		base := uint64(1) << e
		for _, delta := range []uint64{0, 1, 2, 3, 31, 4095, 4096, 4097} {
			check(base + delta)
			check(base - delta)
		}
		if e >= 24 {
			half := uint64(1) << (e - 24)
			for _, k := range []uint64{1, 3, 5, 7} {
				for _, d := range []uint64{k*half - 1, k * half, k*half + 1} {
					check(base + d)
					check(-base - d)
				}
			}
		}
	}
	x := uint64(0x9e3779b97f4a7c15)
	for i := 0; i < 100000; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		check(x)
	}
	t.Logf("PASS: %d vectors, four real conversions each, integer oracle agrees with math/big nearest-even", count)
}
