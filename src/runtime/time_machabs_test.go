// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	"math"
	"math/bits"
	"runtime"
	"testing"
)

// exactMachTime is floor(ticks*numer/denom) computed through a 128-bit product.
// ok is false when the quotient does not fit in 64 bits.
func exactMachTime(ticks uint64, numer, denom uint32) (q uint64, ok bool) {
	hi, lo := bits.Mul64(ticks, uint64(numer))
	if hi >= uint64(denom) {
		return 0, false
	}
	q, _ = bits.Div64(hi, lo, uint64(denom))
	return q, true
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
		// Read from the Power Mac G5 Quad's mach_timebase_info.
		{"g5-measured", 1000000000, 33330863, 33330863},
		// The 33.3 MHz timebase the D12 lane inferred before it was measured.
		{"g5-inferred-33333333", 1000000000, 33333333, 33333333},
		{"25MHz", 1000000000, 25000000, 25000000},
		{"apple-silicon-24MHz", 125, 3, 24000000},
		// numer < denom: a timebase faster than 1 GHz (3/125 ns per tick).
		{"numer-below-denom", 3, 125, 41666666666},
		{"numer-just-below-denom", 0xfffffffe, 0xffffffff, 1000000000},
		{"identity", 1, 1, 1000000000},
		{"equal-ratio", 7, 7, 1000000000},
		{"max-uint32", 0xffffffff, 0xfffffffe, 1000000000},
	}
	uptimes := []uint64{0, 1, 2, 10, 276, 277, 278, 282, 300, 553, 554, 830, 3600, 86400, 30 * 86400, 365 * 86400}
	for _, b := range bases {
		check := func(ticks uint64) {
			t.Helper()
			want, ok := exactMachTime(ticks, b.numer, b.denom)
			if !ok || want > math.MaxInt64 {
				return // not representable as a nanotime
			}
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
		// Straddle the points where ticks*numer reaches 2^63, 2^64 and 3*2^63,
		// where the former int64 product wrapped.
		for k := uint64(1); k <= 3; k++ {
			hi, lo := bits.Mul64(k, 1<<63)
			if hi != 0 {
				continue
			}
			edge := lo / uint64(b.numer)
			for d := uint64(0); d < 4; d++ {
				check(edge - 2 + d)
			}
		}
		// The end of the uint64 range, where it still converts to int64.
		check(math.MaxUint64)
		check(math.MaxUint64 - 1)
		check(math.MaxUint32)
		check(1 << 32)
	}
}

// TestMachTimeToNanosNeverNegativeG5 samples the whole first hour of uptime on
// the G5 timebases: the result must stay non-negative and strictly follow the
// tick count, across the wraps of the old 64-bit product at 277 s and 553 s.
func TestMachTimeToNanosNeverNegativeG5(t *testing.T) {
	for _, b := range []struct{ denom, hz uint64 }{{33330863, 33330863}, {33333333, 33333333}} {
		var prev int64 = -1
		for s := uint64(0); s <= 3600; s++ {
			got := runtime.MachTimeToNanos(s*b.hz, 1000000000, uint32(b.denom))
			if got < 0 || got <= prev {
				t.Fatalf("denom %d, uptime %d s: nanotime %d (previous %d)", b.denom, s, got, prev)
			}
			prev = got
		}
	}
}

// TestMachTimeToNanosZeroTimebase pins the behaviour for degenerate timebases.
// No real mach_timebase_info reports a zero denominator. numer == denom == 0
// used to panic (divide by zero) and now returns the tick count, because
// numer == denom is treated as the identity; numer != denom == 0 still panics.
func TestMachTimeToNanosZeroTimebase(t *testing.T) {
	if got := runtime.MachTimeToNanos(12345, 0, 0); got != 12345 {
		t.Errorf("MachTimeToNanos(12345, 0, 0) = %d, want 12345 (numer == denom is the identity)", got)
	}
	if got := runtime.MachTimeToNanos(12345, 0, 7); got != 0 {
		t.Errorf("MachTimeToNanos(12345, 0, 7) = %d, want 0", got)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("MachTimeToNanos(12345, 1000000000, 0) did not panic")
			}
		}()
		runtime.MachTimeToNanos(12345, 1000000000, 0)
	}()
}

// TestWrappedMachTimeGoesNegativeG5 documents the defect: the former int64
// product-then-divide conversion on the G5 timebase is negative for uptimes in
// [277 s, 553 s) and 553 s too low from there on. It states the premise of the
// tests above; it does not exercise runtime code.
func TestWrappedMachTimeGoesNegativeG5(t *testing.T) {
	const numer, denom, hz = 1000000000, 33330863, 33330863
	if got := wrappedMachTime(100*hz, numer, denom); got < 0 {
		t.Errorf("uptime 100 s: old conversion = %d, want positive", got)
	}
	if got := wrappedMachTime(282*hz, numer, denom); got >= 0 {
		t.Errorf("uptime 282 s: old conversion = %d, want negative (int64 overflow of ticks*numer)", got)
	}
}

