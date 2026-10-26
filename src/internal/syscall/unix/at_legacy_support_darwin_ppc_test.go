// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix_test

// Behavioral controls for the two approved divergences of this backend, and for
// the dispatch fixes that make an absolute name independent of its descriptor.
//
// The zero-flag link cases and the cross-directory rename cases are adopted from
// the MacPorts legacy-support test suite (reference/macports-legacy-support), and
// those parts of this file carry its notice:
//
// Copyright (c) 2026 Frederick H. G. Wright II <fw@fwright.net>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
//
// Each translated case names the C case it comes from, so the adopted assertion
// is discoverable from the Go test. What is NOT claimed here: that a zero-flag
// link is linkat(2)'s "do not follow" form (it follows, see
// TestLinkatFinalSourceSymlinkDarwinPPC), that a cross-directory destination is
// confined (its directory name is reconstructed), or that these tests are a
// security oracle for os.Root. Those are documented differences, not
// exemptions; every case below is required to pass.

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"internal/syscall/unix"
)

// atAbsTree builds the absolute-operand fixture: an "abs" directory (the target
// of every absolute name), a "cwd" directory that holds same-named decoys, and
// the process working directory set to the decoy so that a call which fell back
// to the cwd would be visible as a change to the wrong tree.
func atAbsTree(t *testing.T) (abs, cwd string) {
	t.Helper()
	root := t.TempDir()
	abs = filepath.Join(root, "abs")
	cwd = filepath.Join(root, "cwd")
	for _, dir := range []string{abs, cwd} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(cwd)
	return abs, cwd
}

// atDeadFD returns a descriptor number that is not open.
func atDeadFD(t *testing.T) int {
	t.Helper()
	f, err := os.Open(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fd := int(f.Fd())
	f.Close()
	return fd
}

// atWrite creates a file with contents and fails the test if it cannot.
func atWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

// atMode reports the permission bits of a path.
func atMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return fi.Mode().Perm()
}

// atIno reports the inode of a path.
func atIno(t *testing.T, path string) uint64 {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("%s: no Stat_t", path)
	}
	return st.Ino
}

