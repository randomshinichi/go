// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test hooks for goppc-452y: hold the ppc32 64-bit atomic spinlock across a
// SIGPROF that is directed at the holding thread. See TestSIGPROFWhileLock64Held.

//go:build ppc

package runtime

import _ "unsafe" // for go:linkname

// lock64Acquire and lock64Release are internal/runtime/atomic's, pushed by the
// bare //go:linkname directives there.
//
//go:linkname lock64Acquire internal/runtime/atomic.lock64Acquire
func lock64Acquire()

//go:linkname lock64Release internal/runtime/atomic.lock64Release
func lock64Release()

// HoldLock64AcrossSIGPROF takes the 64-bit atomic spinlock, makes the kernel
// deliver SIGPROF to THIS thread, and releases the lock afterwards. pthread_kill
// to the calling thread delivers the signal before the call returns, so the
// handler runs while this thread holds the lock: the timing is explicit, not
// raced. The caller must have called LockOSThread so "this thread" is stable.
func HoldLock64AcrossSIGPROF() {
	lock64Acquire()
	signalM(getg().m, _SIGPROF)
	lock64Release()
}

// ProfLostAtomic is the count of samples sigprof has dropped because the
// 64-bit atomic lock was held, not yet reported to the profile.
func ProfLostAtomic() uint64 { return cpuprof.lostAtomic }
