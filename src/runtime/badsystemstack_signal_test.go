// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"os"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
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
