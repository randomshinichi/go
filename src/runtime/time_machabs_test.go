// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"math/bits"
	"runtime"
	"testing"
)

// exactMachTime is floor(ticks*numer/denom) computed through a 128-bit product.
func exactMachTime(ticks uint64, numer, denom uint32) uint64 {
	hi, lo := bits.Mul64(ticks, uint64(numer))
	q, _ := bits.Div64(hi, lo, uint64(denom))
	return q
}

// wrappedMachTime is the conversion nanotime1 used to do: t *= numer; t /= denom
// in int64.
func wrappedMachTime(ticks int64, numer, denom uint32) int64 {
	t := ticks
	if numer != 1 {
		t *= int64(numer)
	}
	if denom != 1 {
		t /= int64(denom)
	}
	return t
}

// TestMachTimeToNanos checks the mach_absolute_time conversion against exact
// arithmetic. The Power Mac G5's timebase is 33.3 MHz, so numer is about 1e9
// and ticks*numer leaves int64 range after ~277 s of uptime; the old
// conversion then returned a negative nanotime, which made
// (*scavengerState).sleep call timer.reset with a negative `when` ("timer
// when must be positive").
func TestMachTimeToNanos(t *testing.T) {
	bases := []struct {
		name         string
		numer, denom uint32
		hz           uint64
	}{
		{"g5-33.3MHz", 1000000000, 33333333, 33333333},
		{"25MHz", 1000000000, 25000000, 25000000},
		{"apple-silicon-24MHz", 125, 3, 24000000},
		{"identity", 1, 1, 1000000000},
		{"equal-ratio", 7, 7, 1000000000},
		{"max-uint32", 0xffffffff, 0xfffffffe, 1000000000},
	}
	uptimes := []uint64{0, 1, 2, 10, 276, 277, 278, 282, 300, 553, 554, 830, 3600, 86400, 30 * 86400, 365 * 86400}
	for _, b := range bases {
		check := func(ticks uint64) {
			t.Helper()
			want := exactMachTime(ticks, b.numer, b.denom)
			if got := runtime.MachTimeToNanos(ticks, b.numer, b.denom); got != int64(want) {
				t.Errorf("%s: MachTimeToNanos(%d, %d, %d) = %d, want %d", b.name, ticks, b.numer, b.denom, got, int64(want))
			}
		}
		for _, s := range uptimes {
			check(s * b.hz)
			check(s*b.hz + b.hz/2)
			check(s*b.hz + 1)
			if s > 0 {
				check(s*b.hz - 1)
			}
		}
		// Straddle the point where ticks*numer reaches 2^63.
		edge := uint64(1<<63) / uint64(b.numer)
		for d := uint64(0); d < 4; d++ {
			check(edge - 2 + d)
		}
	}
}

// TestMachTimeToNanosNeverNegativeG5 samples the whole first hour of uptime on
// the G5 timebase: the result must stay non-negative and strictly follow the
// tick count, across the wraps of the old 64-bit product at 277 s and 553 s.
func TestMachTimeToNanosNeverNegativeG5(t *testing.T) {
	const numer, denom, hz = 1000000000, 33333333, 33333333
	var prev int64 = -1
	for s := uint64(0); s <= 3600; s++ {
		got := runtime.MachTimeToNanos(s*hz, numer, denom)
		if got < 0 || got <= prev {
			t.Fatalf("uptime %d s: nanotime %d (previous %d)", s, got, prev)
		}
		prev = got
	}
}

// TestWrappedMachTimeGoesNegativeG5 documents the defect: the former int64
// product-then-divide conversion on the same timebase is negative for uptimes
// in [277 s, 553 s). It states the premise of the test above; it does not
// exercise runtime code.
func TestWrappedMachTimeGoesNegativeG5(t *testing.T) {
	const numer, denom, hz = 1000000000, 33333333, 33333333
	if got := wrappedMachTime(100*hz, numer, denom); got < 0 {
		t.Errorf("uptime 100 s: old conversion = %d, want positive", got)
	}
	if got := wrappedMachTime(282*hz, numer, denom); got >= 0 {
		t.Errorf("uptime 282 s: old conversion = %d, want negative (int64 overflow of ticks*numer)", got)
	}
}
