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
// The two trees hold entries with the same names and different contents, so an
// operation that fell back to the working directory would be visible as a change
// to "f" == "cwd". Each step gets a fresh root tree: several of them mutate it,
// and sharing one made the expectations depend on the order the steps ran in.
//
// RemoveAll stays refused on this platform until the confined directory walk is
// retired (docs/leopard-at-compat-design.md): it is listed here with its refusal
// so that turning it into a working call is a visible edit, and so that the
// no-mutation half is asserted now.
func TestRootOpsConfinedDarwinPPC(t *testing.T) {
	cwd := t.TempDir()
	confinedTree(t, cwd, "cwd")
	t.Chdir(cwd)
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	freshRoot := func(t *testing.T) (*os.Root, string) {
		t.Helper()
		rootDir := t.TempDir()
		confinedTree(t, rootDir, "root")
		r, err := os.OpenRoot(rootDir)
		if err != nil {
			t.Fatalf("OpenRoot(%q): %v", rootDir, err)
		}
		t.Cleanup(func() { r.Close() })
		return r, rootDir
	}

	steps := []struct {
		name string
		run  func(r *os.Root) error
		want error
		// check asserts the step's effect inside the root, and is nil when the
		// step is meant to leave it untouched.
		check func(t *testing.T, rootDir string)
	}{
		{"Open", func(r *os.Root) error {
			f, err := r.Open("f")
			if f != nil {
				f.Close()
			}
			return err
		}, nil, nil},
		{"OpenRoot", func(r *os.Root) error {
			sub, err := r.OpenRoot("sub")
			if sub != nil {
				sub.Close()
			}
			return err
		}, nil, nil},
		{"Stat", func(r *os.Root) error { _, err := r.Stat("f"); return err }, nil, nil},
		{"Lstat", func(r *os.Root) error { _, err := r.Lstat("f"); return err }, nil, nil},
		{"Mkdir", func(r *os.Root) error { return r.Mkdir("m", 0o755) }, nil,
			func(t *testing.T, rootDir string) {
				if fi, err := os.Lstat(filepath.Join(rootDir, "m")); err != nil || !fi.IsDir() {
					t.Errorf("%s/m: %v, %v; want the directory created inside the root", rootDir, fi, err)
				}
			}},
		{"Remove", func(r *os.Root) error { return r.Remove("sub") }, nil,
			func(t *testing.T, rootDir string) {
				if _, err := os.Lstat(filepath.Join(rootDir, "sub")); !os.IsNotExist(err) {
					t.Errorf("%s/sub: %v; want it removed inside the root", rootDir, err)
				}
			}},
		{"Rename", func(r *os.Root) error { return r.Rename("f", "g") }, nil,
			func(t *testing.T, rootDir string) {
				if b, err := os.ReadFile(filepath.Join(rootDir, "g")); err != nil || string(b) != "root" {
					t.Errorf("%s/g = %q, %v; want the root's own file renamed to g", rootDir, b, err)
				}
				if _, err := os.Lstat(filepath.Join(rootDir, "f")); !os.IsNotExist(err) {
					t.Errorf("%s/f: %v; want the rename to have moved it", rootDir, err)
				}
			}},
		{"Symlink", func(r *os.Root) error { return r.Symlink("g", "s") }, nil,
			func(t *testing.T, rootDir string) {
				fi, err := os.Lstat(filepath.Join(rootDir, "s"))
				if err != nil || fi.Mode()&os.ModeSymlink == 0 {
					t.Errorf("%s/s: %v, %v; want the symlink created inside the root", rootDir, fi, err)
				}
			}},
		{"RemoveAll", func(r *os.Root) error { return r.RemoveAll("sub") }, syscall.ENOTSUP,
			func(t *testing.T, rootDir string) {
				if fi, err := os.Lstat(filepath.Join(rootDir, "sub")); err != nil || !fi.IsDir() {
					t.Errorf("%s/sub: %v, %v; the refusal must not remove anything", rootDir, fi, err)
				}
			}},
	}

	for _, step := range steps {
		r, rootDir := freshRoot(t)
		err := step.run(r)
		if step.want == nil {
			if err != nil {
				t.Errorf("Root.%s = %v; want success", step.name, err)
			}
		} else if !errors.Is(err, step.want) {
			t.Errorf("Root.%s = %v; want an error wrapping %v", step.name, err, step.want)
		}
		if after, err := os.Getwd(); err != nil || after != before {
			t.Errorf("Root.%s changed cwd: %q, %v; want %q", step.name, after, err, before)
		}
		if step.check != nil {
			step.check(t, rootDir)
		}

		// The decoy working directory never changes, whatever the step did
		// inside the root.
		if b, err := os.ReadFile(filepath.Join(cwd, "f")); err != nil || string(b) != "cwd" {
			t.Errorf("after Root.%s, %s/f = %q, %v; want the working directory untouched (%q)", step.name, cwd, b, err, "cwd")
		}
		if fi, err := os.Lstat(filepath.Join(cwd, "sub")); err != nil || !fi.IsDir() {
			t.Errorf("after Root.%s, %s/sub: %v, %v; want the working directory untouched", step.name, cwd, fi, err)
		}
		if _, err := os.Lstat(filepath.Join(cwd, "g")); !os.IsNotExist(err) {
			t.Errorf("after Root.%s, %s/g: %v; want the working directory untouched", step.name, cwd, err)
		}
		if _, err := os.Lstat(filepath.Join(cwd, "s")); !os.IsNotExist(err) {
			t.Errorf("after Root.%s, %s/s: %v; want the working directory untouched", step.name, cwd, err)
		}
		if _, err := os.Lstat(filepath.Join(cwd, "m")); !os.IsNotExist(err) {
			t.Errorf("after Root.%s, %s/m: %v; want the working directory untouched", step.name, cwd, err)
		}
	}
}
