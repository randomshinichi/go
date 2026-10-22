// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix

import (
	"strconv"
	"strings"
	"sync"
	"syscall"
	_ "unsafe" // for linkname
)

// Leopard has no *at family and no *at functions in libSystem, but xnu-1228 has
// two kernel interfaces that take a directory descriptor, and one thread-private
// working directory:
//
//   - fcntl F_OPENFROM (56) and F_UNLINKFROM (57): anchored open and anchored
//     unlink (syscall/at_compound_darwin_ppc.go).
//   - syscall 349, __pthread_fchdir: a per-thread cwd that the rest of the
//     legacy path calls resolve against. The same file runs one anchored
//     operation under it inside a single entersyscall/exitsyscall scope.
//
// That makes most of os.Root expressible without reconstructing a pathname.
// The exceptions below are refused, not approximated:
//
//   - Linkat with any real directory descriptor, even two names under the same
//     parent. Leopard's link(2) follows a final source symlink
//     (bsd/vfs/vfs_syscalls.c:2605-2607) while os.Root.Link must link the link
//     itself, and journaled HFS+ can hard-link directories
//     (vfs_syscalls.c:2619-2636). A pre-lstat cannot close the window between
//     the check and the link.
//   - Renameat whose two operands need two *different* anchored directories.
//     There is no race-free two-anchor rename interface here: only one cwd
//     applies to a pathname, and F_GETPATH/volfs rebuild a pathname and keep
//     the replacement race they were meant to remove.
//   - Fchmodat with AT_SYMLINK_NOFOLLOW. The equivalent is setattrlist with
//     FSOPT_NOFOLLOW, whose libc/permission/errno parity has not been
//     established on this kernel; a following chmod would change the wrong
//     object.
//   - Utimensat with a real descriptor. Same setattrlist candidate, and the
//     timestamp edge cases (both times omitted, UTIME_NOW, permission rules)
//     have to match Darwin before it can replace the refusal.
//
// The backend is enabled only for the kernel this was designed and measured
// against: a parsed kern.osrelease of 9.8.x, the declared Leopard 10.5.8
// target. Older 9.x builds are a different xnu; newer releases are unchecked
// and may reuse these private fcntl numbers. On anything else, and if the
// release cannot be read or parsed, the previous refusal (ENOTSUP for a real
// descriptor) is what callers see. Nothing here probes a syscall for
// existence: Darwin's nosys raises SIGSYS, which Go treats as unrecoverable.

// The native primitives, implemented in package syscall next to the ABI they
// depend on (syscall/at_compound_darwin_ppc.go and .s).
//
//go:linkname atOpenFrom syscall.atOpenFrom
func atOpenFrom(anchor int, path string, flags int, mode uint32) (int, error)

//go:linkname atUnlinkFrom syscall.atUnlinkFrom
func atUnlinkFrom(anchor int, path string) error

//go:linkname atOpenatCompound syscall.atOpenatCompound
func atOpenatCompound(anchor int, path string, flags int, mode uint32) (int, error)

//go:linkname atRmdirCompound syscall.atRmdirCompound
func atRmdirCompound(anchor int, path string) error

//go:linkname atReadlinkatCompound syscall.atReadlinkatCompound
func atReadlinkatCompound(anchor int, path string, buf []byte) (int, error)

//go:linkname atMkdiratCompound syscall.atMkdiratCompound
func atMkdiratCompound(anchor int, path string, mode uint32) error

//go:linkname atSymlinkatCompound syscall.atSymlinkatCompound
func atSymlinkatCompound(target string, anchor int, newname string) error

//go:linkname atStatatCompound syscall.atStatatCompound
func atStatatCompound(anchor int, path string, st *syscall.Stat_t, follow bool) error

//go:linkname atChownatCompound syscall.atChownatCompound
func atChownatCompound(anchor int, path string, uid, gid int, follow bool) error

//go:linkname atChmodatCompound syscall.atChmodatCompound
func atChmodatCompound(anchor int, path string, mode uint32) error

//go:linkname atRenameatCompound syscall.atRenameatCompound
func atRenameatCompound(anchor int, oldpath, newpath string) error

// atCompatOnce and atCompatOK cache the kern.osrelease decision: the answer
// cannot change while the process runs, and every *at entry point asks.
var (
	atCompatOnce sync.Once
	atCompatOK   bool
)

// atCompatEnabled reports whether the anchored backend may be used.
func atCompatEnabled() bool {
	atCompatOnce.Do(func() {
		release, err := syscall.Sysctl("kern.osrelease")
		if err != nil {
			return
		}
		atCompatOK, _ = atCompatRelease(release)
	})
	return atCompatOK
}