// TestAtAbsoluteOperandsDarwinPPC requires every single-name operation to
// classify an absolute name before it looks at the descriptor. The descriptor
// passed here is not open at all, so an implementation that consulted it (or
// handed the name to an anchored helper) fails instead of quietly succeeding
// through some other route.
//
// The same test file is compiled into both gate builds: with the anchored backend
// enabled (the measured kernel) and with it forced off by
// goppc_atcompat_disabled. An absolute name reaches the root without any
// anchor, so both builds must behave identically here.
func TestAtAbsoluteOperandsDarwinPPC(t *testing.T) {
	abs, cwd := atAbsTree(t)
	dead := atDeadFD(t)

	t.Run("Openat", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "f"), "abs")
		atWrite(t, filepath.Join(cwd, "f"), "cwd")
		fd, err := unix.Openat(dead, filepath.Join(abs, "f"), syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatalf("Openat: %v", err)
		}
		b := make([]byte, 8)
		n, _ := syscall.Read(fd, b)
		syscall.Close(fd)
		if string(b[:n]) != "abs" {
			t.Errorf("read %q; want %q (the absolute name's file, not the cwd decoy)", b[:n], "abs")
		}
	})

	t.Run("OpenatCreate", func(t *testing.T) {
		fd, err := unix.Openat(dead, filepath.Join(abs, "created"), syscall.O_RDWR|syscall.O_CREAT, 0o600)
		if err != nil {
			t.Fatalf("Openat(O_CREAT): %v", err)
		}
		syscall.Close(fd)
		if _, err := os.Lstat(filepath.Join(abs, "created")); err != nil {
			t.Errorf("created file is not in the absolute directory: %v", err)
		}
		wantAbsent(t, filepath.Join(cwd, "created"))
	})

	t.Run("OpenatExclusive", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "excl"), "original")
		_, err := unix.Openat(dead, filepath.Join(abs, "excl"), syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL, 0o600)
		if !errors.Is(err, syscall.EEXIST) {
			t.Fatalf("Openat(O_CREAT|O_EXCL) on an existing file = %v; want EEXIST", err)
		}
		wantContent(t, filepath.Join(abs, "excl"), "original")
	})

	t.Run("Unlinkat", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "u"), "abs")
		atWrite(t, filepath.Join(cwd, "u"), "cwd")
		if err := unix.Unlinkat(dead, filepath.Join(abs, "u"), 0); err != nil {
			t.Fatalf("Unlinkat: %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "u"))
		wantContent(t, filepath.Join(cwd, "u"), "cwd")
		if err := unix.Unlinkat(dead, filepath.Join(abs, "u"), 0x1); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Unlinkat with an invalid flag = %v; want EINVAL", err)
		}
	})

	t.Run("Rmdir", func(t *testing.T) {
		if err := os.Mkdir(filepath.Join(abs, "d"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(cwd, "d"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := unix.Unlinkat(dead, filepath.Join(abs, "d"), unix.AT_REMOVEDIR); err != nil {
			t.Fatalf("Unlinkat(AT_REMOVEDIR): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "d"))
		if fi, err := os.Lstat(filepath.Join(cwd, "d")); err != nil || !fi.IsDir() {
			t.Errorf("cwd decoy directory: %v, %v; want it untouched", fi, err)
		}
	})

	t.Run("Readlinkat", func(t *testing.T) {
		if err := os.Symlink("literal-target", filepath.Join(abs, "s")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("cwd-target", filepath.Join(cwd, "s")); err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 64)
		n, err := unix.Readlinkat(dead, filepath.Join(abs, "s"), buf)
		if err != nil {
			t.Fatalf("Readlinkat: %v", err)
		}
		if string(buf[:n]) != "literal-target" {
			t.Errorf("Readlinkat = %q; want %q", buf[:n], "literal-target")
		}
		if b, err := os.Readlink(filepath.Join(cwd, "s")); err != nil || b != "cwd-target" {
			t.Errorf("cwd decoy symlink = %q, %v; want it untouched", b, err)
		}
	})

	t.Run("Mkdirat", func(t *testing.T) {
		if err := unix.Mkdirat(dead, filepath.Join(abs, "newdir"), 0o700); err != nil {
			t.Fatalf("Mkdirat: %v", err)
		}
		fi, err := os.Lstat(filepath.Join(abs, "newdir"))
		if err != nil || !fi.IsDir() {
			t.Errorf("Mkdirat result: %v, %v; want a directory", fi, err)
		}
		wantAbsent(t, filepath.Join(cwd, "newdir"))
	})

	t.Run("FstatatFollow", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "target"), "abs")
		if err := os.Symlink(filepath.Join(abs, "target"), filepath.Join(abs, "follow")); err != nil {
			t.Fatal(err)
		}
		var st syscall.Stat_t
		if err := unix.Fstatat(dead, filepath.Join(abs, "follow"), &st, 0); err != nil {
			t.Fatalf("Fstatat: %v", err)
		}
		if st.Mode&syscall.S_IFMT != syscall.S_IFREG {
			t.Errorf("Fstatat followed type = %#o; want a regular file", st.Mode&syscall.S_IFMT)
		}
		if st.Ino != atIno(t, filepath.Join(abs, "target")) {
			t.Errorf("Fstatat inode = %d; want the target's inode", st.Ino)
		}
	})

	t.Run("FstatatNofollow", func(t *testing.T) {
		if err := os.Symlink("nowhere", filepath.Join(abs, "dangling")); err != nil {
			t.Fatal(err)
		}
		var st syscall.Stat_t
		if err := unix.Fstatat(dead, filepath.Join(abs, "dangling"), &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fstatat(AT_SYMLINK_NOFOLLOW): %v", err)
		}
		if st.Mode&syscall.S_IFMT != syscall.S_IFLNK {
			t.Errorf("Fstatat no-follow type = %#o; want a symlink", st.Mode&syscall.S_IFMT)
		}
		if err := unix.Fstatat(dead, filepath.Join(abs, "dangling"), &st, 0x1); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Fstatat with an unknown flag = %v; want EINVAL", err)
		}
	})

	t.Run("FchownatFollow", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "own"), "abs")
		atWrite(t, filepath.Join(cwd, "own"), "cwd")
		uid, gid := os.Getuid(), os.Getgid()
		if err := unix.Fchownat(dead, filepath.Join(abs, "own"), uid, gid, 0); err != nil {
			t.Fatalf("Fchownat: %v", err)
		}
		wantContent(t, filepath.Join(abs, "own"), "abs")
		wantContent(t, filepath.Join(cwd, "own"), "cwd")
	})

	t.Run("FchownatNofollow", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "linktarget"), "abs")
		if err := os.Symlink(filepath.Join(abs, "linktarget"), filepath.Join(abs, "ownlink")); err != nil {
			t.Fatal(err)
		}
		linkIno := atIno(t, filepath.Join(abs, "ownlink"))
		uid, gid := os.Getuid(), os.Getgid()
		if err := unix.Fchownat(dead, filepath.Join(abs, "ownlink"), uid, gid, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchownat(AT_SYMLINK_NOFOLLOW): %v", err)
		}
		if atIno(t, filepath.Join(abs, "ownlink")) != linkIno {
			t.Errorf("Fchownat no-follow replaced the link instead of changing it")
		}
		wantContent(t, filepath.Join(abs, "linktarget"), "abs")
	})

	t.Run("Symlinkat", func(t *testing.T) {
		// Only the NEW name is classified: the first argument is literal link
		// contents and must not be resolved or rewritten here.
		if err := unix.Symlinkat("relative-literal", dead, filepath.Join(abs, "sym")); err != nil {
			t.Fatalf("Symlinkat: %v", err)
		}
		if b, err := os.Readlink(filepath.Join(abs, "sym")); err != nil || b != "relative-literal" {
			t.Errorf("Symlinkat contents = %q, %v; want %q", b, err, "relative-literal")
		}
		wantAbsent(t, filepath.Join(cwd, "sym"))
	})
}

