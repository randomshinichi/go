// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// memequal compares bytes without alignment assumptions, so it is safe for all
// G3/G4 addresses and does not depend on word endianness.
TEXT runtime·memequal(SB), NOSPLIT, $0-13
	LWZ R3, a+0(FP)
	LWZ R4, b+4(FP)
	LWZ R5, size+8(FP)
	CMPWI R5, $0
	BC 12, 2, equal
loop:
	LBZ R6, 0(R3)
	LBZ R7, 0(R4)
	CMPW R6, R7
	BC 4, 2, unequal
	ADDI R3, R3, $1
	ADDI R4, R4, $1
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, loop
equal:
	ADDI R3, R0, $1
	STW R3, ret+12(FP)
	RET
unequal:
	ADDI R3, R0, $0
	STW R3, ret+12(FP)
	RET
