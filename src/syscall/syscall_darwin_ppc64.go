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
// ENOTSUP rather than pretending, as internal/syscall/unix's *at family does
// (at_darwin_ppc64.go).
func fstatat(fd int, path string, stat *Stat_t, flags int) (err error) {
	const atSymlinkNofollow = 0x20 // the value of the modern *at interface; Leopard defines none
	if fd != _AT_FDCWD {
		return ENOTSUP
	}
	if flags&atSymlinkNofollow != 0 {
		return Lstat(path, stat)
	}
	return Stat(path, stat)
}

// utimensat is implemented without libc's utimensat: Leopard does not have it
// (dlprobe reports utimensat and futimens MISSING; the *at family is absent
// altogether), so shared callers (UtimesNano, internal/syscall/unix.Utimensat)
// would otherwise import a symbol dyld cannot resolve. What Leopard does have
// is utimes (path, follows symlinks), futimes and lutimes, all in libSystem's
// exports; none takes a nanosecond or a "leave this one alone" value.
//
// The mapping, and what it does not cover:
//   - dirfd must be _AT_FDCWD (a chosen sentinel, absent on Leopard). Any real
//     directory descriptor returns ENOTSUP, never a fall-back to the
//     path-based call: that would act on the wrong file.
//   - flags 0 is utimes. Undefined bits are EINVAL. AT_SYMLINK_NOFOLLOW would
//     be lutimes, but lutimes's behaviour on Leopard has not been measured, so
//     it returns ENOTSUP rather than assume it does not follow the link.
//   - times == nil, or both entries UTIME_NOW, is utimes(path, NULL).
//   - A UTIME_OMIT entry keeps that time by reading it with stat and writing it
//     back with the other one (the technique fs_wasip1.go's UtimesNano uses).
//     That is a read-modify-write: another process changing the omitted time
//     between the stat and the utimes has its change overwritten. If both are
//     UTIME_OMIT nothing is written and only the stat's error is reported.
//   - UTIME_NOW mixed with any other value is ENOTSUP. utimes needs ownership
//     for an explicit time but only write access for "now"; mixing them cannot
//     be expressed faithfully. Nothing in the standard library does this.
//   - Precision: utimes takes microseconds. Explicit times are rounded up to a
//     microsecond by NsecToTimeval (as UtimesNano's ENOSYS fallback already
//     does on the other BSDs), and a time kept via UTIME_OMIT is passed through
//     the same conversion.
func utimensat(dirfd int, path string, times *[2]Timespec, flags int) (err error) {
	const (
		atSymlinkNofollow = 0x20 // value of the modern *at interface; Leopard defines none
		utimeNow          = -1   // modern Darwin values; Leopard defines neither
		utimeOmit         = -2
	)
	if dirfd != _AT_FDCWD {
		return ENOTSUP
	}
	if flags&^atSymlinkNofollow != 0 {
		return EINVAL
	}
	if flags != 0 {
		return ENOTSUP
	}
	if times == nil {
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
		return utimes(path, nil)
	}
	if now0 || now1 {
		return ENOTSUP
	}
	omit0, omit1 := times[0].Nsec == utimeOmit, times[1].Nsec == utimeOmit
	var tv [2]Timeval
	if omit0 || omit1 {
		var st Stat_t
		if err := Stat(path, &st); err != nil {
			return err
		}
		if omit0 && omit1 {
			return nil
		}
		times = &[2]Timespec{times[0], times[1]}
		if omit0 {
			times[0] = st.Atimespec
		}
		if omit1 {
			times[1] = st.Mtimespec
		}
	}
	tv[0] = NsecToTimeval(TimespecToNsec(times[0]))
	tv[1] = NsecToTimeval(TimespecToNsec(times[1]))
	return utimes(path, &tv)
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
