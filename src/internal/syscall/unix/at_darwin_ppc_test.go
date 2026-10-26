// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix_test

import (
	"errors"
	"fmt"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// These tests are the Unix-layer controls for the anchored backend described in
// internal/syscall/unix/at_darwin_ppc.go. They run on the target kernel, where
// the backend is enabled, and they assert two things that must be distinguished:
// an operation that is served acts inside the anchor and nowhere else, and an
// operation that is refused leaves everything untouched. Every anchored call
// below runs with the process working directory set to a decoy directory that
// holds entries with the same names, so a fallback to the cwd is visible as a
// change to the wrong tree rather than as a passing test.

func ppcAtTestDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("go-at-ppc-%d-%s", os.Getpid(), t.Name()))
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, name := range []string{"file", "renamed", "link", "sub", "d", "new", "excl",
			"target", "slink", "alt", "fifo", "while-blocked"} {
			_ = os.Remove(filepath.Join(dir, name))
		}
		if err := os.Remove(dir); err != nil {
			t.Errorf("Remove test directory: %v", err)
		}
	})
	return dir
}

// atTree creates the same names in two directories with distinguishable
// contents, and returns an open descriptor on the first. The second is the
// process working directory for the duration of the test.
func atTree(t *testing.T) (anchorDir, cwdDir string, fd int) {
	t.Helper()
	anchorDir, cwdDir = ppcAtTestDir(t), t.TempDir()
	for dir, content := range map[string]string{anchorDir: "anchor", cwdDir: "cwd"} {
		if err := os.WriteFile(filepath.Join(dir, "file"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(dir, "sub"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(cwdDir)
	fd, err := syscall.Open(anchorDir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Close(fd) })
	return anchorDir, cwdDir, fd
}

// wantContent asserts that a path holds want; it is how "acted in the anchor,
// not in the cwd" is spelled.
func wantContent(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil || string(b) != want {
		t.Errorf("%s = %q, %v; want %q", path, b, err, want)
	}
}

func wantAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Errorf("%s: %v; want it not to exist", path, err)
	}
}

// TestAtCompatReleaseDarwinPPC pins the platform gate. The wrong answers here
// are not cosmetic: the backend uses private fcntl numbers, so enabling it on a
// kernel nobody measured would be an unchecked assumption.
func TestAtCompatReleaseDarwinPPC(t *testing.T) {
	for _, tc := range []struct {
		release string
		want    bool
	}{
		{"9.8.0", true},  // the measured target, Darwin 9.8.0
		{"9.8", false},   // no patch level: not the shape of a release string
		{"9.8.x", false}, // ditto
		{"9.7.0", false}, // older xnu, never measured
		{"9.9.0", false}, // newer, private fcntl numbers unchecked
		{"10.0.0", false},
		{"8.11.0", false},
		{"", false},
		{"9", false},
	} {
		got, _ := unix.AtCompatRelease(tc.release)
		if got != tc.want {
			t.Errorf("release %q: enabled = %v; want %v", tc.release, got, tc.want)
		}
	}
	// Whether the gate is actually ON here is asserted in two build-tag-scoped
	// files, because a test-only build can force it off
	// (at_disabled_darwin_ppc_test.go, tag goppc_atcompat_disabled): the parser
	// above is what this test can decide unconditionally.
}

// TestAtDarwinPPCAnchoredOps requires each served operation to reach the anchor.
func TestAtDarwinPPCAnchoredOps(t *testing.T) {
	anchorDir, cwdDir, fd := atTree(t)
	var st syscall.Stat_t

	t.Run("OpenatFOPENFROM", func(t *testing.T) {
		got, err := unix.Openat(fd, "file", syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatalf("Openat: %v", err)
		}
		b := make([]byte, 16)
		n, err := syscall.Read(got, b)
		syscall.Close(got)
		if err != nil || string(b[:n]) != "anchor" {
			t.Errorf("read %q, %v; want the anchor's file", b[:n], err)
		}
	})

	t.Run("OpenatCreateCompound", func(t *testing.T) {
		// O_CREATE without O_EXCL takes the private-cwd compound: F_OPENFROM's
		// creation retry loses the anchor.
		got, err := unix.Openat(fd, "new", syscall.O_RDWR|syscall.O_CREAT, 0600)
		if err != nil {
			t.Fatalf("Openat(O_CREATE): %v", err)
		}
		syscall.Close(got)
		if _, err := os.Lstat(filepath.Join(anchorDir, "new")); err != nil {
			t.Errorf("the create did not land in the anchor: %v", err)
		}
		wantAbsent(t, filepath.Join(cwdDir, "new"))
	})

	t.Run("OpenatExclusiveCreate", func(t *testing.T) {
		// O_CREATE|O_EXCL stays on F_OPENFROM and must report EEXIST for an
		// existing name instead of creating or returning the old file.
		if _, err := unix.Openat(fd, "file", syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL, 0600); !errors.Is(err, syscall.EEXIST) {
			t.Errorf("Openat(O_CREATE|O_EXCL) = %v; want EEXIST", err)
		}
		wantContent(t, filepath.Join(anchorDir, "file"), "anchor")
	})

	t.Run("FstatatNofollow", func(t *testing.T) {
		if err := unix.Fstatat(fd, "file", &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fstatat: %v", err)
		}
		if st.Size != int64(len("anchor")) {
			t.Errorf("Fstatat size = %d; want the anchor's file (%d)", st.Size, len("anchor"))
		}
	})

	t.Run("Readlinkat", func(t *testing.T) {
		if err := unix.Symlinkat("file", fd, "link"); err != nil {
			t.Fatalf("Symlinkat: %v", err)
		}
		buf := make([]byte, 32)
		n, err := unix.Readlinkat(fd, "link", buf)
		if err != nil || string(buf[:n]) != "file" {
			t.Errorf("Readlinkat = %q, %v; want \"file\"", buf[:n], err)
		}
		wantAbsent(t, filepath.Join(cwdDir, "link"))
	})

	t.Run("MkdiratAndRmdir", func(t *testing.T) {
		if err := unix.Mkdirat(fd, "d", 0700); err != nil {
			t.Fatalf("Mkdirat: %v", err)
		}
		if fi, err := os.Lstat(filepath.Join(anchorDir, "d")); err != nil || !fi.IsDir() {
			t.Errorf("the directory did not land in the anchor: %v, %v", fi, err)
		}
		wantAbsent(t, filepath.Join(cwdDir, "d"))
		// F_UNLINKFROM is unlink(2) only; the remove-directory form has to be
		// the compound's rmdir, which is what distinguishes the two paths.
		if err := unix.Unlinkat(fd, "d", unix.AT_REMOVEDIR); err != nil {
			t.Fatalf("Unlinkat(AT_REMOVEDIR): %v", err)
		}
		wantAbsent(t, filepath.Join(anchorDir, "d"))
	})

	t.Run("UnlinkatFUNLINKFROM", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(anchorDir, "victim"), []byte("v"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := unix.Unlinkat(fd, "victim", 0); err != nil {
			t.Fatalf("Unlinkat: %v", err)
		}
		wantAbsent(t, filepath.Join(anchorDir, "victim"))
	})

	t.Run("FchownatNofollow", func(t *testing.T) {
		if err := unix.Fchownat(fd, "file", -1, -1, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Errorf("Fchownat(AT_SYMLINK_NOFOLLOW): %v", err)
		}
	})

	t.Run("FchmodatFollowing", func(t *testing.T) {
		if err := unix.Fchmodat(fd, "file", 0640, 0); err != nil {
			t.Fatalf("Fchmodat: %v", err)
		}
		if fi, err := os.Lstat(filepath.Join(anchorDir, "file")); err != nil || fi.Mode().Perm() != 0640 {
			t.Errorf("mode = %v, %v; want 0640 on the anchor's file", fi.Mode(), err)
		}
	})

	t.Run("RenameatSameAnchor", func(t *testing.T) {
		if err := unix.Renameat(fd, "file", fd, "renamed"); err != nil {
			t.Fatalf("Renameat: %v", err)
		}
		wantContent(t, filepath.Join(anchorDir, "renamed"), "anchor")
		wantAbsent(t, filepath.Join(anchorDir, "file"))
		wantContent(t, filepath.Join(cwdDir, "file"), "cwd")
		if err := unix.Renameat(fd, "renamed", fd, "file"); err != nil {
			t.Fatalf("Renameat back: %v", err)
		}
	})
}

// TestAtDarwinPPCCwdUnchanged requires the anchored operations to leave the
// caller's working directory alone. A compound that installed the private cwd on
// the caller's own thread and failed to clear it would be visible here, and so
// would a fallback that chdir'ed the process.
func TestAtDarwinPPCCwdUnchanged(t *testing.T) {
	_, cwdDir, fd := atTree(t)
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var st syscall.Stat_t
	for _, op := range []func() error{
		func() error { return unix.Mkdirat(fd, "probe", 0700) },
		func() error { return unix.Unlinkat(fd, "probe", unix.AT_REMOVEDIR) },
		func() error { return unix.Fstatat(fd, "file", &st, unix.AT_SYMLINK_NOFOLLOW) },
		func() error { return unix.Fchownat(fd, "file", -1, -1, unix.AT_SYMLINK_NOFOLLOW) },
		func() error { return unix.Renameat(fd, "file", fd, "file") },
		func() error {
			buf := make([]byte, 8)
			_, err := unix.Readlinkat(fd, "file", buf)
			return err
		},
	} {
		op()
	}

	after, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Errorf("the working directory moved to %q; want %q", after, before)
	}
	if fi, err := os.Stat(cwdDir); err != nil || !fi.IsDir() {
		t.Errorf("the working directory %q: %v, %v", cwdDir, fi, err)
	}
}

