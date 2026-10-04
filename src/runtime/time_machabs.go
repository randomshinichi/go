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
// result is then negative until the next wrap 553 seconds later. Dividing
// first keeps every intermediate no larger than the result:
// (ticks%denom)*numer is below 2^64 for any pair of uint32 factors.
//
// ticks must be the unsigned mach_absolute_time value.
//
//go:nosplit
func machTimeToNanos(ticks uint64, numer, denom uint32) int64 {
	if numer == denom {
		return int64(ticks)
	}
	n, d := uint64(numer), uint64(denom)
	return int64(ticks/d*n + ticks%d*n/d)
}
