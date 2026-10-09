// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix_test

import (
	"fmt"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func ppcAtTestDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("go-at-ppc-%d-%s", os.Getpid(), t.Name()))
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(filepath.Join(dir, "file"))
		if err := os.Remove(dir); err != nil {
			t.Errorf("Remove test directory: %v", err)
		}
	})
	return dir
}

func TestAtDarwinPPCRealDirfdReturnsENOTSUP(t *testing.T) {
	dir := ppcAtTestDir(t)
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	fd, err := syscall.Open(dir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fd)

	var stat syscall.Stat_t
	checks := []struct {
		name string
		call func() error
	}{
		{"Openat", func() error { _, err := unix.Openat(fd, "file", syscall.O_RDONLY, 0); return err }},
		{"Unlinkat", func() error { return unix.Unlinkat(fd, "file", 0) }},
		{"Readlinkat", func() error { _, err := unix.Readlinkat(fd, "file", make([]byte, 32)); return err }},
		{"Mkdirat", func() error { return unix.Mkdirat(fd, "sub", 0700) }},
		{"Fchmodat", func() error { return unix.Fchmodat(fd, "file", 0600, 0) }},
		{"Fchownat", func() error { return unix.Fchownat(fd, "file", -1, -1, 0) }},
		{"Renameat", func() error { return unix.Renameat(fd, "file", fd, "renamed") }},
		{"Linkat", func() error { return unix.Linkat(fd, "file", fd, "linked", 0) }},
		{"Symlinkat", func() error { return unix.Symlinkat("file", fd, "link") }},
		{"Fstatat", func() error { return unix.Fstatat(fd, "file", &stat, 0) }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.call(); err != syscall.ENOTSUP {
				t.Errorf("got %v, want ENOTSUP", err)
			}
		})
	}
}

func TestOpenatDarwinPPCAtFDCWDFallsBackToPath(t *testing.T) {
	dir := ppcAtTestDir(t)
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Openat(unix.AT_FDCWD, file, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Close(fd); err != nil {
		t.Fatal(err)
	}
}
