// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"os"
	"path/filepath"
	"testing"
)

// No TempDir or RemoveAll: a failure here measures directory-entry statting,
// independently of Leopard's RemoveAll cleanup route.
func TestReaddirStatDarwinPPC(t *testing.T) {
	dir, err := os.MkdirTemp("", "irfvstat-readdir-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, name := range []string{"link", "file", "sub"} {
			if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
				t.Errorf("cleanup %s: %v", name, err)
			}
		}
		if err := os.Remove(dir); err != nil {
			t.Errorf("cleanup directory: %v", err)
		}
	}()
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("file", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	check := func(t *testing.T, infos []os.FileInfo) {
		t.Helper()
		if len(infos) != 3 {
			t.Fatalf("got %d entries; want 3", len(infos))
		}
		seen := make(map[string]bool)
		for _, info := range infos {
			want, err := os.Lstat(filepath.Join(dir, info.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if seen[info.Name()] || !os.SameFile(info, want) || info.Mode() != want.Mode() || info.Size() != want.Size() {
				t.Errorf("entry %s = %v; want Lstat %v", info.Name(), info, want)
			}
			seen[info.Name()] = true
		}
		for _, name := range []string{"file", "sub", "link"} {
			if !seen[name] {
				t.Errorf("missing entry %s", name)
			}
		}
	}
	t.Run("Readdir", func(t *testing.T) {
		f, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		infos, err := f.Readdir(-1)
		if err != nil {
			t.Fatalf("File.Readdir = %v", err)
		}
		check(t, infos)
	})
	t.Run("ReaddirAfterRename", func(t *testing.T) {
		f, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		moved := dir + "-moved"
		if err := os.Rename(dir, moved); err != nil {
			t.Fatal(err)
		}
		dir = moved
		infos, err := f.Readdir(-1)
		if err != nil {
			t.Fatalf("File.Readdir after rename = %v", err)
		}
		check(t, infos)
	})
	for _, method := range []string{"FileReadDir", "ReadDir"} {
		t.Run(method, func(t *testing.T) {
			var entries []os.DirEntry
			var err error
			if method == "ReadDir" {
				entries, err = os.ReadDir(dir)
			} else {
				var f *os.File
				f, err = os.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				entries, err = f.ReadDir(-1)
			}
			if err != nil {
				t.Fatalf("%s = %v", method, err)
			}
			// d_type may let ReadDir itself succeed without statting. Info must
			// still work, including lstat (not stat) semantics for the symlink.
			var infos []os.FileInfo
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					t.Fatalf("%s entry %s Info = %v", method, entry.Name(), err)
				}
				infos = append(infos, info)
			}
			check(t, infos)
		})
	}
}
