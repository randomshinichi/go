// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && amd64

package runtime_test

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

// TestSIGBUSFaultClassification is a regression guard for the _SIGBUS arm of
// sigpanic: a SIGBUS whose si_addr lies in the first page is a nil/wild-pointer
// dereference only when the kernel says so (_BUS_ADRERR). Being at a small
// address is not by itself evidence of a nil dereference.
//
// The counterexample behind this test is real, not hypothetical. On x86 Linux
// the alignment check exception (#AC) is delivered as SIGBUS with si_code =
// BUS_ADRALN (1) and si_addr = NULL even though the address that faulted was an
// ordinary, nonzero one. Upstream's #AC handler hands the fault on with a null
// address --
//
//	do_trap(X86_TRAP_AC, SIGBUS, "alignment check", regs, error_code, BUS_ADRALN, NULL);
//
// -- and that null is what user space reads back as si_addr. The reference is
// upstream's and the test does not rest on it: the
// KernelReportsBUS_ADRALNWithZeroAddr case below reads the triple back out of
// the goroutine on every run, and fails if a kernel or a CPU ever reports
// something else. A gate of the form
//
//	if gp.sigcode1 < 0x1000 { panicmem() }
//
// therefore misclassifies a hardware alignment fault as a nil pointer
// dereference, which turns a fatal fault into a panic that a recover can
// swallow. sigpanic narrows that condition with a compile-time operand
// (BUS_ADRALN is accepted only on darwin/ppc); this test is what stops a future
// refactor from quietly re-broadening it.
//
// The three cases below re-exec this test binary, each with a marker in the
// environment, so that a fault which kills the child cannot take the test down
// with it:
//
//   - HardwareAlignmentFaultIsFatal takes a real #AC fault in a child whose
//     paniconfault flag is off, as in a normal program, and requires the child
//     to die rather than recover. This is the actual guard.
//   - KernelReportsBUS_ADRALNWithZeroAddr takes the same fault in a child told
//     to panic on faults, so it survives long enough to report what the kernel
//     actually delivered. That is the premise the guard rests on, re-measured
//     every run instead of assumed.
//   - ClassifierOnKernelReportedTriple feeds sigpanic the very triple measured
//     by the previous case, in a process where alignment checking is off, so
//     that the complete message of the fatal path can be observed.
//
// A note for anyone reading a child's output: after an #AC fault the thread
// resumes with EFLAGS.AC still set (the kernel does not mask AC on entry to the
// user signal handler), and while AC is set perfectly legal unaligned accesses
// fault too -- notably the 9..16 byte path of runtime.memmove, which moves two
// 8-byte chunks at unaligned offsets. The child of the first case therefore
// dies *inside* the runtime's fatal path, before it can print the whole
// message. That is why the assertion is on the message the classification
// itself produces ("unexpected fault address", printed only after the gate
// declines to panic) plus the absence of a recovery, and why the third case
// exists: with AC off, nothing disturbs the message path.
func TestSIGBUSFaultClassification(t *testing.T) {
	if mode := os.Getenv(sigbusChildEnvVar); mode != "" {
		sigbusChild(mode)
		os.Exit(125) // unreachable
	}

	t.Run("HardwareAlignmentFaultIsFatal", func(t *testing.T) {
		out, err := sigbusRunChild(t, sigbusModeFatal)
		t.Logf("child (%s) output:\n%s", sigbusModeFatal, out)

		if strings.Contains(out, sigbusNoFaultMarker) {
			t.Errorf("the #AC fault did not reproduce on this CPU/kernel: the unaligned load did not fault, so this test proved nothing")
		}
		if err == nil {
			t.Errorf("child exited successfully; a hardware fault was classified as recoverable")
		}
		if strings.Contains(out, sigbusRecoveredMarker) {
			t.Errorf("child recovered from the fault (%s): a SIGBUS with si_addr=0 and si_code=BUS_ADRALN must stay fatal on linux/amd64; the _SIGBUS gate in sigpanic has been re-broadened", sigbusRecoveredMarker)
		}
		if strings.Contains(out, "nil pointer dereference") {
			t.Errorf("the fault was reported as a nil pointer dereference; si_addr=0 does not imply one here, see the comment on TestSIGBUSFaultClassification")
		}
		if !strings.Contains(out, "unexpected fault address") {
			t.Errorf("child output does not contain %q: the fatal path of sigpanic's _SIGBUS arm was not taken", "unexpected fault address")
		}
	})

	t.Run("KernelReportsBUS_ADRALNWithZeroAddr", func(t *testing.T) {
		out, err := sigbusRunChild(t, sigbusModeProbe)
		t.Logf("child (%s) output:\n%s", sigbusModeProbe, out)

		if err != nil {
			t.Errorf("child failed: %v", err)
		}
		if strings.Contains(out, sigbusNoFaultMarker) {
			t.Errorf("the #AC fault did not reproduce on this CPU/kernel")
		}
		if !strings.Contains(out, sigbusProbeOKMarker) {
			t.Errorf("child did not report SIGBUS/BUS_ADRALN/si_addr=0 for a nonzero misaligned address")
		}
	})

	t.Run("ClassifierOnKernelReportedTriple", func(t *testing.T) {
		out, err := sigbusRunChild(t, sigbusModeInject)
		t.Logf("child (%s) output:\n%s", sigbusModeInject, out)

		if err == nil {
			t.Errorf("child exited successfully; (sig=SIGBUS, si_code=BUS_ADRALN, si_addr=0) was classified as recoverable")
		}
		if strings.Contains(out, sigbusRecoveredMarker) {
			t.Errorf("child recovered from (sig=SIGBUS, si_code=BUS_ADRALN, si_addr=0) (%s)", sigbusRecoveredMarker)
		}
		if !strings.Contains(out, "unexpected fault address") {
			t.Errorf("child output does not contain %q", "unexpected fault address")
		}
		if !strings.Contains(out, "fatal error: fault") {
			t.Errorf("child output does not contain %q", "fatal error: fault")
		}
	})
}

