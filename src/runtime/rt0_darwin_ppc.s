// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "go_asm.h"
#include "textflag.h"

// Darwin process entry is not a C function call: argc and argv are on the
// initial stack. The measured C ABI's R3-R10 argument registers apply to calls,
// not to the LC_UNIXTHREAD state that transfers control to this entry.
TEXT _rt0_ppc_darwin(SB), NOSPLIT|NOFRAME, $0-0
	ADDI R0, R0, $0
	// Preserve the kernel's initial stack block FIRST, from the UNMASKED entry
	// SP: argc is at entry SP+0 and argv at entry SP+4 (measured on the G5; the
	// earlier "+8" comment was wrong — argv is the word right after argc).
	LWZ R3, 0(R1)
	ADDI R4, R1, $4
	// The kernel hands an entry SP rounded only to the 4-byte pointer size
	// (xnu-1228 kern_exec.c), so it is NOT guaranteed 16-byte aligned: measured
	// entry SP mod 16 = 4/8/12 across argv/env variants. rt0_go and every later
	// m0/g0 frame inherit that misalignment, breaking the measured 16-byte frame
	// rule (ABI:251) and 8-byte alignment of float scratch (an lfd/stfd could hit
	// a 4-mod-8 address). Align R1 DOWN to a 16-byte boundary before building the
	// frame. R4 is an absolute pointer into the untouched kernel block, so it stays
	// valid. This backend rejects a 5-operand rlwinm in hand asm and mis-tokenizes
	// the trailing dot on "ANDI.", so align with ADDI+AND:
	//   ADDI R6, R0, $-16   -> R6 = 0xFFFFFFF0 = ~15 (addi destination is the 1st
	//                          operand; R0 holds 0, and rA=0 means the literal 0)
	//   AND  R1, R1, R6     -> R1 = R1 & ~15, i.e. round R1 DOWN to a multiple of 16
	// The AND is written with R1 in BOTH the destination and source-A fields, so
	// it computes R1 & R6 into R1 regardless of how this backend orders the two
	// register operands — deliberate, because a boot-path instruction must not
	// depend on an operand-order convention. Mirrors rt0_darwin_ppc64.s's RLDCR
	// mask. Effect confirmed on hardware at runtime.schedinit (gdb), not just assembled.
	ADDI R6, R0, $-16
	AND  R1, R1, R6
	ADDI R1, R1, $-32
	STW R3, 24(R1)
	STW R4, 28(R1)
	B runtime·rt0_go(SB)
