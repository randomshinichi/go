// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall

import (
	"internal/abi"
	"unsafe"
)

// O_CLOEXEC and F_DUPFD_CLOEXEC are absent from Leopard's headers. These zero
// values preserve the shared fallback checks; close-on-exec uses fcntl.
const (
	O_CLOEXEC       = 0
	F_DUPFD_CLOEXEC = 0
)

func setTimespec(sec, nsec int64) Timespec {
	return Timespec{Sec: int32(sec), Nsec: int32(nsec)}
}

func setTimeval(sec, usec int64) Timeval {
	return Timeval{Sec: int32(sec), Usec: int32(usec)}
}

func SetKevent(k *Kevent_t, fd, mode, flags int) {
	k.Ident = uint32(fd)
	k.Filter = int16(mode)
	k.Flags = uint16(flags)
}

func (iov *Iovec) SetLen(length int) { iov.Len = uint32(length) }

func (msghdr *Msghdr) SetControllen(length int) { msghdr.Controllen = uint32(length) }

func (cmsg *Cmsghdr) SetLen(length int) { cmsg.Len = uint32(length) }

// ptrace, mmap, munmap, and lutimes are libc calls; mkasm.go emits their PPC
// trampolines from zsyscall_darwin_ppc.go.
//sys Fstat(fd int, stat *Stat_t) (err error) = SYS_FSTAT64
//sys Fstatfs(fd int, stat *Statfs_t) (err error) = SYS_FSTATFS64
//sysnb Gettimeofday(tp *Timeval) (err error)
//sys Lstat(path string, stat *Stat_t) (err error) = SYS_LSTAT64
//sys Stat(path string, stat *Stat_t) (err error) = SYS_STAT64
//sys Statfs(path string, stat *Statfs_t) (err error) = SYS_STATFS64
//sys ptrace(request int, pid int, addr uintptr, data uintptr) (err error)
//sys lutimes(path string, timeval *[2]Timeval) (err error)
//sys lchmod(path string, mode uint32) (err error)

// used by internal/syscall/unix
//
//go:linkname lchmod

// Leopard does not implement the POSIX *at calls. Preserve the path-based
// fallback contract and return ENOTSUP for real directory descriptors.
func fstatat(fd int, path string, stat *Stat_t, flags int) error {
	const atSymlinkNofollow = 0x20
	if fd != _AT_FDCWD {
		return ENOTSUP
	}
	if flags&^atSymlinkNofollow != 0 {
		return EINVAL
	}
	if flags&atSymlinkNofollow != 0 {
		return Lstat(path, stat)
	}
	return Stat(path, stat)
}

func utimensat(dirfd int, path string, times *[2]Timespec, flags int) error {
	const (
		atSymlinkNofollow = 0x20
		utimeNow          = -1
		utimeOmit         = -2
	)
	if dirfd != _AT_FDCWD {
		return ENOTSUP
	}
	if flags&^atSymlinkNofollow != 0 {
		return EINVAL
	}
	nofollow := flags != 0
	setTimes, statPath := utimes, Stat
	if nofollow {
		setTimes, statPath = lutimes, Lstat
	}
	if times == nil {
		if nofollow {
			return ENOTSUP
		}
		return utimes(path, nil)
	}
	for _, ts := range times {
		if ts.Nsec == utimeNow || ts.Nsec == utimeOmit {
			continue
		}
		if ts.Nsec < 0 || ts.Nsec >= 1e9 {
			return EINVAL
		}
	}
	now0, now1 := times[0].Nsec == utimeNow, times[1].Nsec == utimeNow
	if now0 && now1 {
		if nofollow {
			return ENOTSUP
		}
		return utimes(path, nil)
	}
	if now0 || now1 {
		return ENOTSUP
	}
	omit0, omit1 := times[0].Nsec == utimeOmit, times[1].Nsec == utimeOmit
	if omit0 || omit1 {
		copy := *times
		times = &copy
		var st Stat_t
		if err := statPath(path, &st); err != nil {
			return err
		}
		if omit0 && omit1 {
			return nil
		}
		if omit0 {
			times[0] = st.Atimespec
		}
		if omit1 {
			times[1] = st.Mtimespec
		}
	}
	tv := [2]Timeval{
		NsecToTimeval(TimespecToNsec(times[0])),
		NsecToTimeval(TimespecToNsec(times[1])),
	}
	return setTimes(path, &tv)
}

