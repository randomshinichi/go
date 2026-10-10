// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall

import (
	"internal/abi"
	"unsafe"
)

// Leopard's plain getfsstat fills the legacy 336-byte struct statfs, while
// Statfs_t is the 64-bit-inode struct statfs64. Use the matching libSystem API.
//
//go:cgo_import_dynamic libc_getfsstat64 getfsstat64 "/usr/lib/libSystem.B.dylib"
func libc_getfsstat64_trampoline()

func Getfsstat(buf []Statfs_t, flags int) (n int, err error) {
	var ptr unsafe.Pointer
	var size uintptr
	if len(buf) > 0 {
		ptr = unsafe.Pointer(&buf[0])
		size = unsafe.Sizeof(Statfs_t{}) * uintptr(len(buf))
	}
	r0, _, e1 := syscall(abi.FuncPCABI0(libc_getfsstat64_trampoline), uintptr(ptr), size, uintptr(flags))
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}