// TestAtAbsoluteNofollowMetadataDarwinPPC requires the no-follow metadata forms
// for an absolute name to run the same setattrlist request without the private
// working directory.
//
// This is also the zero-syscall-349 control for that envelope, in the form this
// tree can measure it: the operations below run against a descriptor that is not
// open (so any anchored dispatch fails with EBADF) and, in the
// goppc_atcompat_disabled build, against a gate that refuses every anchored form
// before it can issue syscall 349. Success in both builds, plus the raw
// 349(-1)=EBADF check on the calling thread at the end, is what rules the
// private-cwd machinery out; the instrumentation-mutant variant of the same
// claim would need a counting mutant of the production routine, which this tree
// does not have and which no phase authorizes adding.
func TestAtAbsoluteNofollowMetadataDarwinPPC(t *testing.T) {
	abs, cwd := atAbsTree(t)
	dead := atDeadFD(t)
	timeAt := func(sec int32) syscall.Timespec { return syscall.Timespec{Sec: sec, Nsec: 0} }

	t.Run("ChmodRegular", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "m"), "abs")
		atWrite(t, filepath.Join(cwd, "m"), "cwd")
		if err := unix.Fchmodat(dead, filepath.Join(abs, "m"), 0o640, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchmodat(AT_SYMLINK_NOFOLLOW): %v", err)
		}
		if got := atMode(t, filepath.Join(abs, "m")); got != 0o640 {
			t.Errorf("mode = %#o; want %#o", got, 0o640)
		}
		if got := atMode(t, filepath.Join(cwd, "m")); got == 0o640 {
			t.Errorf("cwd decoy mode = %#o; want it untouched", got)
		}
	})

	t.Run("ChmodSymlink", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "linktarget"), "abs")
		if err := os.Symlink(filepath.Join(abs, "linktarget"), filepath.Join(abs, "modeLink")); err != nil {
			t.Fatal(err)
		}
		linkIno := atIno(t, filepath.Join(abs, "modeLink"))
		targetIno := atIno(t, filepath.Join(abs, "linktarget"))
		if err := unix.Fchmodat(dead, filepath.Join(abs, "modeLink"), 0o701, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchmodat(AT_SYMLINK_NOFOLLOW) on a symlink: %v", err)
		}
		if atMode(t, filepath.Join(abs, "modeLink")) != 0o701 {
			t.Errorf("link mode = %#o; want %#o (the link itself changed)", atMode(t, filepath.Join(abs, "modeLink")), 0o701)
		}
		if atIno(t, filepath.Join(abs, "modeLink")) != linkIno {
			t.Errorf("the symlink was replaced instead of changed")
		}
		if atIno(t, filepath.Join(abs, "target")) != targetIno {
			t.Errorf("the link target was replaced")
		}
		if atMode(t, filepath.Join(abs, "linktarget")) == 0o701 {
			t.Errorf("the link target's mode changed; the no-follow form acted on the referent")
		}
	})

	t.Run("BothTimes", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "t"), "abs")
		times := [2]syscall.Timespec{timeAt(1000000000), timeAt(1000000100)}
		if err := unix.Utimensat(dead, filepath.Join(abs, "t"), &times, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat: %v", err)
		}
		var st syscall.Stat_t
		if err := syscall.Lstat(filepath.Join(abs, "t"), &st); err != nil {
			t.Fatal(err)
		}
		if st.Atimespec.Sec != 1000000000 || st.Mtimespec.Sec != 1000000100 {
			t.Errorf("times = %d/%d; want 1000000000/1000000100", st.Atimespec.Sec, st.Mtimespec.Sec)
		}
	})

	t.Run("ModOnly", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "modonly"), "abs")
		before := [2]syscall.Timespec{timeAt(1000000000), timeAt(1000000000)}
		if err := unix.Utimensat(dead, filepath.Join(abs, "modonly"), &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(baseline): %v", err)
		}
		omit := syscall.Timespec{Sec: 0, Nsec: unix.UTIME_OMIT}
		times := [2]syscall.Timespec{omit, timeAt(1000000500)}
		if err := unix.Utimensat(dead, filepath.Join(abs, "modonly"), &times, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(mod-only): %v", err)
		}
		var st syscall.Stat_t
		if err := syscall.Lstat(filepath.Join(abs, "modonly"), &st); err != nil {
			t.Fatal(err)
		}
		if st.Mtimespec.Sec != 1000000500 {
			t.Errorf("mtime = %d; want 1000000500", st.Mtimespec.Sec)
		}
		if st.Atimespec.Sec != 1000000000 {
			t.Errorf("atime = %d; want the omitted value 1000000000 to be left alone", st.Atimespec.Sec)
		}
	})

	t.Run("BothOmitted", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "both"), "abs")
		base := [2]syscall.Timespec{timeAt(1000000000), timeAt(1000000000)}
		if err := unix.Utimensat(dead, filepath.Join(abs, "both"), &base, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(baseline): %v", err)
		}
		omit := syscall.Timespec{Sec: 0, Nsec: unix.UTIME_OMIT}
		times := [2]syscall.Timespec{omit, omit}
		if err := unix.Utimensat(dead, filepath.Join(abs, "both"), &times, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(both omitted): %v", err)
		}
		var st syscall.Stat_t
		if err := syscall.Lstat(filepath.Join(abs, "both"), &st); err != nil {
			t.Fatal(err)
		}
		if st.Atimespec.Sec != 1000000000 || st.Mtimespec.Sec != 1000000000 {
			t.Errorf("both-omitted request changed times: %d/%d", st.Atimespec.Sec, st.Mtimespec.Sec)
		}
	})

	t.Run("MissingBothOmitted", func(t *testing.T) {
		missing := filepath.Join(abs, "absent")
		omit := syscall.Timespec{Sec: 0, Nsec: unix.UTIME_OMIT}
		times := [2]syscall.Timespec{omit, omit}
		err := unix.Utimensat(dead, missing, &times, unix.AT_SYMLINK_NOFOLLOW)
		if !errors.Is(err, syscall.ENOENT) {
			t.Errorf("both-omitted on a missing name = %v; want ENOENT", err)
		}
		wantAbsent(t, filepath.Join(cwd, "absent"))
	})

	t.Run("NilRefusal", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "nilref"), "abs")
		if err := unix.Utimensat(dead, filepath.Join(abs, "nilref"), nil, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
			t.Errorf("Utimensat(NULL times) = %v; want ENOTSUP", err)
		}
		wantContent(t, filepath.Join(abs, "nilref"), "abs")
	})

	t.Run("NowRefusal", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "now"), "abs")
		now := syscall.Timespec{Sec: 0, Nsec: -1} // UTIME_NOW
		omit := syscall.Timespec{Sec: 0, Nsec: unix.UTIME_OMIT}
		for _, times := range [][2]syscall.Timespec{
			{now, now},
			{now, omit},
			{omit, now},
		} {
			if err := unix.Utimensat(dead, filepath.Join(abs, "now"), &times, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
				t.Errorf("Utimensat(%+v) = %v; want ENOTSUP", times, err)
			}
		}
	})

	t.Run("InvalidNsec", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "nsec"), "abs")
		bad := syscall.Timespec{Sec: 1000000000, Nsec: 1000000000}
		times := [2]syscall.Timespec{bad, bad}
		if err := unix.Utimensat(dead, filepath.Join(abs, "nsec"), &times, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Utimensat with an out-of-range nsec = %v; want EINVAL", err)
		}
		if err := unix.Fchmodat(dead, filepath.Join(abs, "nsec"), 0o640, 0x2); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Fchmodat with an unknown flag = %v; want EINVAL", err)
		}
	})

	// The calling thread must not be holding a private working directory after
	// any of the requests above: raw 349(-1) reports EBADF when there is none to
	// clear, and success when the thread does hold one.
	t.Run("CallerThreadClean", func(t *testing.T) {
		if err := unix.AtPthreadFchdirClearForTest(); !errors.Is(err, syscall.EBADF) {
			t.Errorf("raw 349(-1) on the calling thread = %v; want EBADF (no private cwd was installed)", err)
		}
	})
}