// atCompatRelease reports whether a kern.osrelease value is a 9.8.x release,
// the version this backend was designed and measured against (Darwin 9.8.0 /
// Leopard 10.5.8). A version it cannot parse is not accepted.
func atCompatRelease(release string) (ok bool, parsed bool) {
	major, rest, found := strings.Cut(release, ".")
	if !found {
		return false, false
	}
	minor, rest, found := strings.Cut(rest, ".")
	if !found {
		return false, false
	}
	maj, err := strconv.Atoi(major)
	if err != nil {
		return false, false
	}
	min, err := strconv.Atoi(minor)
	if err != nil {
		return false, false
	}
	if maj != 9 || min != 8 {
		return false, true
	}
	// The patch level is not part of the decision, but it must look like one:
	// an empty or non-numeric tail means this is not the shape of a Darwin
	// release string and no version claim should rest on it.
	patch := rest
	if i := strings.IndexByte(patch, '.'); i >= 0 {
		patch = patch[:i]
	}
	if _, err := strconv.Atoi(patch); err != nil {
		return false, false
	}
	return true, true
}

// atUnsupported is the refusal for a real descriptor: the platform cannot do
// this operation confined, and doing it unconfined would act on another
// directory.
func atUnsupported() error { return syscall.ENOTSUP }

func Unlinkat(dirfd int, path string, flags int) error {
	if dirfd == AT_FDCWD {
		return unlinkatPath(path, flags)
	}
	if flags != 0 && flags != AT_REMOVEDIR {
		return syscall.EINVAL
	}
	if !atCompatEnabled() {
		return atUnsupported()
	}
	if flags == 0 {
		return atUnlinkFrom(dirfd, path)
	}
	return atRmdirCompound(dirfd, path)
}

// unlinkatPath is the AT_FDCWD form: an ordinary path call on the calling
// thread, whose effective cwd is what AT_FDCWD means. The dirfd is ignored, not
// validated, exactly as POSIX requires.
func unlinkatPath(path string, flags int) error {
	switch flags {
	case 0:
		return syscall.Unlink(path)
	case AT_REMOVEDIR:
		return syscall.Rmdir(path)
	}
	return syscall.EINVAL
}

func Openat(dirfd int, path string, flags int, perm uint32) (int, error) {
	if dirfd == AT_FDCWD {
		return syscall.Open(path, flags, perm)
	}
	if !atCompatEnabled() {
		return -1, atUnsupported()
	}
	// F_OPENFROM keeps looking in the caller's flags, but its creation retry
	// drops the anchor and can report a spurious ENOENT for a pathname that is
	// there (bsd/vfs/vfs_vnops.c:294-305). Only the non-create form and the
	// exclusive create form, which never reaches that retry, may use it.
	if flags&syscall.O_CREAT == 0 || flags&syscall.O_EXCL != 0 {
		return atOpenFrom(dirfd, path, flags, perm)
	}
	return atOpenatCompound(dirfd, path, flags, perm)
}

func Readlinkat(dirfd int, path string, buf []byte) (int, error) {
	if dirfd == AT_FDCWD {
		return syscall.Readlink(path, buf)
	}
	if !atCompatEnabled() {
		return 0, atUnsupported()
	}
	return atReadlinkatCompound(dirfd, path, buf)
}

func Mkdirat(dirfd int, path string, mode uint32) error {
	if dirfd == AT_FDCWD {
		return syscall.Mkdir(path, mode)
	}
	if !atCompatEnabled() {
		return atUnsupported()
	}
	return atMkdiratCompound(dirfd, path, mode)
}

