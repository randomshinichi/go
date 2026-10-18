// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test hooks for goppc-452y and goppc-nzqrh: hold ppc32 64-bit atomic spinlock
// stripes across a SIGPROF that is directed at the holding thread. See
// TestSIGPROFWhileLock64Held.

//go:build ppc

package runtime

import "unsafe" // also for go:linkname

// The stripe functions are internal/runtime/atomic's, pushed by the bare
// //go:linkname directives there.
//
//go:linkname atomic64StripeAcquire internal/runtime/atomic.atomic64StripeAcquire
func atomic64StripeAcquire(ptr *uint64)

//go:linkname atomic64StripeRelease internal/runtime/atomic.atomic64StripeRelease
func atomic64StripeRelease(ptr *uint64)

//go:linkname atomic64StripeAcquireAll internal/runtime/atomic.atomic64StripeAcquireAll
func atomic64StripeAcquireAll()

//go:linkname atomic64StripeReleaseAll internal/runtime/atomic.atomic64StripeReleaseAll
func atomic64StripeReleaseAll()

//go:linkname atomic64StripeIndex internal/runtime/atomic.atomic64StripeIndex
func atomic64StripeIndex(ptr *uint64) int

// profBufOperands are the 64-bit words the SIGPROF handler's cpuprof.add
// touches in the profile buffer: the read and write indices and the overflow
// counters. profBufOperands is not proof that the handler touches no other
// 64-bit word; that is why sigprof's guard is "any stripe".
func profBufOperands() []*uint64 {
	b := cpuprof.log
	return []*uint64{
		(*uint64)(unsafe.Pointer(&b.r)),
		(*uint64)(unsafe.Pointer(&b.w)),
		(*uint64)(unsafe.Pointer(&b.overflow)),
		(*uint64)(unsafe.Pointer(&b.overflowTime)),
	}
}

// StripeHoldResult describes which stripe a hook held.
type StripeHoldResult struct {
	Stripe      int  // index of the held stripe
	ProfBufUses bool // whether a profile-buffer word hashes to it
}

// HoldAtomic64StripeAcrossSIGPROF takes one 64-bit atomic stripe, makes the
// kernel deliver SIGPROF to THIS thread, and releases the stripe afterwards.
// pthread_kill to the calling thread delivers the signal before the call
// returns, so the handler runs while this thread holds the stripe: the timing is
// explicit, not raced. The caller must have called LockOSThread so "this thread"
// is stable, and have started the CPU profile so that cpuprof.log exists.
//
// With needed true the held stripe is the one of the profile buffer's write
// index, which cpuprof.add must take: a handler that waits for it waits for its
// own interrupted caller. With needed false it is a stripe none of
// profBufOperands hashes to.
func HoldAtomic64StripeAcrossSIGPROF(needed bool) StripeHoldResult {
	var p *uint64
	if needed {
		p = profBufOperands()[1]
	} else {
		p = stripeOtherThan(profBufOperands())
	}
	r := StripeHoldResult{Stripe: atomic64StripeIndex(p)}
	for _, q := range profBufOperands() {
		if atomic64StripeIndex(q) == r.Stripe {
			r.ProfBufUses = true
		}
	}
	atomic64StripeAcquire(p)
	signalM(getg().m, _SIGPROF)
	atomic64StripeRelease(p)
	return r
}

// ProfBufStripes is a plan to hold the stripe of the profile buffer's READ
// index r across a SIGPROF (goppc-l3r5t). cpuprof.add takes r's stripe as well
// as w's, and an any-stripe guard must cover it: a guard that looked only at the
// stripe of w would let the handler wait for the held r stripe forever. R and W
// are the two stripes, reported so that the control can state, before the
// signal, that they differ. r and w are adjacent 8-byte words of one profBuf, so
// their stripes are adjacent modulo 57 and cannot coincide; Distinct records
// that fact rather than assuming it.
type ProfBufStripes struct {
	R, W     int
	Distinct bool
	r        *uint64
}

// PlanRHeld chooses r. The profile must be running so that cpuprof.log exists.
func PlanRHeld() ProfBufStripes {
	ops := profBufOperands()
	r, w := ops[0], ops[1]
	p := ProfBufStripes{R: atomic64StripeIndex(r), W: atomic64StripeIndex(w), r: r}
	p.Distinct = p.R != p.W
	return p
}

// HoldAcrossSIGPROF takes r's stripe, makes the kernel deliver SIGPROF to THIS
// thread (pthread_kill to the calling thread delivers before the call returns, so
// the handler runs while the stripe is held), and releases the stripe. The caller
// must have called LockOSThread. With sigprof's guard intact the handler drops
// the sample and counts it in ProfLostAtomic; with a guard that ignores r's
// stripe it waits for this thread forever.
func (p ProfBufStripes) HoldAcrossSIGPROF() {
	atomic64StripeAcquire(p.r)
	signalM(getg().m, _SIGPROF)
	atomic64StripeRelease(p.r)
}

// stripeOtherThan returns an operand address whose stripe none of avoid hashes
// to. Consecutive 8-byte words hash to consecutive stripes, so a table of
// more than 57 aligned words reaches all of them.
func stripeOtherThan(avoid []*uint64) *uint64 {
	words := new([2 * 57]uint64)
	base := (uintptr(unsafe.Pointer(&words[0])) + 7) &^ 7
	for i := uintptr(0); i < 57; i++ {
		p := (*uint64)(unsafe.Pointer(base + 8*i))
		ok := true
		for _, q := range avoid {
			if atomic64StripeIndex(p) == atomic64StripeIndex(q) {
				ok = false
			}
		}
		if ok {
			return p
		}
	}
	throw("no free stripe")
	return nil
}

// ProfLostAtomic is the count of samples sigprof has dropped because a
// 64-bit atomic stripe was held, not yet reported to the profile.
func ProfLostAtomic() uint64 { return cpuprof.lostAtomic }
