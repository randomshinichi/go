// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

import "internal/runtime/atomic"

// mutexProfilingFlag mirrors "mutexprofilerate > 0" in 32 bits, because on this
// architecture reading the rate is not cheap: there is no native 64-bit atomic,
// and every 64-bit operation takes a non-reentrant address-hashed stripe lock
// (see internal/runtime/atomic's atomic64Lock). mutexSampleContention, which
// unlock2 calls on every unlock, tests this flag before it reads the rate.
//
// The consequence it removes: every print() ends in printunlock ->
// unlock(&debuglock) -> unlock2 -> mutexSampleContention, so a fatal report or a
// panic report printed on a thread that already held a stripe used to print one
// line and then spin in atomic64Lock forever (if the rate hashed to that stripe) -- the lock's owner is the code
// the report interrupted, and that code cannot run again until the report
// returns (goppc-7tqu).
//
// The flag is written only by SetMutexProfileFraction, and only after the rate
// it mirrors, so a reader that sees it set cannot miss the rate; a reader that
// sees it clear skips at most the samples of the transition.
var mutexProfilingFlag uint32

// mutexProfilingOn reports whether mutex contention profiling is on. It touches
// no 64-bit memory, so it is safe on a signal handler's print path.
//
//go:nosplit
func mutexProfilingOn() bool { return atomic.Load(&mutexProfilingFlag) != 0 }

// setMutexProfiling records that mutex contention profiling is (not) on.
//
//go:nosplit
func setMutexProfiling(on bool) {
	v := uint32(0)
	if on {
		v = 1
	}
	atomic.Store(&mutexProfilingFlag, v)
}
