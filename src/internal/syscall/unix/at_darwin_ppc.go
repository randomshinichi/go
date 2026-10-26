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
//
// A name that is absolute resolves from the root, so no descriptor is consulted
// for it and none is validated: Openat, Unlinkat, Readlinkat, Mkdirat, Fstatat,
// Fchownat, Fchmodat and Utimensat classify the name before they look at a
// descriptor or the kernel gate, and run the ordinary caller-thread call (or, for
// the no-follow metadata forms, the same setattrlist request over a plain raw
// syscall with no private-cwd install/identity/clear). Symlinkat classifies only
// the new name; its first argument is literal link contents. One absolute form is
// deliberately not converted: see the access-time-only note below.
//
// The two-anchor operations, Linkat and Renameat, follow the anchor table in
// atTwoNameDispatch. Their compatibility divergences are stated here because no
// caller can see them from the outside, and none of them is an exemption from a
// test:
//
//   - Zero-flag Linkat FOLLOWS a final source symlink. Leopard's link(2) resolves
//     the first name with LOOKUP|FOLLOW (bsd/vfs/vfs_syscalls.c:2605-2608) and
//     there is no llink(2), so a source symlink is linked by referent rather than
//     by inode. POSIX linkat(2) with no flags links the symlink itself, and
//     os.Root.Link promises exactly that, so through os.Root a zero-flag link can
//     link a file outside the root through a symlink inside it, and a dangling
//     source symlink is ENOENT instead of a new hard link. AT_SYMLINK_NOFOLLOW is
//     therefore refused with ENOTSUP (the C mechanism's dummy llink), and any
//     other flag value, AT_SYMLINK_FOLLOW included, is EINVAL before a name is
//     looked at.
//   - A zero-flag Linkat can create a directory hard link. The kernel permits it
//     on journaled HFS+ with ownership checks; nothing here pre-checks the object
//     type, because a check would not close the window before the link. Creating
//     directory hard links is not exercised by any test and is barred on the
//     shared target machines: their teardown in a shared tree is unreliable
//     (removing one of them can leave the object reachable only through the other,
//     with the directory's own entries still referring to it).
//   - A Rename or Link between two different relative directories reconstructs the
//     DESTINATION name from its descriptor with fcntl F_GETPATH and then uses it.
//     That is a name snapshot, not a held anchor: if the destination directory's
//     own path is renamed or a component of it is replaced between the F_GETPATH
//     and the syscall, the operation acts wherever the reconstructed name points.
//     A rename can therefore overwrite a file outside the directory the caller
//     anchored, and dev/ino comparisons before the lookup do not close it. The
//     source keeps the real private-cwd anchor. The C mechanism this follows makes
//     the same trade (M/src/atfuncs.c:470-475) but checks the combined length with
//     strlcat's return value, which reports a truncated result as success; here the
//     total pathname length is checked before anything is looked up and the
//     operation fails with ERANGE instead.
//   - A relative AT_FDCWD operand in one of those mixed forms is captured by
//     opening "." on the caller's own thread. That is real I/O: a directory the
//     caller may search but not read-open, or a descriptor shortage, fails the
//     operation where the pure path form would not have opened anything. There is
//     no fallback to a pathname.
//   - Rename moves whatever names it resolves. One rename is one atomic namespace
//     operation; it is not an atomic selection of the two directories the caller
//     named.
//
// Fchmodat with AT_SYMLINK_NOFOLLOW and Utimensat with a real descriptor are
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
// An access-time-ONLY request for an absolute name is not converted to the
// no-anchor envelope. The HFS workaround below needs a guard that identifies the
// filesystem and the entry (fstatfs of an anchor, fstat of the same descriptor,
// lstat of the target); the no-anchor form has no anchor to hold, and a guard
// equivalent to the approved one has not been demonstrated, so an absolute
// access-time-only request keeps the anchored dispatch it has today.
// TestAtAbsoluteAtimeOnlyDeferredDarwinPPC pins that (it succeeds with the gate
// on and is refused with the gate forced off, which is only true while the
// anchored dispatch is still the one running).
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

//go:linkname atLinkatCompound syscall.atLinkatCompound
func atLinkatCompound(anchor int, oldpath, newpath string) error

