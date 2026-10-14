// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

#include "textflag.h"

// func rawVforkSyscall(trap, a1, a2, a3 uintptr) (r1 uintptr, err Errno)
TEXT ·rawVforkSyscall(SB), NOSPLIT|NOFRAME, $0-24
	LWZ R3, a1+4(FP)
	LWZ R4, a2+8(FP)
	LWZ R5, a3+12(FP)
	LWZ R0, trap+0(FP)
	SC
	BC 4, 3, vfork_ok
	ADDI R4, R0, $-1
	STW R4, r1+16(FP)
	STW R3, err+20(FP)
	RET
vfork_ok:
	STW R3, r1+16(FP)
	ADDI R4, R0, $0
	STW R4, err+20(FP)
	RET

// func rawSyscallNoError(trap, a1, a2, a3 uintptr) (r1, r2 uintptr)
TEXT ·rawSyscallNoError(SB), NOSPLIT|NOFRAME, $0-24
	LWZ R3, a1+4(FP)
	LWZ R4, a2+8(FP)
	LWZ R5, a3+12(FP)
	LWZ R0, trap+0(FP)
	SC
	STW R3, r1+16(FP)
	STW R4, r2+20(FP)
	RET