// Child-process plumbing. The child is this same test binary re-executed with a
// mode in the environment; it never returns to the test framework.
const (
	sigbusChildEnvVar = "RUNTIME_TEST_SIGBUS_CLASSIFY_CHILD"

	sigbusModeFatal  = "fatal"  // real #AC fault: the classification must be fatal
	sigbusModeProbe  = "probe"  // real #AC fault, survived so the report can be read back
	sigbusModeInject = "inject" // the same report handed to sigpanic directly

	sigbusRecoveredMarker = "RECOVERED_UNEXPECTEDLY"
	sigbusNoFaultMarker   = "SIGBUS-CHILD-NOFAULT"
	sigbusProbeOKMarker   = "SIGBUS-PROBE-OK"
	sigbusProbeBadMarker  = "SIGBUS-PROBE-BAD"
)

// sigbusRunChild re-executes this test binary in the given child mode and
// returns the child's combined output and the error from waiting for it (nil
// only if the child exited 0).
func sigbusRunChild(t *testing.T, mode string) (string, error) {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	// Build the child's environment explicitly instead of appending to ours:
	// an inherited GODEBUG would come first in the environ and would win over
	// the one added here.
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GODEBUG=") || strings.HasPrefix(kv, sigbusChildEnvVar+"=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		sigbusChildEnvVar+"="+mode,
		// The thread must not be preempted between setting AC and taking the
		// fault: the preemption handler is Go code, and with AC set every
		// unaligned access in it faults too, which would turn one controlled
		// fault into a random one.
		"GODEBUG=asyncpreemptoff=1",
	)

	cmd := exec.Command(exe, "-test.run=^TestSIGBUSFaultClassification$")
	cmd.Env = env
	// The child may be killed by a signal and dump core; keep that out of the
	// package directory.
	cmd.Dir = t.TempDir()

	out, err := cmd.CombinedOutput()
	return string(out), err
}

