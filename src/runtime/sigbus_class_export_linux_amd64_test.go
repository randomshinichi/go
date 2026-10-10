// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && amd64

package runtime

// Bridge between TestSIGBUSFaultClassification (which lives in package
// runtime_test, because it has to spawn child processes and therefore has to
// import os/exec) and the unexported runtime internals the child needs: the
// alignment-check assembly helpers in sigbus_class_linux_amd64.s, the signal
// fields the signal handler fills in before it arranges for sigpanic to run,
// and sigpanic itself.
//
// This file is test-only: it is only ever compiled into the runtime test
// binary. See the test for why all of this is on linux/amd64 only.

// Implemented in sigbus_class_linux_amd64.s. Declaring them in a test file
// (rather than in the runtime proper) keeps these two out of anything that is
// not the runtime test binary; the assembly simply defines unused symbols in a
// normal build.
func sigbusTestLoadUnalignedWithAC(p *byte) uint64
func sigbusTestClearAlignmentCheck()

// SigbusTestSigbus is _SIGBUS.
const SigbusTestSigbus = _SIGBUS

// SigbusTestBUSADRALN is BUS_ADRALN. Linux has no name for it in package
// runtime because upstream sigpanic spells the value as a literal; the kernel
// stands behind the name, see the comment on the test.
const SigbusTestBUSADRALN = 1

// SigbusTestLoadUnalignedWithAC loads 8 bytes from p with x86 alignment
// checking turned on. It faults (SIGBUS, BUS_ADRALN, si_addr = 0) unless p
// happens to be 8-byte aligned, in which case it returns the value it read.
func SigbusTestLoadUnalignedWithAC(p *byte) uint64 {
	return sigbusTestLoadUnalignedWithAC(p)
}

// SigbusTestClearAlignmentCheck turns alignment checking back off on the
// calling thread. A thread that has taken an #AC fault resumes with it still
// on, so this must be the first thing any recovering code does.
func SigbusTestClearAlignmentCheck() {
	sigbusTestClearAlignmentCheck()
}

// SigbusTestSetPanicOnFault sets the same flag runtime/debug.SetPanicOnFault
// sets -- getg().paniconfault, which sigpanic consults in its _SIGBUS and
// _SIGSEGV arms -- and returns the previous value.
func SigbusTestSetPanicOnFault(v bool) (old bool) {
	gp := getg()
	old = gp.paniconfault
	gp.paniconfault = v
	return old
}

// SigbusTestFaultRecord returns what the signal handler recorded for the most
// recent synchronous fault taken on this goroutine: the signal number and the
// two words out of siginfo that sigpanic's _SIGBUS gate looks at (si_code and
// si_addr).
func SigbusTestFaultRecord() (sig uint32, code0, code1 uintptr) {
	gp := getg()
	return gp.sig, gp.sigcode0, gp.sigcode1
}

// SigbusTestSigpanicOnFault hands sigpanic the fault record a signal handler
// would have left behind, and does not return: it either panics or throws,
// depending on how sigpanic classifies the record.
func SigbusTestSigpanicOnFault(sig uint32, code0, code1 uintptr) {
	gp := getg()
	gp.sig = sig
	gp.sigcode0 = code0
	gp.sigcode1 = code1
	gp.paniconfault = false
	sigpanic()
}
