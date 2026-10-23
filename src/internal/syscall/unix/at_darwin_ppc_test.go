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
	"time"
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
		for _, name := range []string{"file", "renamed", "link", "sub", "d", "new", "excl"} {
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
	if !unix.AtCompatEnabledForTest() {
		t.Fatal("the anchored backend is disabled on this kernel; these tests only make sense on the reviewed 9.8.x regime")
	}
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

// TestAtDarwinPPCRetainedRefusals is the no-mutation control for the operations
// this backend does not serve. Each must refuse, and must leave both trees
// exactly as they were: an ENOTSUP that still acted would be worse than a wrong
// answer.
func TestAtDarwinPPCRetainedRefusals(t *testing.T) {
	anchorDir, cwdDir, fd := atTree(t)
	otherFd, err := syscall.Open(cwdDir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(otherFd)

	// A same-parent Link is refused even for a regular file: link(2) follows a
	// final source symlink and HFS+ can hard-link directories, so this is not
	// a same-parent exception.
	if err := unix.Linkat(fd, "file", fd, "linked", 0); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Linkat(same parent) = %v; want ENOTSUP", err)
	}
	wantAbsent(t, filepath.Join(anchorDir, "linked"))
	wantAbsent(t, filepath.Join(cwdDir, "linked"))

	// A rename whose two operands need two different anchored directories is
	// not representable here.
	if err := unix.Renameat(fd, "file", otherFd, "moved"); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Renameat(two anchors) = %v; want ENOTSUP", err)
	}
	wantContent(t, filepath.Join(anchorDir, "file"), "anchor")
	wantContent(t, filepath.Join(cwdDir, "file"), "cwd")
	wantAbsent(t, filepath.Join(cwdDir, "moved"))

	// No-follow chmod: a following chmod would change the wrong object, and the
	// setattrlist form has not passed its parity gates.
	if err := unix.Fchmodat(fd, "file", 0600, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Fchmodat(AT_SYMLINK_NOFOLLOW) = %v; want ENOTSUP", err)
	}

	// Timestamps through a real descriptor are refused for the same reason.
	now := syscall.NsecToTimespec(time.Now().UnixNano())
	if err := unix.Utimensat(fd, "file", &[2]syscall.Timespec{now, now}, unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Utimensat(real dirfd) = %v; want ENOTSUP", err)
	}

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
