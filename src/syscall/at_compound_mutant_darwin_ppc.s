// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompound_mutant

#include "textflag.h"

// Two deliberately broken copies of atCompound, compiled only under the
// goppc_atcompound_mutant build tag. They exist so the cleanup oracle can be
// shown to discriminate: the production routine is not built with this tag and
// is not touched by it, so nothing here ships.
//
//   atCompoundSkipClearMutant is atCompound with the clear removed. If the
//   oracle reported the same result for it as for the production routine, the
//   oracle would be measuring nothing.
//
//   atCompoundClearFailMutant clears the private cwd twice: the second clear
//   (the routine's own) finds none and reports EBADF, which is the internal
//   invariant failure that must end the process through the raw exit(127)
//   fail-stop branch before exitsyscall.
//
// Everything except those two edits is copied from at_compound_darwin_ppc.s,
// including register placement and the request offsets; read that file for the
// annotation of each syscall.

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

#define STAT_DEV 0
#define STAT_INO 8

#define SYS_exit_                 1
#define SYS_stat64_               338
#define SYS_fstat64_              339
#define SYS_pthread_fchdir_       349

// func atCompoundSkipClearMutant(req *atCompoundReq)
TEXT ·atCompoundSkipClearMutant(SB),NOSPLIT,$32-4
	BL	runtime·entersyscall<ABIInternal>(SB)

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atSkipClearEnterFail

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_DOT(R10)
	LWZ	R4, ATC_ST1(R10)
	ADDI	R0, R0, $SYS_stat64_
	SC
	B	atSkipClearIdentFail

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	LWZ	R4, ATC_ST2(R10)
	ADDI	R0, R0, $SYS_fstat64_
	SC
	B	atSkipClearIdentFail

	LWZ	R3, req+0(FP)
	LWZ	R4, ATC_ST1(R3)
	LWZ	R5, ATC_ST2(R3)
	LWZ	R6, STAT_DEV(R4)
	LWZ	R7, STAT_DEV(R5)
	CMPW	R6, R7
	BC	4, 2, atSkipClearIdentMismatch
	LWZ	R6, STAT_INO(R4)
	LWZ	R7, STAT_INO(R5)
	CMPW	R6, R7
	BC	4, 2, atSkipClearIdentMismatch
	LWZ	R6, STAT_INO+4(R4)
	LWZ	R7, STAT_INO+4(R5)
	CMPW	R6, R7
	BC	4, 2, atSkipClearIdentMismatch

	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_W1(R10)
	LWZ	R5, ATC_W2(R10)
	LWZ	R6, ATC_W3(R10)
	LWZ	R7, ATC_W4(R10)
	LWZ	R0, ATC_NR(R10)
	LWZ	R3, ATC_W0(R10)
	SC
	B	atSkipClearOpFail

	LWZ	R10, req+0(FP)
	STW	R3, ATC_R0(R10)
	// The defect: no clear of the thread-private cwd.
	B	atSkipClearDone

atSkipClearEnterFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_ENTERERR(R10)
	B	atSkipClearDone

atSkipClearIdentFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_IDENTERR(R10)
	B	atSkipClearDone

atSkipClearIdentMismatch:
	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_ENOTSUP(R10)
	STW	R4, ATC_IDENTERR(R10)
	B	atSkipClearDone

atSkipClearOpFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_OPERR(R10)
	B	atSkipClearDone

atSkipClearDone:
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET

// func atCompoundClearFailMutant(req *atCompoundReq)
TEXT ·atCompoundClearFailMutant(SB),NOSPLIT,$32-4
	BL	runtime·entersyscall<ABIInternal>(SB)

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atClearFailEnterFail

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_DOT(R10)
	LWZ	R4, ATC_ST1(R10)
	ADDI	R0, R0, $SYS_stat64_
	SC
	B	atClearFailIdentFail

	LWZ	R10, req+0(FP)
	LWZ	R3, ATC_ANCHOR(R10)
	LWZ	R4, ATC_ST2(R10)
	ADDI	R0, R0, $SYS_fstat64_
	SC
	B	atClearFailIdentFail

	LWZ	R3, req+0(FP)
	LWZ	R4, ATC_ST1(R3)
	LWZ	R5, ATC_ST2(R3)
	LWZ	R6, STAT_DEV(R4)
	LWZ	R7, STAT_DEV(R5)
	CMPW	R6, R7
	BC	4, 2, atClearFailIdentMismatch
	LWZ	R6, STAT_INO(R4)
	LWZ	R7, STAT_INO(R5)
	CMPW	R6, R7
	BC	4, 2, atClearFailIdentMismatch
	LWZ	R6, STAT_INO+4(R4)
	LWZ	R7, STAT_INO+4(R5)
	CMPW	R6, R7
	BC	4, 2, atClearFailIdentMismatch

	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_W1(R10)
	LWZ	R5, ATC_W2(R10)
	LWZ	R6, ATC_W3(R10)
	LWZ	R7, ATC_W4(R10)
	LWZ	R0, ATC_NR(R10)
	LWZ	R3, ATC_W0(R10)
	SC
	B	atClearFailOpFail

	LWZ	R10, req+0(FP)
	STW	R3, ATC_R0(R10)
	B	atClearFailClear

atClearFailEnterFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_ENTERERR(R10)
	B	atClearFailDone

atClearFailIdentFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_IDENTERR(R10)
	B	atClearFailClear

atClearFailIdentMismatch:
	LWZ	R10, req+0(FP)
	LWZ	R4, ATC_ENOTSUP(R10)
	STW	R4, ATC_IDENTERR(R10)
	B	atClearFailClear

atClearFailOpFail:
	LWZ	R10, req+0(FP)
	STW	R3, ATC_OPERR(R10)
	B	atClearFailClear

	// The injected failure: clear the private cwd here, so the routine's own
	// clear below finds none and reports EBADF.
atClearFailClear:
	ADDI	R3, R0, $-1
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atClearFailDone

	// The routine's own clear. It must now fail, and that failure must take the
	// fail-stop branch: raw exit(127) before exitsyscall.
	ADDI	R3, R0, $-1
	ADDI	R0, R0, $SYS_pthread_fchdir_
	SC
	B	atClearFailFailStop
	B	atClearFailDone

atClearFailFailStop:
	ADDI	R3, R0, $127
	ADDI	R0, R0, $SYS_exit_
	SC
	ADDI	R0, R0, $-1
	SC
atClearFailHang:
	B	atClearFailHang

atClearFailDone:
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
