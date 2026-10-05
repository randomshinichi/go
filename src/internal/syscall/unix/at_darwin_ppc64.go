// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"syscall"
	_ "unsafe" // for linkname
)

// Mac OS X 10.5 (Leopard) has no *at system calls and no *at functions in
// libSystem: unlinkat, openat, readlinkat, faccessat, fstatat, renameat,
// mkdirat, symlinkat are all absent (dlprobe and nm -g of libSystem.B.dylib,
// out/stage2-20261003/parent-native-sched/PARENT-MEASUREMENTS-D9.md). There is
// also no AT_FDCWD: the value -2 in at_sysnum_darwin.go is a sentinel chosen to
// match Apple's later releases, not a value Leopard provides.
//
// So this file implements the *at family for the one case that has a plain
// equivalent, dirfd == AT_FDCWD, by calling the path-based function. For any
// other dirfd it returns ENOTSUP, and it never falls back to the path-based
// call: Leopard cannot resolve a path against a directory descriptor, and
// running the call against the process's working directory instead would
// silently act on the wrong directory. The visible consequence is that
// os.Root, and the fd-based halves of os (openDirAt and friends), fail with
// ENOTSUP on darwin/ppc64 rather than misbehave.
//
// There is one exception to that, and it is a read: LstatatByPath, below, lets
// (*os.File).Readdir lstat the entries of a directory it has open by resolving
// the descriptor to a path. No operation that writes gets such a form.

// Flag combinations the path-based call cannot express are rejected, with
// EINVAL for undefined bits and ENOTSUP for defined ones Leopard cannot honour.

func Unlinkat(dirfd int, path string, flags int) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	switch flags {
	case 0:
		return syscall.Unlink(path)
	case AT_REMOVEDIR:
		return syscall.Rmdir(path)
	}
	return syscall.EINVAL
}

func Openat(dirfd int, path string, flags int, perm uint32) (int, error) {
	if dirfd != AT_FDCWD {
		return -1, syscall.ENOTSUP
	}
	return syscall.Open(path, flags, perm)
}

func Readlinkat(dirfd int, path string, buf []byte) (int, error) {
	if dirfd != AT_FDCWD {
		return 0, syscall.ENOTSUP
	}
	return syscall.Readlink(path, buf)
}

func Mkdirat(dirfd int, path string, mode uint32) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	return syscall.Mkdir(path, mode)
}

func Fchmodat(dirfd int, path string, mode uint32, flags int) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	switch flags {
	case 0:
		return syscall.Chmod(path, mode)
	case AT_SYMLINK_NOFOLLOW:
		// lchmod changes the link itself: measured on the G5, lchmod(link, 0600)
		// made lstat(link) report 0120600 and left stat(target) at 0100644, while
		// the control chmod(link) followed the link (out/stage2-20261003/
		// parent-native-run/BREADTH-AT-NATIVE-RUNG.md item 3).
		return lchmod(path, mode)
	}
	return syscall.EINVAL
}

func Fchownat(dirfd int, path string, uid, gid int, flags int) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	switch flags {
	case 0:
		return syscall.Chown(path, uid, gid)
	case AT_SYMLINK_NOFOLLOW:
		return syscall.Lchown(path, uid, gid)
	}
	return syscall.EINVAL
}

func Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) error {
	if olddirfd != AT_FDCWD || newdirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	return syscall.Rename(oldpath, newpath)
}

func Linkat(olddirfd int, oldpath string, newdirfd int, newpath string, flag int) error {
	// A non-zero flag (AT_SYMLINK_FOLLOW) changes which object is linked and
	// has no path-based equivalent here.
	if olddirfd != AT_FDCWD || newdirfd != AT_FDCWD || flag != 0 {
		return syscall.ENOTSUP
	}
	return syscall.Link(oldpath, newpath)
}

func Symlinkat(oldpath string, newdirfd int, newpath string) error {
	if newdirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	return syscall.Symlink(oldpath, newpath)
}

func Fstatat(dirfd int, path string, stat *syscall.Stat_t, flags int) error {
	return fstatat(dirfd, path, stat, flags)
}

// fstatat is syscall.fstatat (syscall_darwin_ppc64.go), which already follows
// the same rule: stat or lstat for AT_FDCWD, ENOTSUP otherwise.
//
//go:linkname fstatat syscall.fstatat
func fstatat(dirfd int, path string, stat *syscall.Stat_t, flags int) error

// LstatatByPath is lstat of name inside the directory open as dirfd, done by
// resolving dirfd to a path (syscall.lstatatByPath, which documents the
// measurement it rests on and what it costs). It is the one exception to the
// ENOTSUP rule above, and it exists only for (*os.File).Readdir. Fstatat above
// keeps refusing a real dirfd, so os.Root's lstat of an entry stays ENOTSUP:
// Root's whole purpose is the race protection this path form gives up.
//
//go:linkname LstatatByPath syscall.lstatatByPath
func LstatatByPath(dirfd int, name string, stat *syscall.Stat_t) error

// lchmod is syscall.lchmod (zsyscall_darwin_ppc64.go).
//
//go:linkname lchmod syscall.lchmod
func lchmod(path string, mode uint32) error

// faccessat is used by Eaccess, which asks for AT_EACCESS (check with the
// effective user and group ids).
func faccessat(dirfd int, path string, mode uint32, flags int) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	switch flags {
	case 0:
		return syscall.Access(path, mode)
	case AT_EACCESS:
		// Leopard's access checks the real ids and offers no effective-id
		// variant. They only differ in a set-id process; there, report ENOSYS,
		// on which Eaccess's callers fall back to checking the mode bits
		// (os/exec findExecutable), rather than answer for the wrong ids.
		if syscall.Getuid() != syscall.Geteuid() || syscall.Getgid() != syscall.Getegid() {
			return syscall.ENOSYS
		}
		return syscall.Access(path, mode)
	}
	return syscall.EINVAL
}
