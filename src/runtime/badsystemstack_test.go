// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"os"
	"runtime"
	"testing"
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

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case badssTakenArg:
		badssChildReady()
		// g is replaced by a copy at a different address, so it is neither
		// m.gsignal, m.g0 nor m.curg: systemstack must take its fatal branch.
		runtime.BadSystemstackBranchTaken()
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssControlArg:
		badssChildReady()
		// The same call with the untouched g: systemstack must not take the
		// fatal branch, and must return.
		runtime.SystemstackOnRealG()
		os.Stderr.WriteString(badssControlLine + "\n")
		os.Exit(0)
	case badssAbortArg:
		badssChildReady()
		// abort() from the ordinary g, systemstack not involved at all.
		runtime.AbortForTesting()
	}
}

// badssControlExpect: with the real g, systemstack returns and the child exits 0.
var badssControlExpect = badssExpect{
	exitCode:  0,
	markers:   []string{badssControlLine},
	forbidden: []string{badssFatalLine, badssFaultLine},
}

// badssAbortExpect: abort() from the ordinary g faults, the whole report is
// written, and the child exits 2 by itself. The register dump is the last thing
// the report prints, so "link 0x" is the terminal evidence.
var badssAbortExpect = badssExpect{
	exitCode:  2,
	markers:   []string{badssFaultLine, "goroutine 1 gp=", "runtime.abort()", "pc   0x", "link 0x"},
	forbidden: []string{badssFatalLine, badssReturnedLine},
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
//
// The witness grades diagnostic completion, not wall expiry (goppc-0v7x review
// 691deaec). It passes only if the child terminated by itself with exit 4 after
// printing the whole nested report (see badssWitnessExpect). A child that is
// still alive when the budget ends is a wedge only if it had spent its whole CPU
// budget; otherwise the target was too loaded to say, and the test skips with
// INFRASTRUCTURE-INCONCLUSIVE. See badsystemstack_oracle_test.go.
func TestSystemstackBadBranchTaken(t *testing.T) {
	// Control: with the real g, systemstack must return normally. This catches a
	// witness that is fatal no matter what it is handed.
	r := runBadssChild(t, badssControlArg, badssDefaultBudget)
	badssApply(t, "control child (g unmodified)", r, judgeBadss(r, badssControlExpect))

	// Witness: the only difference from the control is the g's address.
	r = runBadssChild(t, badssTakenArg, badssDefaultBudget)
	badssApply(t, "witness child (g replaced by a copy at another address)", r, judgeBadss(r, badssWitnessExpect))
}

// TestAbortEpilogueOnRealG is the witness's second control and it exists to keep
// the report honest about what the witness child does AFTER the fault. abort() is
// called here from the ordinary g with systemstack, badsystemstack and the
// replaced g all out of the picture. If this child also prints its fault report
// and then does not exit, the non-exit is a property of the fatal epilogue on
// ppc32 and not something the systemstack_bad change caused.
func TestAbortEpilogueOnRealG(t *testing.T) {
	r := runBadssChild(t, badssAbortArg, badssDefaultBudget)
	badssApply(t, "abort-only child (real g, no systemstack)", r, judgeBadss(r, badssAbortExpect))
}
