// Copyright 2009,2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && !ppc64

// Getfsstat moved here, unchanged, from syscall_darwin.go so that darwin/ppc64
// can provide its own Getfsstat (syscall_darwin_ppc64.go): on Leopard the plain
// getfsstat symbol fills the legacy 336-byte struct statfs, not the 64-bit-inode
// Statfs_t. The libc_getfsstat_trampoline declaration and its cgo_import_dynamic
// stay in syscall_darwin.go, where mkasm.go reads them.

package syscall

import (
	"internal/abi"
	"unsafe"
)

func Getfsstat(buf []Statfs_t, flags int) (n int, err error) {
	var _p0 unsafe.Pointer
	var bufsize uintptr
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
		bufsize = unsafe.Sizeof(Statfs_t{}) * uintptr(len(buf))
	}
	r0, _, e1 := syscall(abi.FuncPCABI0(libc_getfsstat_trampoline), uintptr(_p0), bufsize, uintptr(flags))
	n = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}
