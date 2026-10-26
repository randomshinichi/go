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

// atTwoDirs builds two sibling directories with an open descriptor each, which is
// the fixture for every two-anchor row of the dispatch table.
func atTwoDirs(t *testing.T) (aDir, bDir string, a, b int) {
	t.Helper()
	root := t.TempDir()
	aDir = filepath.Join(root, "a")
	bDir = filepath.Join(root, "b")
	for _, dir := range []string{aDir, bDir} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	if a, err = syscall.Open(aDir, syscall.O_RDONLY, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Close(a) })
	if b, err = syscall.Open(bDir, syscall.O_RDONLY, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Close(b) })
	return aDir, bDir, a, b
}

// atNlink reports the link count of a path.
func atNlink(t *testing.T, path string) uint16 {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("%s: no Stat_t", path)
	}
	return st.Nlink
}

// TestLinkatOperandsDarwinPPC covers every row of the two-name anchor table for
// link: the rows that need no anchor at all, the ones a single held directory
// serves, the mixed AT_FDCWD forms (which capture the caller's own working
// directory), and the two-different-directories row whose destination is
// reconstructed by name.
//
// Every success case checks the link really exists, is the same object as the
// source (dev/ino), raised nlink, and left the source in place: a nil error is not
// evidence that the right thing happened. The reconstruction row is the approved
// divergence and is labelled as such; it is not a claim of confinement.
func TestLinkatOperandsDarwinPPC(t *testing.T) {
	t.Run("AbsAbs", func(t *testing.T) {
		aDir, _, _, _ := atTwoDirs(t)
		dead := atDeadFD(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		srcIno := atIno(t, filepath.Join(aDir, "src"))
		if err := unix.Linkat(dead, filepath.Join(aDir, "src"), dead, filepath.Join(aDir, "dst"), 0); err != nil {
			t.Fatalf("Linkat(abs, abs): %v", err)
		}
		if got := atIno(t, filepath.Join(aDir, "dst")); got != srcIno {
			t.Errorf("link inode = %d; want the source inode %d", got, srcIno)
		}
		if got := atNlink(t, filepath.Join(aDir, "src")); got != 2 {
			t.Errorf("nlink = %d; want 2", got)
		}
		wantContent(t, filepath.Join(aDir, "src"), "abs")
	})

	t.Run("AbsRelFDCWD", func(t *testing.T) {
		aDir, bDir, _, _ := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		t.Chdir(bDir)
		if err := unix.Linkat(unix.AT_FDCWD, filepath.Join(aDir, "src"), unix.AT_FDCWD, "dst", 0); err != nil {
			t.Fatalf("Linkat(abs, rel AT_FDCWD): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "abs")
		wantAbsent(t, filepath.Join(aDir, "dst"))
	})

	t.Run("RelFDCWDAbs", func(t *testing.T) {
		aDir, bDir, _, _ := atTwoDirs(t)
		atWrite(t, filepath.Join(bDir, "src"), "cwd")
		t.Chdir(bDir)
		if err := unix.Linkat(unix.AT_FDCWD, "src", unix.AT_FDCWD, filepath.Join(aDir, "dst"), 0); err != nil {
			t.Fatalf("Linkat(rel AT_FDCWD, abs): %v", err)
		}
		wantContent(t, filepath.Join(aDir, "dst"), "cwd")
	})

	t.Run("AbsRelRealFD", func(t *testing.T) {
		aDir, bDir, _, b := atTwoDirs(t)
		dead := atDeadFD(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		if err := unix.Linkat(dead, filepath.Join(aDir, "src"), b, "dst", 0); err != nil {
			t.Fatalf("Linkat(abs, rel real fd): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "abs")
	})

	t.Run("RelRealFDAbs", func(t *testing.T) {
		_, bDir, a, _ := atTwoDirs(t)
		dead := atDeadFD(t)
		atWrite(t, filepath.Join(bDir, "src"), "abs")
		if err := unix.Linkat(a, "src", dead, filepath.Join(bDir, "dst"), 0); err != nil {
			t.Fatalf("Linkat(rel real fd, abs): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "abs")
	})

	t.Run("RelRelFDCWD", func(t *testing.T) {
		_, bDir, _, _ := atTwoDirs(t)
		atWrite(t, filepath.Join(bDir, "src"), "cwd")
		t.Chdir(bDir)
		if err := unix.Linkat(unix.AT_FDCWD, "src", unix.AT_FDCWD, "dst", 0); err != nil {
			t.Fatalf("Linkat(rel AT_FDCWD, rel AT_FDCWD): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "cwd")
	})

	t.Run("RelRelSameFD", func(t *testing.T) {
		aDir, _, a, _ := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		if err := unix.Linkat(a, "src", a, "dst", 0); err != nil {
			t.Fatalf("Linkat(one fd): %v", err)
		}
		wantContent(t, filepath.Join(aDir, "dst"), "abs")
	})

	t.Run("RelRelTwoFDsSameDirectory", func(t *testing.T) {
		aDir, _, _, _ := atTwoDirs(t)
		first, err := syscall.Open(aDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(first)
		second, err := syscall.Open(aDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Close(second)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		if err := unix.Linkat(first, "src", second, "dst", 0); err != nil {
			t.Fatalf("Linkat(two fds on one directory): %v", err)
		}
		wantContent(t, filepath.Join(aDir, "dst"), "abs")
	})

	t.Run("RelRelTwoDirectoriesReconstructed", func(t *testing.T) {
		// The approved divergence: the destination directory is not held, its name
		// is reconstructed from the descriptor. The link itself must still be a link
		// of the source inode, with the source left in place.
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		srcIno := atIno(t, filepath.Join(aDir, "src"))
		if err := unix.Linkat(a, "src", b, "dst", 0); err != nil {
			t.Fatalf("Linkat(two directories): %v", err)
		}
		if got := atIno(t, filepath.Join(bDir, "dst")); got != srcIno {
			t.Errorf("link inode = %d; want the source inode %d", got, srcIno)
		}
		if got := atNlink(t, filepath.Join(aDir, "src")); got != 2 {
			t.Errorf("nlink = %d; want 2", got)
		}
		wantContent(t, filepath.Join(aDir, "src"), "abs")
		wantAbsent(t, filepath.Join(bDir, "src"))
	})

	t.Run("RelFDCWDToRealFD", func(t *testing.T) {
		// The caller's own working directory is captured for the AT_FDCWD operand
		// and the destination anchor is the caller's descriptor.
		_, bDir, _, b := atTwoDirs(t)
		atWrite(t, filepath.Join(bDir, "src"), "cwd")
		t.Chdir(bDir)
		if err := unix.Linkat(unix.AT_FDCWD, "src", b, "dst", 0); err != nil {
			t.Fatalf("Linkat(AT_FDCWD, real fd): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "cwd")
	})

	t.Run("RelRealFDToFDCWD", func(t *testing.T) {
		aDir, bDir, a, _ := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		t.Chdir(bDir)
		if err := unix.Linkat(a, "src", unix.AT_FDCWD, "dst", 0); err != nil {
			t.Fatalf("Linkat(real fd, AT_FDCWD): %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "abs")
		wantAbsent(t, filepath.Join(aDir, "dst"))
	})
}

// TestLinkatFlagsDarwinPPC pins the flag policy: zero is the legacy following
// link, AT_SYMLINK_NOFOLLOW is an explicit refusal, and every other value is
// EINVAL. The refusals must happen before anything is looked at, so the source
// survives them and no name is created.
func TestLinkatFlagsDarwinPPC(t *testing.T) {
	aDir, bDir, a, b := atTwoDirs(t)
	atWrite(t, filepath.Join(aDir, "src"), "abs")

	for _, flag := range []int{unix.AT_SYMLINK_NOFOLLOW, 0x40, 0x1, -1, unix.AT_SYMLINK_NOFOLLOW | 0x40} {
		err := unix.Linkat(a, "src", b, "dst", flag)
		want := syscall.EINVAL
		if flag == unix.AT_SYMLINK_NOFOLLOW {
			want = syscall.ENOTSUP
		}
		if !errors.Is(err, want) {
			t.Errorf("Linkat flag %#x = %v; want %v", flag, err, want)
		}
		wantAbsent(t, filepath.Join(bDir, "dst"))
		wantContent(t, filepath.Join(aDir, "src"), "abs")
	}

	// The same refusal for the no-anchor forms: the flags are decided before the
	// anchor table, so an absolute pair is refused identically.
	if err := unix.Linkat(atDeadFD(t), filepath.Join(aDir, "src"), atDeadFD(t), filepath.Join(bDir, "dst"), unix.AT_SYMLINK_NOFOLLOW); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Linkat(abs, abs) with AT_SYMLINK_NOFOLLOW = %v; want ENOTSUP", err)
	}
	if err := unix.Linkat(a, "src", b, "dst", 0); err != nil {
		t.Errorf("Linkat with flag 0 = %v; want success", err)
	}
}

// TestLinkatFinalSourceSymlinkDarwinPPC documents and pins the first approved
// divergence: a zero-flag link follows a final source symlink.
//
// The sibling referent is created inside the same temporary tree, so the
// divergence is demonstrated without touching any file the test does not own.
// This is exactly the property that makes zero-flag Linkat unusable as
// os.Root.Link's "link the symlink inode" form; the test does not pretend
// otherwise, and no case here is skippable.
func TestLinkatFinalSourceSymlinkDarwinPPC(t *testing.T) {
	aDir, bDir, a, b := atTwoDirs(t)

	t.Run("RelativeReferent", func(t *testing.T) {
		atWrite(t, filepath.Join(aDir, "referent"), "referent")
		if err := os.Symlink("referent", filepath.Join(aDir, "link")); err != nil {
			t.Fatal(err)
		}
		linkIno := atIno(t, filepath.Join(aDir, "link"))
		refIno := atIno(t, filepath.Join(aDir, "referent"))
		if err := unix.Linkat(a, "link", b, "made", 0); err != nil {
			t.Fatalf("Linkat through a source symlink: %v", err)
		}
		if got := atIno(t, filepath.Join(bDir, "made")); got == linkIno {
			t.Errorf("the new link shares the symlink's inode: this would be POSIX semantics, not this port's")
		} else if got != refIno {
			t.Errorf("link inode = %d; want the referent's inode %d (the kernel follows the final source)", got, refIno)
		}
		wantContent(t, filepath.Join(bDir, "made"), "referent")
	})

	t.Run("AbsoluteReferent", func(t *testing.T) {
		atWrite(t, filepath.Join(aDir, "absreferent"), "abs")
		if err := os.Symlink(filepath.Join(aDir, "absreferent"), filepath.Join(aDir, "abslink")); err != nil {
			t.Fatal(err)
		}
		if err := unix.Linkat(a, "abslink", b, "absmade", 0); err != nil {
			t.Fatalf("Linkat through an absolute source symlink: %v", err)
		}
		if got, want := atIno(t, filepath.Join(bDir, "absmade")), atIno(t, filepath.Join(aDir, "absreferent")); got != want {
			t.Errorf("link inode = %d; want the referent's inode %d", got, want)
		}
	})

	t.Run("DanglingSource", func(t *testing.T) {
		if err := os.Symlink("nowhere", filepath.Join(aDir, "dangling")); err != nil {
			t.Fatal(err)
		}
		err := unix.Linkat(a, "dangling", b, "danglingmade", 0)
		if !errors.Is(err, syscall.ENOENT) {
			t.Errorf("Linkat of a dangling source symlink = %v; want ENOENT (POSIX would link the link itself)", err)
		}
		wantAbsent(t, filepath.Join(bDir, "danglingmade"))
	})
}

// TestAtReconstructedPathLengthDarwinPPC pins the length rule of the
// destination reconstruction.
//
// The C mechanism it replaces checks the combined length with strlcat's return
// value and can report a truncated path as success (M/src/atfuncs.c:45-64). Here
// the total is compared with MAXPATHLEN before the operation starts, so an
// over-long combination is ERANGE with nothing mutated. The subtests include short
// components whose SUM overflows, which a per-component check would miss.
func TestAtReconstructedPathLengthDarwinPPC(t *testing.T) {
	t.Run("BelowLimit", func(t *testing.T) {
		got, err := unix.AtConcatPathForTest("/tmp/x", "y")
		if err != nil || got != "/tmp/x/y" {
			t.Errorf("concat = %q, %v; want %q", got, err, "/tmp/x/y")
		}
	})

	t.Run("AtLimit", func(t *testing.T) {
		// base + '/' + suffix + NUL must fit in 1024 bytes exactly.
		base := "/" + repeatByte('b', 400)
		suffix := repeatByte('s', 1024-1-400-1-1)
		got, err := unix.AtConcatPathForTest(base, suffix)
		if err != nil {
			t.Fatalf("concat at the limit: %v", err)
		}
		if len(got)+1 != 1024 {
			t.Errorf("combined length = %d (+NUL); want exactly 1024", len(got)+1)
		}
	})

	t.Run("TotalOverflow", func(t *testing.T) {
		// Each component is short enough for a per-component check; their sum is not.
		base := "/" + repeatByte('b', 400)
		suffix := repeatByte('s', 700)
		if _, err := unix.AtConcatPathForTest(base, suffix); !errors.Is(err, syscall.ERANGE) {
			t.Errorf("concat of short components that sum over the limit = %v; want ERANGE", err)
		}
	})

	t.Run("RealF_GETPATH", func(t *testing.T) {
		aDir, _, a, _ := atTwoDirs(t)
		got, err := unix.AtPathFromFDForTest(a, "name")
		if err != nil {
			t.Fatalf("AtPathFromFD: %v", err)
		}
		if want := aDir + "/name"; got != want {
			t.Errorf("reconstructed path = %q; want %q", got, want)
		}
		// A suffix that makes the real directory's own path overflow is refused
		// before anything is looked up.
		if _, err := unix.AtPathFromFDForTest(a, repeatByte('s', 1024)); !errors.Is(err, syscall.ERANGE) {
			t.Errorf("reconstruction with an over-long suffix = %v; want ERANGE", err)
		}
	})

	t.Run("EmbeddedNUL", func(t *testing.T) {
		// The concatenation itself does not scan for NUL; the name conversion for
		// the kernel does, and that must happen before any mutation.
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		if err := unix.Linkat(a, "src", b, "bad\x00name", 0); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("Linkat with an embedded NUL destination = %v; want EINVAL", err)
		}
		wantAbsent(t, filepath.Join(bDir, "bad"))
		wantContent(t, filepath.Join(aDir, "src"), "abs")
	})
}

// repeatByte returns n copies of c as a string.
func repeatByte(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

// TestRenameatCrossDirectoryDarwinPPC covers the rename rows that reconstruct the
// destination: two relative names under two different held directories. They are
// here, and not in TestRenameatOperandsDarwinPPC, because they are the only rows
// that need the reconstruction seam; the dispatch rows that need no reconstruction
// stay in the file that pins the dispatch order.
func TestRenameatCrossDirectoryDarwinPPC(t *testing.T) {

	t.Run("RelRelTwoDirectories", func(t *testing.T) {
		// The approved divergence: the destination directory's name is
		// reconstructed from its descriptor.
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "abs")
		ino := atIno(t, filepath.Join(aDir, "src"))
		if err := unix.Renameat(a, "src", b, "dst"); err != nil {
			t.Fatalf("Renameat(two directories): %v", err)
		}
		wantAbsent(t, filepath.Join(aDir, "src"))
		if got := atIno(t, filepath.Join(bDir, "dst")); got != ino {
			t.Errorf("destination inode = %d; want the moved inode %d", got, ino)
		}
	})

	t.Run("RelRelTwoDirectoriesReplacement", func(t *testing.T) {
		// A separate fixture from the non-replacement case: replacing an existing
		// destination object and creating a new one are different kernel paths.
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "src")
		atWrite(t, filepath.Join(bDir, "dst"), "old")
		srcIno, oldIno := atIno(t, filepath.Join(aDir, "src")), atIno(t, filepath.Join(bDir, "dst"))
		if err := unix.Renameat(a, "src", b, "dst"); err != nil {
			t.Fatalf("Renameat replacing the destination: %v", err)
		}
		wantAbsent(t, filepath.Join(aDir, "src"))
		wantContent(t, filepath.Join(bDir, "dst"), "src")
		if got := atIno(t, filepath.Join(bDir, "dst")); got != srcIno || got == oldIno {
			t.Errorf("destination inode = %d; want the source inode %d (the old object %d is gone)", got, srcIno, oldIno)
		}
	})

	t.Run("RelRelMissingSourceNoMutation", func(t *testing.T) {
		_, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(bDir, "dst"), "keep")
		err := unix.Renameat(a, "absent", b, "dst")
		if !errors.Is(err, syscall.ENOENT) {
			t.Errorf("Renameat of a missing source across directories = %v; want ENOENT", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "keep")
	})

	t.Run("RelRelSymlinkFinalSourceMovesLink", func(t *testing.T) {
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "referent"), "referent")
		if err := os.Symlink("referent", filepath.Join(aDir, "link")); err != nil {
			t.Fatal(err)
		}
		linkIno := atIno(t, filepath.Join(aDir, "link"))
		if err := unix.Renameat(a, "link", b, "movedlink"); err != nil {
			t.Fatalf("Renameat of a final symlink: %v", err)
		}
		if got := atIno(t, filepath.Join(bDir, "movedlink")); got != linkIno {
			t.Errorf("moved object inode = %d; want the symlink's inode %d (rename does not follow the final source)", got, linkIno)
		}
		if target, err := os.Readlink(filepath.Join(bDir, "movedlink")); err != nil || target != "referent" {
			t.Errorf("moved link = %q, %v; want %q", target, err, "referent")
		}
		wantContent(t, filepath.Join(aDir, "referent"), "referent")
	})

	t.Run("RelRelSymlinkFinalDestinationReplacedNotFollowed", func(t *testing.T) {
		aDir, bDir, a, b := atTwoDirs(t)
		atWrite(t, filepath.Join(aDir, "src"), "src")
		atWrite(t, filepath.Join(bDir, "referent"), "referent")
		if err := os.Symlink("referent", filepath.Join(bDir, "dst")); err != nil {
			t.Fatal(err)
		}
		refIno := atIno(t, filepath.Join(bDir, "referent"))
		if err := unix.Renameat(a, "src", b, "dst"); err != nil {
			t.Fatalf("Renameat onto a final symlink: %v", err)
		}
		wantContent(t, filepath.Join(bDir, "dst"), "src")
		if got := atIno(t, filepath.Join(bDir, "referent")); got != refIno {
			t.Errorf("the destination's referent was replaced (%d -> %d); rename must replace the link, not follow it", refIno, got)
		}
		wantContent(t, filepath.Join(bDir, "referent"), "referent")
	})
}
