// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"
#include "funcdata.h"

// Darwin ppc32 has a 24-byte fixed frame; outgoing arguments start at SP+24.
// The frame is 48 bytes so the local byte and call area preserve 16-byte
// alignment. Function-value context arrives in R11.
#define ARG_CTXT 24
#define ARG_FRAME 28
#define ARG_RETVALID 32
#define ARG_REGARGS 36
#define LOCAL_RETVALID 40

// makeFuncStub is the code half of the function returned by MakeFunc.
TEXT ·makeFuncStub(SB),(NOSPLIT|WRAPPER),$48
	NO_LOCAL_POINTERS
	ADDI	R0, R0, $0	// R0 is not inherently zero on PPC32.
	STW	R11, ARG_CTXT(R1)
	ADDI	R3, R1, $argframe+0(FP)
	STW	R3, ARG_FRAME(R1)
	STB	R0, LOCAL_RETVALID(R1)
	ADDI	R3, R1, $LOCAL_RETVALID
	STW	R3, ARG_RETVALID(R1)
	STW	R0, ARG_REGARGS(R1)
	BL	·callReflect(SB)
	RET

// methodValueCall is the code half of the function returned by makeMethodValue.
TEXT ·methodValueCall(SB),(NOSPLIT|WRAPPER),$48
	NO_LOCAL_POINTERS
	ADDI	R0, R0, $0	// R0 is not inherently zero on PPC32.
	STW	R11, ARG_CTXT(R1)
	ADDI	R3, R1, $argframe+0(FP)
	STW	R3, ARG_FRAME(R1)
	STB	R0, LOCAL_RETVALID(R1)
	ADDI	R3, R1, $LOCAL_RETVALID
	STW	R3, ARG_RETVALID(R1)
	STW	R0, ARG_REGARGS(R1)
	BL	·callMethod(SB)
	RET
