// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

// Test hooks for goppc-7tqu: run a fatal report on the thread that holds the
// ppc32 64-bit atomic spinlock. lock64Acquire and lock64Release, linknamed to
// internal/runtime/atomic, are declared in export_lock64_test.go in this package.

// HoldLock64AcrossSIGQUIT takes the 64-bit atomic spinlock and makes the kernel
// deliver SIGQUIT to THIS thread, so the fatal report runs on the lock's holder.
// pthread_kill to the calling thread delivers the signal before the call
// returns, which makes the timing explicit rather than raced. The caller must
// have called LockOSThread so "this thread" is stable.
func HoldLock64AcrossSIGQUIT() {
	lock64Acquire()
	signalM(getg().m, _SIGQUIT)
	lock64Release()
}

// SIGQUITCurrentThread delivers SIGQUIT to the calling thread with no lock held:
// the control for HoldLock64AcrossSIGQUIT.
func SIGQUITCurrentThread() {
	signalM(getg().m, _SIGQUIT)
}

// MutexProfilingFlag is the value of ppc32's 32-bit mirror of
// "mutexprofilerate > 0" (mutexprofiling_ppc.go), which mutexSampleContention
// tests before it reads the rate. Read non-atomically: the test that uses it is
// single-threaded at the point of the read.
func MutexProfilingFlag() uint32 { return mutexProfilingFlag }

// HoldLock64AcrossPrint takes the 64-bit atomic spinlock and prints a line while
// holding it. Every print() ends in printunlock -> unlock(&debuglock) ->
// unlock2 -> mutexSampleContention, so this drives the rate-zero path of unlock2
// directly: if that path reads the 64-bit rate it takes lock64, which this
// thread is holding, and the call never returns. That is the dynamic form of the
// "the rate-zero path never calls Load64" check.
func HoldLock64AcrossPrint() {
	lock64Acquire()
	println("lock64: printed while holding lock64")
	lock64Release()
}
