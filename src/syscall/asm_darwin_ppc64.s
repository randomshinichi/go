// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//
// System call support for PPC64, Darwin (Mac OS X 10.5 Leopard)
//
// Convention, measured on the G5 (out/stage2-20261003/D7-syscall/g5-request/
// out-d7/06-rawsc.txt, and libSystem's own stubs in 07-selected.txt, e.g.
// _write: li r0,4; sc; b <cerror path>; b <return>; ...):
//   - the system call number is in R0, arguments in R3, R4, ...
//   - the kernel resumes at the instruction after sc on failure, with the
//     errno value in R3, and two instructions after sc on success, with the
//     result in R3 and a second result in R4. Failure is therefore the
//     branch immediately following SYSCALL; success skips it.
//   - SYSCALL with no operand is a bare sc: the register forms used on linux
//     emit a move before and an R0 restore after, which would move the two
//     resume points.
//
// R0 holds the system call number across the sc, and Go code expects R0 to be
// zero, so every path clears it again before using it as zero.

// func Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
TEXT	·Syscall(SB),NOSPLIT,$0-56
	BL	runtime·entersyscall<ABIInternal>(SB)
	MOVD	a1+8(FP), R3
	MOVD	a2+16(FP), R4
	MOVD	a3+24(FP), R5
	MOVD	trap+0(FP), R0
	SYSCALL
	BR	error
	XOR	R0, R0, R0
	MOVD	R3, r1+32(FP)
	MOVD	R4, r2+40(FP)
	MOVD	R0, err+48(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
error:
	XOR	R0, R0, R0
	MOVD	R3, err+48(FP)
	MOVD	$-1, R3
	MOVD	R3, r1+32(FP)
	MOVD	R0, r2+40(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET

// func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
TEXT	·Syscall6(SB),NOSPLIT,$0-80
	BL	runtime·entersyscall<ABIInternal>(SB)
	MOVD	a1+8(FP), R3
	MOVD	a2+16(FP), R4
	MOVD	a3+24(FP), R5
	MOVD	a4+32(FP), R6
	MOVD	a5+40(FP), R7
	MOVD	a6+48(FP), R8
	MOVD	trap+0(FP), R0
	SYSCALL
	BR	error6
	XOR	R0, R0, R0
	MOVD	R3, r1+56(FP)
	MOVD	R4, r2+64(FP)
	MOVD	R0, err+72(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET
error6:
	XOR	R0, R0, R0
	MOVD	R3, err+72(FP)
	MOVD	$-1, R3
	MOVD	R3, r1+56(FP)
	MOVD	R0, r2+64(FP)
	BL	runtime·exitsyscall<ABIInternal>(SB)
	RET

// func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
TEXT	·RawSyscall(SB),NOSPLIT,$0-56
	MOVD	a1+8(FP), R3
	MOVD	a2+16(FP), R4
	MOVD	a3+24(FP), R5
	MOVD	trap+0(FP), R0
	SYSCALL
	BR	rawerror
	XOR	R0, R0, R0
	MOVD	R3, r1+32(FP)
	MOVD	R4, r2+40(FP)
	MOVD	R0, err+48(FP)
	RET
rawerror:
	XOR	R0, R0, R0
	MOVD	R3, err+48(FP)
	MOVD	$-1, R3
	MOVD	R3, r1+32(FP)
	MOVD	R0, r2+40(FP)
	RET

// func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err Errno)
TEXT	·RawSyscall6(SB),NOSPLIT,$0-80
	MOVD	a1+8(FP), R3
	MOVD	a2+16(FP), R4
	MOVD	a3+24(FP), R5
	MOVD	a4+32(FP), R6
	MOVD	a5+40(FP), R7
	MOVD	a6+48(FP), R8
	MOVD	trap+0(FP), R0
	SYSCALL
	BR	rawerror6
	XOR	R0, R0, R0
	MOVD	R3, r1+56(FP)
	MOVD	R4, r2+64(FP)
	MOVD	R0, err+72(FP)
	RET
rawerror6:
	XOR	R0, R0, R0
	MOVD	R3, err+72(FP)
	MOVD	$-1, R3
	MOVD	R3, r1+56(FP)
	MOVD	R0, r2+64(FP)
	RET
