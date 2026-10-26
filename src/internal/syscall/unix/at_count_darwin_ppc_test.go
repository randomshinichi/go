// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcount_test

package unix_test

import (
	"fmt"
	"internal/syscall/unix"
	"runtime"
	"syscall"
	"testing"
	_ "unsafe" // for linkname
)

//go:linkname at349CountRead syscall.at349CountRead
func at349CountRead() uint32

func at349ZeroError(before uint32) error {
	if n := at349CountRead() - before; n != 0 {
		return fmt.Errorf("syscall-349 count = %d; want ZERO", n)
	}
	return nil
}

func atBegin349Observation(t *testing.T) func() {
	t.Helper()
	before := at349CountRead()
	return func() {
		t.Helper()
		t.Logf("full-operation syscall-349 count=%d", at349CountRead()-before)
		if err := at349ZeroError(before); err != nil {
			t.Error(err)
		}
	}
}

func TestAt349CountRoutesDarwinPPC(t *testing.T) {
	dead := uintptr(atDeadFD(t))
	for _, call := range []struct {
		name string
		run  func()
	}{
		{"Syscall", func() { syscall.Syscall(349, dead, 0, 0) }},
		{"Syscall6", func() { syscall.Syscall6(349, dead, 0, 0, 0, 0, 0) }},
		{"RawSyscall", func() { syscall.RawSyscall(349, dead, 0, 0) }},
		{"RawSyscall6", func() { syscall.RawSyscall6(349, dead, 0, 0, 0, 0, 0) }},
	} {
		before := at349CountRead()
		call.run()
		if n := at349CountRead() - before; n != 1 {
			t.Errorf("%s count=%d want=1", call.name, n)
		}
	}
	_, _, fd := atTree(t)
	before := at349CountRead()
	var st syscall.Stat_t
	if err := unix.Fstatat(fd, "file", &st, 0); err != nil {
		t.Fatal(err)
	}
	if n := at349CountRead() - before; n != 2 {
		t.Fatalf("compound count=%d want=2", n)
	}
}

// The helper M installs and clears its private cwd. Both helper and caller end
// with their original cwd, yet the zero-call oracle MUST reject its two traps.
func TestAt349InstallClearMutantDarwinPPC(t *testing.T) {
	_, _, fd := atTree(t)
	var callerBefore, callerAfter syscall.Stat_t
	if err := syscall.Stat(".", &callerBefore); err != nil {
		t.Fatal(err)
	}
	before := at349CountRead()
	done := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var original, final syscall.Stat_t
		if err := syscall.Stat(".", &original); err != nil {
			done <- err
			return
		}
		if err := rawPthreadFchdir(fd); err != nil {
			done <- err
			return
		}
		if err := rawPthreadFchdir(-1); err != nil {
			done <- err
			return
		}
		if err := syscall.Stat(".", &final); err != nil {
			done <- err
			return
		}
		if original.Dev != final.Dev || original.Ino != final.Ino {
			done <- fmt.Errorf("helper cwd not restored")
			return
		}
		done <- nil
	}()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := syscall.Stat(".", &callerAfter); err != nil {
		t.Fatal(err)
	}
	if callerBefore.Dev != callerAfter.Dev || callerBefore.Ino != callerAfter.Ino {
		t.Fatal("caller cwd changed")
	}
	err := at349ZeroError(before)
	t.Logf("install/clear mutant: final cwd clean; count=%d; oracle=%v", at349CountRead()-before, err)
	if n := at349CountRead() - before; n != 2 || err == nil {
		t.Fatal("install/clear mutant not detected")
	}
}