// machReading is one raw mach_absolute_time value with the timebase in effect.
type machReading struct {
	ticks        uint64
	numer, denom uint32
}

// nanotimeBracketProblem checks a nanotime reading against two raw readings
// taken around it: nanotime must lie between the exact conversions of the two
// tick counts. No wrong conversion of the tick count can satisfy that, whatever
// its form. It returns a description of the violation, or "", and reports
// whether the former inline formula (wrappedMachTime) would itself have failed
// this check at these readings. When it would not (uptime too low for
// ticks*numer to reach 2^63), a pass says nothing about the call site.
func nanotimeBracketProblem(read func() machReading, now func() int64) (problem string, oldFormulaWouldFail bool) {
	before := read()
	got := now()
	after := read()
	if before.denom == 0 || before.numer != after.numer || before.denom != after.denom {
		return fmt.Sprintf("timebase unusable or changed between readings: %+v then %+v", before, after), false
	}
	lo, okLo := exactMachTime(before.ticks, before.numer, before.denom)
	hi, okHi := exactMachTime(after.ticks, after.numer, after.denom)
	if !okLo || !okHi {
		return fmt.Sprintf("readings %+v and %+v do not convert to a 64-bit nanosecond count", before, after), false
	}
	oldFormulaWouldFail = uint64(wrappedMachTime(int64(before.ticks), before.numer, before.denom)) != lo &&
		uint64(wrappedMachTime(int64(after.ticks), after.numer, after.denom)) != hi
	if got < 0 || uint64(got) < lo || uint64(got) > hi {
		return fmt.Sprintf("nanotime = %d ns, want within [%d, %d] (raw ticks %d..%d, timebase %d/%d)",
			got, lo, hi, before.ticks, after.ticks, before.numer, before.denom), oldFormulaWouldFail
	}
	return "", oldFormulaWouldFail
}

// simulatedNanotime drives nanotimeBracketProblem from a simulated tick counter
// that starts at uptime startSeconds, converting with conv.
func simulatedNanotime(startSeconds, numer, denom, hz uint64, conv func(ticks int64, numer, denom uint32) int64) (string, bool) {
	cur := startSeconds * hz
	read := func() machReading {
		cur += 3
		return machReading{cur, uint32(numer), uint32(denom)}
	}
	now := func() int64 {
		cur += 3
		return conv(int64(cur), uint32(numer), uint32(denom))
	}
	return nanotimeBracketProblem(read, now)
}

// TestNanotimeBracketProblem exercises the check that the Darwin call-site test
// (TestNanotimeCallSite) applies to the real nanotime, on a simulated G5 clock:
// the fixed conversion passes it, the former inline formula fails it, and a
// reading from before the old formula overflows is reported as unable to tell
// them apart.
func TestNanotimeBracketProblem(t *testing.T) {
	newConv := func(ticks int64, numer, denom uint32) int64 {
		return runtime.MachTimeToNanos(uint64(ticks), numer, denom)
	}
	const numer, denom, hz = 1000000000, 33330863, 33330863
	for _, s := range []uint64{282, 300, 553, 600, 830, 3600, 86400} {
		if problem, informative := simulatedNanotime(s, numer, denom, hz, newConv); problem != "" || !informative {
			t.Errorf("uptime %d s, fixed conversion: problem %q informative %v, want no problem and informative", s, problem, informative)
		}
		if problem, informative := simulatedNanotime(s, numer, denom, hz, wrappedMachTime); problem == "" || !informative {
			t.Errorf("uptime %d s, former conversion: problem %q informative %v, want a problem and informative", s, problem, informative)
		}
	}
	// Before 2^63/numer ticks (277 s) the former formula is exact, so the
	// check passes it and says so.
	if problem, informative := simulatedNanotime(100, numer, denom, hz, wrappedMachTime); problem != "" || informative {
		t.Errorf("uptime 100 s, former conversion: problem %q informative %v, want no problem and not informative", problem, informative)
	}
	// A timebase that changes between the readings is reported.
	calls := uint32(0)
	read := func() machReading { calls++; return machReading{1, 1, calls} }
	if problem, _ := nanotimeBracketProblem(read, func() int64 { return 1 }); problem == "" {
		t.Errorf("changing timebase was not reported")
	}
}
