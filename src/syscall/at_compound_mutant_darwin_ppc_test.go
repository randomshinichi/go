// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompound_mutant

package syscall_test

// Controls that need a deliberately broken copy of the native compound. They are
// built only with the goppc_atcompound_mutant tag, from
// syscall/at_compound_mutant_darwin_ppc.s, so no shipped binary contains them.
//
// Why they exist: TestAtCompoundSameMClearDarwinPPC says "the private cwd was
// cleared on this M". A test that cannot fail is not a control, so the same
// observation has to be shown to report the opposite for a routine that does not
// clear (the skip-clear mutant), and the fail-stop branch that protects a
// contaminated M has to be shown to actually fire (the clear-failure mutant).

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// sameDirDarwinPPC reports whether two paths name the same existing directory.
func sameDirDarwinPPC(t *testing.T, a, b string) bool {
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

func writeFixtureDarwinPPC(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestAtCompoundSkipClearMutantDarwinPPC requires the cleanup oracle to notice a
// routine that installs the private cwd and never clears it.
//
// The observations are the ones the production test makes, on the same thread:
// after the mutant call, syscall.Getwd must report the ANCHOR (the leak), and
// raw 349(-1) must report success — there was a private cwd to clear — where the
// production routine leaves EBADF. If the oracle could not tell the two apart,
// the production test would be green for any routine at all.
func TestAtCompoundSkipClearMutantDarwinPPC(t *testing.T) {
	anchorDir := t.TempDir()
	if err := syscall.Mkdir(anchorDir+"/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixtureDarwinPPC(t, anchorDir+"/file", "anchor")
	prev, err := syscall.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Chdir(t.TempDir()); err != nil {
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

	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		defer close(done)
		// Never leave a private cwd behind, even if an assertion fails.
		defer syscall.AtPthreadFchdirForTest(-1)

		if got, err := syscall.Getwd(); err != nil || got != decoy {
			t.Errorf("before the call this thread's cwd is %q, %v; want the process directory %q", got, err, decoy)
			return
		}
		enter, ident, op := syscall.AtCompoundSkipClearMutantForTest(anchorFd, "file")
		if enter != 0 || ident != 0 || op != 0 {
			t.Errorf("skip-clear mutant = enter %v, ident %v, op %v; want 0, 0, 0", enter, ident, op)
			return
		}
		// The leak must be visible as the thread's cwd.
		got, err := syscall.Getwd()
		if err != nil {
			t.Errorf("Getwd after the mutant call: %v", err)
			return
		}
		// Compare directories, not spellings: /tmp is a symlink to /private/tmp
		// on Darwin, so the same directory has two names.
		if sameDirDarwinPPC(t, got, anchorDir) {
			// Expected: the anchor is the thread's cwd.
		} else {
			t.Errorf("after the skip-clear mutant this thread's cwd is %q; want the anchor %q — the oracle cannot see a missing clear", got, anchorDir)
		}
		// And as a private cwd that 349(-1) can still clear.
		if err := syscall.AtPthreadFchdirForTest(-1); err != nil {
			t.Errorf("349(-1) after the skip-clear mutant = %v; want nil (a private cwd was installed and not cleared)", err)
		}
	}()
	<-done
}

// failStopEnv selects the child half of TestAtCompoundClearFailMutantDarwinPPC.
const failStopEnv = "GOPPC_ATCOMPOUND_CLEARFAIL_CHILD"

// TestAtCompoundClearFailMutantDarwinPPC requires an unexpected clear failure to
// end the process through the native fail-stop branch before exitsyscall, not to
// be reported to Go on a thread that still carries a private cwd.
//
// The mutant clears the private cwd itself and then lets the routine's own clear
// fail with EBADF. The child (this same binary, re-executed) records that it
// armed the request, then dies. What proves the ordering is:
//
//   - the exit status is exactly 127, which only the raw exit(2) branch produces
//     (a Go-level report would exit 0, 1 or panic with 2);
//   - the "returned" marker, written immediately after the call, is absent, so no
//     Go code ran after the failed clear;
//   - the child never prints anything, so there was no diagnostic write in the
//     native scope before the exit.
func TestAtCompoundClearFailMutantDarwinPPC(t *testing.T) {
	if os.Getenv(failStopEnv) == "1" {
		childClearFailDarwinPPC(t)
		return
	}

	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run", "TestAtCompoundClearFailMutantDarwinPPC$", "-test.v")
	cmd.Env = append(os.Environ(), failStopEnv+"=1", "GOPPC_ATCOMPOUND_MARKER="+marker)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("the clear-failure mutant exited 0; want the native fail-stop exit 127. output: %s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("running the mutant child: %v", err)
	}
	if code := exitErr.ExitCode(); code != 127 {
		t.Errorf("the clear-failure mutant exited %d; want 127 from the raw exit branch. output: %s", code, out)
	}
	b, rerr := os.ReadFile(marker)
	if rerr != nil {
		t.Fatalf("reading the marker: %v", rerr)
	}
	text := string(b)
	if !strings.Contains(text, "armed") {
		t.Errorf("the child never armed the request (marker %q); the exit status proves nothing", text)
	}
	if strings.Contains(text, "returned") {
		t.Errorf("the child wrote the post-call marker %q; Go code ran after the failed clear, so the fail-stop branch did not fire before exitsyscall", text)
	}
	if strings.Contains(string(out), "FAIL") || strings.Contains(string(out), "PASS") {
		t.Errorf("the child produced testing output %q; the process should have exited inside the native scope", out)
	}
}

// childClearFailDarwinPPC is the child half: arm the request, then run the mutant,
// which must not return.
func childClearFailDarwinPPC(t *testing.T) {
	marker := os.Getenv("GOPPC_ATCOMPOUND_MARKER")
	note := func(s string) {
		f, err := os.OpenFile(marker, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
		if err != nil {
			os.Exit(90)
		}
		f.WriteString(s + "\n")
		f.Close()
	}
	anchorDir := t.TempDir()
	writeFixtureDarwinPPC(t, anchorDir+"/file", "anchor")
	anchorFd, err := syscall.Open(anchorDir, syscall.O_RDONLY, 0)
	if err != nil {
		os.Exit(91)
	}
	// A goroutine that keeps marking liveness: if the contaminated M were handed
	// back to Go, this goroutine could run again on it. Its continued activity is
	// not proof either way, so the test does not assert on it; it is here so a
	// human reading the marker file can see how much ran.
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(time.Millisecond):
				note("tick")
			}
		}
	}()
	note("armed")
	syscall.AtCompoundClearFailMutantForTest(anchorFd, "file")
	note("returned")
	close(stop)
	os.Exit(92)
}