// sendfile takes an off_t (64 bits on Leopard) followed by pointer arguments.
// PPC's measured big-endian ABI passes the off_t in R5:R6 without a skipped
// register; syscall9 forwards those words and the remaining pointer arguments.
func sendfile(outfd, infd int, offset *int64, count int) (written int, err error) {
	length := int64(count)
	off := uint64(*offset)
	_, _, e1 := syscall9(abi.FuncPCABI0(libc_sendfile_trampoline),
		uintptr(infd), uintptr(outfd), uintptr(uint32(off>>32)), uintptr(uint32(off)),
		uintptr(unsafe.Pointer(&length)), 0, 0, 0, 0)
	written = int(length)
	if e1 != 0 {
		err = e1
	}
	return
}

func libc_sendfile_trampoline()

//go:cgo_import_dynamic libc_sendfile sendfile "/usr/lib/libSystem.B.dylib"

// Exec cannot work on darwin/ppc Leopard, and this is a platform limit, not a
// port defect. Leopard's kernel (xnu-1228) refuses execve (and posix_spawn with
// POSIX_SPAWN_SETEXEC) from any process that has more than one thread, with
// errno 45 (ENOTSUP), before it looks at the path: execve of a path that does
// not exist also returns 45. A Go process always has several threads, so
// syscall.Exec always returns ENOTSUP here, from libc execve and from the raw
// SYS_EXECVE alike. Measured on the G4 (runs/diag-dsvm-pbg4-20261026T204019Z:
// a plain Go process shows 3 threads in ps -M; Exec of self, /bin/sh, /bin/echo
// and a nonexistent path return 45; a C execve from one thread succeeds, with a
// second thread returns 45, and posix_spawn with POSIX_SPAWN_SETEXEC from two
// threads returns 45). fork followed by execve in the child works, which is
// what ForkExec, StartProcess and os/exec use. TestExec fails here for this
// reason and is not weakened.

// Plain readdir_r fills the legacy 32-bit-inode dirent. Use the 64-bit-inode
// libSystem entry to match the measured 1048-byte Dirent layout.
func readdir_r(dir uintptr, entry *Dirent, result **Dirent) (res Errno) {
	r0, _, _ := syscall(abi.FuncPCABI0(libc_readdir_r_inode64_trampoline), uintptr(dir), uintptr(unsafe.Pointer(entry)), uintptr(unsafe.Pointer(result)))
	return Errno(r0)
}

func libc_readdir_r_inode64_trampoline()

//go:cgo_import_dynamic libc_readdir_r_inode64 readdir_r$INODE64 "/usr/lib/libSystem.B.dylib"

// Getdirentries uses the Leopard syscall that returns 64-bit-inode directory
// records, matching Dirent on ppc. SYS_getdirentries (196) returns a legacy
// 32-bit-inode record and is not layout-compatible. The 64-bit record layout
// was measured on the G5 for ppc64 (ino@0, seekoff@8, reclen@16, namlen@18,
// type@20, name@21); the inode record is independent of userspace pointer size.
//
// The syscall's position argument is an off_t *: the kernel writes 8 bytes
// through it. darwin/ppc32's *uintptr is only 4 bytes and cannot express that,
// so a real int64 receives the kernel's value and only its low 32 bits are
// stored back into *basep. The value is therefore truncated on write-back, and
// a directory offset above 4 GiB is unrepresentable through this API.
func Getdirentries(fd int, buf []byte, basep *uintptr) (n int, err error) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	} else {
		p = unsafe.Pointer(&_zero)
	}
	var base int64
	r0, _, e1 := Syscall6(SYS_GETDIRENTRIES64, uintptr(fd), uintptr(p), uintptr(len(buf)), uintptr(unsafe.Pointer(&base)), 0, 0)
	n = int(r0)
	if e1 != 0 {
		err = errnoErr(e1)
	}
	if basep != nil {
		*basep = uintptr(uint32(base))
	}
	return
}
