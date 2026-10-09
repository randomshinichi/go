// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// machTimeToNanos converts a mach_absolute_time reading to nanoseconds:
// ticks * numer / denom, where numer and denom come from mach_timebase_info.
//
// The product is not formed in 64 bits. A tick rate far below 1 GHz makes
// numer large: with a 33.3 MHz timebase (Power Mac G5) and numer = 1e9,
// ticks*numer overflows int64 after about 277 seconds of uptime, and the
// result is then negative until the next wrap 553 seconds later. Writing
// ticks = q*denom + r with r < denom, the first term q*numer never exceeds the
// result. The second, r*numer, can (when ticks < denom, r = ticks and the
// result is smaller than r*numer), but it is below 2^64 for any pair of uint32
// factors: r*numer <= (2^32-2)*(2^32-1). The sum is the exact floor of
// ticks*numer/denom and is meaningful while that value is below 2^63
// nanoseconds, about 292 years.
//
// numer == denom returns ticks unchanged. That includes the pair 0/0, which no
// real mach_timebase_info reports; the former conversion divided by zero and
// panicked there, and this one deliberately does not (see
// TestMachTimeToNanosZeroTimebase). numer != 0 with denom == 0 still panics
// with an integer divide error, as before.
//
// ticks must be the unsigned mach_absolute_time value.
//
// The divisions go through machDivmod rather than / and %. nanotime1 is
// nosplit and runs where a stack check is fatal: casgstatus calls nanotime
// inside exitsyscall, while stackguard0 is stackPreempt and throwsplit is set,
// so any function with a stack check that it reaches dies with "stack split
// at bad time". On 32-bit targets a 64-bit / or % compiles to a call to
// runtime.uint64div or runtime.uint64mod, which have stack checks.
//
//go:nosplit
func machTimeToNanos(ticks uint64, numer, denom uint32) int64 {
	if numer == denom {
		return int64(ticks)
	}
	n := uint64(numer)
	q, r := machDivmod(ticks, denom)
	f, _ := machDivmod(r*n, denom)
	return int64(q*n + f)
}
