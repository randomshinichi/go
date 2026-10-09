// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// Darwin supplies sig, info, ctxt in R3-R5; Go's PPC32 ABI is stack-only.
// Like PPC64's host-to-Go bridge, preserve the C callee-saved GPRs/FPRs
// until returning to libc's signal trampoline, which then restores the
// interrupted machine context. R13 is reserved and never changed by Go.
// 24-byte linkage + 12-byte arguments + saved registers, rounded to 256.
// Manage the frame explicitly so the generic Go epilogue does not clobber
// C's callee-saved R31 when loading LR.
TEXT runtime·sigtramp(SB), NOSPLIT|NOFRAME|TOPFRAME, $0-0
	MFLR R0
	STW R0, 8(R1)
	ADDI R0, R0, $0
	STWU R1, -256(R1)
	STW R3, 24(R1)
	STW R4, 28(R1)
	STW R5, 32(R1)
	STW R14, 40(R1)
	STW R15, 44(R1)
	STW R16, 48(R1)
	STW R17, 52(R1)
	STW R18, 56(R1)
	STW R19, 60(R1)
	STW R20, 64(R1)
	STW R21, 68(R1)
	STW R22, 72(R1)
	STW R23, 76(R1)
	STW R24, 80(R1)
	STW R25, 84(R1)
	STW R26, 88(R1)
	STW R27, 92(R1)
	STW R28, 96(R1)
	STW R29, 100(R1)
	STW R30, 104(R1)
	STW R31, 108(R1)
	MFCR R31
	STW R31, 36(R1)
	STFD F14, 112(R1)
	STFD F15, 120(R1)
	STFD F16, 128(R1)
	STFD F17, 136(R1)
	STFD F18, 144(R1)
	STFD F19, 152(R1)
	STFD F20, 160(R1)
	STFD F21, 168(R1)
	STFD F22, 176(R1)
	STFD F23, 184(R1)
	STFD F24, 192(R1)
	STFD F25, 200(R1)
	STFD F26, 208(R1)
	STFD F27, 216(R1)
	STFD F28, 224(R1)
	STFD F29, 232(R1)
	STFD F30, 240(R1)
	STFD F31, 248(R1)
	// Call indirectly, as PPC64 does: this is an OS-provided signal stack,
	// not a small Go stack to which the linker's nosplit budget applies.
	ADDI $runtime·sigtrampgo<ABIInternal>(SB), R12
	MTCTR R12
	BCTRL
	LFD 112(R1), F14
	LFD 120(R1), F15
	LFD 128(R1), F16
	LFD 136(R1), F17
	LFD 144(R1), F18
	LFD 152(R1), F19
	LFD 160(R1), F20
	LFD 168(R1), F21
	LFD 176(R1), F22
	LFD 184(R1), F23
	LFD 192(R1), F24
	LFD 200(R1), F25
	LFD 208(R1), F26
	LFD 216(R1), F27
	LFD 224(R1), F28
	LFD 232(R1), F29
	LFD 240(R1), F30
	LFD 248(R1), F31
	LWZ R14, 40(R1)
	LWZ R15, 44(R1)
	LWZ R16, 48(R1)
	LWZ R17, 52(R1)
	LWZ R18, 56(R1)
	LWZ R19, 60(R1)
	LWZ R20, 64(R1)
	LWZ R21, 68(R1)
	LWZ R22, 72(R1)
	LWZ R23, 76(R1)
	LWZ R24, 80(R1)
	LWZ R25, 84(R1)
	LWZ R26, 88(R1)
	LWZ R27, 92(R1)
	LWZ R28, 96(R1)
	LWZ R29, 100(R1)
	LWZ R30, 104(R1)
	LWZ R31, 36(R1)
	MTCRF $0xff, R31
	LWZ R31, 108(R1)
	LWZ R0, 264(R1)
	MTLR R0
	ADDI R1, R1, $256
	RET

// cgo is not supported by the ppc32 Darwin port; forward without a frame.
TEXT runtime·cgoSigtramp(SB), NOSPLIT|NOFRAME, $0-0
	B runtime·sigtramp(SB)
