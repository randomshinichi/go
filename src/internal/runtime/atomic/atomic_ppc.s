// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// Keep the pointer arguments visible to escape analysis in the Go declarations.
// The 32-bit pointer ABI slots match the raw word operations.
TEXT ·Casp1(SB),NOSPLIT,$0-13
	B ·casRaw(SB)

TEXT ·StorepNoWB(SB),NOSPLIT,$0-8
	B ·storeRaw(SB)

TEXT ·casRaw(SB),NOSPLIT,$0-13
	LWZ R3, ptr+0(FP)
	LWZ R4, old+4(FP)
	LWZ R5, new+8(FP)
	// R0 is an ordinary indexed-address register here; establish zero before
	// using it as the second operand to lwarx/stwcx.
	ADDI R0, R0, $0
	SYNC
cas_again:
	LWARX R6, R3, R0
	CMPW R6, R4
	BC 4, 2, cas_fail
	STWCX R5, R3, R0
	BC 4, 2, cas_again
	SYNC
	ADDI R3, R0, $1
	STB R3, ret+12(FP)
	RET
cas_fail:
	SYNC
	STB R0, ret+12(FP)
	RET

