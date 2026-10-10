// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && amd64

#include "textflag.h"

// Support code for TestSIGBUSFaultClassification, in
// sigbus_class_linux_amd64_test.go. These two functions exist so that the test
// can put the current thread into x86 alignment-checking mode (#AC) and take it
// back out again; nothing else calls them, so they are dead code in a real
// program (the linker drops them when they are not referenced).

// func sigbusTestLoadUnalignedWithAC(p *byte) uint64
//
// Sets EFLAGS.AC (bit 18) and then loads 8 bytes from p. CR0.AM is set on every
// modern Linux, so at CPL 3 the AC flag on its own makes that load raise #AC
// when p is not 8-byte aligned. Linux turns #AC into SIGBUS with si_code =
// BUS_ADRALN and si_addr = NULL (the kernel's handler passes addr = NULL, it
// does not report the faulting address).
//
// The whole experiment is in assembly on purpose: between POPFQ and the load
// the thread is running with AC set, and with AC set every unaligned access in
// ordinary Go code faults too, including the 9..16 byte path of
// runtime.memmove. Keeping the window free of Go code keeps the fault the test
// takes the only fault it takes.
//
// If the load does not fault, AC is cleared again before returning.
TEXT ·sigbusTestLoadUnalignedWithAC(SB), NOSPLIT|NOFRAME, $0-16
	MOVQ	p+0(FP), AX
	PUSHFQ
	POPQ	CX
	ORQ	$0x40000, CX		// EFLAGS.AC, bit 18
	PUSHQ	CX
	POPFQ
	MOVQ	(AX), AX		// unaligned: #AC -> SIGBUS / BUS_ADRALN / si_addr = 0
	PUSHFQ
	POPQ	CX
	ANDQ	$-0x40001, CX
	PUSHQ	CX
	POPFQ
	MOVQ	AX, ret+8(FP)
	RET

// func sigbusTestClearAlignmentCheck()
//
// Clears EFLAGS.AC on the calling thread.
//
// A thread that has taken an #AC fault resumes with AC still set: the kernel
// stores the faulting context, including its RFLAGS, in the signal frame and
// does not mask AC when it enters the user handler. So any Go code that runs
// after such a fault -- a deferred recover, for instance -- has to clear the
// flag first, or its own perfectly legal unaligned accesses will fault as well.
TEXT ·sigbusTestClearAlignmentCheck(SB), NOSPLIT|NOFRAME, $0-0
	PUSHFQ
	POPQ	AX
	ANDQ	$-0x40001, AX
	PUSHQ	AX
	POPFQ
	RET
