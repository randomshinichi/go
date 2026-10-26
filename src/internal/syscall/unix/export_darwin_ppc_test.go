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

// AtConcatPathForTest is the length-checked concatenation the cross-directory
// dispatch uses to build a destination name from a descriptor's path and a
// relative name. It is exported so the length rule can be pinned without needing
// a directory whose path is 1023 bytes long.
func AtConcatPathForTest(base, suffix string) (string, error) { return atConcatPath(base, suffix) }

// AtPathFromFDForTest resolves a descriptor's directory path with fcntl
// F_GETPATH and appends a relative name, with the same length rule.
func AtPathFromFDForTest(dirfd int, suffix string) (string, error) {
	return atPathFromFD(dirfd, suffix)
}

// AtFaccessatForTest is the unix-layer faccessat dispatch: no caller in the
// standard library passes a real descriptor or flags other than AT_EACCESS, so
// the refusal paths can only be asserted from outside the package.
func AtFaccessatForTest(dirfd int, path string, mode uint32, flags int) error {
	return faccessat(dirfd, path, mode, flags)
}
