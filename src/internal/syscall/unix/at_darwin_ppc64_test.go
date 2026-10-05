// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package unix_test

import (
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestAtFDCWDDarwinPPC64 checks the *at family on Mac OS X 10.5, where it is
// implemented only for AT_FDCWD (at_darwin_ppc64.go).
func TestAtFDCWDDarwinPPC64(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := unix.Mkdirat(unix.AT_FDCWD, "d", 0o755); err != nil {
		t.Fatalf("Mkdirat: %v", err)
	}
	fd, err := unix.Openat(unix.AT_FDCWD, "d/f", syscall.O_CREAT|syscall.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("Openat: %v", err)
	}
	syscall.Close(fd)
	if err := unix.Symlinkat("f", unix.AT_FDCWD, "d/l"); err != nil {
		t.Fatalf("Symlinkat: %v", err)
	}
	buf := make([]byte, 16)
	n, err := unix.Readlinkat(unix.AT_FDCWD, "d/l", buf)
	if err != nil || string(buf[:n]) != "f" {
		t.Fatalf("Readlinkat = %q, %v; want \"f\", nil", buf[:n], err)
	}
	if err := unix.Renameat(unix.AT_FDCWD, "d/f", unix.AT_FDCWD, "d/g"); err != nil {
		t.Fatalf("Renameat: %v", err)
	}
	var st syscall.Stat_t
	if err := unix.Fstatat(unix.AT_FDCWD, "d/g", &st, 0); err != nil {
		t.Fatalf("Fstatat: %v", err)
	}
	if err := unix.Fchmodat(unix.AT_FDCWD, "d/g", 0o600, 0); err != nil {
		t.Fatalf("Fchmodat: %v", err)
	}
	if err := unix.Eaccess("d/g", unix.R_OK); err != nil {
		t.Fatalf("Eaccess (AT_EACCESS): %v", err)
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "d", 0); err == nil {
		t.Fatalf("Unlinkat(non-empty dir, 0) succeeded")
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "d/l", 0); err != nil {
		t.Fatalf("Unlinkat(link): %v", err)
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "d/g", 0); err != nil {
		t.Fatalf("Unlinkat(file): %v", err)
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "d", unix.AT_REMOVEDIR); err != nil {
		t.Fatalf("Unlinkat(AT_REMOVEDIR): %v", err)
	}
	if _, err := os.Lstat("d"); !os.IsNotExist(err) {
		t.Fatalf("d still exists: %v", err)
	}
}

// TestAtRealDirfdNotSupportedDarwinPPC64 checks that a real directory
// descriptor is refused with ENOTSUP and, more to the point, that nothing
// happens to the working directory instead: the wrong-directory fallback is
// the failure this refusal exists to prevent.
func TestAtRealDirfdNotSupportedDarwinPPC64(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()
	for _, d := range []string{dirA, dirB} {
		if err := os.WriteFile(filepath.Join(d, "f"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dirB) // the "wrong directory" a careless fallback would act on
	dirfd, err := syscall.Open(dirA, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(dirfd)

	var st syscall.Stat_t
	buf := make([]byte, 8)
	for name, err := range map[string]error{
		"Unlinkat":  unix.Unlinkat(dirfd, "f", 0),
		"Mkdirat":   unix.Mkdirat(dirfd, "m", 0o755),
		"Fchmodat":  unix.Fchmodat(dirfd, "f", 0o600, 0),
		"Fchownat":  unix.Fchownat(dirfd, "f", -1, -1, 0),
		"Renameat":  unix.Renameat(dirfd, "f", dirfd, "g"),
		"Linkat":    unix.Linkat(dirfd, "f", dirfd, "g", 0),
		"Symlinkat": unix.Symlinkat("f", dirfd, "s"),
		"Fstatat":   unix.Fstatat(dirfd, "f", &st, 0),
		// The path form that serves Readdir is a separate function; Fstatat must not
		// have acquired it, because os.Root's Lstat reaches Fstatat.
		"Fstatat(NOFOLLOW)":   unix.Fstatat(dirfd, "f", &st, unix.AT_SYMLINK_NOFOLLOW),
		"Fchmodat(NOFOLLOW)":  unix.Fchmodat(dirfd, "f", 0o600, unix.AT_SYMLINK_NOFOLLOW),
		"Utimensat(NOFOLLOW)": unix.Utimensat(dirfd, "f", &[2]syscall.Timespec{{Sec: 1e9}, {Sec: 1e9}}, unix.AT_SYMLINK_NOFOLLOW),
		"Utimensat":           unix.Utimensat(dirfd, "f", &[2]syscall.Timespec{{Sec: 1e9}, {Sec: 1e9}}, 0),
	} {
		if err != syscall.ENOTSUP {
			t.Errorf("%s(real dirfd) = %v; want ENOTSUP", name, err)
		}
	}
	if _, err := unix.Openat(dirfd, "f", syscall.O_RDONLY, 0); err != syscall.ENOTSUP {
		t.Errorf("Openat(real dirfd) = %v; want ENOTSUP", err)
	}
	if _, err := unix.Readlinkat(dirfd, "f", buf); err != syscall.ENOTSUP {
		t.Errorf("Readlinkat(real dirfd) = %v; want ENOTSUP", err)
	}
	// Neither directory changed.
	for _, d := range []string{dirA, dirB} {
		entries, err := os.ReadDir(d)
		if err != nil || len(entries) != 1 || entries[0].Name() != "f" {
			t.Errorf("directory %s changed: %v, %v", d, entries, err)
		}
	}
	// Utimensat(real dirfd) touched neither file's modification time.
	for _, d := range []string{dirA, dirB} {
		fi, err := os.Stat(filepath.Join(d, "f"))
		if err != nil || fi.ModTime().Unix() == 1e9 {
			t.Errorf("%s/f modification time changed by Utimensat(real dirfd): %v, %v", d, fi, err)
		}
	}
}

// TestUtimensatDarwinPPC64 checks syscall.utimensat on Mac OS X 10.5, which has
// no utimensat and maps AT_FDCWD onto utimes (syscall_darwin_ppc64.go). Times
// are whole seconds because the file system, not the call, decides the finer
// resolution.
func TestUtimensatDarwinPPC64(t *testing.T) {
	const utimeNow = -1 // modern Darwin value; Leopard defines none
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("f", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	at := func(sec int64) syscall.Timespec { return syscall.Timespec{Sec: sec} }
	omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
	now := syscall.Timespec{Nsec: utimeNow}
	times := func() (atime, mtime int64) {
		t.Helper()
		var st syscall.Stat_t
		if err := syscall.Stat("f", &st); err != nil {
			t.Fatal(err)
		}
		return st.Atimespec.Sec, st.Mtimespec.Sec
	}
	set := func(ts [2]syscall.Timespec, flags int) error {
		return unix.Utimensat(unix.AT_FDCWD, "f", &ts, flags)
	}

	if err := set([2]syscall.Timespec{at(1_000_000_000), at(1_100_000_000)}, 0); err != nil {
		t.Fatalf("explicit times: %v", err)
	}
	if a, m := times(); a != 1_000_000_000 || m != 1_100_000_000 {
		t.Fatalf("explicit times: got atime %d mtime %d", a, m)
	}
	if err := set([2]syscall.Timespec{omit, at(1_200_000_000)}, 0); err != nil {
		t.Fatalf("omit atime: %v", err)
	}
	if a, m := times(); a != 1_000_000_000 || m != 1_200_000_000 {
		t.Fatalf("omit atime: got atime %d mtime %d; want atime kept", a, m)
	}
	if err := set([2]syscall.Timespec{at(1_300_000_000), omit}, 0); err != nil {
		t.Fatalf("omit mtime: %v", err)
	}
	if a, m := times(); a != 1_300_000_000 || m != 1_200_000_000 {
		t.Fatalf("omit mtime: got atime %d mtime %d; want mtime kept", a, m)
	}
	if err := set([2]syscall.Timespec{omit, omit}, 0); err != nil {
		t.Fatalf("omit both: %v", err)
	}
	if a, m := times(); a != 1_300_000_000 || m != 1_200_000_000 {
		t.Fatalf("omit both: got atime %d mtime %d; want both kept", a, m)
	}
	if err := unix.Utimensat(unix.AT_FDCWD, "missing", &[2]syscall.Timespec{omit, omit}, 0); err != syscall.ENOENT {
		t.Errorf("omit both on a missing file = %v; want ENOENT", err)
	}
	if err := unix.Utimensat(unix.AT_FDCWD, "f", nil, 0); err != nil {
		t.Fatalf("nil times: %v", err)
	}
	if a, m := times(); a < 1_400_000_000 || m < 1_400_000_000 {
		t.Fatalf("nil times: got atime %d mtime %d; want the current time", a, m)
	}
	if err := set([2]syscall.Timespec{at(1_000_000_000), at(1_000_000_000)}, 0); err != nil {
		t.Fatal(err)
	}
	if err := set([2]syscall.Timespec{now, now}, 0); err != nil {
		t.Fatalf("both now: %v", err)
	}
	if a, m := times(); a < 1_400_000_000 || m < 1_400_000_000 {
		t.Fatalf("both now: got atime %d mtime %d; want the current time", a, m)
	}

	// Refusals leave the file alone.
	if err := set([2]syscall.Timespec{at(1_000_000_000), at(1_000_000_000)}, 0); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		ts    [2]syscall.Timespec
		flags int
		want  error
	}{
		// lutimes(NULL) was never measured, so "now" with NOFOLLOW is refused.
		"nofollow and now": {[2]syscall.Timespec{now, now}, unix.AT_SYMLINK_NOFOLLOW, syscall.ENOTSUP},
		"undefined bit":    {[2]syscall.Timespec{at(5), at(5)}, 0x1, syscall.EINVAL},
		"now and time":     {[2]syscall.Timespec{now, at(5)}, 0, syscall.ENOTSUP},
		"now and omit":     {[2]syscall.Timespec{omit, now}, 0, syscall.ENOTSUP},
		"nsec too big":     {[2]syscall.Timespec{{Sec: 5, Nsec: 1e9}, at(5)}, 0, syscall.EINVAL},
		"nsec negative":    {[2]syscall.Timespec{at(5), {Sec: 5, Nsec: -3}}, 0, syscall.EINVAL},
	} {
		if err := set(c.ts, c.flags); err != c.want {
			t.Errorf("%s: got %v; want %v", name, err, c.want)
		}
	}
	if err := unix.Utimensat(unix.AT_FDCWD, "f", nil, unix.AT_SYMLINK_NOFOLLOW); err != syscall.ENOTSUP {
		t.Errorf("nil times with NOFOLLOW = %v; want ENOTSUP", err)
	}
	if a, m := times(); a != 1_000_000_000 || m != 1_000_000_000 {
		t.Errorf("refusals changed the file: atime %d mtime %d", a, m)
	}
}

// TestUtimensatNofollowDarwinPPC64 checks AT_SYMLINK_NOFOLLOW, which
// syscall.utimensat maps onto lutimes: the link's own times change and the
// target's do not. It fails if NOFOLLOW is refused (the previous behaviour,
// ENOTSUP) and if it is mapped onto utimes, which would change the target.
func TestUtimensatNofollowDarwinPPC64(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("target", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", "link"); err != nil {
		t.Fatal(err)
	}
	at := func(sec int64) syscall.Timespec { return syscall.Timespec{Sec: sec} }
	omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}
	times := func(path string, lstat bool) (atime, mtime int64) {
		t.Helper()
		var st syscall.Stat_t
		var err error
		if lstat {
			err = syscall.Lstat(path, &st)
		} else {
			err = syscall.Stat(path, &st)
		}
		if err != nil {
			t.Fatal(err)
		}
		return st.Atimespec.Sec, st.Mtimespec.Sec
	}
	const (
		targetAtime, targetMtime = 1_000_000_000, 1_100_000_000
		linkAtime, linkMtime     = 1_300_000_000, 1_400_000_000
	)
	if err := unix.Utimensat(unix.AT_FDCWD, "target", &[2]syscall.Timespec{at(targetAtime), at(targetMtime)}, 0); err != nil {
		t.Fatal(err)
	}
	if err := unix.Utimensat(unix.AT_FDCWD, "link", &[2]syscall.Timespec{at(linkAtime), at(linkMtime)}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatalf("NOFOLLOW explicit times: %v", err)
	}
	if a, m := times("link", true); a != linkAtime || m != linkMtime {
		t.Errorf("link: atime %d mtime %d; want %d %d (the link itself)", a, m, linkAtime, linkMtime)
	}
	if a, m := times("target", false); a != targetAtime || m != targetMtime {
		t.Errorf("target: atime %d mtime %d; want %d %d untouched", a, m, targetAtime, targetMtime)
	}

	// UTIME_OMIT with NOFOLLOW keeps the LINK's own time, read with lstat, not
	// the target's that stat would return.
	const newMtime = 1_500_000_000
	if err := unix.Utimensat(unix.AT_FDCWD, "link", &[2]syscall.Timespec{omit, at(newMtime)}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatalf("NOFOLLOW omit atime: %v", err)
	}
	if a, m := times("link", true); a != linkAtime || m != newMtime {
		t.Errorf("link after omit atime: atime %d mtime %d; want %d %d", a, m, linkAtime, newMtime)
	}
	if a, m := times("target", false); a != targetAtime || m != targetMtime {
		t.Errorf("target after omit: atime %d mtime %d; want %d %d untouched", a, m, targetAtime, targetMtime)
	}

	// Without NOFOLLOW the same call acts on the target.
	if err := unix.Utimensat(unix.AT_FDCWD, "link", &[2]syscall.Timespec{at(1_600_000_000), at(1_700_000_000)}, 0); err != nil {
		t.Fatal(err)
	}
	if a, m := times("target", false); a != 1_600_000_000 || m != 1_700_000_000 {
		t.Errorf("target after follow: atime %d mtime %d", a, m)
	}
	if a, m := times("link", true); a != linkAtime || m != newMtime {
		t.Errorf("link after follow: atime %d mtime %d; want it unchanged (%d %d)", a, m, linkAtime, newMtime)
	}
}

// TestFchmodatNofollowAtFDCWDDarwinPPC64 checks Fchmodat(AT_FDCWD, link, mode,
// AT_SYMLINK_NOFOLLOW), which maps onto lchmod: the link's mode changes and the
// target's does not. It fails if NOFOLLOW is refused (the previous behaviour,
// ENOTSUP) and if it is mapped onto chmod, which would change the target. This
// is the honourable half of what TestFchmodatSymlinkNofollow (fchmodat_test.go)
// exercises; that test also calls Fchmodat with a real directory descriptor
// before it reaches NOFOLLOW, which stays ENOTSUP here.
func TestFchmodatNofollowAtFDCWDDarwinPPC64(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("target", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod("target", 0o644); err != nil { // not subject to the umask
		t.Fatal(err)
	}
	if err := os.Symlink("target", "link"); err != nil {
		t.Fatal(err)
	}
	linkMode := func() os.FileMode {
		t.Helper()
		fi, err := os.Lstat("link")
		if err != nil {
			t.Fatal(err)
		}
		return fi.Mode()
	}
	targetMode := func() os.FileMode {
		t.Helper()
		fi, err := os.Stat("target")
		if err != nil {
			t.Fatal(err)
		}
		return fi.Mode().Perm()
	}
	before := linkMode().Perm()
	if err := unix.Fchmodat(unix.AT_FDCWD, "link", 0o600, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatalf("Fchmodat(AT_FDCWD, link, NOFOLLOW): %v", err)
	}
	if got := linkMode(); got&os.ModeSymlink == 0 || got.Perm() != 0o600 {
		t.Errorf("link mode = %v (was perm %v); want a symlink with perm 0600", got, before)
	}
	if got := targetMode(); got != 0o644 {
		t.Errorf("target mode = %v; want 0644 untouched", got)
	}

	// The control: without NOFOLLOW the same call follows the link.
	if err := unix.Fchmodat(unix.AT_FDCWD, "link", 0o640, 0); err != nil {
		t.Fatalf("Fchmodat(AT_FDCWD, link, 0): %v", err)
	}
	if got := targetMode(); got != 0o640 {
		t.Errorf("target mode after follow = %v; want 0640", got)
	}
	if got := linkMode(); got.Perm() != 0o600 {
		t.Errorf("link perm after follow = %v; want it unchanged (0600)", got.Perm())
	}
	if err := unix.Fchmodat(unix.AT_FDCWD, "link", 0o600, 0x1); err != syscall.EINVAL {
		t.Errorf("Fchmodat with an undefined flag = %v; want EINVAL", err)
	}
}

// TestLstatatByPathDarwinPPC64 checks the path form that serves
// (*os.File).Readdir: lstat of an entry of a directory that is open as a
// descriptor, by resolving the descriptor with F_GETPATH. The working
// directory holds entries of the same names with different contents, so a
// lookup that fell back to it would be seen.
func TestLstatatByPathDarwinPPC64(t *testing.T) {
	dir, cwd := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "f"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("f", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("elsewhere", filepath.Join(cwd, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	dirfd, err := syscall.Open(dir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(dirfd)

	var st syscall.Stat_t
	if err := unix.LstatatByPath(dirfd, "f", &st); err != nil || st.Size != 5 || st.Mode&syscall.S_IFMT != syscall.S_IFREG {
		t.Errorf("LstatatByPath(f) = %v, size %d, mode %o; want size 5 regular", err, st.Size, st.Mode)
	}
	st = syscall.Stat_t{}
	if err := unix.LstatatByPath(dirfd, "link", &st); err != nil || st.Mode&syscall.S_IFMT != syscall.S_IFLNK {
		t.Errorf("LstatatByPath(link) = %v, mode %o; want the symbolic link itself (S_IFLNK)", err, st.Mode)
	}
	st = syscall.Stat_t{}
	if err := unix.LstatatByPath(dirfd, "sub", &st); err != nil || st.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		t.Errorf("LstatatByPath(sub) = %v, mode %o; want a directory", err, st.Mode)
	}
	if err := unix.LstatatByPath(dirfd, "missing", &st); err != syscall.ENOENT {
		t.Errorf("LstatatByPath(missing) = %v; want ENOENT", err)
	}
	// A name that is not a single entry name is refused, never resolved.
	for _, name := range []string{"sub/..", "../f", "sub/x", "/f"} {
		if err := unix.LstatatByPath(dirfd, name, &st); err != syscall.EINVAL {
			t.Errorf("LstatatByPath(%q) = %v; want EINVAL", name, err)
		}
	}
	if err := unix.LstatatByPath(dirfd, "", &st); err != syscall.ENOENT {
		t.Errorf("LstatatByPath(empty name) = %v; want ENOENT", err)
	}
	if err := unix.LstatatByPath(-1, "f", &st); err != syscall.EBADF {
		t.Errorf("LstatatByPath(bad fd) = %v; want EBADF", err)
	}
	// AT_FDCWD is plain lstat in the working directory.
	st = syscall.Stat_t{}
	if err := unix.LstatatByPath(unix.AT_FDCWD, "f", &st); err != nil || st.Size != 1 {
		t.Errorf("LstatatByPath(AT_FDCWD, f) = %v, size %d; want size 1", err, st.Size)
	}
}

// TestARC4RandomDarwinPPC64 checks the arc4random-based buffer fill,
// including lengths that are not a multiple of the four bytes one draw yields.
func TestARC4RandomDarwinPPC64(t *testing.T) {
	for n := 0; n <= 37; n++ {
		a, b := make([]byte, n+8), make([]byte, n+8)
		unix.ARC4Random(a[4 : 4+n])
		unix.ARC4Random(b[4 : 4+n])
		for i := range a {
			if (i < 4 || i >= 4+n) && (a[i] != 0 || b[i] != 0) {
				t.Fatalf("n=%d: wrote outside the slice at %d", n, i)
			}
		}
		if n >= 16 && string(a) == string(b) {
			t.Errorf("n=%d: two fills are identical", n)
		}
	}
}
