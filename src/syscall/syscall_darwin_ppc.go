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

// Plain readdir_r fills the legacy 32-bit-inode dirent. Use the 64-bit-inode
// libSystem entry to match the measured 1048-byte Dirent layout.
func readdir_r(dir uintptr, entry *Dirent, result **Dirent) (res Errno) {
	r0, _, _ := syscall(abi.FuncPCABI0(libc_readdir_r_inode64_trampoline), uintptr(dir), uintptr(unsafe.Pointer(entry)), uintptr(unsafe.Pointer(result)))
	return Errno(r0)
}

func libc_readdir_r_inode64_trampoline()

//go:cgo_import_dynamic libc_readdir_r_inode64 readdir_r$INODE64 "/usr/lib/libSystem.B.dylib"
