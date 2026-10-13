// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test hooks that make systemstack's fatal "bad g" branch reachable. They live
// in package runtime so they can call getg, setg and systemstack directly; see
// TestSystemstackBadBranchTaken for how they are driven.

//go:build ppc

package runtime

import "unsafe"

// gBadSystemstack is the g the witness hands to setg. Only its ADDRESS matters:
// systemstack branches to systemstack_bad when the current g is neither
// m.gsignal, m.g0 nor m.curg, so any g that is none of those three reaches the
// branch. Its content is replaced by a copy of the real g before it is used.
var gBadSystemstack g

// BadSystemstackBranchTaken makes systemstack's fatal "bad g" branch execute and
// never returns. badsystemstack writes
//
//	fatal: systemstack called from unexpected goroutine
//
// through writeErrStr, and systemstack_bad then calls abort, which faults on
// purpose.
//
// The g is a byte-for-byte copy of the current one, so every field that the
// fatal path touches afterwards - m, the stack bounds, the stack guards - is the
// real value. Only the identity differs, which is exactly the invariant
// systemstack is checking.
//
//go:noinline
func BadSystemstackBranchTaken() {
	real := getg()
	// Make the report's nested fault deterministic. The fatal report walks m.curg
	// from curg.sched (tracebacksomeothers), and a running g's saved sp is only
	// meaningful if it was parked: sched.sp == 0 sends the unwinder through
	// *(0+8), a hardware fault raised INSIDE the signal handler. That nested fault
	// is exactly what unblockFaultSignals exists to survive (goppc-0v7x), and
	// without it the witness would not exercise the fix. Left to itself the value
	// is whatever the last syscall or park on this g saved - the "child is ready"
	// line the test harness writes first is enough to change it - so zero it.
	real.sched = gobuf{}
	fake := &gBadSystemstack
	// memmove, not *fake = *real: g embeds a noCopy atomic, which vet rejects.
	memmove(unsafe.Pointer(fake), unsafe.Pointer(real), unsafe.Sizeof(*real))
	// A pending preemption is a sentinel placed in stackguard0. morestack on a g
	// that is not in allgs is not survivable, so do not inherit it.
	if fake.stackguard0 == stackPreempt {
		fake.stackguard0 = fake.stack.lo + stackGuard
	}
	setg(fake)
	systemstack(badSystemstackNoop)
	// Only reachable if the fatal branch was not taken; the test treats that as
	// a failure. Put g back first so the caller can report it.
	setg(real)
}

// SystemstackOnRealG is the not-taken control for BadSystemstackBranchTaken: the
// same systemstack call with the unmodified g. The current g is m.curg here, so
// systemstack takes its ordinary switch path and returns.
//
//go:noinline
func SystemstackOnRealG() {
	systemstack(badSystemstackNoop)
}

//go:noinline
func badSystemstackNoop() {}

// AbortForTesting is the control for the fatal epilogue. It calls from the
// ordinary g the same runtime.abort() that systemstack_bad calls, and reaches it
// without going anywhere near systemstack, badsystemstack or a replaced g. What
// the process does after the fault is therefore a property of abort() alone.
// It never returns.
//
//go:noinline
func AbortForTesting() {
	abort()
}

// ReadOnSystemstack reads one word at addr while running on the system stack, so a
// fault there is a hardware fault on g0. The runtime does not recover those: the
// signal handler reports the signal and exits. It is the ordinary fatal-fault
// control for the fault-signal unblocking in sighandler (goppc-0v7x): a
// legitimate fatal signal that is NOT a nested one.
//
//go:noinline
func ReadOnSystemstack(addr uintptr) uintptr {
	var v uintptr
	systemstack(func() { v = *(*uintptr)(unsafe.Pointer(addr)) })
	return v
}