// TestAtAbsoluteAtimeOnlyDeferredDarwinPPC pins the one absolute metadata form
// that is NOT converted to the no-anchor envelope.
//
// An access-time-only request needs the HFS workaround's guard: the filesystem of
// the anchor, the anchor's own dev, and the target's dev (see
// atChgTimeForAccessTime). There is no anchor in the no-anchor envelope, and a
// guard equivalent to the approved one has not been demonstrated, so the
// anchored dispatch is kept. The assertion below is what makes that visible: with
// the gate on the request is served, and with the gate forced off
// (goppc_atcompat_disabled) the same request is refused with ENOTSUP, which can
// only happen while it still goes through a gated anchored path.
func TestAtAbsoluteAtimeOnlyDeferredDarwinPPC(t *testing.T) {
	abs, _ := atAbsTree(t)
	dead := atDeadFD(t)
	atWrite(t, filepath.Join(abs, "atime"), "abs")
	base := [2]syscall.Timespec{{Sec: 1000000000}, {Sec: 1000000000}}
	if err := unix.Utimensat(dead, filepath.Join(abs, "atime"), &base, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatalf("Utimensat(baseline): %v", err)
	}
	times := [2]syscall.Timespec{{Sec: 1000000700}, {Sec: 0, Nsec: unix.UTIME_OMIT}}
	err := unix.Utimensat(dead, filepath.Join(abs, "atime"), &times, unix.AT_SYMLINK_NOFOLLOW)
	if unix.AtCompatEnabledForTest() {
		if err != nil {
			t.Fatalf("access-time-only request with the anchored backend enabled: %v; want it served by the anchored dispatch", err)
		}
		var st syscall.Stat_t
		if err := syscall.Lstat(filepath.Join(abs, "atime"), &st); err != nil {
			t.Fatal(err)
		}
		if st.Mtimespec.Sec != 1000000000 {
			t.Errorf("mtime = %d; want the omitted time to be left alone", st.Mtimespec.Sec)
		}
		return
	}
	if !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("access-time-only request with the anchored backend disabled = %v; want ENOTSUP (it still uses the anchored dispatch)", err)
	}
}

