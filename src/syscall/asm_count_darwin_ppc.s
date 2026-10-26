// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcount_test

// Test-only counting copy of asm_darwin_ppc.s. Keep native instructions
// identical except for the non-mutating 349 counter before each trap.
// No signal path, allocation or Go call is added in the native scope.
#include "textflag.h"

// Darwin/ppc32 syscall entry points. On G5 hardware, `sc` returns errno in R3
// on failure and the kernel's resume point distinguishes failure by executing
// the instruction immediately after `sc`; success skips that instruction.
// The syscall number is in R0 and syscall arguments start in R3.

// func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
TEXT ·Syscall(SB),NOSPLIT,$0-28
	BL	runtime·entersyscall<ABIInternal>(SB)
	LWZ	R3, a1+4(FP)
	LWZ	R4, a2+8(FP)
	LWZ	R5, a3+12(FP)
	LWZ	R0, trap+0(FP)
	// Count only syscall 349; preserve R0 and all syscall operands.
	ADDI R12, R0, $349
	CMPW R0, R12
	BC	4, 2, atCountTrap1
	ADDI	$·at349Count(SB), R9
	ADDI	R12, R0, $0
atCountAgain1:
	LWARX	R11, R9, R12
	ADDI	R11, R11, $1
	STWCX	R11, R9, R12
	BC	4, 2, atCountAgain1
atCountTrap1:
	SC
	// Darwin/ppc resumes here on failure (R3=errno) and skips this branch on success.
	// The same two-resume-point behavior is exercised by scratch/rawsc_ppc.c on pmg5.
	B	error
	XOR	R0, R0, R0
	STW	R3, r1+16(FP)
	STW	R4, r2+20(FP)
	STW	R0, err+24(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
error:
	XOR	R0, R0, R0
	STW	R3, err+24(FP)
	ADDI	R3, R0, $-1
	STW	R3, r1+16(FP)
	STW	R0, r2+20(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET

// func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
TEXT ·Syscall6(SB),NOSPLIT,$0-40
	BL	runtime·entersyscall<ABIInternal>(SB)
	LWZ	R3, a1+4(FP)
	LWZ	R4, a2+8(FP)
	LWZ	R5, a3+12(FP)
	LWZ	R6, a4+16(FP)
	LWZ	R7, a5+20(FP)
	LWZ	R8, a6+24(FP)
	LWZ	R0, trap+0(FP)
	// Count only syscall 349; preserve R0 and all syscall operands.
	ADDI R12, R0, $349
	CMPW R0, R12
	BC	4, 2, atCountTrap2
	ADDI	$·at349Count(SB), R9
	ADDI	R12, R0, $0
atCountAgain2:
	LWARX	R11, R9, R12
	ADDI	R11, R11, $1
	STWCX	R11, R9, R12
	BC	4, 2, atCountAgain2
atCountTrap2:
	SC
	B	error6
	XOR	R0, R0, R0
	STW	R3, r1+28(FP)
	STW	R4, r2+32(FP)
	STW	R0, err+36(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
error6:
	XOR	R0, R0, R0
	STW	R3, err+36(FP)
	ADDI	R3, R0, $-1
	STW	R3, r1+28(FP)
	STW	R0, r2+32(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET

// func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
TEXT ·RawSyscall(SB),NOSPLIT,$0-28
	LWZ	R3, a1+4(FP)
	LWZ	R4, a2+8(FP)
	LWZ	R5, a3+12(FP)
	LWZ	R0, trap+0(FP)
	// Count only syscall 349; preserve R0 and all syscall operands.
	ADDI R12, R0, $349
	CMPW R0, R12
	BC	4, 2, atCountTrap3
	ADDI	$·at349Count(SB), R9
	ADDI	R12, R0, $0
atCountAgain3:
	LWARX	R11, R9, R12
	ADDI	R11, R11, $1
	STWCX	R11, R9, R12
	BC	4, 2, atCountAgain3
atCountTrap3:
	SC
	B	rawerror
	XOR	R0, R0, R0
	STW	R3, r1+16(FP)
	STW	R4, r2+20(FP)
	STW	R0, err+24(FP)
	RET
rawerror:
	XOR	R0, R0, R0
	STW	R3, err+24(FP)
	ADDI	R3, R0, $-1
	STW	R3, r1+16(FP)
	STW	R0, r2+20(FP)
	RET

// func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
TEXT ·RawSyscall6(SB),NOSPLIT,$0-40
	LWZ	R3, a1+4(FP)
	LWZ	R4, a2+8(FP)
	LWZ	R5, a3+12(FP)
	LWZ	R6, a4+16(FP)
	LWZ	R7, a5+20(FP)
	LWZ	R8, a6+24(FP)
	LWZ	R0, trap+0(FP)
	// Count only syscall 349; preserve R0 and all syscall operands.
	ADDI R12, R0, $349
	CMPW R0, R12
	BC	4, 2, atCountTrap4
	ADDI	$·at349Count(SB), R9
	ADDI	R12, R0, $0
atCountAgain4:
	LWARX	R11, R9, R12
	ADDI	R11, R11, $1
	STWCX	R11, R9, R12
	BC	4, 2, atCountAgain4
atCountTrap4:
	SC
	B	rawerror6
	XOR	R0, R0, R0
	STW	R3, r1+28(FP)
	STW	R4, r2+32(FP)
	STW	R0, err+36(FP)
	RET
rawerror6:
	XOR	R0, R0, R0
	STW	R3, err+36(FP)
	ADDI	R3, R0, $-1
	STW	R3, r1+28(FP)
	STW	R0, r2+32(FP)
	RET
