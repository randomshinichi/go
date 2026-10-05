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

// lutimes and lchmod act on a symbolic link itself instead of its target.
// Both are defined in Leopard's libSystem (nm -g, D10 launch-gate data
// leopard-libsystem-exports.txt: "T _lutimes", "T _lchmod"), and both were
// measured on the G5 to change the link and leave the target alone
// (out/stage2-20261003/parent-native-run/BREADTH-AT-NATIVE-RUNG.md items 2 and 3).
// They exist only to serve AT_SYMLINK_NOFOLLOW in utimensat here and in
// internal/syscall/unix's Fchmodat; neither is exported from package syscall.
// lchmod and lstatatByPath are reached from internal/syscall/unix by linkname,
// so each is marked here as used by it (the linker refuses the reference
// otherwise); linkname_darwin.go does the same for fstatat but is shared with
// darwin targets that do not define these.

//sys	lutimes(path string, timeval *[2]Timeval) (err error)
//sys	lchmod(path string, mode uint32) (err error)

// used by internal/syscall/unix
//
//go:linkname lchmod
//go:linkname lstatatByPath

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

// lstatatByPath is the one *at operation that is implemented for a real
// directory descriptor, and it is implemented by resolving a path: it asks the
// kernel for the directory's current path (fcntl F_GETPATH, which Leopard
// has: measured on the G5, fcntl(dirfd, 50, buf) returned 0 with
// path='/private/tmp/utp', and lstat(dirpath + "/alink") returned the symbolic
// link itself, mode 0120755; out/stage2-20261003/parent-native-run/
// BREADTH-AT-NATIVE-RUNG.md item 1, probe-readdir/fgetpath.c) and then calls
// lstat on dirpath + "/" + name. It exists for (*os.File).Readdir, which
// stats every entry of a directory it has open, and nothing else.
//
// What that costs: this is two steps on a path (look the path up, then use it),
// not one operation on the descriptor, so it loses the descriptor-based call's
// symlink-race (TOCTOU) protection. If the directory is renamed, removed, or
// replaced (by a symbolic link, say) between the F_GETPATH and the lstat, the
// lstat describes whatever is at the old path now; the caller gets metadata
// for the wrong file, never a write to it. That is the same trade already
// accepted and documented for os.RemoveAll on this platform
// (os/removeall_darwin_ppc64.go), and it is why no
// operation that writes (chmod, chown, unlink, rename, ...) is given this form:
// they keep returning ENOTSUP for a real descriptor (internal/syscall/unix,
// at_darwin_ppc64.go), as does os.Root. The path must also fit in the
// MAXPATHLEN bytes F_GETPATH writes and in lstat's limit, so a very deep
// directory fails with ENAMETOOLONG here where the descriptor version would
// not.
//
// name must be a single directory-entry name: an empty name or one containing
// '/' is refused, so the call cannot be steered out of the directory. With
// _AT_FDCWD (a chosen sentinel, absent on Leopard) it is plain lstat.
func lstatatByPath(dirfd int, name string, stat *Stat_t) error {
	if dirfd == _AT_FDCWD {
		return Lstat(name, stat)
	}
	if name == "" {
		return ENOENT
	}
	for i := 0; i < len(name); i++ {
		if name[i] == '/' {
			return EINVAL
		}
	}
	const maxPathLen = 1024 // MAXPATHLEN, the buffer size F_GETPATH requires
	var buf [maxPathLen]byte
	if _, err := fcntlPtr(dirfd, F_GETPATH, unsafe.Pointer(&buf[0])); err != nil {
		return err
	}
	n := clen(buf[:])
	if n == 0 {
		return ENOENT
	}
	return Lstat(string(buf[:n])+"/"+name, stat)
}

