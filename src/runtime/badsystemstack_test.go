// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"internal/testenv"
)

// systemstack_bad is fatal by construction: badsystemstack writes its message
// and systemstack_bad then calls abort, which faults on purpose. It therefore
// cannot be reached from inside a test, and the witness has to be driven as a
// subprocess of the test binary itself - the same re-exec trick
// crash_unix_test.go uses, which needs no `go build` on the target.
const (
	badssTakenArg   = "badssBranchTaken"
	badssControlArg = "badssBranchNotTaken"
	badssAbortArg   = "badssAbortOnly"
)

// Checked by TestSystemstackBadBranchTaken and TestAbortEpilogueOnRealG.
const (
	badssFatalLine    = "fatal: systemstack called from unexpected goroutine"
	badssReturnedLine = "badss: returned without taking the fatal branch"
	badssControlLine  = "badss: control returned from systemstack"
	badssFaultLine    = "SIGBUS: bus error"
)

// badssChildGrace bounds the wait for a child. Every mode here ends in the fatal
// path, which faults on purpose; on ppc32 the fault is reported and the process
// is then not observed to exit (bead goppc-0v7x). The witness has to read the
// child's OUTPUT, so it must not be turned into a test timeout by that.
const badssChildGrace = 30 * time.Second

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case badssTakenArg:
		// g is replaced by a copy at a different address, so it is neither
		// m.gsignal, m.g0 nor m.curg: systemstack must take its fatal branch.
		runtime.BadSystemstackBranchTaken()
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssControlArg:
		// The same call with the untouched g: systemstack must not take the
		// fatal branch, and must return.
		runtime.SystemstackOnRealG()
		os.Stderr.WriteString(badssControlLine + "\n")
		os.Exit(0)
	case badssAbortArg:
		// abort() from the ordinary g, systemstack not involved at all.
		runtime.AbortForTesting()
	}
}

// runBadssChild re-execs this test binary in one of the witness modes and returns
// everything the child printed on stderr, its process state, and whether it had
// to be killed. killed is true when the child neither exited nor died within
// badssChildGrace.
func runBadssChild(t *testing.T, arg string) (out string, state os.ProcessState, killed bool) {
	t.Helper()
	cmd := exec.Command(testenv.Executable(t), arg)
	cmd = testenv.CleanCmdEnv(cmd)
	// Only stderr: the fatal path and every marker here write to fd 2, and one
	// shared bytes.Buffer for both streams would be written by two copiers.
	var buf bytes.Buffer
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting %s: %v", arg, err)
	}
	done := make(chan struct{})
	go func() { cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(badssChildGrace):
		killed = true
		cmd.Process.Kill()
		<-done
	}
	return buf.String(), *cmd.ProcessState, killed
}

// TestSystemstackBadBranchTaken is the witness for the landed ppc32 change that
// replaced the direct BL runtime·badsystemstack(SB) in systemstack_bad with an
// indirect call. That change is only proven at link time; this proves the three
// instructions also WORK when the branch is taken.
//
// The control child differs from the witness child in exactly one thing: whether
// the g is replaced by an address systemstack does not accept. So the test cannot
// pass unless the fatal branch really ran: a witness that never reaches
// systemstack_bad prints badssReturnedLine and exits 97 instead of printing the
// fatal message.
func TestSystemstackBadBranchTaken(t *testing.T) {
	// Control: with the real g, systemstack must return normally. This catches a
	// witness that is fatal no matter what it is handed.
	out, state, killed := runBadssChild(t, badssControlArg)
	switch {
	case killed:
		t.Errorf("control child had to be killed; systemstack must return with an unmodified g\n%s", out)
	case !state.Success():
		t.Errorf("control child did not exit 0: %v\n%s", state, out)
	}
	if !strings.Contains(out, badssControlLine) {
		t.Errorf("control child did not reach the end of the control path\n%s", out)
	}
	if strings.Contains(out, badssFatalLine) {
		t.Errorf("control child took the fatal branch with an unmodified g\n%s", out)
	}
	t.Logf("control child (g unmodified), exit %v, killed=%v:\n%s", state, killed, out)

	// Witness: the only difference from the control is the g's address.
	out, state, killed = runBadssChild(t, badssTakenArg)
	if !strings.Contains(out, badssFatalLine) {
		t.Errorf("systemstack_bad did not run: want %q in the child's output\n"+
			"exit %v, killed=%v\n%s", badssFatalLine, state, killed, out)
	}
	if strings.Contains(out, badssReturnedLine) {
		t.Errorf("the witness returned instead of aborting; the fatal branch was not taken\n%s", out)
	}
	if !killed && state.Success() {
		t.Errorf("the witness exited 0; systemstack_bad's abort() must never return\n%s", out)
	}
	t.Logf("witness child (g replaced by a copy at another address), exit %v, killed=%v:\n%s",
		state, killed, out)
}

// TestAbortEpilogueOnRealG is the witness's second control and it exists to keep
// the report honest about what the witness child does AFTER the fault. abort() is
// called here from the ordinary g with systemstack, badsystemstack and the
// replaced g all out of the picture. If this child also prints its fault report
// and then does not exit, the non-exit is a property of the fatal epilogue on
// ppc32 and not something the systemstack_bad change caused.
func TestAbortEpilogueOnRealG(t *testing.T) {
	out, state, killed := runBadssChild(t, badssAbortArg)
	if !strings.Contains(out, badssFaultLine) {
		t.Errorf("abort() did not fault as designed: want %q\n"+
			"exit %v, killed=%v\n%s", badssFaultLine, state, killed, out)
	}
	t.Logf("abort-only child (real g, no systemstack), exit %v, killed=%v:\n%s", state, killed, out)
}
