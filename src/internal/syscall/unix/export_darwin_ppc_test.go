// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix

import "syscall"

// Exports for the external tests: the platform gate decides whether the
// anchored backend may run at all, so its decision and its parser are asserted
// directly rather than being inferred from the behaviour of ten call sites.

func AtCompatRelease(release string) (ok, parsed bool) { return atCompatRelease(release) }

func AtCompatEnabledForTest() bool { return atCompatEnabled() }

// AtPthreadFchdirClearForTest issues raw syscall 349(-1) on the calling
// goroutine's own thread: it clears a thread-private working directory if one is
// installed there and reports EBADF when there is none. It is how a test asks
// whether the private-cwd machinery was involved at all: a request that never
// installs one leaves this thread with nothing to clear.
func AtPthreadFchdirClearForTest() error {
	_, _, errno := syscall.RawSyscall(syscall.SYS___PTHREAD_FCHDIR, ^uintptr(0), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
