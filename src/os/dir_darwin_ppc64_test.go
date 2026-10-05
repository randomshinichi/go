// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package os_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReaddirDarwinPPC64 checks (*File).Readdir on Mac OS X 10.5, which stats
// every entry through a path resolved from the directory's descriptor
// (statat_darwin_ppc64.go). Before that substitution Readdir failed with
// ENOTSUP, because it reached fstatat with the directory's own descriptor.
// The working directory holds entries of the same names with other contents, so
// a lookup that fell back to it would be seen.
func TestReaddirDarwinPPC64(t *testing.T) {
	dir, cwd := t.TempDir(), t.TempDir()
	write := func(base, name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(base, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(dir, "f", "12345")
	write(cwd, "f", "1")
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

	d, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	infos, err := d.Readdir(-1)
	if err != nil {
		t.Fatalf("Readdir: %v", err)
	}
	got := map[string]os.FileInfo{}
	for _, fi := range infos {
		got[fi.Name()] = fi
	}
	if len(got) != 3 {
		t.Fatalf("Readdir returned %d entries (%v); want f, link, sub", len(got), got)
	}
	if fi := got["f"]; fi == nil || fi.Size() != 5 || !fi.Mode().IsRegular() {
		t.Errorf("f = %v; want a regular file of 5 bytes", fi)
	}
	if fi := got["link"]; fi == nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link = %v; want the symbolic link itself, not its target", fi)
	}
	if fi := got["sub"]; fi == nil || !fi.IsDir() {
		t.Errorf("sub = %v; want a directory", fi)
	}

	// File.ReadDir and Info() on the entries, which stat by path on their own.
	d2, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer d2.Close()
	entries, err := d2.ReadDir(-1)
	if err != nil || len(entries) != 3 {
		t.Fatalf("ReadDir = %v, %v; want 3 entries", entries, err)
	}
	for _, e := range entries {
		if _, err := e.Info(); err != nil {
			t.Errorf("%s: Info: %v", e.Name(), err)
		}
	}
}