// utimensat is implemented without libc's utimensat: Leopard does not have it
// (dlprobe reports utimensat and futimens MISSING; the *at family is absent
// altogether), so shared callers (UtimesNano, internal/syscall/unix.Utimensat)
// would otherwise import a symbol dyld cannot resolve. What Leopard does have
// is utimes (path, follows symlinks), futimes and lutimes (path, acts on the
// link itself), all in libSystem's exports; none takes a nanosecond or a
// "leave this one alone" value.
//
// The mapping, and what it does not cover:
//   - dirfd must be _AT_FDCWD (a chosen sentinel, absent on Leopard). Any real
//     directory descriptor returns ENOTSUP, never a fall-back to the
//     path-based call: that would act on the wrong file.
//   - flags 0 is utimes. Undefined bits are EINVAL. AT_SYMLINK_NOFOLLOW is
//     lutimes, and a UTIME_OMIT time is then read with lstat instead of stat.
//     The lutimes case rests on a measurement on the G5 (parent-native-run/
//     BREADTH-AT-NATIVE-RUNG.md item 2): with explicit times it changed the
//     link's modification time and left the target's untouched.
//   - times == nil, or both entries UTIME_NOW, is utimes(path, NULL). With
//     AT_SYMLINK_NOFOLLOW it is ENOTSUP: lutimes(path, NULL) was never
//     measured, and "now" cannot be replaced by an explicit current time
//     because utimes needs ownership for an explicit time but only write
//     access for "now".
//   - A UTIME_OMIT entry keeps that time by reading it with stat (lstat when
//     not following) and writing it back with the other one (the technique
//     fs_wasip1.go's UtimesNano uses). That is a read-modify-write: another
//     process changing the omitted time between the stat and the utimes has
//     its change overwritten. If both are UTIME_OMIT nothing is written and
//     only the stat's error is reported.
//   - UTIME_NOW mixed with any other value is ENOTSUP, for the same reason.
//     Nothing in the standard library does this.
//   - Precision: utimes takes microseconds. Explicit times are rounded up to a
//     microsecond by NsecToTimeval (as UtimesNano's ENOSYS fallback already
//     does on the other BSDs), and a time kept via UTIME_OMIT is passed through
//     the same conversion. The file system on the G5 keeps whole seconds only
//     (measured: utimes round-tripped mtime_nsec = 0, BREADTH-AT-NATIVE-RUNG.md
//     item 4), so on that machine the microsecond limit and the
//     stat-then-utimes round trip cost no precision the file system had.
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
	var tv [2]Timeval
	if omit0 || omit1 {
		var st Stat_t
		if err := statPath(path, &st); err != nil {
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
	return setTimes(path, &tv)
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

// Exec cannot work on darwin/ppc64 Leopard, and this is a platform limit, not a
// port defect. Leopard's kernel refuses execve (and posix_spawn with
// POSIX_SPAWN_SETEXEC) from any process that has more than one thread, with
// errno 45 (ENOTSUP), before it looks at the path: execve of a path that does
// not exist also returns 45. A Go process always has several threads, so
// syscall.Exec always returns ENOTSUP here, from libc execve and from the raw
// SYS_EXECVE alike. Measured on the G5 (out/stage2-20261003/D16-statfs-exec/
// native-g5/out-d16: c-e1 count=1 succeeds; c-e-mt-main, c-e-mt4-main,
// c-e-mt-helper, c-e-mt-enoent, c-e-spawn-mt return 45; go-rawexec 45;
// go-threads shows 5 threads). fork followed by execve in the child works
// (c-e-mt-fork), which is what ForkExec, StartProcess and os/exec use.
// TestExec fails here for this reason and is not weakened.

// getfsstat64 is the libc entry point that fills the 64-bit-inode Statfs_t
// (struct statfs64, 2168 bytes). Leopard's plain getfsstat, which the shared
// Getfsstat (syscall_darwin_getfsstat.go) imports, fills the legacy 336-byte
// struct statfs instead (sys/mount.h, !__DARWIN_64_BIT_INO_T); read through a
// Statfs_t slice that is a wrong stride, so only entry 0 is partly meaningful and
// the rest is zero. getfsstat64 is exported by Leopard's libSystem
// (out/stage2-20261003/D7-syscall/g5-request/out-d7/08-dlprobe.txt).
//
//sys	getfsstat64(buf unsafe.Pointer, size uintptr, flags int) (n int, err error)

// Getfsstat on darwin/ppc64 reads the mount table through getfsstat64 so that
// the records match Statfs_t, as Statfs and Fstatfs already do (SYS_statfs64,
// SYS_fstatfs64 above).
func Getfsstat(buf []Statfs_t, flags int) (n int, err error) {
	var _p0 unsafe.Pointer
	var bufsize uintptr
	if len(buf) > 0 {
		_p0 = unsafe.Pointer(&buf[0])
		bufsize = unsafe.Sizeof(Statfs_t{}) * uintptr(len(buf))
	}
	return getfsstat64(_p0, bufsize, flags)
}

// readdir_r reads one directory entry through libc's readdir_r$INODE64. It is
// exported through //go:linkname (linkname_darwin.go) but nothing in this tree
// reaches it on darwin/ppc64: Getdirentries below is raw SYS_getdirentries64.
//
// The shared //sys declaration imports plain readdir_r. On Leopard that is the
// legacy function: it fills the 32-bit-inode record (reclen 12 for ".",
// sizeof(struct dirent) 264; out/stage2-20261003/D7-syscall/g5-request2/
// out-d7r2/C-dirent-test.txt) while Dirent describes the 64-bit-inode record
// (1048 bytes). A caller using it would read garbage and not see an error.
// readdir_r$INODE64 is exported by Leopard's libSystem (parent measurement,
// D16, nm -g and dlsym); mksyscall.pl cannot spell the $ in the import name, so
// the wrapper and its cgo_import_dynamic are written by hand here, and
// mkasm.go finds the trampoline declaration in this file.
func readdir_r(dir uintptr, entry *Dirent, result **Dirent) (res Errno) {
	r0, _, _ := syscall(abi.FuncPCABI0(libc_readdir_r_inode64_trampoline), uintptr(dir), uintptr(unsafe.Pointer(entry)), uintptr(unsafe.Pointer(result)))
	res = Errno(r0)
	return
}

func libc_readdir_r_inode64_trampoline()

//go:cgo_import_dynamic libc_readdir_r_inode64 readdir_r$INODE64 "/usr/lib/libSystem.B.dylib"

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
