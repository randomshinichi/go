// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

// goppc-7tqu. ppc32 has no native 64-bit atomic: internal/runtime/atomic
// implements every 64-bit operation through one global, non-reentrant spinlock,
// lock64. runtime.unlock2 calls mutexSampleContention on every unlock, and that
// function's first act was to Load64 mutexprofilerate. Every print() ends in
// printunlock -> unlock(&debuglock), so a fatal report or a panic report printed
// on a thread that already holds lock64 printed its first line and then spun in
// lock64Acquire forever -- the owner is the code the report interrupted, and it
// cannot run again until the report returns.
//
// The four children below run in a subprocess, because the failure is a hang
// that no in-process timeout can recover. The parent bounds each child with a
// wall-clock kill and reports the child's user and system CPU: a spin in the
// print path is USER time (about the wall time, system near zero), which is how
// it is told apart from goppc-0v7x, a kernel re-fault loop that burns SYSTEM
// time.
//
// The named mutant for this test file is "M-no-flag-check": the tree with the
// mutexProfilingOn() test removed from mutexSampleContention (the flag and its
// maintenance in SetMutexProfileFraction left in place). Every child that runs a
// fatal report while lock64 is held must wedge on it, every time.
const (
	fatalSIGQUITLock64Arg = "fatalSIGQUITLock64"
	fatalSIGQUITNoHoldArg = "fatalSIGQUITNoHold"
	fatalLoad64NilArg     = "fatalLoad64Nil"
	fatalLoadNilByteArg   = "fatalLoadNilByte"
	fatalPrintHeldArg     = "fatalPrintHeld"

	fatalChildWall   = 20 * time.Second
	fatalChildStatus = 2

	fatalStartMarker = "fatal child:"
	fatalReturned    = "fatal child: the report returned without exiting"
)

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case fatalSIGQUITLock64Arg:
		fatalSIGQUITLock64Child()
	case fatalSIGQUITNoHoldArg:
		fatalSIGQUITNoHoldChild()
	case fatalLoad64NilArg:
		fatalLoad64NilChild()
	case fatalLoadNilByteArg:
		fatalLoadNilByteChild()
	case fatalPrintHeldArg:
		fatalPrintHeldChild()
	}
}

// fatalSIGQUITLock64Child is the deterministic reproducer: this thread holds
// lock64 when the kernel delivers SIGQUIT to it, so the fatal report runs on the
// lock's holder. Before the fix the child printed "SIGQUIT: quit" and then spun.
func fatalSIGQUITLock64Child() {
	runtime.LockOSThread()
	fmt.Println(fatalStartMarker, fatalSIGQUITLock64Arg)
	runtime.HoldLock64AcrossSIGQUIT()
	fmt.Println(fatalReturned)
	os.Exit(4)
}

// fatalSIGQUITNoHoldChild is the leg-B control: SIGQUIT with lock64 free. It must
// print the same full report and exit 2 on the mutant as well as on the fix, so
// that a change to the fatal path itself cannot pass for the fix.
func fatalSIGQUITNoHoldChild() {
	runtime.LockOSThread()
	fmt.Println(fatalStartMarker, fatalSIGQUITNoHoldArg)
	runtime.SIGQUITCurrentThread()
	fmt.Println(fatalReturned)
	os.Exit(4)
}

// fatalLoad64NilChild reaches the same hazard with no signal, no profiler and no
// recover(): an unrecovered 64-bit atomic on an invalid operand. Whether this one
// wedges depends on the fault being taken before lock64Acquire, which
// atomic.check64 arranges (a1a6ce2918); it is kept here because it is the leg
// that needs no signal to reproduce the class.
func fatalLoad64NilChild() {
	fmt.Println(fatalStartMarker, fatalLoad64NilArg)
	atomic.LoadUint64(nil)
	fmt.Println(fatalReturned)
	os.Exit(4)
}

// fatalLoadNilByteChild is the leg-A control: a faulting load of the same width
// as address zero that is not a 64-bit atomic. Same report path, no lock64.
func fatalLoadNilByteChild() {
	fmt.Println(fatalStartMarker, fatalLoadNilByteArg)
	_ = loadByteAt((*byte)(unsafe.Pointer(fatalZeroAddress)))
	fmt.Println(fatalReturned)
	os.Exit(4)
}

// fatalPrintHeldChild prints while holding lock64 and exits 0 if it gets that
// far. It is the direct dynamic form of criterion 3: the print's own unlock of
// debuglock runs the rate-zero path of mutexSampleContention, so a read of the
// 64-bit rate there would deadlock this child. It is a hang rather than a report,
// so it is run as a child like the others.
func fatalPrintHeldChild() {
	runtime.LockOSThread()
	fmt.Println(fatalStartMarker, fatalPrintHeldArg)
	runtime.HoldLock64AcrossPrint()
	fmt.Println("fatal child: printed and returned while holding lock64")
	os.Exit(0)
}

// fatalZeroAddress is zero and never written, but that is not something the
// compiler may assume: the load through it has to reach the hardware.
var fatalZeroAddress uintptr

//go:noinline
func loadByteAt(p *byte) byte { return *p }

type fatalChildResult struct {
	out    string
	status int
	user   time.Duration
	sys    time.Duration
}