// TestRenameatOperandsDarwinPPC covers the rename dispatch rows that need no
// destination reconstruction: every combination where an absolute name or a
// single directory object decides the anchor.
func TestRenameatOperandsDarwinPPC(t *testing.T) {
	abs, cwd := atAbsTree(t)
	dead := atDeadFD(t)

	t.Run("AbsAbsIgnoredFDs", func(t *testing.T) {
		atWrite(t, filepath.Join(abs, "src"), "abs")
		ino := atIno(t, filepath.Join(abs, "src"))
		if err := unix.Renameat(dead, filepath.Join(abs, "src"), dead, filepath.Join(abs, "dst")); err != nil {
			t.Fatalf("Renameat(abs, abs): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "src"))
		if got := atIno(t, filepath.Join(abs, "dst")); got != ino {
			t.Errorf("destination inode = %d; want the moved inode %d", got, ino)
		}
	})

	t.Run("AbsRelIgnoredOldFD", func(t *testing.T) {
		anchor, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(anchor)
		atWrite(t, filepath.Join(abs, "movesrc"), "abs")
		ino := atIno(t, filepath.Join(abs, "movesrc"))
		if err := unix.Renameat(dead, filepath.Join(abs, "movesrc"), anchor, "movedst"); err != nil {
			t.Fatalf("Renameat(abs, rel): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "movesrc"))
		if got := atIno(t, filepath.Join(abs, "movedst")); got != ino {
			t.Errorf("destination inode = %d; want %d", got, ino)
		}
	})

	t.Run("RelAbsIgnoredNewFD", func(t *testing.T) {
		anchor, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(anchor)
		atWrite(t, filepath.Join(abs, "relsrc"), "abs")
		ino := atIno(t, filepath.Join(abs, "relsrc"))
		if err := unix.Renameat(anchor, "relsrc", dead, filepath.Join(abs, "reldst")); err != nil {
			t.Fatalf("Renameat(rel, abs): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "relsrc"))
		if got := atIno(t, filepath.Join(abs, "reldst")); got != ino {
			t.Errorf("destination inode = %d; want %d", got, ino)
		}
	})

	t.Run("AbsWithIrrelevantFDCWD", func(t *testing.T) {
		// The absolute operand ignores the other side's AT_FDCWD; no mixed-case
		// refusal may fire.
		atWrite(t, filepath.Join(cwd, "cwdsrc"), "cwd")
		if err := unix.Renameat(unix.AT_FDCWD, "cwdsrc", unix.AT_FDCWD, filepath.Join(abs, "fromcwd")); err != nil {
			t.Fatalf("Renameat(relative AT_FDCWD, absolute): %v", err)
		}
		wantAbsent(t, filepath.Join(cwd, "cwdsrc"))
		wantContent(t, filepath.Join(abs, "fromcwd"), "cwd")
		atWrite(t, filepath.Join(abs, "abs2"), "abs")
		if err := unix.Renameat(unix.AT_FDCWD, filepath.Join(abs, "abs2"), unix.AT_FDCWD, "tocwd"); err != nil {
			t.Fatalf("Renameat(absolute, relative AT_FDCWD): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "abs2"))
		wantContent(t, filepath.Join(cwd, "tocwd"), "abs")
	})

	t.Run("AbsAbsDisabled", func(t *testing.T) {
		// Two absolute names consult no descriptor and no kernel interface, so
		// this works in the enabled and in the disabled build alike.
		atWrite(t, filepath.Join(abs, "anygate"), "abs")
		if err := unix.Renameat(dead, filepath.Join(abs, "anygate"), dead, filepath.Join(abs, "anygatedst")); err != nil {
			t.Fatalf("Renameat(abs, abs): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "anygate"))
		wantContent(t, filepath.Join(abs, "anygatedst"), "abs")
	})

	t.Run("SameFD", func(t *testing.T) {
		anchor, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(anchor)
		atWrite(t, filepath.Join(abs, "samefd"), "abs")
		ino := atIno(t, filepath.Join(abs, "samefd"))
		if err := unix.Renameat(anchor, "samefd", anchor, "samefddst"); err != nil {
			t.Fatalf("Renameat(same fd): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "samefd"))
		if got := atIno(t, filepath.Join(abs, "samefddst")); got != ino {
			t.Errorf("destination inode = %d; want %d", got, ino)
		}
	})

	t.Run("TwoFDsSameDirectory", func(t *testing.T) {
		a, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(a)
		b, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(b)
		if a == b {
			t.Fatalf("two Open calls returned the same descriptor %d", a)
		}
		atWrite(t, filepath.Join(abs, "twofd"), "abs")
		if err := unix.Renameat(a, "twofd", b, "twofddst"); err != nil {
			t.Fatalf("Renameat(two fds on one directory): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "twofd"))
		wantContent(t, filepath.Join(abs, "twofddst"), "abs")
	})

	t.Run("TwoFDsSameDirectoryAfterRename", func(t *testing.T) {
		// The identity that decides "one anchor" is the live directory object, not
		// the path the descriptors were opened from: renaming the directory under
		// the descriptors leaves both of them pointing at the same object.
		dir := filepath.Join(abs, "moved")
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		a, err := syscall.Open(dir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(a)
		b, err := syscall.Open(dir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(b)
		if err := os.Rename(dir, filepath.Join(abs, "moved2")); err != nil {
			t.Fatal(err)
		}
		atWrite(t, filepath.Join(abs, "moved2", "x"), "abs")
		if err := unix.Renameat(a, "x", b, "y"); err != nil {
			t.Fatalf("Renameat(two fds on one renamed directory): %v", err)
		}
		wantAbsent(t, filepath.Join(abs, "moved2", "x"))
		wantContent(t, filepath.Join(abs, "moved2", "y"), "abs")
	})

	t.Run("NonDirectory", func(t *testing.T) {
		file := filepath.Join(abs, "notadir")
		atWrite(t, file, "abs")
		f, err := os.OpenFile(file, os.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		fd := int(f.Fd())
		defer f.Close()
		if err := unix.Renameat(fd, "x", fd, "y"); !errors.Is(err, syscall.ENOTDIR) {
			t.Errorf("Renameat with a regular-file descriptor = %v; want ENOTDIR", err)
		}
	})

	t.Run("InvalidRelevantFD", func(t *testing.T) {
		anchor, err := syscall.Open(abs, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(anchor)
		atWrite(t, filepath.Join(abs, "inv"), "abs")
		if err := unix.Renameat(dead, "inv", anchor, "invdst"); !errors.Is(err, syscall.EBADF) {
			t.Errorf("Renameat with an unopened relevant descriptor = %v; want EBADF", err)
		}
		wantContent(t, filepath.Join(abs, "inv"), "abs")
		wantAbsent(t, filepath.Join(abs, "invdst"))
	})
}

// TestRenameatDestinationAnchorDarwinPPC describes what the reconstructed
// destination anchor is and is not.
//
// F_GETPATH describes the directory OBJECT the caller anchored, so a directory
// that has been renamed since it was opened is still found, and a path that has
// been replaced by a different directory does not capture the operation. That is
// the property that makes the reconstruction usable. What it does NOT give is a
// held anchor for the duration of the syscall: the reconstructed name is looked up
// again, so a component of it can change between the reconstruction and the rename.
// There is no maintained seam for that interleaving, so it is described here rather
// than pretended to be tested; the exit criteria record it as a documented
// difference, not as a passing security oracle.
func TestRenameatDestinationAnchorDarwinPPC(t *testing.T) {
	t.Run("MovedDirectoryIsFollowed", func(t *testing.T) {
		root := t.TempDir()
		srcDir := filepath.Join(root, "src")
		dstDir := filepath.Join(root, "dst")
		movedDir := filepath.Join(root, "moved")
		for _, dir := range []string{srcDir, dstDir, movedDir} {
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		srcFd, err := syscall.Open(srcDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(srcFd)
		dstFd, err := syscall.Open(dstDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(dstFd)
		// Move the destination directory out of the name it was opened at.
		if err := os.Rename(dstDir, movedDir); err != nil {
			t.Fatal(err)
		}
		atWrite(t, filepath.Join(srcDir, "src"), "src")
		if err := unix.Renameat(srcFd, "src", dstFd, "dst"); err != nil {
			t.Fatalf("Renameat into a moved destination directory: %v", err)
		}
		wantContent(t, filepath.Join(movedDir, "dst"), "src")
	})

	t.Run("ReplacedPathDoesNotCapture", func(t *testing.T) {
		root := t.TempDir()
		srcDir := filepath.Join(root, "src")
		dstDir := filepath.Join(root, "dst")
		aside := filepath.Join(root, "aside")
		for _, dir := range []string{srcDir, dstDir, aside} {
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		srcFd, err := syscall.Open(srcDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(srcFd)
		dstFd, err := syscall.Open(dstDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(dstFd)
		// Keep the anchored object, but put a different directory at its old path.
		if err := os.Rename(dstDir, aside); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(dstDir, 0o700); err != nil {
			t.Fatal(err)
		}
		atWrite(t, filepath.Join(srcDir, "src"), "src")
		if err := unix.Renameat(srcFd, "src", dstFd, "dst"); err != nil {
			t.Fatalf("Renameat into a moved destination directory: %v", err)
		}
		wantContent(t, filepath.Join(aside, "dst"), "src")
		wantAbsent(t, filepath.Join(dstDir, "dst"))
	})
}
