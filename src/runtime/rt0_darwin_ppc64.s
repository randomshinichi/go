// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "go_asm.h"
#include "go_tls.h"
#include "asm_ppc64x.h"

// _rt0_ppc64_darwin is the executable's entry point: the srr0 of the
// LC_UNIXTHREAD load command, which the linker must point here.
//
// On entry the stack holds the 32-bit argc at 0(R1) and the argv pointers from
// 8(R1) (the first instructions of Apple's own crt1 `start`, disassembled from
// t_ppc64: `lwz r3,0(r26)` with r26 = the entry r1, and `addi r4,r26,8`).
// Unlike the other Darwin ports there is no argc/argv in registers, and no
// libc initialisation to wait for: dyld has already bound libSystem.
//
// Before rt0_go can call save_g, runtime·tls_g must hold the g slot offset, so
// this runs tlsinit first. It is NOSPLIT because g is not yet set.
TEXT _rt0_ppc64_darwin(SB),NOSPLIT,$48
	MOVD	$0, R0			// Go expects R0 == 0
	MOVD	R1, R14
	MOVWZ	0(R14), R3		// argc
	ADD	$8, R14, R4		// argv
	MOVD	R3, 48(R1)
	MOVD	R4, 56(R1)
	MOVD	R14, 64(R1)

	MOVD	$0, g			// make sure g is not junk
	MOVD	$runtime·tls_g(SB), R3
	MOVD	R3, 32(R1)		// arg 1: &tls_g
	MOVD	R13, 40(R1)		// arg 2: thread pointer
	BL	·tlsinit(SB)

	MOVD	48(R1), R3
	MOVD	56(R1), R4
	MOVD	64(R1), R1		// back to the entry stack pointer
	MOVD	$runtime·rt0_go(SB), R12
	MOVD	R12, CTR
	BR	(CTR)
