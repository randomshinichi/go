// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package os_test

import (
	"errors"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestRootNotSupportedDarwinPPC64 checks the boundary that keeps the Leopard
// *at substitution honest: os.Root needs real directory-descriptor semantics,
// which Mac OS X 10.5 lacks, so every operation fails with ENOTSUP and none of
// them acts on the process's working directory instead.
func TestRootNotSupportedDarwinPPC64(t *testing.T) {
	rootDir, cwd := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(rootDir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "f"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	ops := map[string]func() error{
		"Open":      func() error { _, err := r.Open("f"); return err },
		"Create":    func() error { _, err := r.Create("new"); return err },
		"Mkdir":     func() error { return r.Mkdir("m", 0o755) },
		"Remove":    func() error { return r.Remove("f") },
		"RemoveAll": func() error { return r.RemoveAll("f") },
		"Stat":      func() error { _, err := r.Stat("f"); return err },
		"Lstat":     func() error { _, err := r.Lstat("f"); return err },
		"Rename":    func() error { return r.Rename("f", "g") },
		"Symlink":   func() error { return r.Symlink("f", "s") },
	}
	for name, op := range ops {
		if err := op(); !errors.Is(err, syscall.ENOTSUP) {
			t.Errorf("Root.%s = %v; want an error wrapping ENOTSUP", name, err)
		}
	}
	// Neither directory was touched: still exactly the file "f", with its own content.
	for dir, want := range map[string]string{rootDir: "x", cwd: "y"} {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 1 || entries[0].Name() != "f" {
			t.Errorf("directory %s changed: %v, %v", dir, entries, err)
		}
		if b, err := os.ReadFile(filepath.Join(dir, "f")); err != nil || string(b) != want {
			t.Errorf("%s/f = %q, %v; want %q", dir, b, err, want)
		}
	}
}

// TestRemoveAllNonEmptyDarwinPPC64 checks RemoveAll on a populated tree. With
// the descriptor-based removeall_at.go this fails on Leopard (ENOTSUP from the
// *at calls); removeall_noat.go is used instead.
func TestRemoveAllNonEmptyDarwinPPC64(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	tree := filepath.Join(base, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "a", "b", "c"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"x", "a/y", "a/b/z", "a/b/c/w"} {
		if err := os.WriteFile(filepath.Join(tree, f), []byte(f), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A symlink must be removed, not followed.
	if err := os.Symlink(outside, filepath.Join(tree, "a", "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(tree); err != nil {
		t.Fatalf("RemoveAll(non-empty tree) = %v", err)
	}
	if _, err := os.Lstat(tree); !os.IsNotExist(err) {
		t.Errorf("tree still exists: %v", err)
	}
	if b, err := os.ReadFile(outside); err != nil || string(b) != "keep" {
		t.Errorf("symlink target was touched: %q, %v", b, err)
	}
}

// TestCloseOnExecDarwinPPC64 checks that descriptors opened by os carry
// FD_CLOEXEC even though Leopard has no O_CLOEXEC: supportsCloseOnExec is
// false there, so os follows each open with fcntl(F_SETFD, FD_CLOEXEC).
func TestCloseOnExecDarwinPPC64(t *testing.T) {
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, "f"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	for name, fd := range map[string]uintptr{"os.Create": f.Fd(), "os.Open(dir)": d.Fd()} {
		flags, err := unix.Fcntl(int(fd), syscall.F_GETFD, 0)
		if err != nil {
			t.Fatalf("%s: fcntl(F_GETFD): %v", name, err)
		}
		if flags&syscall.FD_CLOEXEC == 0 {
			t.Errorf("%s: FD_CLOEXEC not set on fd %d", name, fd)
		}
	}
}