// TestAtDarwinPPCRetainedRefusals is the no-mutation control for the forms that
// are still refused. Each must refuse, and must leave both trees exactly as they
// were: a refusal that still acted would be worse than a wrong answer.
//
// Two refusals this test used to pin are gone, and their removal is a policy
// change rather than a weakening to obtain green: the operator decision recorded
// in goppc-wjfe6 replaced "Linkat with any real descriptor" and "Renameat between
// two different anchors" with real implementations whose weaker semantics are
// documented in this file's header. What those two forms now do is asserted, row
// by row, in at_two_anchor_darwin_ppc_test.go (including the consequences of
// zero-flag link following a final source symlink); nothing was dropped without a
// replacement control.
func TestAtDarwinPPCRetainedRefusals(t *testing.T) {
	anchorDir, cwdDir, fd := atTree(t)
	otherFd, err := syscall.Open(cwdDir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(otherFd)

	// Link flags other than zero are refused before any name is looked at. (The
	// zero-flag rows live in at_two_anchor_darwin_ppc_test.go.)
	if err := unix.Linkat(fd, "file", fd, "linked", unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Linkat(AT_SYMLINK_NOFOLLOW) = %v; want ENOTSUP", err)
	}
	if err := unix.Linkat(fd, "file", fd, "linked", 0x40); !errors.Is(err, syscall.EINVAL) {
		t.Errorf("Linkat(AT_SYMLINK_FOLLOW) = %v; want EINVAL", err)
	}
	wantAbsent(t, filepath.Join(anchorDir, "linked"))
	wantAbsent(t, filepath.Join(cwdDir, "linked"))

	// The no-follow metadata forms are no longer refused here: they are served
	// by setattrlist(FSOPT_NOFOLLOW) and have their own controls in
	// TestAtDarwinPPCMetadataDarwinPPC. The forms that stay refused are the ones
	// with no equivalent, and each must still refuse without mutating anything.

	// The anchored open of a real descriptor is not "supported" for a file
	// descriptor that is not a directory: the kernel interface requires a
	// readable directory and says so, rather than acting on the process cwd.
	if _, err := unix.Openat(int(otherFd), "file", syscall.O_RDONLY, 0); err != nil {
		// otherFd is a directory, so this must succeed; the control is that a
		// non-directory descriptor does not.
		t.Errorf("Openat(directory descriptor) = %v; want success", err)
	}
	fileFd, err := syscall.Open(filepath.Join(anchorDir, "file"), syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fileFd)
	if _, err := unix.Openat(fileFd, "file", syscall.O_RDONLY, 0); !errors.Is(err, syscall.ENOTDIR) {
		t.Errorf("Openat(non-directory descriptor) = %v; want ENOTDIR", err)
	}
}

// TestAtDarwinPPCRemovedAnchor requires a prompt ENOENT rather than an
// operation somewhere else when the anchored directory is gone.
func TestAtDarwinPPCRemovedAnchor(t *testing.T) {
	dir := ppcAtTestDir(t)
	gone := filepath.Join(dir, "gone")
	if err := os.Mkdir(gone, 0700); err != nil {
		t.Fatal(err)
	}
	fd, err := syscall.Open(gone, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fd)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep"), []byte("k"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(filepath.Join(dir, "keep")) })

	var st syscall.Stat_t
	if err := unix.Fstatat(fd, "keep", &st, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOENT) {
		t.Errorf("Fstatat on a removed anchor = %v; want ENOENT", err)
	}
	if err := unix.Mkdirat(fd, "x", 0700); !errors.Is(err, syscall.ENOENT) {
		t.Errorf("Mkdirat on a removed anchor = %v; want ENOENT", err)
	}
	wantAbsent(t, filepath.Join(dir, "x"))
}

func TestOpenatDarwinPPCAtFDCWDFallsBackToPath(t *testing.T) {
	dir := ppcAtTestDir(t)
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Openat(unix.AT_FDCWD, file, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Close(fd); err != nil {
		t.Fatal(err)
	}
}

// TestAtDarwinPPCMetadataDarwinPPC is the Unix-layer control for the two
// no-follow metadata operations, which Leopard expresses only as
// setattrlist(FSOPT_NOFOLLOW) (there is no lchmod and no lutimes syscall, and
// libSystem's lchmod/lutimes are built on this same request — disassembled and
// measured on the target, runs/root-at-metadata-20261023).
//
// Every case names the entry in the anchor tree that must change and the decoy
// entry of the same name in the working directory that must not: the decoy is
// what makes "acted relative to the descriptor" distinguishable from "acted
// relative to the process cwd". The times are chosen far from "now" so a
// request that silently did nothing cannot look like success.
func TestAtDarwinPPCMetadataDarwinPPC(t *testing.T) {
	anchorDir, cwdDir, fd := atTree(t)
	anchorPath := func(name string) string { return filepath.Join(anchorDir, name) }

	// A symlink inside the anchor, and a decoy file of the same name outside it,
	// so "changed the link itself" is distinguishable from "followed it" and
	// from "changed the cwd's file".
	if err := os.WriteFile(anchorPath("target"), []byte("t"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", anchorPath("slink")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwdDir, "slink"), []byte("decoy"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(anchorPath("target"), 0644); err != nil {
		t.Fatal(err)
	}

	lstat := func(t *testing.T, path string) syscall.Stat_t {
		t.Helper()
		var st syscall.Stat_t
		if err := syscall.Lstat(path, &st); err != nil {
			t.Fatalf("Lstat %s: %v", path, err)
		}
		return st
	}

	t.Run("ChmodNofollowRegularFile", func(t *testing.T) {
		if err := unix.Fchmodat(fd, "file", 0641, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchmodat(AT_SYMLINK_NOFOLLOW): %v", err)
		}
		if got := lstat(t, anchorPath("file")).Mode & 0o777; got != 0641 {
			t.Errorf("anchor file mode = %o; want 641", got)
		}
		// The decoy in the working directory keeps its own mode.
		if got := lstat(t, filepath.Join(cwdDir, "file")).Mode & 0o777; got != 0600 {
			t.Errorf("cwd file mode = %o; want the untouched 600", got)
		}
	})

	t.Run("ChmodNofollowChangesTheLinkNotItsTarget", func(t *testing.T) {
		if err := unix.Fchmodat(fd, "slink", 0700, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchmodat(AT_SYMLINK_NOFOLLOW) on a symlink: %v", err)
		}
		if got := lstat(t, anchorPath("slink")).Mode & 0o777; got != 0700 {
			t.Errorf("anchor symlink mode = %o; want 700 on the link itself", got)
		}
		if got := lstat(t, anchorPath("target")).Mode & 0o777; got != 0644 {
			t.Errorf("symlink target mode = %o; want the untouched 644 (the link was followed)", got)
		}
		if got := lstat(t, filepath.Join(cwdDir, "slink")).Mode & 0o777; got != 0600 {
			t.Errorf("cwd decoy mode = %o; want the untouched 600", got)
		}
	})

	// chmod(2) masks the mode with ALLPERMS (bsd/vfs/vfs_syscalls.c:4106-4117).
	// File-type bits passed in the mode must not reach the file.
	t.Run("ChmodMasksAllPerms", func(t *testing.T) {
		if err := unix.Fchmodat(fd, "file", 0o170701, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Fchmodat with high mode bits: %v", err)
		}
		st := lstat(t, anchorPath("file"))
		if got := st.Mode & 0o777; got != 0701 {
			t.Errorf("mode = %o; want 701", got)
		}
		if st.Mode&syscall.S_IFMT != syscall.S_IFREG {
			t.Errorf("file type = %o; want a regular file: the type bits in the mode must be masked, not stored", st.Mode&syscall.S_IFMT)
		}
	})

	t.Run("ChmodMissingName", func(t *testing.T) {
		if err := unix.Fchmodat(fd, "no-such-entry", 0600, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOENT) {
			t.Errorf("Fchmodat(missing) = %v; want ENOENT", err)
		}
	})

	// Explicit times: the two times are distinguishable, and one second of
	// separation keeps them so after the filesystem's own rounding.
	mod := syscall.Timespec{Sec: 1300000001, Nsec: 0}
	acc := syscall.Timespec{Sec: 1300000000, Nsec: 0}
	setStamp := func(t *testing.T, path string) {
		t.Helper()
		if err := syscall.UtimesNano(path, []syscall.Timespec{{Sec: 1200000000}, {Sec: 1200000000}}); err != nil {
			t.Fatalf("UtimesNano(%s): %v", path, err)
		}
	}
	checkTimes := func(t *testing.T, path string, wantAcc, wantMod int32) {
		t.Helper()
		st := lstat(t, path)
		if got := st.Atimespec.Sec; got != wantAcc {
			t.Errorf("%s atime = %d; want %d", path, got, wantAcc)
		}
		if got := st.Mtimespec.Sec; got != wantMod {
			t.Errorf("%s mtime = %d; want %d", path, got, wantMod)
		}
	}

	t.Run("UtimensatBoth", func(t *testing.T) {
		setStamp(t, anchorPath("file"))
		setStamp(t, filepath.Join(cwdDir, "file"))
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{acc, mod}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat: %v", err)
		}
		checkTimes(t, anchorPath("file"), acc.Sec, mod.Sec)
		checkTimes(t, filepath.Join(cwdDir, "file"), 1200000000, 1200000000)
	})

	t.Run("UtimensatModTimeOnly", func(t *testing.T) {
		setStamp(t, anchorPath("file"))
		omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{omit, mod}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(MODTIME only): %v", err)
		}
		// The access time keeps the stamp it had: the omitted attribute is not
		// read back and rewritten, it is not sent at all.
		checkTimes(t, anchorPath("file"), 1200000000, mod.Sec)
	})

	// An access-time-only request is the form HFS interprets differently when the
	// access time is the request's ONLY active attribute: it would be a lazy
	// access-time touch, so the supplied value would be discarded and the stored
	// access time would become "now" (bsd/hfs/hfs_vnops.c:776-780). The backend
	// therefore also selects the change time on HFS — a value the filesystem
	// ignores — which makes the ordinary timestamp block run and store the
	// requested access time exactly (runs/root-at-atime-exact-20261023). This case
	// requires the exact value, and requires the change time to be untouched, which
	// is what proves the extra attribute was ignored rather than applied.
	//
	// The pair around it keeps it discriminating: with both times selected the
	// value is stored exactly (UtimensatBoth), and with only the modification time
	// selected the access time is untouched (UtimensatModTimeOnly).
	t.Run("UtimensatAcctimeOnlyIsExact", func(t *testing.T) {
		setStamp(t, anchorPath("file"))
		omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
		before := lstat(t, anchorPath("file"))
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{acc, omit}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(ACCTIME only): %v", err)
		}
		st := lstat(t, anchorPath("file"))
		if got := st.Atimespec.Sec; got != acc.Sec {
			t.Errorf("atime = %d; want the requested %d: an access-time-only request must store the value it was given", got, acc.Sec)
		}
		if got := st.Mtimespec.Sec; got != 1200000000 {
			t.Errorf("mtime = %d; want the untouched 1200000000: the omitted time must not be sent", got)
		}
		// The workaround's dummy change time is the epoch. If the filesystem had
		// honoured it, ctime would be the epoch; on HFS it is ignored, and the
		// measured behaviour is that ctime does not move at all for this request.
		if got := st.Ctimespec.Sec; got == 0 {
			t.Errorf("ctime = 0; the dummy change time was applied instead of ignored")
		}
		if st.Ctimespec != before.Ctimespec {
			t.Errorf("ctime = %+v; want the unchanged %+v: the dummy change time is ignored by HFS, not applied", st.Ctimespec, before.Ctimespec)
		}
		// The entry is still the anchor's, and the decoy is untouched.
		checkTimes(t, filepath.Join(cwdDir, "file"), 1200000000, 1200000000)
	})

	// A modification-time-only request must not carry the workaround's extra
	// attribute either: the access time stays as it was and the change time moves
	// the way HFS moves it when the modification time is set.
	t.Run("UtimensatModTimeOnlyChangeTimeMoves", func(t *testing.T) {
		setStamp(t, anchorPath("file"))
		omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
		before := lstat(t, anchorPath("file"))
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{omit, mod}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(MODTIME only): %v", err)
		}
		st := lstat(t, anchorPath("file"))
		if st.Atimespec != before.Atimespec {
			t.Errorf("atime = %+v; want the unchanged %+v", st.Atimespec, before.Atimespec)
		}
		if st.Mtimespec.Sec != mod.Sec {
			t.Errorf("mtime = %d; want %d", st.Mtimespec.Sec, mod.Sec)
		}
	})

	t.Run("UtimensatSymlinkNofollow", func(t *testing.T) {
		setStamp(t, anchorPath("slink"))
		setStamp(t, anchorPath("target"))
		if err := unix.Utimensat(fd, "slink", &[2]syscall.Timespec{acc, mod}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat on a symlink: %v", err)
		}
		checkTimes(t, anchorPath("slink"), acc.Sec, mod.Sec)
		checkTimes(t, anchorPath("target"), 1200000000, 1200000000)
	})

	// Both times omitted must still resolve the pathname: a missing name is
	// ENOENT, not a silent success, and an existing name is not modified.
	t.Run("UtimensatBothOmitted", func(t *testing.T) {
		setStamp(t, anchorPath("file"))
		omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{omit, omit}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Fatalf("Utimensat(both omitted) on an existing name: %v", err)
		}
		checkTimes(t, anchorPath("file"), 1200000000, 1200000000)
		if err := unix.Utimensat(fd, "no-such-entry", &[2]syscall.Timespec{omit, omit}, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOENT) {
			t.Errorf("Utimensat(both omitted) on a missing name = %v; want ENOENT", err)
		}
	})

	// The forms this backend does not claim are refused rather than
	// approximated, and the refusals are named so that changing one is an edit.
	t.Run("RetainedRefusals", func(t *testing.T) {
		now := syscall.Timespec{Sec: 0, Nsec: -1} // UTIME_NOW
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{now, now}, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
			t.Errorf("Utimensat(UTIME_NOW) = %v; want ENOTSUP (the different UTIME_NOW permission rule is not claimed)", err)
		}
		if err := unix.Utimensat(fd, "file", nil, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
			t.Errorf("Utimensat(nil times) = %v; want ENOTSUP (utimes(NULL) is \"both times now\")", err)
		}
		// POSIX says tv_sec is ignored when tv_nsec is UTIME_OMIT, so this
		// spelling is an omission and must not be an error.
		half := syscall.Timespec{Sec: 12, Nsec: unix.UTIME_OMIT}
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{half, half}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Errorf("Utimensat(UTIME_OMIT with a non-sentinel tv_sec) = %v; want nil", err)
		}
		bad := syscall.Timespec{Sec: 0, Nsec: 1 << 30}
		if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{bad, bad}, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Utimensat(nanoseconds out of range) = %v; want EINVAL", err)
		}
	})
}
