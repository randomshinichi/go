// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// sigtramp is entered with the Darwin C signal-handler arguments in R3-R5.
// On this port, the current Go g is kept in callee-saved R30, not in R13.
// Preserve Darwin's callee-saved integer and floating-point registers around
// the Go ABIInternal call; the kernel restores the interrupted caller state
// from ucontext when libc returns from the signal callback.
TEXT runtime·sigtramp(SB), NOSPLIT, $232-0
	STW	R2, 24(R1)
	STW	R13, 28(R1)
	STW	R14, 32(R1)
	STW	R15, 36(R1)
	STW	R16, 40(R1)
	STW	R17, 44(R1)
	STW	R18, 48(R1)
	STW	R19, 52(R1)
	STW	R20, 56(R1)
	STW	R21, 60(R1)
	STW	R22, 64(R1)
	STW	R23, 68(R1)
	STW	R24, 72(R1)
	STW	R25, 76(R1)
	STW	R26, 80(R1)
	STW	R27, 84(R1)
	STW	R28, 88(R1)
	STW	R29, 92(R1)
	STW	R30, 96(R1)
	STW	R31, 100(R1)
	STFD	F14, 104(R1)
	STFD	F15, 112(R1)
	STFD	F16, 120(R1)
	STFD	F17, 128(R1)
	STFD	F18, 136(R1)
	STFD	F19, 144(R1)
	STFD	F20, 152(R1)
	STFD	F21, 160(R1)
	STFD	F22, 168(R1)
	STFD	F23, 176(R1)
	STFD	F24, 184(R1)
	STFD	F25, 192(R1)
	STFD	F26, 200(R1)
	STFD	F27, 208(R1)
	STFD	F28, 216(R1)
	STFD	F29, 224(R1)
	STFD	F30, 232(R1)
	STFD	F31, 240(R1)
	BL	runtime·sigtrampgo<ABIInternal>(SB)
	LFD	240(R1), F31
	LFD	232(R1), F30
	LFD	224(R1), F29
	LFD	216(R1), F28
	LFD	208(R1), F27
	LFD	200(R1), F26
	LFD	192(R1), F25
	LFD	184(R1), F24
	LFD	176(R1), F23
	LFD	168(R1), F22
	LFD	160(R1), F21
	LFD	152(R1), F20
	LFD	144(R1), F19
	LFD	136(R1), F18
	LFD	128(R1), F17
	LFD	120(R1), F16
	LFD	112(R1), F15
	LFD	104(R1), F14
	LWZ	R31, 100(R1)
	LWZ	R30, 96(R1)
	LWZ	R29, 92(R1)
	LWZ	R28, 88(R1)
	LWZ	R27, 84(R1)
	LWZ	R26, 80(R1)
	LWZ	R25, 76(R1)
	LWZ	R24, 72(R1)
	LWZ	R23, 68(R1)
	LWZ	R22, 64(R1)
	LWZ	R21, 60(R1)
	LWZ	R20, 56(R1)
	LWZ	R19, 52(R1)
	LWZ	R18, 48(R1)
	LWZ	R17, 44(R1)
	LWZ	R16, 40(R1)
	LWZ	R15, 36(R1)
	LWZ	R14, 32(R1)
	LWZ	R13, 28(R1)
	LWZ	R2, 24(R1)
	RET

// cgo is not supported by the ppc32 Darwin port; keep the entry ABI complete
// for pure-Go configurations by forwarding to the identical signal handler.
TEXT runtime·cgoSigtramp(SB), NOSPLIT, $8-0
	B	runtime·sigtramp(SB)
