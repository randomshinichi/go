// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall_test

import (
	"errors"
	"runtime"
	"syscall"
	"testing"
)

// sameMDarwinPPC runs f on a goroutine locked to its thread, which is the only
// way to ask whether that thread holds a private working directory afterwards.
func sameMDarwinPPC(t *testing.T, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)
		f()
	}()
	<-done
}

// TestAtCompoundSameMClearDarwinPPC is the cleanup oracle for the native
// compound, run through the direct test export rather than the production
// dispatch path, which uses a helper goroutine.
//
// Leopard's private cwd (syscall 349) is per-thread state, and getcwd resolves
// against it (bsd/vfs/kpi_vfs.c:1240-1257). So on a goroutine locked to one
// thread, syscall.Getwd() reports that thread's private cwd when one is
// installed, and the process cwd when none is. That makes Getwd the observation:
// after a compound call on this same goroutine it must still report the process
// directory, which fails if the routine installed the anchor and did not clear
// it. A helper-goroutine call could not show this, and a process-wide chdir
// would move the endpoint for everyone, so both alternatives are excluded by
// construction.
//
// The second half of the design's G2 control -- raw 349(-1) must return EBADF
// for the production routine and 0 for a deliberate skip-clear variant -- is
// asserted here for production (below) and by
// TestAtCompoundSkipClearMutantDarwinPPC for a mutant built with the
// goppc_atcompound_mutant tag. The mutant is a separate assembly copy used only
// by that test build; the production routine carries no switch for it.
func TestAtCompoundSameMClearDarwinPPC(t *testing.T) {
	anchorDir := t.TempDir()
	decoyDir := t.TempDir()
	if err := syscall.Mkdir(anchorDir+"/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeFileDarwinPPC(anchorDir+"/file", "anchor"); err != nil {
		t.Fatal(err)
	}
	if err := writeFileDarwinPPC(decoyDir+"/file", "decoy"); err != nil {
		t.Fatal(err)
	}
	prev, err := syscall.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Chdir(decoyDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { syscall.Chdir(prev) })
	decoy, err := syscall.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	anchorFd, err := syscall.Open(anchorDir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(anchorFd)
	fileFd, err := syscall.Open(anchorDir+"/file", syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fileFd)

	// The observation must be live before it can clear anything: with no private
	// cwd installed, the thread sees the process directory. (This also fails if
	// a previous test in this process leaked a private cwd onto this thread.)
	cwd := func(t *testing.T) string {
		t.Helper()
		got, err := syscall.Getwd()
		if err != nil {
			t.Fatalf("Getwd: %v", err)
		}
		return got
	}
	sameMDarwinPPC(t, func() {
		if got := cwd(t); got != decoy {
			t.Fatalf("before any compound call this thread's cwd is %q; want the process directory %q", got, decoy)
		}
		// With no private cwd installed, 349(-1) reports EBADF. This is the
		// second, independent observation of "there is nothing to clear on this
		// thread": Getwd cannot distinguish a private cwd equal to the process
		// directory, and raw 349(-1) can.
		if err := syscall.AtPthreadFchdirForTest(-1); !errors.Is(err, syscall.EBADF) {
			t.Fatalf("349(-1) before any compound call = %v; want EBADF (this thread must start with no private cwd)", err)
		}

		// A served operation: enter, identity, lstat64, clear.
		enter, ident, op := syscall.AtCompoundDirectForTest(anchorFd, "file")
		if enter != 0 || ident != 0 || op != 0 {
			t.Errorf("compound lstat64 = enter %v, ident %v, op %v; want 0, 0, 0 (a failure here also means the native call put its arguments in the wrong registers)", enter, ident, op)
		}
		if got := cwd(t); got != decoy {
			t.Errorf("after a successful compound call the thread's cwd is %q; want %q (the private cwd was not cleared on this M)", got, decoy)
		}
		if err := syscall.AtPthreadFchdirForTest(-1); !errors.Is(err, syscall.EBADF) {
			t.Errorf("349(-1) after a successful compound call = %v; want EBADF: the routine must have cleared the private cwd it installed", err)
		}

		// An operation failure must still clear.
		enter, ident, op = syscall.AtCompoundDirectForTest(anchorFd, "missing")
		if enter != 0 || ident != 0 {
			t.Errorf("compound of a missing name = enter %v, ident %v; want 0, 0", enter, ident)
		}
		if op == 0 {
			t.Errorf("compound lstat64 of a missing name succeeded; want an operation error")
		}
		if got := cwd(t); got != decoy {
			t.Errorf("after a failed operation the thread's cwd is %q; want %q", got, decoy)
		}
		if err := syscall.AtPthreadFchdirForTest(-1); !errors.Is(err, syscall.EBADF) {
			t.Errorf("349(-1) after a failed operation = %v; want EBADF: an operation failure must still clear", err)
		}

		// An enter failure changes nothing and must not attempt a clear.
		enter, _, _ = syscall.AtCompoundDirectForTest(fileFd, "file")
		if enter == 0 {
			t.Errorf("compound anchored at a non-directory failed to report an enter error")
		}
		if got := cwd(t); got != decoy {
			t.Errorf("after a failed enter the thread's cwd is %q; want %q", got, decoy)
		}
	})
}

// writeFileDarwinPPC creates path with contents, using only package syscall so
// this test needs no os import (which would import syscall).
func writeFileDarwinPPC(path, contents string) error {
	fd, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := syscall.Write(fd, []byte(contents)); err != nil {
		syscall.Close(fd)
		return err
	}
	return syscall.Close(fd)
}
