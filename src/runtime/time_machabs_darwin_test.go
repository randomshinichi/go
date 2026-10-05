// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
)

// TestNanotimeCallSite guards the place the clock defect lived: the conversion
// inside nanotime1, which TestMachTimeToNanos cannot see because it calls the
// helper directly. It brackets runtime.Nanotime between two raw
// mach_absolute_time readings taken through the same trampoline and requires
// the result to lie between the exact (128-bit) conversions of those readings.
// The former inline formula formed ticks*numer in int64 and fails this on the
// Power Mac G5 at every uptime past ~277 s: negative in [277 s, 553 s), 553 s
// too low in [553 s, 830 s) and so on.
//
// Below that uptime, and wherever ticks*numer cannot reach 2^63 (the 1/1
// timebase of Intel Macs), the former formula passes the bracket too; the test
// then reports SKIP rather than PASS, because a pass would not show the call
// site is sound.
func TestNanotimeCallSite(t *testing.T) {
	read := func() machReading {
		ticks, numer, denom := runtime.MachAbsoluteTime()
		return machReading{ticks, numer, denom}
	}
	var first machReading
	informative := 0
	const reps = 2000
	for i := 0; i < reps; i++ {
		if i == 0 {
			first = read()
		}
		problem, oldFormulaWouldFail := nanotimeBracketProblem(read, runtime.Nanotime)
		if problem != "" {
			t.Fatalf("rep %d: %s", i, problem)
		}
		if oldFormulaWouldFail {
			informative++
		}
	}
	t.Logf("timebase %d/%d, first raw ticks %d (uptime ~%.1f s); %d of %d brackets would have failed the former inline formula",
		first.numer, first.denom, first.ticks, float64(first.ticks)*float64(first.numer)/float64(first.denom)/1e9, informative, reps)
	if informative == 0 {
		t.Skipf("nanotime is within its bracket, but the former inline formula would pass here too " +
			"(ticks*numer < 2^63): this run does not distinguish the call site. Rerun after uptime exceeds 2^63/numer ticks (~277 s on the G5)")
	}
}
