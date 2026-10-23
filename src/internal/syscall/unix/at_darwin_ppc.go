// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix

import (
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
//   - Fchmodat with AT_SYMLINK_NOFOLLOW and Utimensat with a real descriptor are
//     not refused: they run setattrlist with FSOPT_NOFOLLOW, selecting
//     ATTR_CMN_ACCESSMASK or the MODTIME/CHGTIME/ACCTIME timespecs. Leopard has
//     no lchmod and no lutimes syscall; libSystem's lchmod/lutimes are built on
//     this same request (disassembled and measured on the target in
//     runs/root-at-metadata-20261023), which is what makes it the honest
//     equivalent rather than a following chmod that would change the wrong
//     object.
//   - Utimensat with a real descriptor and a nil times pointer (utimes(NULL) is
//     "both times now"), any UTIME_NOW timestamp, and the general UTIME_* forms:
//     refused, because their permission rules differ from the explicit-time form
//     this backend implements. os.Root.Chtimes never sends them.
//
// One timestamp case needs help rather than refusal. An access-time-ONLY request
// is a lazy access-time touch on HFS, which stores "now" instead of the supplied
// value; the workaround and the guard that keeps its dummy change time off other
// filesystems are in atUtimensatNofollow and syscall.atChgTimeForAccessTime.
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

//go:linkname atChmodNofollowCompound syscall.atChmodNofollowCompound
func atChmodNofollowCompound(anchor int, path string, mode uint32) error

//go:linkname atUtimensatCompound syscall.atUtimensatCompound
func atUtimensatCompound(anchor int, path string, mod, chg, acc *syscall.Timespec) error

//go:linkname atChgTimeForAccessTime syscall.atChgTimeForAccessTime
func atChgTimeForAccessTime(anchor int, path string) (syscall.Timespec, bool, error)

//go:linkname atRenameatCompound syscall.atRenameatCompound
func atRenameatCompound(anchor int, oldpath, newpath string) error

// atCompatOnce and atCompatOK cache the kern.osrelease decision: the answer
// cannot change while the process runs, and every *at entry point asks.
var (
	atCompatOnce sync.Once
	atCompatOK   bool
)

// atCompatEnabled reports whether the anchored backend may be used.
//
// atCompatForceDisabled is a compile-time constant, true only in a build made
// with the goppc_atcompat_disabled tag (two files, one per tag value). It exists
// so the disabled-backend dispatch can be executed and asserted on a kernel where
// the backend would otherwise be on — the refusal path is otherwise reachable
// only on a machine nobody has. It is not a runtime switch and no production
// build contains the true value.
func atCompatEnabled() bool {
	if atCompatForceDisabled {
		return false
	}
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
//
// This is hand-rolled on bytes rather than using strings and strconv on
// purpose: internal/syscall/unix belongs to the SYSCALL layer of the go/build
// dependency lattice (src/go/build/deps_test.go), which forbids both packages
// here, and TestDependencies enforces that lattice. Do not "tidy" this back
// into strings.Cut/strconv.Atoi. The grammar is the one that matters for the
// gate: three dot-separated decimal components, of which the first two decide
// and the third only has to look numeric.
func atCompatRelease(release string) (ok bool, parsed bool) {
	major, next, haveMajor := atReleaseComponent(release, 0)
	if !haveMajor || next >= len(release) || release[next] != '.' {
		return false, false
	}
	minor, next, haveMinor := atReleaseComponent(release, next+1)
	if !haveMinor || next >= len(release) || release[next] != '.' {
		return false, false
	}
	// The patch level is not part of the decision, but it must look like one:
	// an empty or non-numeric tail means this is not the shape of a Darwin
	// release string and no version claim should rest on it.
	_, next, havePatch := atReleaseComponent(release, next+1)
	if !havePatch {
		return false, false
	}
	if next < len(release) && release[next] != '.' {
		return false, false
	}
	if major != 9 || minor != 8 {
		return false, true
	}
	return true, true
}

// atReleaseComponent reads one decimal component of a release string starting
// at offset i. It reports the component's value, the offset just past it and
// whether any digit was seen at all.
//
// The value saturates well above any real release number instead of wrapping:
// only equality with 9 and 8 is ever asked, and a wrapped value that happened
// to land on one of them would enable the backend on a string nobody meant as
// a version.
func atReleaseComponent(s string, i int) (value int, next int, ok bool) {
	start := i
	v := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		if v < 1<<20 {
			v = v*10 + int(s[i]-'0')
		}
		i++
	}
	if i == start {
		return 0, i, false
	}
	return v, i, true
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
		// os.Root.Chmod's form. With AT_FDCWD it keeps the libc lchmod path it
		// has today, which itself resolves the final component and uses
		// setattrlist (with FSOPT_NOFOLLOW) only for a symlink.
		if dirfd == AT_FDCWD {
			return lchmod(path, mode)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return atChmodNofollow(dirfd, path, mode)
	}
	return syscall.EINVAL
}

// atChmodNofollow is Fchmodat(AT_SYMLINK_NOFOLLOW) for a real directory
// descriptor.
func atChmodNofollow(anchor int, path string, mode uint32) error {
	// chmod(2) masks the mode with ALLPERMS before the kernel sees it
	// (bsd/vfs/vfs_syscalls.c:4106-4117); setattrlist's ATTR_CMN_ACCESSMASK is
	// unpacked as an unmasked 32-bit value, so the mask is applied here to keep
	// the two forms agreeing on the file-mode semantics the kernel documents.
	err := atChmodNofollowCompound(anchor, path, mode&0o7777)
	return atAuthErrno(err, true)
}

// atUtimensatNofollow is Utimensat(AT_SYMLINK_NOFOLLOW) for a real directory
// descriptor, for the explicit-time form os.Root.Chtimes uses.
//
// An access-time-only request needs one extra step, and it is the HFS workaround
// the measured behaviour forced (syscall.atChgTimeForAccessTime documents the
// kernel source and the measurements): HFS discards the supplied access time
// when the access time is the only active attribute, and stores it exactly as
// soon as any other attribute is active. So on HFS — and only there, checked
// through the descriptor for the anchor and through st_dev for the entry, so a
// mount point inside the root cannot smuggle the dummy change time onto a
// filesystem that honours it — the request also selects ATTR_CMN_CHGTIME with a
// value the filesystem ignores. On every other filesystem the ordinary selective
// request is sent, unchanged, and on a filesystem query error nothing is sent:
// the error is returned rather than guessing which form is right.
//
// UTIME_OMIT omits its attribute rather than reading the old value and writing
// it back, so there is no read-modify-write window and no chance of rewriting an
// ambient change. Both times omitted is a legal request: it still resolves the
// pathname (which is how a missing name is reported) and authorizes nothing.
func atUtimensatNofollow(anchor int, path string, times *[2]syscall.Timespec) error {
	if times == nil {
		// utimensat(..., NULL) means both times set to the current time, which
		// needs a clock read inside the native scope and follows utimes(NULL)'s
		// different permission rule. Refused rather than approximated.
		return atUnsupported()
	}
	const utimeNow = -1
	// times[0] is the access time and times[1] the modification time.
	var sel [2]*syscall.Timespec
	for i, ts := range times {
		switch ts.Nsec {
		case UTIME_OMIT:
			// POSIX says tv_sec is ignored when tv_nsec is UTIME_OMIT, and
			// package os spells an omitted time as both fields set to the
			// sentinel. Either spelling means "do not send this attribute".
			continue
		case utimeNow:
			// UTIME_NOW is not claimed for this form; see the file comment.
			return atUnsupported()
		}
		if ts.Nsec < 0 || ts.Nsec >= 1e9 {
			return syscall.EINVAL
		}
		ts := ts
		sel[i] = &ts
	}

	// sel[0] is the access time, sel[1] the modification time. Only the
	// access-time-alone form needs the HFS workaround; with the modification
	// time selected the ordinary timestamp block already stores the access time.
	var chg *syscall.Timespec
	if sel[0] != nil && sel[1] == nil {
		dummy, needed, err := atChgTimeForAccessTime(anchor, path)
		if err != nil {
			// The filesystem could not be identified, so which form is exact is
			// unknown. Report that instead of changing an attribute with an
			// unverified request.
			return err
		}
		if needed {
			chg = &dummy
		}
	}
	err := atUtimensatCompound(anchor, path, sel[1], chg, sel[0])
	return atAuthErrno(err, sel[0] != nil || sel[1] != nil)
}

// atAuthErrno maps the authorization failure a metadata change reports to the
// errno the equivalent second-order libc call reports.
//
// The kernel returns EACCES from vnode_authorize for a metadata change the
// caller may not make (bsd/vfs/vfs_subr.c:5965-6115), while chmod(2) and the
// explicit-time utimes(2) turn exactly that failure into EPERM
// (vfs_syscalls.c:3998-4006,4386-4394). libc's lchmod/lutimes reach those forms
// for everything that is not a symlink, which is the only object os.Root.Chmod
// and os.Root.Chtimes reach: both resolve a symlink before the call, so the
// no-follow form here sees a non-symlink except for the documented
// replaced-after-the-check race. Measured on the target: lchmod/lutimes on a
// file the caller does not own return EPERM, and the same setattrlist request
// returns EACCES (runs/root-at-metadata-20261023).
//
// auth is false when the request changes nothing, so no authorization step can
// run and an EACCES can only come from pathname resolution, where the path forms
// also keep EACCES.
//
// Residual, recorded rather than hidden: the kernel does not tell us whether an
// EACCES came from resolving the name or from authorizing the change, so a
// search-permission failure on the final component is reported as EPERM here
// where the path-based call keeps EACCES. Both satisfy errors.Is(err,
// fs.ErrPermission).
func atAuthErrno(err error, auth bool) error {
	if auth && err == syscall.EACCES {
		return syscall.EPERM
	}
	return err
}

// AnchoredAvailable reports whether the anchored backend is enabled on this
// kernel. Package os uses it to choose the confined descriptor form over the
// pathname substitute for an ordinary directory File's entries; on a kernel
// where the backend is off the substitute is the only form that works.
func AnchoredAvailable() bool { return atCompatEnabled() }

// Utimensat is utimensat(2) for this platform. Leopard has no utimensat, so the
// shared implementation is syscall.utimensat's path form; the one case it cannot
// express is a real directory descriptor with AT_SYMLINK_NOFOLLOW, which runs in
// the private-cwd compound instead.
func Utimensat(dirfd int, path string, times *[2]syscall.Timespec, flag int) error {
	if dirfd == AT_FDCWD || flag != AT_SYMLINK_NOFOLLOW || !atCompatEnabled() {
		return utimensatPath(dirfd, path, times, flag)
	}
	return atUtimensatNofollow(dirfd, path, times)
}

// utimensatPath is the path-based implementation: syscall.utimensat handles
// UTIME_OMIT by stat-then-substitute for AT_FDCWD, which stays correct on the
// caller's own thread.
//
//go:linkname utimensatPath syscall.utimensat
func utimensatPath(dirfd int, path string, times *[2]syscall.Timespec, flag int) error

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
