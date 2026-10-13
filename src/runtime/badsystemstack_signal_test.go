// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"internal/testenv"
)

// Controls for the fault-signal unblocking in sighandler (goppc-0v7x). sighandler
// now unblocks SIGILL, SIGTRAP, SIGFPE, SIGBUS and SIGSEGV before it starts the
// fatal report, so that a fault raised BY the report is delivered and walks the
// m.dying ladder instead of being retried by the kernel forever. These children
// are the legitimate cases that change must leave alone: a fatal fault that is
// not nested, a recoverable one, and a SIGTRAP that is not a Go fault at all.
const (
	badssSegvG0Arg    = "badssSegvOnSystemstack"
	badssNilG0Arg     = "badssNilOnSystemstack"
	badssNilPanicArg  = "badssNilPanicRecovered"
	badssTrapArg      = "badssTrapToSelf"
	badssTrapDoneLine = "badss: SIGTRAP returned to the program"
	badssNilDoneLine  = "badss: recovered the nil pointer panic"
)

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case badssSegvG0Arg:
		badssChildReady()
		// A page that was mapped and is not any more: the kernel reports an
		// invalid address (SIGSEGV), not the protection failure page zero gives
		// (SIGBUS).
		mem, err := syscall.Mmap(-1, 0, 4096, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			os.Stderr.WriteString("badss: mmap: " + err.Error() + "\n")
			os.Exit(badssPrologueFailedExit)
		}
		addr := uintptr(unsafe.Pointer(&mem[0]))
		if err := syscall.Munmap(mem); err != nil {
			os.Stderr.WriteString("badss: munmap: " + err.Error() + "\n")
			os.Exit(badssPrologueFailedExit)
		}
		runtime.ReadOnSystemstack(addr)
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssNilG0Arg:
		badssChildReady()
		runtime.ReadOnSystemstack(8)
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssNilPanicArg:
		badssChildReady()
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(runtime.Error); ok && e.Error() == "runtime error: invalid memory address or nil pointer dereference" {
						os.Stderr.WriteString(badssNilDoneLine + "\n")
						os.Exit(0)
					}
					os.Stderr.WriteString("badss: recovered something else\n")
					os.Exit(98)
				}
			}()
			var p *int
			*p = 1
		}()
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssTrapArg:
		badssChildReady()
		syscall.Kill(syscall.Getpid(), syscall.SIGTRAP)
		os.Stderr.WriteString(badssTrapDoneLine + "\n")
		os.Exit(0)
	}
}

// A fatal fault that is not nested. The report must run to its end exactly once:
// no "panic during panic", and the original fault is the only one attributed.
// The fault is on g0, so the report is the signal handler's own (not sigpanic's
// "unexpected signal during runtime execution", which is for faults on a user g).
func badssPlainFatalFault(signal string) badssExpect {
	return badssExpect{
		exitCode:  2,
		markers:   []string{signal, "PC=0x", "goroutine 0 gp=", "goroutine 1 gp=", "link 0x"},
		forbidden: []string{badssReturnedLine, "panic during panic", "stack trace unavailable"},
		once:      []string{signal},
	}
}

func TestFatalFaultOnSystemstackKeepsItsReport(t *testing.T) {
	for _, c := range []struct {
		name, arg, signal string
	}{
		{"unmapped-address", badssSegvG0Arg, "SIGSEGV: segmentation violation"},
		{"nil-address", badssNilG0Arg, "SIGBUS: bus error"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := runBadssChild(t, c.arg, badssWitnessBudget(t))
			badssApply(t, c.name+" fault on g0", r, judgeBadss(r, badssPlainFatalFault(c.signal)))
		})
	}
}

func TestNilPanicStillRecovers(t *testing.T) {
	r := runBadssChild(t, badssNilPanicArg, badssWitnessBudget(t))
	badssApply(t, "recoverable nil pointer panic", r, judgeBadss(r, badssExpect{
		exitCode:  0,
		markers:   []string{badssNilDoneLine},
		forbidden: []string{badssReturnedLine, "fatal error", "SIGBUS", "SIGSEGV"},
	}))
}

// A SIGTRAP sent to the process by something other than a debugger consuming it is
// still the Go runtime's fatal signal, and the report must say SIGTRAP.
func TestSigtrapWithoutDebuggerIsReportedAsSigtrap(t *testing.T) {
	r := runBadssChild(t, badssTrapArg, badssWitnessBudget(t))
	badssApply(t, "SIGTRAP with no debugger", r, judgeBadss(r, badssExpect{
		exitCode:  2,
		markers:   []string{"SIGTRAP: trace trap", "goroutine ", "link 0x"},
		forbidden: []string{badssTrapDoneLine, "panic during panic", "SIGBUS", "SIGSEGV"},
		once:      []string{"SIGTRAP: trace trap"},
	}))
}