func Fchmodat(dirfd int, path string, mode uint32, flags int) error {
	switch flags {
	case 0:
		if dirfd == AT_FDCWD {
			return syscall.Chmod(path, mode)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return atChmodatCompound(dirfd, path, mode)
	case AT_SYMLINK_NOFOLLOW:
		// os.Root.Chmod's form. Refused on a real descriptor: see the file
		// comment. With AT_FDCWD it keeps the libc lchmod path it has today.
		if dirfd != AT_FDCWD {
			return atUnsupported()
		}
		return lchmod(path, mode)
	}
	return syscall.EINVAL
}

func Fchownat(dirfd int, path string, uid, gid int, flags int) error {
	switch flags {
	case 0:
		if dirfd == AT_FDCWD {
			return syscall.Chown(path, uid, gid)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return atChownatCompound(dirfd, path, uid, gid, true)
	case AT_SYMLINK_NOFOLLOW:
		if dirfd == AT_FDCWD {
			return syscall.Lchown(path, uid, gid)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		// lchown is a real syscall on this kernel (364), not a libc shim.
		return atChownatCompound(dirfd, path, uid, gid, false)
	}
	return syscall.EINVAL
}

func Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) error {
	oldAbs, newAbs := isAbsPath(oldpath), isAbsPath(newpath)

	// AT_FDCWD means the calling thread's effective cwd, which the helper that
	// runs an anchored operation does not have. Every operand that resolves
	// from it therefore has to stay on a path call, and that is exact only when
	// no operand needs a real anchored directory.
	if olddirfd == AT_FDCWD || newdirfd == AT_FDCWD {
		if (olddirfd == AT_FDCWD || oldAbs) && (newdirfd == AT_FDCWD || newAbs) {
			return syscall.Rename(oldpath, newpath)
		}
		return atUnsupported()
	}

	if !atCompatEnabled() {
		return atUnsupported()
	}
	if olddirfd != newdirfd {
		// Two descriptors are one anchor only if they are the same directory
		// object. Compare live identity, not the names the descriptors were
		// opened from.
		same, err := sameDirectory(olddirfd, newdirfd)
		if err != nil {
			return err
		}
		if !same {
			return atUnsupported()
		}
	}
	switch {
	case !oldAbs && !newAbs:
		return atRenameatCompound(olddirfd, oldpath, newpath)
	case oldAbs && newAbs:
		// An absolute pathname needs no anchor, so nothing here is confined by
		// one; the plain call is exact.
		return syscall.Rename(oldpath, newpath)
	case oldAbs:
		return atRenameatCompound(newdirfd, oldpath, newpath)
	default:
		return atRenameatCompound(olddirfd, oldpath, newpath)
	}
}

// sameDirectory reports whether two descriptors refer to the same directory
// object, by device and inode.
func sameDirectory(a, b int) (bool, error) {
	var sa, sb syscall.Stat_t
	if err := syscall.Fstat(a, &sa); err != nil {
		return false, err
	}
	if err := syscall.Fstat(b, &sb); err != nil {
		return false, err
	}
	return sa.Dev == sb.Dev && sa.Ino == sb.Ino, nil
}

func Linkat(olddirfd int, oldpath string, newdirfd int, newpath string, flag int) error {
	if olddirfd != AT_FDCWD || newdirfd != AT_FDCWD || flag != 0 {
		// Every real-descriptor form is refused, including two names under one
		// parent: see the file comment.
		return atUnsupported()
	}
	return syscall.Link(oldpath, newpath)
}

func Symlinkat(oldpath string, newdirfd int, newpath string) error {
	if newdirfd == AT_FDCWD {
		return syscall.Symlink(oldpath, newpath)
	}
	if !atCompatEnabled() {
		return atUnsupported()
	}
	// Only the new name is resolved against the descriptor; oldpath is literal
	// link contents and is never looked up here.
	return atSymlinkatCompound(oldpath, newdirfd, newpath)
}

func Fstatat(dirfd int, path string, stat *syscall.Stat_t, flags int) error {
	if dirfd == AT_FDCWD {
		// The calling thread's own cwd, so the path-based pair applies; it also
		// owns the flag validation.
		return fstatat(dirfd, path, stat, flags)
	}
	if flags != 0 && flags != AT_SYMLINK_NOFOLLOW {
		return syscall.EINVAL
	}
	if !atCompatEnabled() {
		return atUnsupported()
	}
	return atStatatCompound(dirfd, path, stat, flags == 0)
}

// fstatat is the AT_FDCWD form: stat or lstat of a path relative to the calling
// thread's cwd.
//
//go:linkname fstatat syscall.fstatat
func fstatat(dirfd int, path string, stat *syscall.Stat_t, flags int) error

// LstatatByPath is lstat of name inside the directory open as dirfd, done by
// resolving dirfd to a path (syscall.lstatatByPath, which documents the
// measurement it rests on and what it costs). It exists for the directory
// entries of an ordinary (*os.File).Readdir, which the confined form could also
// serve now, and it is the remaining weaker substitute there: it loses the
// descriptor-based lookup's race protection. A File opened through an os.Root
// does not use it (os/statat_darwin_ppc.go).
//
//go:linkname LstatatByPath syscall.lstatatByPath
func LstatatByPath(dirfd int, name string, stat *syscall.Stat_t) error

// lchmod changes the link itself for AT_SYMLINK_NOFOLLOW.
//
//go:linkname lchmod syscall.lchmod
func lchmod(path string, mode uint32) error

// isAbsPath reports whether path is absolute, so that no descriptor's directory
// is consulted for it.
func isAbsPath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

// faccessat is used by Eaccess, which requests AT_EACCESS. No caller in the
// standard library passes a real descriptor, so it needs no anchored form; if
// that changes, this is a refusal to re-audit rather than a missing feature.
func faccessat(dirfd int, path string, mode uint32, flags int) error {
	if dirfd != AT_FDCWD {
		return atUnsupported()
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
