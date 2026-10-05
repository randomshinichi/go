// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package unix_test

import (
	"internal/syscall/unix"
	"strings"
	"sync"
	"syscall"
	"testing"
	"unsafe"
)

func TestPtsnameStringDarwinPPC64(t *testing.T) {
	for _, n := range []int{0, 1, 253, 254, 255} {
		buf := append([]byte(strings.Repeat("x", n)), 0)
		got, err := unix.PtsnameStringForTest(unsafe.Pointer(&buf[0]))
		if n == 255 {
			if got != "" || err != syscall.ERANGE {
				t.Fatalf("length %d: got %q, %v; want ERANGE", n, got, err)
			}
		} else if err != nil || got != string(buf[:n]) {
			t.Fatalf("length %d: got %q, %v", n, got, err)
		}
	}
	if got, err := unix.PtsnameStringForTest(nil); got != "" || err != syscall.EINVAL {
		t.Fatalf("nil: got %q, %v; want EINVAL", got, err)
	}
}

func TestPtsnameDarwinPPC64(t *testing.T) {
	if got, err := unix.Ptsname(-1); got != "" || err == nil {
		t.Fatalf("invalid fd: got %q, %v", got, err)
	}
	var fds [2]int
	var names [2]string
	for i := range fds {
		fd, err := unix.PosixOpenpt(syscall.O_RDWR | syscall.O_NOCTTY)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(fd)
		fds[i] = fd
		if err := unix.Grantpt(fd); err != nil {
			t.Fatal(err)
		}
		if err := unix.Unlockpt(fd); err != nil {
			t.Fatal(err)
		}
		names[i], err = unix.Ptsname(fd)
		if err != nil || !strings.HasPrefix(names[i], "/dev/") {
			t.Fatalf("slave name: %q, %v", names[i], err)
		}
		slave, err := syscall.Open(names[i], syscall.O_RDWR|syscall.O_NOCTTY, 0)
		if err != nil {
			t.Fatal(err)
		}
		syscall.Close(slave)
	}
	if names[0] == names[1] {
		t.Fatal("two live masters have the same slave name")
	}
	var wg sync.WaitGroup
	for i := range fds {
		wg.Go(func() {
			for j := 0; j < 100; j++ {
				got, err := unix.Ptsname(fds[i])
				if err != nil || got != names[i] {
					t.Errorf("master %d: got %q, %v; want %q", i, got, err, names[i])
				}
			}
		})
	}
	wg.Wait()
}
