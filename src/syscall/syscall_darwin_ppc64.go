// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import (
	"internal/abi"
	"unsafe"
)

// O_CLOEXEC does not exist on Leopard (it arrived in OS X 10.7; the headers and
// the compiler on the G5 both reject it, docs/darwin-abi-reference.md section 8),
// but shared code names it. Zero is the encoding js/wasm use for a platform
// without the flag: it contributes no bits to open's flags, so nothing is passed
// to the kernel that it does not define, and callers that need close-on-exec
// must set it with fcntl(F_SETFD, FD_CLOEXEC) after the open.
const O_CLOEXEC = 0

// F_DUPFD_CLOEXEC is likewise absent from Leopard's headers. Zero is the value
// internal/poll tests for ("syscall.F_DUPFD_CLOEXEC != 0") before it falls back
// to dup followed by fcntl(F_SETFD, FD_CLOEXEC); aix defines it the same way.
const F_DUPFD_CLOEXEC = 0

func setTimespec(sec, nsec int64) Timespec {
	return Timespec{Sec: sec, Nsec: nsec}
}

func setTimeval(sec, usec int64) Timeval {
	return Timeval{Sec: sec, Usec: int32(usec)}
}

// The stat family uses the *64 libc entry points, which exist in Leopard's
// libSystem (dlprobe, out/stage2-20261003/D7-syscall/g5-request/out-d7/08-dlprobe.txt)
// and fill the 64-bit-inode Stat_t/Statfs_t that ztypes_darwin_ppc64.go
// measured. The plain stat/fstat/lstat symbols fill the legacy 32-bit-inode record.

//sys	Fstat(fd int, stat *Stat_t) (err error) = SYS_fstat64
//sys	Fstatfs(fd int, stat *Statfs_t) (err error) = SYS_fstatfs64
//sysnb	Gettimeofday(tp *Timeval) (err error)
//sys	Lstat(path string, stat *Stat_t) (err error) = SYS_lstat64
//sys	Stat(path string, stat *Stat_t) (err error) = SYS_stat64
//sys	Statfs(path string, stat *Statfs_t) (err error) = SYS_statfs64
//sys   ptrace(request int, pid int, addr uintptr, data uintptr) (err error)

// fstatat is implemented without libc: Leopard has no fstatat (nor any other
// *at function; fstatat64 is among the names dlprobe reports MISSING). With
// the current directory as the base it is stat or lstat. With any other
// directory descriptor there is no equivalent to build on, so it fails with
// ENOSYS rather than pretending.
func fstatat(fd int, path string, stat *Stat_t, flags int) (err error) {
	const atSymlinkNofollow = 0x20 // the value of the modern *at interface; Leopard defines none
	if fd != _AT_FDCWD {
		return ENOSYS
	}
	if flags&atSymlinkNofollow != 0 {
		return Lstat(path, stat)
	}
	return Stat(path, stat)
}

func SetKevent(k *Kevent_t, fd, mode, flags int) {
	k.Ident = uint64(fd)
	k.Filter = int16(mode)
	k.Flags = uint16(flags)
}

func (iov *Iovec) SetLen(length int) {
	iov.Len = uint64(length)
}

func (msghdr *Msghdr) SetControllen(length int) {
	msghdr.Controllen = uint32(length)
}

func (cmsg *Cmsghdr) SetLen(length int) {
	cmsg.Len = uint32(length)
}

func sendfile(outfd int, infd int, offset *int64, count int) (written int, err error) {
	var length = uint64(count)

	_, _, e1 := syscall6(abi.FuncPCABI0(libc_sendfile_trampoline), uintptr(infd), uintptr(outfd), uintptr(*offset), uintptr(unsafe.Pointer(&length)), 0, 0)

	written = int(length)

	if e1 != 0 {
		err = e1
	}
	return
}

func libc_sendfile_trampoline()

//go:cgo_import_dynamic libc_sendfile sendfile "/usr/lib/libSystem.B.dylib"

// Getdirentries is a SUBSTITUTED mechanism on darwin/ppc64. The shared
// implementation (syscall_darwin_libcdir.go) simulates it with libc fdopendir,
// openat, readdir_r and closedir; fdopendir and openat do not exist on Leopard
// (dlprobe, out/stage2-20261003/D7-syscall/g5-request/out-d7/08-dlprobe.txt), and
// plain readdir_r returns the legacy 32-bit-inode record rather than the Dirent
// of ztypes_darwin_ppc64.go.
//
// Instead this issues SYS_getdirentries64 (344) directly. Measured on the G5
// (out/stage2-20261003/D7-syscall/g5-request2/out-d7r2/C-dirent-test.txt) it
// returns exactly the 64-bit-inode record that Dirent describes: ino@0,
// seekoff@8, reclen@16, namlen@18, type@20, name@21. SYS_getdirentries (196)
// returns a different, legacy record and must not be used here.
//
// seekoff in those records is a filesystem cookie (values such as
// 2208261730205695 were observed), not a byte offset; nothing here interprets it.
// *basep receives the kernel's position cookie and is likewise opaque.
func Getdirentries(fd int, buf []byte, basep *uintptr) (n int, err error) {
	var _p0 unsafe.Pointer
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
	} else {
		_p0 = unsafe.Pointer(&_zero)
	}
	var base uintptr
	if basep == nil {
		basep = &base
	}
	r0, _, e1 := Syscall6(SYS_GETDIRENTRIES64, uintptr(fd), uintptr(_p0), uintptr(len(buf)), uintptr(unsafe.Pointer(basep)), 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}
