// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix

import (
	"syscall"
	_ "unsafe" // for linkname
)

// Leopard has no *at system calls or *at functions in libSystem. Implement
// the *at family for AT_FDCWD by calling path-based functions. For any other
// dirfd, return ENOTSUP: using the process working directory instead would
// silently operate on the wrong directory and defeat os.Root's confinement.
// A path-based fallback is therefore safe only for AT_FDCWD.
//
// Flags that path-based calls cannot express are rejected rather than ignored.

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

// syscall.fstatat is hand-written for darwin/ppc and already rejects real dirfds.
//
//go:linkname fstatat syscall.fstatat
func fstatat(dirfd int, path string, stat *syscall.Stat_t, flags int) error

// LstatatByPath is lstat of name inside the directory open as dirfd, done by
// resolving dirfd to a path (syscall.lstatatByPath, which documents the
// measurement it rests on and what it costs). It is the one exception to the
// ENOTSUP rule and exists only for (*os.File).Readdir. Fstatat above keeps
// refusing a real dirfd, so os.Root's lstat of an entry stays ENOTSUP: Root's
// whole purpose is the race protection this path form gives up. Mirrors
// at_darwin_ppc64.go.
//
//go:linkname LstatatByPath syscall.lstatatByPath
func LstatatByPath(dirfd int, name string, stat *syscall.Stat_t) error

// lchmod changes the link itself for AT_SYMLINK_NOFOLLOW.
//
//go:linkname lchmod syscall.lchmod
func lchmod(path string, mode uint32) error

// faccessat is used by Eaccess, which requests AT_EACCESS.
func faccessat(dirfd int, path string, mode uint32, flags int) error {
	if dirfd != AT_FDCWD {
		return syscall.ENOTSUP
	}
	switch flags {
	case 0:
		return syscall.Access(path, mode)
	case AT_EACCESS:
		if syscall.Getuid() != syscall.Geteuid() || syscall.Getgid() != syscall.Getegid() {
			return syscall.ENOSYS
		}
		return syscall.Access(path, mode)
	}
	return syscall.EINVAL
}
