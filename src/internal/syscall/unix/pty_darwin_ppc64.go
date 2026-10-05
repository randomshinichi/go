// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

// D18: Leopard lacks ptsname_r; ptsname is measured present in
// out/stage2-20261003/parent-native-run/D18-FRAMEWORK-MEASUREMENTS.md.
// Serialize ptsname and copy its static result before releasing the lock.
// This is not libc-wide reentrancy: foreign callers of ptsname are not locked.
// CGO is unsupported on this target. ttyname_r cannot replace ptsname because
// it names the master descriptor, not its slave.
package unix

import (
	"internal/abi"
	"sync"
	"syscall"
	"unsafe"
)

//go:cgo_import_dynamic libc_grantpt grantpt "/usr/lib/libSystem.B.dylib"
func libc_grantpt_trampoline()

func Grantpt(fd int) error {
	_, _, errno := syscall_syscall6(abi.FuncPCABI0(libc_grantpt_trampoline), uintptr(fd), 0, 0, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

//go:cgo_import_dynamic libc_unlockpt unlockpt "/usr/lib/libSystem.B.dylib"
func libc_unlockpt_trampoline()

func Unlockpt(fd int) error {
	_, _, errno := syscall_syscall6(abi.FuncPCABI0(libc_unlockpt_trampoline), uintptr(fd), 0, 0, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

//go:cgo_import_dynamic libc_ptsname ptsname "/usr/lib/libSystem.B.dylib"
func libc_ptsname_trampoline()

var ptsnameMu sync.Mutex

func Ptsname(fd int) (string, error) {
	ptsnameMu.Lock()
	defer ptsnameMu.Unlock()
	p, _, errno := syscall_syscallPtr(abi.FuncPCABI0(libc_ptsname_trampoline), uintptr(fd), 0, 0)
	if errno != 0 {
		return "", errno
	}
	return ptsnameString(unsafe.Pointer(p))
}

func ptsnameString(p unsafe.Pointer) (string, error) {
	if p == nil {
		return "", syscall.EINVAL
	}
	// The modern wrapper supplies 255 bytes including the terminating NUL.
	// Read only through the terminator; do not form a slice beyond a short
	// libc allocation. At most 254 name bytes fit, otherwise return ERANGE.
	var buf [254]byte
	for i := 0; i <= len(buf); i++ {
		c := *(*byte)(unsafe.Add(p, i))
		if c == 0 {
			return string(buf[:i]), nil
		}
		if i < len(buf) {
			buf[i] = c
		}
	}
	return "", syscall.ERANGE
}

//go:cgo_import_dynamic libc_posix_openpt posix_openpt "/usr/lib/libSystem.B.dylib"
func libc_posix_openpt_trampoline()

func PosixOpenpt(flag int) (fd int, err error) {
	ufd, _, errno := syscall_syscall6(abi.FuncPCABI0(libc_posix_openpt_trampoline), uintptr(flag), 0, 0, 0, 0, 0)
	if errno != 0 {
		return -1, errno
	}
	return int(ufd), nil
}
