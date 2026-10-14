// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

package runtime

import (
	"internal/runtime/syscall/linux"
	"unsafe"
)

// The PowerPC time-base register is not present on every supported CPU.
func cputicks() int64 { return nanotime() }

//go:nosplit
func nanotime1() int64 {
	if use64bitsTimeOn32bits {
		var ts timespec
		_, _, errno := linux.Syscall6(_SYS_CLOCK_GETTIME64, 1, uintptr(unsafe.Pointer(&ts)), 0, 0, 0, 0)
		if errno == 0 {
			return ts.tv_sec*1e9 + ts.tv_nsec
		}
	}
	var ts timespec32
	_, _, errno := linux.Syscall6(_SYS_CLOCK_GETTIME32, 1, uintptr(unsafe.Pointer(&ts)), 0, 0, 0, 0)
	if errno != 0 {
		return 0
	}
	return int64(ts.tv_sec)*1e9 + int64(ts.tv_nsec)
}

//go:nosplit
func walltime() (sec int64, nsec int32) {
	if use64bitsTimeOn32bits {
		var ts timespec
		_, _, errno := linux.Syscall6(_SYS_CLOCK_GETTIME64, 0, uintptr(unsafe.Pointer(&ts)), 0, 0, 0, 0)
		if errno == 0 {
			return ts.tv_sec, int32(ts.tv_nsec)
		}
	}
	var ts timespec32
	_, _, errno := linux.Syscall6(_SYS_CLOCK_GETTIME32, 0, uintptr(unsafe.Pointer(&ts)), 0, 0, 0, 0)
	if errno != 0 {
		return 0, 0
	}
	return int64(ts.tv_sec), ts.tv_nsec
}