// runFatalChild runs this test binary as the named child and kills it after
// fatalChildWall. A killed child is a child that hung.
func runFatalChild(t *testing.T, arg string) fatalChildResult {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe, arg)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", arg, err)
	}
	var wasKilled atomic.Bool
	timer := time.AfterFunc(fatalChildWall, func() { wasKilled.Store(true); cmd.Process.Kill() })
	werr := cmd.Wait()
	timer.Stop()
	if wasKilled.Load() {
		t.Fatalf("child %s HUNG: still running after %v wall (kill sent); output so far:\n%s",
			arg, fatalChildWall, buf.String())
	}
	if werr != nil {
		if _, ok := werr.(*exec.ExitError); !ok {
			t.Fatalf("child %s: %v", arg, werr)
		}
	}
	r := fatalChildResult{out: buf.String(), status: cmd.ProcessState.ExitCode()}
	r.user, r.sys = cmd.ProcessState.UserTime(), cmd.ProcessState.SystemTime()
	return r
}

// requireFatalReport is the shared assertion: the child reached its own exit
// status, and its report is complete rather than a first line.
func requireFatalReport(t *testing.T, arg, firstLine string, r fatalChildResult) {
	t.Helper()
	t.Logf("child %s: status=%d user=%v sys=%v\n%s", arg, r.status, r.user, r.sys, r.out)
	if strings.Contains(r.out, fatalReturned) {
		t.Errorf("child %s: the fatal report returned to its caller instead of exiting", arg)
	}
	if r.status != fatalChildStatus {
		t.Errorf("child %s: exit status %d, want %d", arg, r.status, fatalChildStatus)
	}
	if !strings.Contains(r.out, firstLine) {
		t.Errorf("child %s: output does not contain %q", arg, firstLine)
	}
	if !strings.Contains(r.out, "goroutine 1") {
		t.Errorf("child %s: the report stops before the goroutine dump; no \"goroutine 1\" in output", arg)
	}
}

// TestMutexProfilingFlagMirrorsRate runs the instrument the fix rests on: the
// 32-bit flag mutexSampleContention tests before it reads the rate must follow
// SetMutexProfileFraction in both directions. If the flag were stuck clear, the
// fatal-report tests below would pass for the wrong reason -- mutex profiling
// would be silently off for every program.
func TestMutexProfilingFlagMirrorsRate(t *testing.T) {
	old := runtime.SetMutexProfileFraction(-1)
	defer runtime.SetMutexProfileFraction(old)
	for _, step := range []struct {
		rate int
		flag uint32
	}{
		{0, 0},
		{1, 1},
		{100, 1},
		{0, 0},
	} {
		runtime.SetMutexProfileFraction(step.rate)
		if got := runtime.MutexProfilingFlag(); got != step.flag {
			t.Errorf("after SetMutexProfileFraction(%d): flag=%d, want %d", step.rate, got, step.flag)
		}
		if got := runtime.SetMutexProfileFraction(-1); got != step.rate {
			t.Errorf("after SetMutexProfileFraction(%d): rate reads back %d, want %d", step.rate, got, step.rate)
		}
	}
}

// TestFatalReportWhileLock64Held is criterion 1 of goppc-7tqu: the fatal report
// for a signal delivered to the thread that holds lock64 must print in full and
// exit 2, instead of printing "SIGQUIT: quit" and spinning in lock64Acquire.
func TestFatalReportWhileLock64Held(t *testing.T) {
	r := runFatalChild(t, fatalSIGQUITLock64Arg)
	requireFatalReport(t, fatalSIGQUITLock64Arg, "SIGQUIT: quit", r)
}

// TestPrintWhileLock64Held requires the print path itself to survive holding
// lock64. printunlock -> unlock(&debuglock) -> unlock2 calls mutexSampleContention
// on every unlock, so this is the rate-zero path of unlock2 under test, with no
// signal and no dump in the way: only the flag check can make it return.
func TestPrintWhileLock64Held(t *testing.T) {
	r := runFatalChild(t, fatalPrintHeldArg)
	t.Logf("child %s: status=%d user=%v sys=%v\n%s", fatalPrintHeldArg, r.status, r.user, r.sys, r.out)
	if r.status != 0 {
		t.Errorf("child %s: exit status %d, want 0", fatalPrintHeldArg, r.status)
	}
	if !strings.Contains(r.out, "printed and returned while holding lock64") {
		t.Errorf("child %s: the print did not return", fatalPrintHeldArg)
	}
}

// TestFatalReportInvalidAtomic64Operand is the same requirement for the leg that
// needs no signal: an unrecovered 64-bit atomic on an invalid operand.
func TestFatalReportInvalidAtomic64Operand(t *testing.T) {
	r := runFatalChild(t, fatalLoad64NilArg)
	requireFatalReport(t, fatalLoad64NilArg, "panic: runtime error: invalid memory address or nil pointer dereference", r)
}

// TestFatalReportControls are the controls: the same reports with lock64 free
// (leg B) and with a plain byte load instead of a 64-bit atomic (leg A). They
// must pass on the mutant too -- that is what makes the two tests above evidence
// about lock64 rather than about the report path in general.
func TestFatalReportControls(t *testing.T) {
	t.Run(fatalSIGQUITNoHoldArg, func(t *testing.T) {
		r := runFatalChild(t, fatalSIGQUITNoHoldArg)
		requireFatalReport(t, fatalSIGQUITNoHoldArg, "SIGQUIT: quit", r)
	})
	t.Run(fatalLoadNilByteArg, func(t *testing.T) {
		r := runFatalChild(t, fatalLoadNilByteArg)
		requireFatalReport(t, fatalLoadNilByteArg, "panic: runtime error: invalid memory address or nil pointer dereference", r)
	})
}
