// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

TEXT ·casRaw(SB),NOSPLIT,$0-13
	LWZ R3, ptr+0(FP)
	LWZ R4, old+4(FP)
	LWZ R5, new+8(FP)
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

