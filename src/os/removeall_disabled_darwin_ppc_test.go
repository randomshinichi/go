// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompat_disabled

package os_test

// The os-level half of the disabled-backend dispatch control, compiled only with
// the goppc_atcompat_disabled build tag (see internal/syscall/unix/
// at_disabled_darwin_ppc_test.go for the Unix-layer half and the reason the tag
// exists). Two promises are at stake when the anchored backend is off, which is
// the state of every Darwin/ppc kernel other than the measured 9.8.x:
//
//   - ordinary RemoveAll keeps working, through the path-based algorithm, so
//     t.TempDir teardown and every caller does not break on such a kernel;
//   - Root refuses rather than approximating: Root.RemoveAll returns ENOTSUP and
//     removes nothing, because a path-based substitute inside a Root would give
//     up the confinement the Root exists for.

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestRemoveAllFallbackWhenBackendDisabledDarwinPPC requires the path-based
// RemoveAll to still remove a tree, including a symlink and a nested directory,
// while the anchored backend is disabled.
func TestRemoveAllFallbackWhenBackendDisabledDarwinPPC(t *testing.T) {
	base := t.TempDir()
	tree := filepath.Join(base, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"x", "a/y", "a/b/z"} {
		if err := os.WriteFile(filepath.Join(tree, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(outside, "keep")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(tree, "link-out")); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(tree); err != nil {
		t.Fatalf("RemoveAll with the anchored backend disabled: %v", err)
	}
	if _, err := os.Lstat(tree); !os.IsNotExist(err) {
		t.Errorf("%s: %v; want the tree removed by the path-based algorithm", tree, err)
	}
	if b, err := os.ReadFile(keep); err != nil || string(b) != "keep" {
		t.Errorf("%s = %q, %v; want the symlink's target untouched", keep, b, err)
	}
}

// TestRootRefusesWhenBackendDisabledDarwinPPC requires a Root on this kernel to
// refuse its descriptor-relative operations and to leave the tree alone.
func TestRootRefusesWhenBackendDisabledDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	tree := filepath.Join(dir, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(tree, "f")
	if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := os.OpenRoot(tree)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer root.Close()

	if err := root.RemoveAll("sub"); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Root.RemoveAll with the anchored backend disabled = %v; want ENOTSUP", err)
	}
	if fi, err := os.Lstat(filepath.Join(tree, "sub")); err != nil || !fi.IsDir() {
		t.Errorf("%s/sub: %v, %v; want the directory left in place", tree, fi, err)
	}
	if err := root.Chtimes("f", time.Unix(1300000000, 0), time.Unix(1300000001, 0)); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Root.Chtimes with the anchored backend disabled = %v; want ENOTSUP", err)
	}
	if err := root.Chmod("f", 0o600); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Root.Chmod with the anchored backend disabled = %v; want ENOTSUP", err)
	}
	if err := root.Remove("f"); !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("Root.Remove with the anchored backend disabled = %v; want ENOTSUP", err)
	}
	if b, err := os.ReadFile(file); err != nil || string(b) != "data" {
		t.Errorf("%s = %q, %v; want the file unchanged", file, b, err)
	}
	if fi, err := os.Lstat(file); err != nil || fi.Mode().Perm() != 0o644 {
		t.Errorf("%s mode: %v, %v; want the untouched 644", file, fi, err)
	}
}
