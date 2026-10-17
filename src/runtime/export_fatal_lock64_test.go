// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

// Test hooks for goppc-7tqu: run a fatal report on the thread that holds the
// ppc32 64-bit atomic spinlock stripes. The linknamed internal/runtime/atomic
// stripe functions are declared in export_lock64_test.go in this package. These
// hooks hold ALL stripes, so they prove that no 64-bit atomic runs on the path
// (the guarantee the one global lock used to give), not only none that hashes to
// one stripe.

// HoldAllAtomic64StripesAcrossSIGQUIT takes every 64-bit atomic stripe and makes the kernel
// deliver SIGQUIT to THIS thread, so the fatal report runs on the stripes' holder.
// pthread_kill to the calling thread delivers the signal before the call
// returns, which makes the timing explicit rather than raced. The caller must
// have called LockOSThread so "this thread" is stable.
func HoldAllAtomic64StripesAcrossSIGQUIT() {
	atomic64StripeAcquireAll()
	signalM(getg().m, _SIGQUIT)
	atomic64StripeReleaseAll()
}

// SIGQUITCurrentThread delivers SIGQUIT to the calling thread with no lock held:
// the control for HoldAllAtomic64StripesAcrossSIGQUIT.
func SIGQUITCurrentThread() {
	signalM(getg().m, _SIGQUIT)
}

// MutexProfilingFlag is the value of ppc32's 32-bit mirror of
// "mutexprofilerate > 0" (mutexprofiling_ppc.go), which mutexSampleContention
// tests before it reads the rate. Read non-atomically: the test that uses it is
// single-threaded at the point of the read.
func MutexProfilingFlag() uint32 { return mutexProfilingFlag }

// HoldAllAtomic64StripesAcrossPrint takes every 64-bit atomic stripe and prints a line while
// holding it. Every print() ends in printunlock -> unlock(&debuglock) ->
// unlock2 -> mutexSampleContention, so this drives the rate-zero path of unlock2
// directly: if that path reads the 64-bit rate it takes a stripe, which this
// thread is holding, and the call never returns. That is the dynamic form of the
// "the rate-zero path never calls Load64" check.
func HoldAllAtomic64StripesAcrossPrint() {
	atomic64StripeAcquireAll()
	println("lock64: printed while holding every stripe")
	atomic64StripeReleaseAll()
}
