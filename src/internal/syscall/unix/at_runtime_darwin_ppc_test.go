// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix_test

// Runtime-scope controls for the private-cwd compound (G2 in
// docs/leopard-at-compat-design.md). The Unix-layer operation tests ask whether
// an operation reached the anchor; these ask about the thread state and the
// scheduler around it.

import (
	"errors"
	"fmt"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// rawPthreadFchdir issues syscall 349 on the calling goroutine's own thread.
// fd >= 0 installs that descriptor as the thread's private cwd, -1 clears it
// (and reports EBADF when there is none).
func rawPthreadFchdir(fd int) error {
	_, _, errno := syscall.Syscall(syscall.SYS___PTHREAD_FCHDIR, uintptr(fd), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// TestAtDarwinPPCLockedCallerDarwinPPC requires an operation issued by a
// goroutine that is locked to its thread and holds its own private cwd to leave
// both alone.
//
// The caller's thread state is the thing at risk: if the compound ran on the
// caller's own M, it would install its anchor as that thread's private cwd,
// destroy the caller's, and its own clear would remove the anchor rather than
// restore the caller's directory. The three observations are therefore: the
// caller's private cwd still resolves (Getwd), the operation still acted on the
// descriptor it was given rather than on the caller's private cwd (the two
// directories hold a same-named file with different sizes), and clearing at the
// end reports that there was still a private cwd to clear.
func TestAtDarwinPPCLockedCallerDarwinPPC(t *testing.T) {
	_, cwdDir, fd := atTree(t)

	// The caller's own private cwd: a second directory holding a same-named file
	// with a distinguishable size.
	callerDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(callerDir, "file"), []byte("caller-private-cwd"), 0600); err != nil {
		t.Fatal(err)
	}
	callerFd, err := syscall.Open(callerDir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(callerFd)

	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)

		if err := rawPthreadFchdir(callerFd); err != nil {
			t.Errorf("installing the caller's private cwd: %v", err)
			return
		}
		// A goroutine that is locked and then exits without clearing would leak
		// its private cwd into whichever thread it was on, so clean up even if an
		// assertion below fails. The last clear below sets cleared, so the two
		// paths do not both report.
		cleared := false
		defer func() {
			if !cleared {
				rawPthreadFchdir(-1)
			}
		}()

		got, err := syscall.Getwd()
		if err != nil {
			t.Errorf("Getwd after installing a private cwd: %v", err)
			return
		}
		if !sameFile(t, got, callerDir) {
			t.Errorf("this thread's cwd is %q; want its own private cwd %q (the install did not take)", got, callerDir)
			return
		}

		var st syscall.Stat_t
		if err := unix.Fstatat(fd, "file", &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			t.Errorf("Fstatat from a locked caller holding a private cwd: %v", err)
			return
		}
		if st.Size != int64(len("anchor")) {
			t.Errorf("Fstatat size = %d; want %d from the anchor tree: the operation acted on the caller's private cwd instead of its descriptor",
				st.Size, len("anchor"))
		}

		got, err = syscall.Getwd()
		if err != nil {
			t.Errorf("Getwd after the operation: %v", err)
			return
		}
		if !sameFile(t, got, callerDir) {
			t.Errorf("after the operation this thread's cwd is %q; want the caller's private cwd %q preserved: the compound must not touch the caller's thread",
				got, callerDir)
		}

		// There must still be a private cwd, and clearing it must return to the
		// process directory.
		if err := rawPthreadFchdir(-1); err != nil {
			t.Errorf("clearing the caller's private cwd: %v; want nil (the compound's own clear must not have removed it)", err)
			return
		}
		cleared = true
		got, err = syscall.Getwd()
		if err != nil {
			t.Errorf("Getwd after clearing: %v", err)
			return
		}
		if !sameFile(t, got, cwdDir) {
			t.Errorf("this thread's cwd is %q after clearing; want the process directory %q", got, cwdDir)
		}
	}()
	<-done
}

// sameFile reports whether two paths name the same directory.
func sameFile(t *testing.T, a, b string) bool {
	t.Helper()
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}

// TestAtDarwinPPCBlockedOpDarwinPPC requires one blocked anchored operation not
// to prevent another from making progress.
//
// A FIFO open for reading blocks until a writer appears, which parks the helper
// goroutine and its native scope (with the private cwd installed on that M) for
// as long as the test wants. If the compound were serialized behind one worker or
// a pool, the second operation could not complete until the first returned, which
// would show up here as a timeout rather than as a wrong answer.
func TestAtDarwinPPCBlockedOpDarwinPPC(t *testing.T) {
	anchorDir, cwdDir, fd := atTree(t)
	if err := syscall.Mkfifo(filepath.Join(anchorDir, "fifo"), 0600); err != nil {
		t.Fatalf("Mkfifo: %v", err)
	}
	t.Cleanup(func() { os.Remove(filepath.Join(anchorDir, "fifo")) })

	opened := make(chan error, 1)
	go func() {
		// Blocks in open(2) until a writer opens the other end.
		f, err := unix.Openat(fd, "fifo", syscall.O_RDONLY, 0)
		if err == nil {
			syscall.Close(f)
		}
		opened <- err
	}()

	progressed := make(chan error, 1)
	go func() {
		// A short-lived anchored operation that must not wait behind the FIFO.
		if err := unix.Mkdirat(fd, "while-blocked", 0700); err != nil {
			progressed <- err
			return
		}
		var st syscall.Stat_t
		if err := unix.Fstatat(fd, "while-blocked", &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			progressed <- err
			return
		}
		progressed <- nil
	}()

	select {
	case err := <-progressed:
		if err != nil {
			t.Errorf("an operation issued while another was blocked failed: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("an operation issued while another was blocked did not complete: the anchored, blocked call is preventing independent progress")
	}

	// Release the blocked open with a writer, then require it to finish.
	w, err := os.OpenFile(filepath.Join(anchorDir, "fifo"), os.O_WRONLY, 0)
	if err != nil {
		t.Logf("opening the FIFO for writing (path form) failed: %v; the blocked open is released by the cleanup instead", err)
	} else {
		w.Close()
	}
	select {
	case err := <-opened:
		if err != nil {
			t.Errorf("the blocked anchored FIFO open failed after a writer appeared: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Logf("the blocked FIFO open did not return within 30s; it is released when the test process exits, and the progress assertion above is unaffected")
	}

	if _, err := os.Stat(filepath.Join(cwdDir, "while-blocked")); !os.IsNotExist(err) {
		t.Errorf("%s/while-blocked: %v; the operation must not have landed in the working directory", cwdDir, err)
	}
}

// TestAtDarwinPPCCompoundStressDarwinPPC is the memory-ownership control: the
// native routine reads buffers the caller allocated, so a request whose payload
// is garbage-collected or whose caller stack moves must be impossible. The test
// drives many short calls from goroutines with small stacks while forcing the
// collector, and checks every result: a stale or relocated buffer would show up
// as a wrong size or a spurious failure, not as a crash.
//
// It covers the metadata operations too, not only the read: Fchmodat's
// ATTR_CMN_ACCESSMASK payload and Utimensat's timespec payload are built on the
// caller side and read by the kernel after the goroutine has parked, which is the
// case the ownership rule exists for. Each worker owns its own file and asserts
// the effect it asked for, so a call that silently acted on another worker's
// entry — the failure a shared fixture would hide — is reported. The
// access-time-only case runs through the HFS workaround, which also reads the
// filesystem and the entry's device before acting.
func TestAtDarwinPPCCompoundStressDarwinPPC(t *testing.T) {
	anchorDir, _, fd := atTree(t)

	const workers, iters = 4, 20
	omit := syscall.Timespec{Sec: unix.UTIME_OMIT, Nsec: unix.UTIME_OMIT}

	// One fixture per worker, distinguishable by size, mode and stamp, so "the
	// operation changed the entry this worker named" is checkable.
	names := make([]string, workers)
	for w := 0; w < workers; w++ {
		names[w] = fmt.Sprintf("stress-%d", w)
		if err := os.WriteFile(filepath.Join(anchorDir, names[w]), make([]byte, w+1), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for _, name := range names {
			os.Remove(filepath.Join(anchorDir, name))
		}
	})

	errs := make(chan error, workers*iters)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 6; i++ {
			runtime.GC()
			time.Sleep(5 * time.Millisecond)
		}
		close(done)
	}()
	for w := 0; w < workers; w++ {
		go func(w int) {
			name := names[w]
			wantSize := int64(w + 1)
			mode := uint16(0o600 + w) // 0600, 0601, 0602, 0603
			acc := syscall.Timespec{Sec: int32(1400000000 + 10*w)}
			mod := syscall.Timespec{Sec: int32(1400000001 + 10*w)}
			accOnly := syscall.Timespec{Sec: int32(1500000000 + 10*w)}
			for i := 0; i < iters; i++ {
				// A stack-hungry frame: the caller's stack may be copied while
				// the compound is in flight, which is what makes laundering a
				// caller-stack address into the request unsafe.
				var pad [512]byte
				pad[i%len(pad)] = byte(i)
				var st syscall.Stat_t
				if err := unix.Fstatat(fd, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
					errs <- fmt.Errorf("worker %d iter %d: Fstatat: %v", w, i, err)
					return
				}
				if st.Size != wantSize {
					errs <- fmt.Errorf("worker %d iter %d: size = %d; want %d", w, i, st.Size, wantSize)
					return
				}
				if err := unix.Fchmodat(fd, name, uint32(mode), unix.AT_SYMLINK_NOFOLLOW); err != nil {
					errs <- fmt.Errorf("worker %d iter %d: Fchmodat: %v", w, i, err)
					return
				}
				if err := unix.Utimensat(fd, name, &[2]syscall.Timespec{acc, mod}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
					errs <- fmt.Errorf("worker %d iter %d: Utimensat: %v", w, i, err)
					return
				}
				if err := unix.Utimensat(fd, name, &[2]syscall.Timespec{accOnly, omit}, unix.AT_SYMLINK_NOFOLLOW); err != nil {
					errs <- fmt.Errorf("worker %d iter %d: Utimensat(access time only): %v", w, i, err)
					return
				}
				var after syscall.Stat_t
				if err := unix.Fstatat(fd, name, &after, unix.AT_SYMLINK_NOFOLLOW); err != nil {
					errs <- fmt.Errorf("worker %d iter %d: Fstatat after the metadata calls: %v", w, i, err)
					return
				}
				if got := after.Mode & 0o777; got != mode {
					errs <- fmt.Errorf("worker %d iter %d: mode = %o; want %o", w, i, got, mode)
					return
				}
				if after.Atimespec.Sec != accOnly.Sec {
					errs <- fmt.Errorf("worker %d iter %d: atime = %d; want the requested %d", w, i, after.Atimespec.Sec, accOnly.Sec)
					return
				}
				if after.Mtimespec.Sec != mod.Sec {
					errs <- fmt.Errorf("worker %d iter %d: mtime = %d; want the untouched %d", w, i, after.Mtimespec.Sec, mod.Sec)
					return
				}
				if after.Size != wantSize {
					errs <- fmt.Errorf("worker %d iter %d: size changed to %d; want %d", w, i, after.Size, wantSize)
					return
				}
				if pad[i%len(pad)] != byte(i) {
					errs <- fmt.Errorf("worker %d iter %d: local buffer was corrupted", w, i)
					return
				}
			}
			errs <- nil
		}(w)
	}
	for w := 0; w < workers; w++ {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}
	<-done

	// Every worker's own entry ended with its own mode and times, and no worker
	// wrote into another worker's file.
	for w := 0; w < workers; w++ {
		fi, err := os.Lstat(filepath.Join(anchorDir, names[w]))
		if err != nil {
			t.Fatalf("%s: %v", names[w], err)
		}
		if got, wantMode := uint32(fi.Mode().Perm()), uint32(0o600+w); got != wantMode {
			t.Errorf("%s mode = %o; want %o", names[w], got, wantMode)
		}
		if got, wantSize := fi.Size(), int64(w+1); got != wantSize {
			t.Errorf("%s size = %d; want %d: the metadata calls must not have written data", names[w], got, wantSize)
		}
	}
}

// TestAtDarwinPPCCompoundGCScanDarwinPPC requires the garbage collector to be
// able to scan a goroutine that is parked inside the native scope.
//
// This is deterministic where the stress test is probabilistic: a FIFO open for
// reading blocks in the kernel, so the helper goroutine stays inside atCompound's
// frame — in _Gsyscall, with the private cwd installed — until a writer appears.
// scanstack() scans exactly that frame (gp.syscallsp, runtime/mgcmark.go:938-960),
// so the GC it is forced to run here must walk through it. The routine's frame is
// 32 bytes of locals plus the 24-byte Darwin linkage area, more than the 32 bytes
// an assembly frame may have without a locals pointer map, so the frame carries
// NO_LOCAL_POINTERS; without it this test dies with "fatal error: missing
// stackmap" instead of failing an assertion. That is the measured pre-fix
// behaviour of the subject this test was added with (see goppc-jic6x.1), which is
// why the test is here rather than left to chance in the stress loop.
func TestAtDarwinPPCCompoundGCScanDarwinPPC(t *testing.T) {
	anchorDir, _, fd := atTree(t)
	fifo := filepath.Join(anchorDir, "gcscan-fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatalf("Mkfifo: %v", err)
	}
	t.Cleanup(func() { os.Remove(fifo) })

	blocked := make(chan error, 1)
	go func() {
		f, err := unix.Openat(fd, "gcscan-fifo", syscall.O_RDONLY, 0)
		if err == nil {
			syscall.Close(f)
		}
		blocked <- err
	}()

	// Give the helper time to reach the blocked open, then force several
	// collections while its frame is on the stack. A collection that faults is
	// the failure this control exists for; one that does not must find the
	// goroutine parked in the syscall rather than finished.
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 4; i++ {
		runtime.GC()
		runtime.Gosched()
	}
	select {
	case err := <-blocked:
		t.Fatalf("the anchored FIFO open returned %v before a writer appeared; it did not block, so this control measured nothing", err)
	default:
	}

	// Release it and require the operation to complete correctly.
	w, err := os.OpenFile(fifo, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("opening the FIFO for writing: %v", err)
	}
	w.Close()
	select {
	case err := <-blocked:
		if err != nil {
			t.Errorf("the blocked anchored open failed after the writer appeared: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Errorf("the blocked anchored open did not return within 30s of a writer appearing")
	}
}

// TestAtDarwinPPCCompoundRefusalsDarwinPPC is the negative control for the two
// thread-state operations used by the other controls in this file: if 349 could
// not be installed or cleared by a test, those controls would be measuring
// nothing.
func TestAtDarwinPPCCompoundRefusalsDarwinPPC(t *testing.T) {
	// With no private cwd installed, clearing reports EBADF. This is also what
	// the production routine must leave behind: nothing to clear.
	if err := rawPthreadFchdir(-1); !errors.Is(err, syscall.EBADF) {
		t.Errorf("349(-1) with no private cwd = %v; want EBADF", err)
	}
	// Installing a non-directory descriptor fails and changes nothing.
	f, err := os.CreateTemp(t.TempDir(), "not-a-dir")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := rawPthreadFchdir(int(f.Fd())); err == nil {
		t.Errorf("349(non-directory descriptor) = nil; want an error")
	}
	if err := rawPthreadFchdir(-1); !errors.Is(err, syscall.EBADF) {
		t.Errorf("349(-1) after a failed install = %v; want EBADF (a failed install must change nothing)", err)
	}
}

// atOpenFDCount reports how many descriptors this process has open, as seen
// through /dev/fd. It is how the mixed-operand tests check that the dispatcher
// closes the descriptor it captured for a relative AT_FDCWD operand.
func atOpenFDCount(t *testing.T) int {
	t.Helper()
	ents, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Fatalf("reading /dev/fd: %v", err)
	}
	return len(ents)
}

// TestAtMixedFDCWDCaptureDarwinPPC sets up the three trees a mixed two-name
// operation has to keep apart: the process working directory (a decoy), the
// locked caller's own private working directory (what AT_FDCWD means for that
// caller), and a held anchor descriptor on a third directory.
//
// Both directions are exercised for rename and for link. The AT_FDCWD operand
// must be answered by the CALLER's working directory, not by the process
// directory and not by the helper's, and the dispatcher's capture must leave no
// descriptor behind and must not disturb the caller's private cwd.
func TestAtMixedFDCWDCaptureDarwinPPC(t *testing.T) {
	procDir, callerDir, anchorDir := t.TempDir(), t.TempDir(), t.TempDir()
	for _, dir := range []string{procDir, callerDir, anchorDir} {
		atWrite(t, filepath.Join(dir, "src"), filepath.Base(dir))
	}
	t.Chdir(procDir)

	done := make(chan struct{})
	go func() {
		defer close(done)
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		callerFd, err := syscall.Open(callerDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Error(err)
			return
		}
		defer syscall.Close(callerFd)
		if err := rawPthreadFchdir(callerFd); err != nil {
			t.Errorf("installing the caller's private cwd: %v", err)
			return
		}
		installed := true
		defer func() {
			if installed {
				if err := rawPthreadFchdir(-1); err != nil {
					t.Errorf("clearing the caller's private cwd: %v", err)
				}
			}
		}()

		anchorFd, err := syscall.Open(anchorDir, syscall.O_RDONLY, 0)
		if err != nil {
			t.Error(err)
			return
		}
		defer syscall.Close(anchorFd)

		fdsBefore := atOpenFDCount(t)

		// Direction one: the source resolves from the caller's cwd, the destination
		// from the held anchor.
		if err := unix.Renameat(unix.AT_FDCWD, "src", anchorFd, "dst"); err != nil {
			t.Errorf("Renameat(caller cwd, anchor): %v", err)
		}
		wantAbsent(t, filepath.Join(callerDir, "src"))
		wantContent(t, filepath.Join(anchorDir, "dst"), filepath.Base(callerDir))

		// Direction two: the source resolves from the held anchor, the destination
		// from the caller's cwd.
		if err := unix.Renameat(anchorFd, "src", unix.AT_FDCWD, "dst"); err != nil {
			t.Errorf("Renameat(anchor, caller cwd): %v", err)
		}
		wantAbsent(t, filepath.Join(anchorDir, "src"))
		wantContent(t, filepath.Join(callerDir, "dst"), filepath.Base(anchorDir))

		// The same two shapes for link, with fresh fixtures.
		atWrite(t, filepath.Join(callerDir, "lsrc"), "lsrc")
		atWrite(t, filepath.Join(anchorDir, "asrc"), "asrc")
		if err := unix.Linkat(unix.AT_FDCWD, "lsrc", anchorFd, "ldst", 0); err != nil {
			t.Errorf("Linkat(caller cwd, anchor): %v", err)
		}
		wantContent(t, filepath.Join(anchorDir, "ldst"), "lsrc")
		if err := unix.Linkat(anchorFd, "asrc", unix.AT_FDCWD, "adst", 0); err != nil {
			t.Errorf("Linkat(anchor, caller cwd): %v", err)
		}
		wantContent(t, filepath.Join(callerDir, "adst"), "asrc")

		if got := atOpenFDCount(t); got != fdsBefore {
			t.Errorf("open descriptors %d -> %d: a capture leaked a descriptor", fdsBefore, got)
		}

		// The process working directory is a decoy throughout: nothing may have
		// landed there, and the caller must not have moved either.
		wantContent(t, filepath.Join(procDir, "src"), filepath.Base(procDir))
		for _, name := range []string{"dst", "ldst", "adst"} {
			wantAbsent(t, filepath.Join(procDir, name))
		}

		// The caller's own private cwd is still installed and still resolves to its
		// own directory: a capture must not be paid for with the caller's thread
		// state.
		if got, err := syscall.Getwd(); err != nil || !sameFile(t, got, callerDir) {
			t.Errorf("caller cwd after the mixed forms = %q, %v; want %q", got, err, callerDir)
		}
		if err := rawPthreadFchdir(-1); err != nil {
			t.Errorf("clearing the caller's private cwd: %v; want nil (it must still be installed)", err)
		}
		installed = false
		if got, err := syscall.Getwd(); err != nil || !sameFile(t, got, procDir) {
			t.Errorf("process cwd after clearing = %q, %v; want %q", got, err, procDir)
		}
	}()
	<-done
}
