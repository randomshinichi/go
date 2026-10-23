// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "textflag.h"

// The single native scope of one anchored operation. See
// at_compound_darwin_ppc.go for the lifecycle and for why the pieces have to
// live inside one entersyscall/exitsyscall scope; this file is only the
// mechanics, and every design claim above is enforced here.
//
// Branch encoding used below: BC 12, 2, label is branch-if-equal and
// BC 4, 2, label is branch-if-not-equal on CR0, which is how the port's other
// ppc assembly spells BEQ/BNE (runtime/sys_darwin_ppc.s, runtime/asm_ppc.s).
//
// Darwin ppc `sc` returns with the saved PC advanced past the instruction
// following it on success, and at that instruction (the `B <errlabel>` below)
// with R3 = errno on failure (bsd/dev/ppc/systemcalls.c:174-181 and :205-210).
// So each syscall is followed by exactly one branch to its error handler, and
// the success path starts after that branch.
//
// The frame exists for one reason: an assembly function that has a frame has
// its LR saved by the generated prologue and restored by RET, and this routine
// calls runtime.entersyscall and runtime.exitsyscall with BL. No local is used,
// so the 24-byte link area and the frame are all the space taken.
//
// Register use stays inside the volatile set: R3-R10 are the kernel's argument
// and result registers, R30 (g) and R31 (LR save, and the ppc32 RET epilogue's
// scratch) are never written. R10 holds the request base wherever a syscall's
// own arguments have to occupy R3 and up.
//
// Darwin raw system call convention: the number is in R0, the arguments begin in
// R3 (r3, r4, ... in order), and the result returns in R3 (bsd/dev/ppc/
// systemcalls.c munges from regs->save_r3 onwards). Every `SC` below therefore
// has to be audited as "R0 = number, R3.. = arguments", not as Go-level
// parameters:
//
//	SC  address   R0    R3         R4      R5      R6      R7
//	enter         349   anchor      -       -       -       -
//	stat64        338   dot         st1     -       -       -
//	fstat64       339   anchor      st2     -       -       -
//	selected op   nr    w0          w1      w2      w3      w4
//	clear         349   -1          -       -       -       -
//	exit fail-stop 1    127         -       -       -       -
//
// `ADDI R0, R0, $N` loads N here, it does not add to whatever R0 held: on
// PowerPC an addi whose RA field is 0 uses the literal zero as the addend, so
// the instruction is `li r0,N` no matter what R0 contains (and the kernel
// restores the user's R0 across `sc`, so relying on a cleared R0 would be
// wrong anyway). TestAtCompoundLayoutDarwinPPC pins request offsets and
// TestAtCompoundSameMClearDarwinPPC exercises cleanup. The linked register audit
// is run evidence, not a maintained test; see goppc-jic6x.1's corrections receipt.

// atCompoundReq field offsets. These must match the atReq* constants in
// at_compound_darwin_ppc.go; TestAtCompoundLayoutDarwinPPC fails if they drift.
#define ATC_NR       0
#define ATC_ANCHOR   4
#define ATC_W0       8
#define ATC_W1       12
#define ATC_W2       16
#define ATC_W3       20
#define ATC_W4       24
#define ATC_DOT      28
#define ATC_ST1      32
#define ATC_ST2      36
#define ATC_ENOTSUP  40
#define ATC_R0       44
#define ATC_ENTERERR 48
#define ATC_IDENTERR 52
#define ATC_OPERR    56

// Stat_t field offsets compared by the anchor identity check (atStatDev and
// atStatIno in the Go file).
#define STAT_DEV 0
#define STAT_INO 8

// Raw syscall numbers used by the fixed part of the sequence.
#define SYS_exit_                 1
#define SYS_stat64_               338
#define SYS_fstat64_              339
#define SYS_pthread_fchdir_       349

// func atCompound(req *atCompoundReq)
TEXT ·atCompound(SB),NOSPLIT,$32-4
	BL	runtime·entersyscall<ABIInternal>(SB)

	// Install the anchor as this thread's private cwd: __pthread_fchdir(anchor).
	// Nothing has changed if this fails, so the error is reported without a
	// clear attempt.
	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atCompoundEnterFail

	// stat64(".", st1): what the private cwd actually resolved to.
	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_DOT(R10)
	LWZ	R4, ATC_ST1(R10)
	ADDI	R0, R0, $SYS_stat64_
	SC
	B	atCompoundIdentFail

	// fstat64(anchor, st2): the object this call is anchored to.
	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	LWZ	R4, ATC_ST2(R10)
	ADDI	R0, R0, $SYS_fstat64_
	SC
	B	atCompoundIdentFail

	// The dev/ino comparison. A covering mount point, or a directory replaced
	// between the fchdir and here, makes these differ; that is refused instead
	// of being acted on through the private cwd.
	LWZ	R3, req+0(FP)
	LWZ	R4, ATC_ST1(R3)
	LWZ	R5, ATC_ST2(R3)
	LWZ	R6, STAT_DEV(R4)
	LWZ	R7, STAT_DEV(R5)
	CMPW	R6, R7
	BC	4, 2, atCompoundIdentMismatch
	LWZ	R6, STAT_INO(R4)
	LWZ	R7, STAT_INO(R5)
	CMPW	R6, R7
	BC	4, 2, atCompoundIdentMismatch
	LWZ	R6, STAT_INO+4(R4)
	LWZ	R7, STAT_INO+4(R5)
	CMPW	R6, R7
	BC	4, 2, atCompoundIdentMismatch

	// The selected operation: R3-R7 from the request's argument words, the
	// number from its nr field. R10 is the request base, so R0 (the number)
	// can be loaded last without a register move.
	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_W1(R10)
	LWZ	R5, ATC_W2(R10)
	LWZ	R6, ATC_W3(R10)
	LWZ	R7, ATC_W4(R10)
	LWZ	R0, ATC_NR(R10)
	LWZ	R3, ATC_W0(R10)
	SC
	B	atCompoundOpFail

	// Success: publish the result word, then clear the private cwd.
	LWZ	R10, req+0(FP)
	STW	R3, ATC_R0(R10)
	B	atCompoundClear

atCompoundEnterFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_ENTERERR(R10)
	B	atCompoundDone

atCompoundIdentFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_IDENTERR(R10)
	B	atCompoundClear

atCompoundIdentMismatch:
	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_ENOTSUP(R10)
	STW	R4, ATC_IDENTERR(R10)
	B	atCompoundClear

atCompoundOpFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_OPERR(R10)
	B	atCompoundClear

	// 349(-1) clears the private cwd and returns 0. Any other result after a
	// successful install is an internal invariant failure.
atCompoundClear:
	ADDI	R3, R0, $-1		// RA=0 is the literal zero: r3 = -1
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atCompoundClearFail
	B	atCompoundDone

	// Fail-stop, before exitsyscall: this M must not return to Go carrying a
	// private cwd, because the next goroutine scheduled on it would resolve
	// relative paths against the wrong directory. No diagnostic write, no
	// allocation, no Go code: raw exit(127) only. If exit somehow returns, an
	// invalid syscall number raises SIGSYS instead of continuing.
atCompoundClearFail:
	ADDI	R3, R0, $127		// RA=0 is the literal zero: r3 = 127
	ADDI	R0, R0, $SYS_exit_
	SC
	ADDI	R0, R0, $-1		// invalid syscall number
	SC
atCompoundHang:
	B	atCompoundHang

atCompoundDone:
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
