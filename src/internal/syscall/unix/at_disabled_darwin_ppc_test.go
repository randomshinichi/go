// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompat_disabled

package unix_test

// The disabled-backend dispatch, executed rather than argued about.
//
// atCompatEnabled is false on every kernel except the measured one, and that is
// the state a Tiger or an older Leopard machine is in. It is otherwise
// unreachable on hardware that exists here, so this file is compiled only with
// the goppc_atcompat_disabled build tag (internal/syscall/unix/
// at_compat_disabled_darwin_ppc.go makes the gate a compile-time constant; no
// production binary contains it and no test can race it).
//
// What must hold while it is disabled:
//
//   - every operation with a REAL directory descriptor refuses with ENOTSUP and
//     mutates nothing — no partial action in the anchor, no fallback to the
//     process working directory;
//   - the AT_FDCWD and absolute-path forms keep working, so the refusal is
//     descriptor-scoped and the rest of os keeps running on such a kernel.
//
// The tree it acts on carries a decoy of every name in the working directory, so
// a fallback that resolved against the cwd instead of the descriptor would be
// visible as a change to the wrong tree.

import (
	"errors"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// snapshotDarwinPPC records every name in dir with its size and mode, so "nothing
// changed" is a comparison rather than an opinion.
func snapshotDarwinPPC(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		fi, err := os.Lstat(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("Lstat %s: %v", e.Name(), err)
		}
		st, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			t.Fatalf("no syscall.Stat_t for %s", e.Name())
		}
		out = append(out, strings.Join([]string{
			e.Name(), fi.Mode().String(),
			itoaDarwinPPC(st.Size), itoaDarwinPPC(int64(st.Ino)),
			itoaDarwinPPC(int64(st.Atimespec.Sec)), itoaDarwinPPC(int64(st.Mtimespec.Sec)),
			itoaDarwinPPC(int64(st.Ctimespec.Sec)),
		}, " "))
	}
	sort.Strings(out)
	return out
}

func itoaDarwinPPC(v int64) string { return strconv.FormatInt(v, 10) }

