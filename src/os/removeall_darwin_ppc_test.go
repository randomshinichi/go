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

// TestRootOpsConfinedDarwinPPC keeps the fd-relative boundary honest from the
// other side: a Root's operations must act on the root's own entries and never
// on the process working directory, and they must do that without changing the
// working directory, even transiently (the endpoint comparison below cannot see
// a swap that is restored before the call returns; the chdir recorder in
// root_darwin_ppc_test.go does).
//
// The two trees have entries with the same names and different contents, so an
// operation that fell back to the working directory would be visible as a
// change to "f" == "cwd". RemoveAll stays refused on this platform until the
// confined directory walk is retired (docs/leopard-at-compat-design.md); the
// refusal is asserted here so that turning it into a working call is a visible
// edit.
func TestRootOpsConfinedDarwinPPC(t *testing.T) {
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
		"Mkdir":   func() error { return r.Mkdir("m", 0o755) },
		"Remove":  func() error { return r.Remove("m") },
		"Stat":    func() error { _, err := r.Stat("f"); return err },
		"Lstat":   func() error { _, err := r.Lstat("f"); return err },
		"Rename":  func() error { return r.Rename("f", "g") },
		"Symlink": func() error { return r.Symlink("g", "s") },
	}
	for name, op := range ops {
		if err := op(); err != nil {
			t.Errorf("Root.%s = %v; want success", name, err)
		}
		if after, err := os.Getwd(); err != nil || after != before {
			t.Errorf("Root.%s changed cwd: %q, %v; want %q", name, after, err, before)
		}
	}
	if err := r.RemoveAll("sub"); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Root.RemoveAll = %v; want an error wrapping ENOTSUP (the confined directory walk is not retired yet)", err)
	}

	// The root's own tree took the changes; the working directory did not.
	if b, err := os.ReadFile(filepath.Join(rootDir, "g")); err != nil || string(b) != "root" {
		t.Errorf("%s/g = %q, %v; want the root's own file renamed to g", rootDir, b, err)
	}
	if _, err := os.Lstat(filepath.Join(rootDir, "s")); err != nil {
		t.Errorf("%s/s: %v; want the symlink created inside the root", rootDir, err)
	}
	if _, err := os.Lstat(filepath.Join(rootDir, "m")); !os.IsNotExist(err) {
		t.Errorf("%s/m: %v; want it removed inside the root", rootDir, err)
	}
	for dir, want := range map[string]string{rootDir: "root", cwd: "cwd"} {
		if b, err := os.ReadFile(filepath.Join(dir, "f")); err != nil || string(b) != want {
			t.Errorf("%s/f = %q, %v; want %q", dir, b, err, want)
		}
	}
	if fi, err := os.Lstat(filepath.Join(cwd, "sub")); err != nil || !fi.IsDir() {
		t.Errorf("%s/sub changed: %v, %v", cwd, fi, err)
	}
	if _, err := os.Lstat(filepath.Join(cwd, "g")); !os.IsNotExist(err) {
		t.Errorf("%s/g: %v; want the working directory untouched by Root.Rename", cwd, err)
	}
	if _, err := os.Lstat(filepath.Join(cwd, "s")); !os.IsNotExist(err) {
		t.Errorf("%s/s: %v; want the working directory untouched by Root.Symlink", cwd, err)
	}
}
