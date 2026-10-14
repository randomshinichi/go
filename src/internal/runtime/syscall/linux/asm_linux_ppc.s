// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// PowerPC Linux passes system call numbers in r0, arguments in r3-r8,
// and reports errors by setting CR0.SO and placing errno in r3.
//
// func Syscall6(num, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, errno uintptr)
TEXT ·Syscall6<ABIInternal>(SB), NOSPLIT|NOFRAME, $0-40
	LWZ R0, num+0(FP)
	LWZ R3, a1+4(FP)
	LWZ R4, a2+8(FP)
	LWZ R5, a3+12(FP)
	LWZ R6, a4+16(FP)
	LWZ R7, a5+20(FP)
	LWZ R8, a6+24(FP)
	SC
	BC 4, 3, syscall_ok // CR0.SO clear means the syscall succeeded.
	STW R3, errno+36(FP)
	ADDI R3, R0, $-1
	B syscall_done
syscall_ok:
	ADDI R0, R0, $0
	STW R0, errno+36(FP)
syscall_done:
	STW R3, r1+28(FP)
	STW R4, r2+32(FP)
	RET
