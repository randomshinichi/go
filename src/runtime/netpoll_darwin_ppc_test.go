// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"syscall"
	"testing"
)

func TestKeventErrnoDarwinPPC(t *testing.T) {
	// libc returns -1 for this invalid descriptor. Runtime callers need
	// -errno, not that sentinel (notably for EINTR in netpoll).
	if got, want := runtime.KeventInvalidDescriptor(), -int32(syscall.EBADF); got != want {
		t.Fatalf("kevent(-1) = %d, want %d (-EBADF)", got, want)
	}
}