// The no-anchor forms, for a name that resolves from the root and therefore
// needs no private working directory: the same request the anchored form sends,
// over a plain raw syscall.
//
//go:linkname atChmodNofollowNoAnchor syscall.atChmodNofollowNoAnchor
func atChmodNofollowNoAnchor(path string, mode uint32) error

//go:linkname atUtimensatNoAnchor syscall.atUtimensatNoAnchor
func atUtimensatNoAnchor(path string, mod, chg, acc *syscall.Timespec) error

// atPathFromFD reconstructs an absolute pathname for a name relative to the
// directory open as dirfd (fcntl F_GETPATH); atConcatPath is the length check it
// uses, exported separately so both can be tested on their own.
//
//go:linkname atPathFromFD syscall.atPathFromFD
func atPathFromFD(dirfd int, suffix string) (string, error)

//go:linkname atConcatPath syscall.atConcatPath
func atConcatPath(base, suffix string) (string, error)

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
	if dirfd == AT_FDCWD || isAbsPath(path) {
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
	if dirfd == AT_FDCWD || isAbsPath(path) {
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
	if dirfd == AT_FDCWD || isAbsPath(path) {
		return syscall.Readlink(path, buf)
	}
	if !atCompatEnabled() {
		return 0, atUnsupported()
	}
	return atReadlinkatCompound(dirfd, path, buf)
}

func Mkdirat(dirfd int, path string, mode uint32) error {
	if dirfd == AT_FDCWD || isAbsPath(path) {
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
		if dirfd == AT_FDCWD || isAbsPath(path) {
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
		if isAbsPath(path) {
			// No anchor is involved in an absolute name, so the same request can
			// be sent over a plain raw syscall instead of the private-cwd
			// compound: there is nothing to install and nothing to identify.
			return atChmodNofollowNoAnchorPath(path, mode)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return atChmodNofollow(dirfd, path, mode)
	}
	return syscall.EINVAL
}

// atChmodNofollowNoAnchorPath is atChmodNofollow for a name that needs no
// anchor. The request, the ALLPERMS mask and the authorization errno mapping are
// the same; only the private-cwd compound is omitted.
func atChmodNofollowNoAnchorPath(path string, mode uint32) error {
	return atAuthErrno(atChmodNofollowNoAnchor(path, mode&0o7777), true)
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
	sel, err := atTimesSelect(times)
	if err != nil {
		return err
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
	err = atUtimensatCompound(anchor, path, sel[1], chg, sel[0])
	return atAuthErrno(err, sel[0] != nil || sel[1] != nil)
}

// atTimesSelect validates a utimensat times argument for the no-follow form and
// returns the two times that are actually being set: a nil entry is an omitted
// attribute, which the request builder leaves out of the bitmap instead of
// reading the old value and writing it back.
//
// The three refusals it returns are the semantics this backend claims: a nil
// times pointer ("both times now") needs a clock read and utimes(NULL)'s
// different permission rule, UTIME_NOW is not approximated from a sampled clock,
// and an out-of-range nanosecond field is EINVAL. Handling them here keeps the
// anchored and no-anchor forms from drifting apart.
func atTimesSelect(times *[2]syscall.Timespec) (sel [2]*syscall.Timespec, err error) {
	if times == nil {
		return sel, atUnsupported()
	}
	const utimeNow = -1
	// times[0] is the access time and times[1] the modification time.
	for i, ts := range times {
		switch ts.Nsec {
		case UTIME_OMIT:
			// POSIX says tv_sec is ignored when tv_nsec is UTIME_OMIT, and
			// package os spells an omitted time as both fields set to the
			// sentinel. Either spelling means "do not send this attribute".
			continue
		case utimeNow:
			// UTIME_NOW is not claimed for this form; see the file comment.
			return sel, atUnsupported()
		}
		if ts.Nsec < 0 || ts.Nsec >= 1e9 {
			return sel, syscall.EINVAL
		}
		ts := ts
		sel[i] = &ts
	}
	return sel, nil
}

// atUtimensatNoAnchorPath sends an already-selected explicit-time request for a
// name that needs no anchor. The access-time-only form does not come here: its
// HFS guard is written against an anchor descriptor (see Utimensat).
func atUtimensatNoAnchorPath(path string, sel [2]*syscall.Timespec) error {
	err := atUtimensatNoAnchor(path, sel[1], nil, sel[0])
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
	if isAbsPath(path) {
		// An absolute name needs no working directory, so the same
		// setattrlist request can go out over a plain raw syscall. The one
		// exception is the access-time-only form below, whose HFS guard is a
		// named deferral for the no-anchor envelope.
		if sel, err := atTimesSelect(times); err == nil && !(sel[0] != nil && sel[1] == nil) {
			return atUtimensatNoAnchorPath(path, sel)
		}
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
		if dirfd == AT_FDCWD || isAbsPath(path) {
			return syscall.Chown(path, uid, gid)
		}
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return atChownatCompound(dirfd, path, uid, gid, true)
	case AT_SYMLINK_NOFOLLOW:
		if dirfd == AT_FDCWD || isAbsPath(path) {
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

// atTwoNameOp is one of the two compatibility operations that take two names
// which may each resolve through a different directory. They share one anchor
// table, so the selection below is written once and each operation supplies its
// own kernel call.
type atTwoNameOp struct {
	// legacy is the caller-thread path call, used whenever no operand needs a
	// real anchored directory.
	legacy func(oldpath, newpath string) error
	// anchored runs the operation in one private-cwd scope on anchor.
	anchored func(anchor int, oldpath, newpath string) error
}

var (
	atRenameOp = atTwoNameOp{legacy: syscall.Rename, anchored: atRenameatCompound}
	atLinkOp   = atTwoNameOp{legacy: syscall.Link, anchored: atLinkatCompound}
)

// atTwoNameDispatch selects how a two-name operation runs.
//
// The order of the tests is the whole point. An absolute name resolves from the
// root, so its descriptor is ignored rather than validated, and a descriptor
// that happens to be AT_FDCWD on the other side is not a "mixed" case: it has
// nothing to anchor and the ordinary path call is exact. Everything that needs
// a private working directory is decided after that, so the gate is only
// consulted when one is actually required.
//
// Relative names that each need a real, different directory cannot both be
// anchored at once. Rename and link are served by reconstructing the second name
// from its descriptor (fcntl F_GETPATH) in Phase 2, and refused until then; a
// relative AT_FDCWD operand is captured by opening the caller's own working
// directory, which is what makes the mixed forms expressible.
func atTwoNameDispatch(op *atTwoNameOp, olddirfd int, oldpath string, newdirfd int, newpath string) error {
	oldAbs, newAbs := isAbsPath(oldpath), isAbsPath(newpath)

	switch {
	case oldAbs && newAbs:
		// No descriptor is consulted, so this works with any fds and on a
		// kernel where the anchored backend is off.
		return op.legacy(oldpath, newpath)
	case oldAbs && newdirfd == AT_FDCWD:
		return op.legacy(oldpath, newpath)
	case newAbs && olddirfd == AT_FDCWD:
		return op.legacy(oldpath, newpath)
	case oldAbs:
		// Absolute source, relative destination: anchor only the destination.
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return op.anchored(newdirfd, oldpath, newpath)
	case newAbs:
		if !atCompatEnabled() {
			return atUnsupported()
		}
		return op.anchored(olddirfd, oldpath, newpath)
	case olddirfd == AT_FDCWD && newdirfd == AT_FDCWD:
		// Both names resolve from the caller's own working directory, which the
		// calling thread has; one path call is exact.
		return op.legacy(oldpath, newpath)
	}

	// Both names are relative and at least one descriptor is real.
	if !atCompatEnabled() {
		return atUnsupported()
	}
	oldAnchor, closeOld, err := atRelativeAnchor(olddirfd)
	if err != nil {
		return err
	}
	defer closeOld()
	newAnchor, closeNew, err := atRelativeAnchor(newdirfd)
	if err != nil {
		return err
	}
	defer closeNew()

	same, err := sameDirectory(oldAnchor, newAnchor)
	if err != nil {
		return err
	}
	if same {
		return op.anchored(oldAnchor, oldpath, newpath)
	}
	// Two different directories, one private cwd. The destination's directory
	// name is reconstructed from its descriptor and the destination name is made
	// absolute against it, so the source keeps the real anchor and the second name
	// needs none. This is a name snapshot, not a held anchor: see
	// atPathFromFD and the file header for what that costs.
	dst, err := atPathFromFD(newAnchor, newpath)
	if err != nil {
		return err
	}
	return op.anchored(oldAnchor, oldpath, dst)
}

// atRelativeAnchor turns one operand's descriptor into a usable anchor. A
// relative AT_FDCWD operand means the CALLING thread's working directory, and the
// helper that runs an anchored operation runs on another thread, so it is
// captured here, on the caller's own thread, by opening ".". The captured
// descriptor is then just another held directory.
//
// The capture is real I/O with real failure modes: a directory the caller may
// search but not read open, or an exhausted descriptor table, fails the operation
// with that error rather than falling back to a pathname. The returned close
// function only closes a descriptor this function opened, and the caller defers it
// so no path out of the dispatch leaks one.
func atRelativeAnchor(dirfd int) (int, func(), error) {
	if dirfd != AT_FDCWD {
		return dirfd, func() {}, nil
	}
	fd, err := syscall.Open(".", syscall.O_RDONLY, 0)
	if err != nil {
		return -1, func() {}, err
	}
	return fd, func() { syscall.Close(fd) }, nil
}

// Renameat is renameat(2) for this platform.
func Renameat(olddirfd int, oldpath string, newdirfd int, newpath string) error {
	return atTwoNameDispatch(&atRenameOp, olddirfd, oldpath, newdirfd, newpath)
}

// sameDirectory reports whether two descriptors refer to the same directory
// object, by device and inode.
//
// Both descriptors have to be directories. A regular file's descriptor is not an
// anchor for a name, so it is ENOTDIR rather than a dev/ino comparison that
// might accidentally match; an unusable descriptor is EBADF from fstat. The
// compare is on live identity, not on the names the descriptors were opened
// from, which is what makes two different descriptors on one directory work.
func sameDirectory(a, b int) (bool, error) {
	var sa, sb syscall.Stat_t
	if err := syscall.Fstat(a, &sa); err != nil {
		return false, err
	}
	if err := syscall.Fstat(b, &sb); err != nil {
		return false, err
	}
	if sa.Mode&syscall.S_IFMT != syscall.S_IFDIR || sb.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		return false, syscall.ENOTDIR
	}
	return sa.Dev == sb.Dev && sa.Ino == sb.Ino, nil
}

// Linkat is linkat(2) for this platform, with the flag policy of the approved
// compatibility mechanism and not the POSIX one.
//
// The flags are decided before any anchor is selected, any name is captured or
// anything is mutated:
//
//	flag 0                     legacy kernel link; it FOLLOWS a final source
//	                           symlink (bsd/vfs/vfs_syscalls.c:2605-2608)
//	exactly AT_SYMLINK_NOFOLLOW  ENOTSUP: this kernel has no llink(2) and no
//	                           anchored link that would not follow
//	anything else              EINVAL, including AT_SYMLINK_FOLLOW, unknown bits,
//	                           a negative value and NOFOLLOW combined with another
//	                           bit
//
// The zero-flag rule is the documented divergence: POSIX linkat(2) with no flags
// links the source INODE, so a source symlink produces a hard link to the symlink
// itself, and os.Root.Link relies on that (os/root.go's Link promise). Here the
// kernel links the referent, so a zero-flag Linkat through os.Root can hard-link
// a file that is outside the root through a symlink inside it, and a dangling
// source symlink reports ENOENT instead. Nothing in this backend pretends
// otherwise; the adopted tests assert both consequences explicitly.
func Linkat(olddirfd int, oldpath string, newdirfd int, newpath string, flag int) error {
	switch flag {
	case 0:
	case AT_SYMLINK_NOFOLLOW:
		return atUnsupported()
	default:
		return syscall.EINVAL
	}
	return atTwoNameDispatch(&atLinkOp, olddirfd, oldpath, newdirfd, newpath)
}

func Symlinkat(oldpath string, newdirfd int, newpath string) error {
	// Only the new name is resolved against the descriptor; oldpath is literal
	// link contents and never classified as absolute or relative here.
	if newdirfd == AT_FDCWD || isAbsPath(newpath) {
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
	if isAbsPath(path) {
		// An absolute name never consults the descriptor, so the path-based pair
		// is exact for it even though the descriptor is ignored, not validated.
		if flags == 0 {
			return syscall.Stat(path, stat)
		}
		return syscall.Lstat(path, stat)
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
