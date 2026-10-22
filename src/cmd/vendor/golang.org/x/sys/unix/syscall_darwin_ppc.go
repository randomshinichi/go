// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

// Port-local (goppc-jic6x.3), not part of upstream golang.org/x/sys.
// darwin/ppc is 32-bit PowerPC Mac OS X 10.5 Leopard. This file is input to
// mksyscall.go and mkasm.go; see src/syscall/testdata/darwin_ppc32/xsys/README.

package unix

import "unsafe"

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

func (iov *Iovec) SetLen(length int) {
	iov.Len = uint32(length)
}

func (msghdr *Msghdr) SetControllen(length int) {
	msghdr.Controllen = uint32(length)
}

func (msghdr *Msghdr) SetIovlen(length int) {
	msghdr.Iovlen = int32(length)
}

func (cmsg *Cmsghdr) SetLen(length int) {
	cmsg.Len = uint32(length)
}

//sys	Fstat(fd int, stat *Stat_t) (err error) = SYS_FSTAT64
//sys	Fstatfs(fd int, stat *Statfs_t) (err error) = SYS_FSTATFS64
//sys	getfsstat(buf unsafe.Pointer, size uintptr, flags int) (n int, err error) = SYS_GETFSSTAT64
//sys	Lstat(path string, stat *Stat_t) (err error) = SYS_LSTAT64
//sys	ptrace1(request int, pid int, addr uintptr, data uintptr) (err error) = SYS_ptrace
//sys	Stat(path string, stat *Stat_t) (err error) = SYS_STAT64
//sys	Statfs(path string, stat *Statfs_t) (err error) = SYS_STATFS64
//sys	lchmod(path string, mode uint32) (err error)

// Getdirentries reads directory entries with Leopard's getdirentries64
// system call: SYS_GETDIRENTRIES64 = 344 in MacOSX10.5.sdk
// usr/include/sys/syscall.h, prototype in xnu-1228 bsd/kern/syscalls.master:
//
//	user_ssize_t getdirentries64(int fd, void *buf, user_size_t bufsize, off_t *position)
//
// Its records are the 64-bit-inode struct dirent (__DARWIN_STRUCT_DIRENTRY in
// sys/dirent.h), matching Dirent. Leopard's libSystem has no fdopendir or
// openat, so the libc simulation in syscall_darwin_libcdir.go cannot be used.
// The kernel stores a full 8-byte off_t through position
// (bsd/vfs/vfs_syscalls.c), so it gets a local int64, never basep.
func Getdirentries(fd int, buf []byte, basep *uintptr) (n int, err error) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	} else {
		p = unsafe.Pointer(&_zero)
	}
	var pos int64
	r0, _, e1 := Syscall6(SYS_GETDIRENTRIES64, uintptr(fd), uintptr(p), uintptr(len(buf)), uintptr(unsafe.Pointer(&pos)), 0, 0)
	n = int(r0)
	if e1 != 0 {
		return n, errnoErr(e1)
	}
	if basep != nil {
		*basep = uintptr(pos)
	}
	return n, nil
}

// Leopard has no *at system calls and no *at functions in libSystem. As in
// the port's internal/syscall/unix/at_darwin_ppc.go, the *at family is
// implemented for AT_FDCWD with the path-based calls. Any other dirfd returns
// ENOTSUP, because resolving it against the working directory would silently
// operate on the wrong file. Flags the path-based calls cannot express are
// rejected rather than ignored.

func Openat(dirfd int, path string, mode int, perm uint32) (fd int, err error) {
	if dirfd != AT_FDCWD {
		return -1, ENOTSUP
	}
	return Open(path, mode, perm)
}

func Fstatat(fd int, path string, stat *Stat_t, flags int) (err error) {
	if fd != AT_FDCWD {
		return ENOTSUP
	}
	switch flags {
	case 0:
		return Stat(path, stat)
	case AT_SYMLINK_NOFOLLOW:
		return Lstat(path, stat)
	}
	return EINVAL
}

func Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	if dirfd != AT_FDCWD {
		return ENOTSUP
	}
	switch flags {
	case 0:
		return Access(path, mode)
	case AT_EACCESS:
		if Getuid() != Geteuid() || Getgid() != Getegid() {
			return ENOSYS
		}
		return Access(path, mode)
	}
	return EINVAL
}

func Fchmodat(dirfd int, path string, mode uint32, flags int) (err error) {
	if dirfd != AT_FDCWD {
		return ENOTSUP
	}
	switch flags {
	case 0:
		return Chmod(path, mode)
	case AT_SYMLINK_NOFOLLOW:
		return lchmod(path, mode)
	}
	return EINVAL
}

func Fchownat(dirfd int, path string, uid int, gid int, flags int) (err error) {
	if dirfd != AT_FDCWD {
		return ENOTSUP
	}
	switch flags {
	case 0:
		return Chown(path, uid, gid)
	case AT_SYMLINK_NOFOLLOW:
		return Lchown(path, uid, gid)
	}
	return EINVAL
}

func Linkat(pathfd int, path string, linkfd int, link string, flags int) (err error) {
	if pathfd != AT_FDCWD || linkfd != AT_FDCWD || flags != 0 {
		return ENOTSUP
	}
	return Link(path, link)
}

func Mkdirat(dirfd int, path string, mode uint32) (err error) {
	if dirfd != AT_FDCWD {
		return ENOTSUP
	}
	return Mkdir(path, mode)
}

func Readlinkat(dirfd int, path string, buf []byte) (n int, err error) {
	if dirfd != AT_FDCWD {
		return 0, ENOTSUP
	}
	return Readlink(path, buf)
}

func Renameat(fromfd int, from string, tofd int, to string) (err error) {
	if fromfd != AT_FDCWD || tofd != AT_FDCWD {
		return ENOTSUP
	}
	return Rename(from, to)
}

func Symlinkat(oldpath string, newdirfd int, newpath string) (err error) {
	if newdirfd != AT_FDCWD {
		return ENOTSUP
	}
	return Symlink(oldpath, newpath)
}

func Unlinkat(dirfd int, path string, flags int) (err error) {
	if dirfd != AT_FDCWD {
		return ENOTSUP
	}
	switch flags {
	case 0:
		return Unlink(path)
	case AT_REMOVEDIR:
		return Rmdir(path)
	}
	return EINVAL
}

// The following libSystem functions do not exist on Leopard: they are absent
// from `nm -g -arch ppc /usr/lib/libSystem.B.dylib` on Mac OS X 10.5.8. The Go
// linker binds darwin imports at launch, so importing them would stop any
// program that reaches them from starting. They fail with ENOSYS instead.

// utimensat: UtimesNano in syscall_bsd.go falls back to utimes on ENOSYS.
func utimensat(dirfd int, path string, times *[2]Timespec, flags int) (err error) {
	return ENOSYS
}

func renamexNp(from string, to string, flag uint32) (err error) { return ENOSYS }

func renameatxNp(fromfd int, from string, tofd int, to string, flag uint32) (err error) {
	return ENOSYS
}

func pthread_chdir_np(path string) (err error) { return ENOSYS }

func pthread_fchdir_np(fd int) (err error) { return ENOSYS }

func connectx(fd int, endpoints *SaEndpoints, associd SaeAssocID, flags uint32, iov []Iovec, n *uintptr, connid *SaeConnID) (err error) {
	return ENOSYS
}

func ClockGettime(clockid int32, time *Timespec) (err error) { return ENOSYS }

func Clonefile(src string, dst string, flags int) (err error) { return ENOSYS }

func Clonefileat(srcDirfd int, src string, dstDirfd int, dst string, flags int) (err error) {
	return ENOSYS
}

func Fclonefileat(srcDirfd int, dstDirfd int, dst string, flags int) (err error) {
	return ENOSYS
}

func preadv(fd int, iovecs []Iovec, offset int64) (n int, err error) { return 0, ENOSYS }

func pwritev(fd int, iovecs []Iovec, offset int64) (n int, err error) { return 0, ENOSYS }