// sigbusChild runs in the re-executed child process and never returns.
func sigbusChild(mode string) {
	switch mode {
	case sigbusModeFatal:
		sigbusChildFault(false)
	case sigbusModeProbe:
		sigbusChildFault(true)
	case sigbusModeInject:
		sigbusChildInject()
	default:
		fmt.Printf("SIGBUS-CHILD: unknown mode %q\n", mode)
		os.Exit(127)
	}
	fmt.Println("SIGBUS-CHILD: returned without dying or exiting")
	os.Exit(126)
}

// sigbusChildFault takes a real hardware alignment fault -- SIGBUS with
// si_code=BUS_ADRALN and si_addr=0 -- on a deliberately misaligned, nonzero
// address.
//
// With probe false, paniconfault stays off as it is in a normal program, so the
// classification decides the outcome: on the narrowed tree the fault is fatal,
// on a re-broadened tree panicmem runs and the deferred recover below gets to
// swallow it.
//
// With probe true, paniconfault is turned on instead, which routes the fault to
// panicmemAddr in either tree, so the child survives a fault it is not supposed
// to survive just long enough to say what the kernel reported.
func sigbusChildFault(probe bool) {
	// Heap, not stack: the address handed to assembly must not move, and the
	// pointer argument keeps the object alive for the collector.
	buf := new([64]byte)
	p := &buf[1]
	addr := uintptr(unsafe.Pointer(p))

	runtime.SigbusTestSetPanicOnFault(probe)

	defer func() {
		// AC is still set on this thread, so turn it off before running any
		// Go code -- including the recover below, which is fine either way,
		// and the prints, which are not.
		runtime.SigbusTestClearAlignmentCheck()
		r := recover()
		if r == nil {
			return
		}
		if !probe {
			// The fault was swallowed. That is the defect: say so on the way
			// out, and leave through the front door so the parent sees a
			// clean, signalless exit status rather than a crash.
			fmt.Printf("%s: %v\n", sigbusRecoveredMarker, r)
			os.Exit(0)
		}
		// panicmemAddr panicked, as it was told to. The signal handler fills
		// in gp.sig, gp.sigcode0 and gp.sigcode1 before it arranges for
		// sigpanic to run, so they still hold this fault's report.
		sig, code0, code1 := runtime.SigbusTestFaultRecord()
		fmt.Printf("SIGBUS-PROBE sig=%d code0=%d code1=%d faultaddr=%#x addrnonzero=%v misaligned=%v\n",
			sig, code0, code1, addr, addr != 0, addr%8 != 0)
		if sig == runtime.SigbusTestSigbus && code0 == runtime.SigbusTestBUSADRALN && code1 == 0 && addr != 0 && addr%8 != 0 {
			fmt.Println(sigbusProbeOKMarker)
		} else {
			fmt.Println(sigbusProbeBadMarker)
		}
		os.Exit(0)
	}()

	fmt.Printf("SIGBUS-CHILD faulting on an 8-byte load at %#x\n", addr)
	v := runtime.SigbusTestLoadUnalignedWithAC(p)

	// Reached only if the CPU did not raise #AC, in which case the experiment
	// did not happen and the parent has to hear about that rather than read a
	// pass.
	runtime.SigbusTestClearAlignmentCheck()
	fmt.Printf("%s: the unaligned load completed, value=%#x\n", sigbusNoFaultMarker, v)
	os.Exit(3)
}

// sigbusChildInject hands sigpanic the (sig, si_code, si_addr) triple that the
// kernel reports for #AC -- measured by the probe case, not assumed -- and lets
// it classify. Alignment checking is off here, so nothing in the runtime's own
// message path is disturbed and the whole fatal message can be observed.
func sigbusChildInject() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s: %v\n", sigbusRecoveredMarker, r)
			os.Exit(0)
		}
	}()

	fmt.Printf("SIGBUS-CHILD injecting sig=%d code0=%d code1=%d\n",
		runtime.SigbusTestSigbus, runtime.SigbusTestBUSADRALN, 0)
	runtime.SigbusTestSigpanicOnFault(runtime.SigbusTestSigbus, runtime.SigbusTestBUSADRALN, 0)

	fmt.Println("SIGBUS-CHILD: sigpanic returned without panicking or throwing")
	os.Exit(6)
}
