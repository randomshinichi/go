// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The GOARCH value ppc is not one of the suffixes the toolchain understands
// implicitly, so the build constraint has to be written out.
//go:build darwin && ppc

package syscall_test

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

// getdirentriesCanary is the caller-owned memory handed to Getdirentries: the
// 4-byte *uintptr position word, immediately followed by a guard word. The
// trailing pad keeps the whole object larger than the 8 bytes the kernel
// writes, so an overrun shows up as a guard mismatch rather than as corruption
// of an adjacent heap object.
type getdirentriesCanary struct {
	base  uintptr
	guard uint32
	pad   [4]uintptr
}

// getdirentriesGuard is a value no real directory offset would take, typed
// uint32 because that is the width of the field it sits in.
const getdirentriesGuard uint32 = 0xfeedface

// TestGetdirentriesBasepWidth checks that Getdirentries honours the width of
// the *uintptr it was given. darwin/ppc32 implements it with
// SYS_GETDIRENTRIES64, whose position argument is an off_t *: the kernel writes
// 8 bytes through the caller's pointer. A basep that points at one uintptr
// therefore loses 4 bytes of its neighbours to the syscall unless the wrapper
// buffers the write itself.
func TestGetdirentriesBasepWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 4 || unsafe.Offsetof(getdirentriesCanary{}.guard) != 4 {
		t.Fatalf("canary layout is wrong for this target: uintptr=%d bytes, guard offset=%d",
			unsafe.Sizeof(uintptr(0)), unsafe.Offsetof(getdirentriesCanary{}.guard))
	}

	d := t.TempDir()
	for i := 0; i < 100; i++ {
		if err := os.WriteFile(filepath.Join(d, fmt.Sprintf("file%03d", i)), []byte("data"), 0); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	fd, err := syscall.Open(d, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer syscall.Close(fd)

	canary := new(getdirentriesCanary)
	canary.guard = getdirentriesGuard

	// A Dirent on darwin/ppc is 1048 bytes, so this buffer yields one entry per
	// call and roughly one call per file: the wrapper is exercised repeatedly,
	// not once.
	var buf [2048]byte
	calls := 0
	for {
		n, err := syscall.Getdirentries(fd, buf[:], &canary.base)
		if err != nil {
			t.Fatalf("Getdirentries: %v", err)
		}
		calls++
		if canary.guard != getdirentriesGuard {
			t.Fatalf("Getdirentries wrote past the 4-byte basep: after call %d guard = %#x, want %#x (base = %#x)",
				calls, canary.guard, getdirentriesGuard, canary.base)
		}
		if n == 0 {
			break
		}
	}
	if calls < 2 {
		t.Fatalf("expected Getdirentries to be called more than once, got %d", calls)
	}
}
