// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os_test

// Copyright (c) 2026 Frederick H. G. Wright II <fw@fwright.net>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

import (
	"os"
	"syscall"
	"testing"
)

// fdBitsDarwinPPC reports the descriptor flags currently set on fd, and whether
// FD_CLOEXEC is one of them.
func fdBitsDarwinPPC(t *testing.T, fd int) (flags uintptr, cloexec bool) {
	t.Helper()
	const fGetFD = 1
	const fdCloexec = 1
	r0, _, errno := syscall.RawSyscall(syscall.SYS_FCNTL, uintptr(fd), fGetFD, 0)
	if errno != 0 {
		t.Fatalf("fcntl(%d, F_GETFD): %v", fd, errno)
	}
	return r0, r0&fdCloexec != 0
}

// TestLegacySupportRootCloseOnExecDarwinPPC adopts the close-on-exec row of the
// MacPorts legacy-support suite (M/src/open.c:44-71, parity P03): a Root's
// directory descriptor and a File opened through a Root both carry FD_CLOEXEC on
// this platform.
//
// The guarantee is the one this port can actually make. O_CLOEXEC is 0 on
// darwin/ppc and os falls back to syscall.CloseOnExec after the open
// (root_unix.go, supportsCloseOnExec in sys_darwin_ppc.go), so a descriptor is
// briefly non-close-on-exec before the flag is applied: a concurrent fork+exec in
// another goroutine can inherit it. That fork race is the documented policy, not
// an atomic guarantee, and this test asserts only what exists.
func TestLegacySupportRootCloseOnExecDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/file", []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	rootFd := os.RootFd(root)
	if flags, cloexec := fdBitsDarwinPPC(t, rootFd); !cloexec {
		t.Errorf("Root descriptor %d flags = %#x; want FD_CLOEXEC set", rootFd, flags)
	}

	f, err := root.Open("file")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fd := int(f.Fd())
	if flags, cloexec := fdBitsDarwinPPC(t, fd); !cloexec {
		t.Errorf("Root.OpenFile descriptor %d flags = %#x; want FD_CLOEXEC set", fd, flags)
	}
}
