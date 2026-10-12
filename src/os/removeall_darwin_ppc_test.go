// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestRemoveAllNonEmptyDarwinPPC checks the path-based Leopard implementation,
// including directory symlinks whose targets must survive removal.
func TestRemoveAllNonEmptyDarwinPPC(t *testing.T) {
	base := t.TempDir()
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(outside, "keep")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
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
	if err := os.Symlink(outside, filepath.Join(tree, "a", "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(tree); err != nil {
		t.Fatalf("RemoveAll(non-empty tree) = %v", err)
	}
	if _, err := os.Lstat(tree); !os.IsNotExist(err) {
		t.Errorf("tree still exists: %v", err)
	}
	// Also exercise a symlink as the top-level argument, not just an entry.
	link := filepath.Join(base, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(link); err != nil {
		t.Fatalf("RemoveAll(directory symlink) = %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Errorf("link still exists: %v", err)
	}
	if b, err := os.ReadFile(keep); err != nil || string(b) != "keep" {
		t.Errorf("symlink target was touched: %q, %v", b, err)
	}
}

// TestRootNotSupportedDarwinPPC keeps the fd-relative boundary honest. Opening
// a Root handle may succeed, but confined operations must report ENOTSUP, not
// act on the process working directory (or temporarily swap it).
func TestRootNotSupportedDarwinPPC(t *testing.T) {
	rootDir, cwd := t.TempDir(), t.TempDir()
	for dir, content := range map[string]string{rootDir: "root", cwd: "cwd"} {
		if err := os.WriteFile(filepath.Join(dir, "f"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(cwd)
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()
	ops := map[string]func() error{
		"Open": func() error {
			f, err := r.Open("f")
			if f != nil {
				f.Close()
			}
			return err
		},
		"OpenRoot": func() error {
			sub, err := r.OpenRoot("sub")
			if sub != nil {
				sub.Close()
			}
			return err
		},
		"Mkdir":     func() error { return r.Mkdir("m", 0o755) },
		"Remove":    func() error { return r.Remove("f") },
		"RemoveAll": func() error { return r.RemoveAll("sub") },
		"Stat":      func() error { _, err := r.Stat("f"); return err },
		"Lstat":     func() error { _, err := r.Lstat("f"); return err },
		"Rename":    func() error { return r.Rename("f", "g") },
		"Symlink":   func() error { return r.Symlink("f", "s") },
	}
	for name, op := range ops {
		if err := op(); !errors.Is(err, syscall.ENOTSUP) {
			t.Errorf("Root.%s = %v; want an error wrapping ENOTSUP", name, err)
		}
		if after, err := os.Getwd(); err != nil || after != before {
			t.Errorf("Root.%s changed cwd: %q, %v; want %q", name, after, err, before)
		}
	}
	for dir, want := range map[string]string{rootDir: "root", cwd: "cwd"} {
		// Readdirnames needs no statat support; Readdir is a separate port gap.
		f, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		names, err := f.Readdirnames(-1)
		f.Close()
		if err != nil || len(names) != 2 {
			t.Errorf("directory %s changed: %v, %v", dir, names, err)
		}
		if b, err := os.ReadFile(filepath.Join(dir, "f")); err != nil || string(b) != want {
			t.Errorf("%s/f = %q, %v; want %q", dir, b, err, want)
		}
		if fi, err := os.Lstat(filepath.Join(dir, "sub")); err != nil || !fi.IsDir() {
			t.Errorf("%s/sub changed: %v, %v", dir, fi, err)
		}
	}
}
