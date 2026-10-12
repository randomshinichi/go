// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os_test

// TestReaddirDecoyCwdDarwinPPC is the ppc32 counterpart of
// dir_darwin_ppc64_test.go: the directory-handle semantics that were exercised
// on ppc64 only.
//
// On Leopard an entry of a directory open by path is stat'ed by resolving the
// descriptor to its path with F_GETPATH and then lstat'ing below that path
// (statat_darwin_ppc.go -> syscall.lstatatByPath). The process working directory
// is the one wrong answer that would still look plausible: it exists, and it can
// hold entries of the same names. So the working directory here contains "f"
// with other contents and a symlink "link" pointing elsewhere, and the
// assertions are on content (5 bytes, not 1) and on identity, not merely on the
// names being findable. A resolution that fell back to the working directory
// would produce wrong FileInfos rather than a missing-file error.
//
// TestReaddirStatDarwinPPC already covers entry-by-entry agreement with Lstat
// for a directory whose entries do not exist in the working directory; this one
// makes the wrong answer available, which is the case that distinguishes
// "resolved the descriptor" from "happened to find nothing to confuse it".

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReaddirDecoyCwdDarwinPPC(t *testing.T) {
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
	if fi := got["f"]; fi == nil || !fi.Mode().IsRegular() || fi.Size() != 5 {
		t.Errorf("f = %v; want the 5-byte regular file in %s, not the 1-byte file of the same name in the working directory %s", fi, dir, cwd)
	}
	if fi := got["link"]; fi == nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link = %v; want the symbolic link itself, not its target", fi)
	}
	if fi := got["sub"]; fi == nil || !fi.IsDir() {
		t.Errorf("sub = %v; want a directory", fi)
	}
	for name, fi := range got {
		want, err := os.Lstat(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if fi == nil || !os.SameFile(fi, want) {
			t.Errorf("%s = %v; want the entry of %s (%v)", name, fi, dir, want)
		}
	}

	// File.ReadDir leaves the entry's FileInfo lazy: Info() stats by path on
	// its own, so it is a second route to the same resolution.
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
		fi, err := e.Info()
		if err != nil {
			t.Errorf("%s: Info: %v", e.Name(), err)
			continue
		}
		want, err := os.Lstat(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(fi, want) {
			t.Errorf("%s: Info = %v; want the entry of %s (%v), not the one in the working directory %s", e.Name(), fi, dir, want, cwd)
		}
	}
}