// The debugger controls. A debugger sits in front of the program's signals: it
// sees each one first and either CONSUMES it (gdb "nopass": the program never
// learns of it) or PASSES it on (gdb "pass": Go's handler runs). sighandler now
// unblocks SIGTRAP before it arms the m.dying ladder, so these children pin down
// what each case must look like, run under the real gdb 6.3 of Leopard. The
// no-unblock build (the state before the fix) is the comparison; see the bead.
//
// The debugger's own exit status is not graded - gdb -batch exits 0 or 1 for
// reasons unrelated to the program - the program's fate is read from gdb's
// "Program exited" line and from the report the program itself wrote.
func TestDebuggerTrapControls(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("the debugger controls are for the darwin fault-signal path")
	}
	gdb, err := exec.LookPath("gdb")
	if err != nil {
		t.Skip("gdb not found; consumed/passed trap controls need it")
	}
	exe := testenv.Executable(t)
	gdbStop := "Program received signal SIGTRAP"

	cases := []struct {
		name, mode, script string
		expect             badssExpect
		// after is an extra check on the output, for outcomes a marker list
		// cannot state.
		after func(t *testing.T, out string)
	}{
		{
			// gdb takes the trap and does not pass it: the program must carry on.
			name: "consumed",
			mode: badssTrapArg,
			script: "handle SIGTRAP stop print nopass\n" +
				"run\ncontinue\n",
			expect: badssExpect{
				exitCode:  -1,
				markers:   []string{gdbStop, badssTrapDoneLine, "Program exited normally."},
				forbidden: []string{"SIGTRAP: trace trap", "panic during panic", "goroutine ", "SIGBUS", "SIGSEGV"},
				once:      []string{gdbStop},
			},
		},
		{
			// gdb passes the trap on: Go's fatal handler runs, and the report must
			// name SIGTRAP, once, as the fault - it is the original fault here.
			name: "passed",
			mode: badssTrapArg,
			script: "handle SIGTRAP stop print pass\n" +
				"run\ncontinue\n",
			expect: badssExpect{
				exitCode:  -1,
				markers:   []string{gdbStop, "SIGTRAP: trace trap", "goroutine ", "link 0x", "Program exited with code 02."},
				forbidden: []string{badssTrapDoneLine, "panic during panic", "SIGBUS", "SIGSEGV"},
				once:      []string{gdbStop, "SIGTRAP: trace trap"},
			},
		},
		{
			// A trap queued exactly in the window between unblockFaultSignals and
			// startpanic_m, with the real fault (abort's SIGBUS) already inside
			// Go's handler. The outcome that must hold is that the program ends
			// by itself with exit 2 and ONE fatal report. WHICH signal that report
			// names is logged, not graded: see the note in the test body.
			name: "trap-queued-before-the-ladder-is-armed",
			mode: badssAbortArg,
			script: "handle SIGTRAP nostop noprint pass\n" +
				"handle EXC_BAD_ACCESS nostop noprint pass\n" +
				"break *" + fmt.Sprintf("%#x", runtime.StartpanicMPC()) + "\n" +
				"run\n" +
				"echo \\n=== INSIDE-HANDLER: stopped at startpanic_m\\n\n" +
				"delete breakpoints\n" +
				"call (int)kill((int)getpid(), 5)\n" +
				"continue\n",
			expect: badssExpect{
				exitCode:  -1,
				markers:   []string{"=== INSIDE-HANDLER", "goroutine ", "Program exited with code 02."},
				forbidden: []string{badssReturnedLine, "panic during panic", "stack trace unavailable"},
			},
			after: func(t *testing.T, out string) {
				bus := strings.Count(out, "SIGBUS: bus error")
				trap := strings.Count(out, "SIGTRAP: trace trap")
				if bus+trap != 1 {
					t.Errorf("want exactly one fatal report header, got SIGBUS x%d and SIGTRAP x%d", bus, trap)
				}
				t.Logf("ATTRIBUTION: the report names SIGBUS x%d, SIGTRAP x%d. The real fault here was "+
					"abort's SIGBUS; a report that names SIGTRAP means the queued trap took over before "+
					"the m.dying ladder was armed.", bus, trap)
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := filepath.Join(t.TempDir(), c.name+".gdb")
			text := "set confirm off\nset height 0\n" + c.script
			if err := os.WriteFile(script, []byte(text), 0o600); err != nil {
				t.Fatal(err)
			}
			r := runBadssCommand(t, "gdb "+c.name, badssWitnessBudget(t), gdb, "-batch", "-x", script, "--args", exe, c.mode)
			badssApply(t, "debugger control "+c.name, r, judgeBadss(r, c.expect))
			if c.after != nil {
				c.after(t, r.out)
			}
		})
	}
}