// TestAtDarwinPPCDisabledBackendRefusesDarwinPPC requires every real-descriptor
// operation to refuse without touching either tree.
func TestAtDarwinPPCDisabledBackendRefusesDarwinPPC(t *testing.T) {
	if unix.AnchoredAvailable() {
		t.Fatalf("AnchoredAvailable() = true in a goppc_atcompat_disabled build; the gate override is broken and this control proves nothing")
	}
	anchorDir, cwdDir, fd := atTree(t)
	if err := os.Symlink("file", filepath.Join(anchorDir, "slink")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("file", filepath.Join(cwdDir, "slink")); err != nil {
		t.Fatal(err)
	}

	beforeAnchor := snapshotDarwinPPC(t, anchorDir)
	beforeCwd := snapshotDarwinPPC(t, cwdDir)

	times := [2]syscall.Timespec{{Sec: 1300000000}, {Sec: 1300000001}}
	var st syscall.Stat_t
	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"Openat", func() error {
			f, err := unix.Openat(fd, "file", syscall.O_RDONLY, 0)
			if err == nil {
				syscall.Close(f)
			}
			return err
		}},
		{"OpenatCreate", func() error {
			f, err := unix.Openat(fd, "created", syscall.O_CREAT|syscall.O_WRONLY, 0600)
			if err == nil {
				syscall.Close(f)
			}
			return err
		}},
		{"OpenatCreateExcl", func() error {
			f, err := unix.Openat(fd, "excl", syscall.O_CREAT|syscall.O_EXCL|syscall.O_WRONLY, 0600)
			if err == nil {
				syscall.Close(f)
			}
			return err
		}},
		{"OpenatDirectory", func() error {
			f, err := unix.Openat(fd, "sub", syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
			if err == nil {
				syscall.Close(f)
			}
			return err
		}},
		{"Unlinkat", func() error { return unix.Unlinkat(fd, "file", 0) }},
		{"UnlinkatRemoveDir", func() error { return unix.Unlinkat(fd, "sub", unix.AT_REMOVEDIR) }},
		{"Readlinkat", func() error {
			_, err := unix.Readlinkat(fd, "slink", make([]byte, 64))
			return err
		}},
		{"Mkdirat", func() error { return unix.Mkdirat(fd, "newdir", 0700) }},
		{"Fstatat", func() error { return unix.Fstatat(fd, "file", &st, 0) }},
		{"FstatatNofollow", func() error { return unix.Fstatat(fd, "slink", &st, unix.AT_SYMLINK_NOFOLLOW) }},
		{"Fchmodat", func() error { return unix.Fchmodat(fd, "file", 0644, 0) }},
		{"FchmodatNofollow", func() error { return unix.Fchmodat(fd, "file", 0644, unix.AT_SYMLINK_NOFOLLOW) }},
		{"Fchownat", func() error { return unix.Fchownat(fd, "file", os.Getuid(), os.Getgid(), 0) }},
		{"FchownatNofollow", func() error { return unix.Fchownat(fd, "file", os.Getuid(), os.Getgid(), unix.AT_SYMLINK_NOFOLLOW) }},
		{"Utimensat", func() error { return unix.Utimensat(fd, "file", &times, unix.AT_SYMLINK_NOFOLLOW) }},
		{"Symlinkat", func() error { return unix.Symlinkat("file", fd, "newlink") }},
		{"Renameat", func() error { return unix.Renameat(fd, "file", fd, "renamed") }},
		{"Linkat", func() error { return unix.Linkat(fd, "file", fd, "linked", 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !errors.Is(err, syscall.ENOTSUP) {
				t.Errorf("%s with a real directory descriptor = %v; want ENOTSUP while the backend is disabled", tc.name, err)
			}
			// Each refusal is checked immediately, so a mutation cannot be
			// attributed to a later case.
			if got := snapshotDarwinPPC(t, anchorDir); !equalDarwinPPC(got, beforeAnchor) {
				t.Errorf("the anchor tree changed after a refused %s:\n got %v\nwant %v", tc.name, got, beforeAnchor)
			}
			if got := snapshotDarwinPPC(t, cwdDir); !equalDarwinPPC(got, beforeCwd) {
				t.Errorf("the working directory changed after a refused %s:\n got %v\nwant %v", tc.name, got, beforeCwd)
			}
		})
	}
}

func equalDarwinPPC(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestAtDarwinPPCDisabledBackendPathFormsDarwinPPC requires the path forms to keep
// working while the backend is disabled. Without this, the refusals above would
// be indistinguishable from a build in which every *at call is broken.
func TestAtDarwinPPCDisabledBackendPathFormsDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// AT_FDCWD: the calling thread's own cwd, exactly as POSIX defines it.
	f, err := unix.Openat(unix.AT_FDCWD, "file", syscall.O_CREAT|syscall.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Openat(AT_FDCWD, create): %v", err)
	}
	if _, err := syscall.Write(f, []byte("x")); err != nil {
		t.Fatal(err)
	}
	syscall.Close(f)

	times := [2]syscall.Timespec{{Sec: 1300000000}, {Sec: 1300000001}}
	if err := unix.Utimensat(unix.AT_FDCWD, "file", &times, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Errorf("Utimensat(AT_FDCWD): %v", err)
	}
	if err := unix.Fchmodat(unix.AT_FDCWD, "file", 0640, 0); err != nil {
		t.Errorf("Fchmodat(AT_FDCWD): %v", err)
	}
	var st syscall.Stat_t
	if err := unix.Fstatat(unix.AT_FDCWD, "file", &st, 0); err != nil {
		t.Errorf("Fstatat(AT_FDCWD): %v", err)
	}
	if got := st.Atimespec.Sec; got != 1300000000 {
		t.Errorf("atime = %d; want 1300000000 after the path-form Utimensat", got)
	}
	if err := unix.Mkdirat(unix.AT_FDCWD, "d", 0700); err != nil {
		t.Errorf("Mkdirat(AT_FDCWD): %v", err)
	}
	if err := unix.Symlinkat("file", unix.AT_FDCWD, "l"); err != nil {
		t.Errorf("Symlinkat(AT_FDCWD): %v", err)
	}
	if err := unix.Renameat(unix.AT_FDCWD, "file", unix.AT_FDCWD, "renamed"); err != nil {
		t.Errorf("Renameat(AT_FDCWD): %v", err)
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "renamed", 0); err != nil {
		t.Errorf("Unlinkat(AT_FDCWD): %v", err)
	}
	if err := unix.Unlinkat(unix.AT_FDCWD, "d", unix.AT_REMOVEDIR); err != nil {
		t.Errorf("Unlinkat(AT_FDCWD, AT_REMOVEDIR): %v", err)
	}
}
