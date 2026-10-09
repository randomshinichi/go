// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"
#include "funcdata.h"

// Darwin ppc32 uses the measured 24-byte linkage area. ABI0 arguments start
// at SP+24, and the caller's stack is 16-byte aligned.
#define PPC32_FIXED_FRAME 24

#define DISPATCH(NAME, HIGH, LOW) \
	ADDIS R4, R0, $HIGH; \
	ADDI R4, R4, $LOW; \
	CMPLW R3, R4; \
	BC 12, 1, 2(PC); \
	B runtime·NAME(SB)

// reflectcall dispatches to a fixed-size frame so the GC has a static frame
// map. The arguments already use stack-only D-2 layout; callN copies them to
// the target's incoming argument area and copies results back afterward.
TEXT runtime·reflectcall(SB), NOSPLIT|NOFRAME, $0-28
	LWZ R3, frameSize+20(FP)
	DISPATCH(call16, 0, 16)
	DISPATCH(call32, 0, 32)
	DISPATCH(call64, 0, 64)
	DISPATCH(call128, 0, 128)
	DISPATCH(call256, 0, 256)
	DISPATCH(call512, 0, 512)
	DISPATCH(call1024, 0, 1024)
	DISPATCH(call2048, 0, 2048)
	DISPATCH(call4096, 0, 4096)
	DISPATCH(call8192, 0, 8192)
	DISPATCH(call16384, 0, 16384)
	DISPATCH(call32768, 1, -32768)
	DISPATCH(call65536, 1, 0)
	DISPATCH(call131072, 2, 0)
	DISPATCH(call262144, 4, 0)
	DISPATCH(call524288, 8, 0)
	DISPATCH(call1048576, 16, 0)
	DISPATCH(call2097152, 32, 0)
	DISPATCH(call4194304, 64, 0)
	DISPATCH(call8388608, 128, 0)
	DISPATCH(call16777216, 256, 0)
	DISPATCH(call33554432, 512, 0)
	DISPATCH(call67108864, 1024, 0)
	DISPATCH(call134217728, 2048, 0)
	DISPATCH(call268435456, 4096, 0)
	DISPATCH(call536870912, 8192, 0)
	DISPATCH(call1073741824, 16384, 0)
	B runtime·badreflectcall(SB)

#define CALLFN(NAME, MAXSIZE) \
TEXT NAME(SB), WRAPPER, $MAXSIZE-28; \
	NO_LOCAL_POINTERS; \
	/* Copy stack arguments into the target's SP+24 argument area. */ \
	LWZ R3, stackArgs+8(FP); \
	LWZ R4, stackArgsSize+12(FP); \
	ADDI R5, R1, $PPC32_FIXED_FRAME; \
	CMPWI R4, $0; \
	BC 12, 2, callfn; \
copyargs: \
	LBZ R6, 0(R3); \
	STB R6, 0(R5); \
	ADDI R3, R3, $1; \
	ADDI R5, R5, $1; \
	ADDI R4, R4, $-1; \
	CMPWI R4, $0; \
	BC 12, 1, copyargs; \
callfn: \
	LWZ R11, f+4(FP); \
	LWZ R12, 0(R11); \
	MTCTR R12; \
	PCDATA $PCDATA_StackMapIndex, $0; \
	BCTRL; \
	/* Original reflectcall arguments remain in the caller's frame. */ \
	LWZ R7, stackArgsType+0(FP); \
	STW R7, PPC32_FIXED_FRAME+0(R1); \
	LWZ R3, stackArgs+8(FP); \
	LWZ R4, stackRetOffset+16(FP); \
	ADD R3, R3, R4; \
	STW R3, PPC32_FIXED_FRAME+4(R1); \
	ADDI R5, R1, $PPC32_FIXED_FRAME; \
	ADD R5, R5, R4; \
	STW R5, PPC32_FIXED_FRAME+8(R1); \
	LWZ R6, stackArgsSize+12(FP); \
	SUBF R6, R4, R6; \
	STW R6, PPC32_FIXED_FRAME+12(R1); \
	STW R0, PPC32_FIXED_FRAME+16(R1); \
	/* reflectcallmove applies write barriers while copying return values. */ \
	BL runtime·callRet(SB); \
	RET


CALLFN(runtime·call16, 16)
CALLFN(runtime·call32, 32)
CALLFN(runtime·call64, 64)
CALLFN(runtime·call128, 128)
CALLFN(runtime·call256, 256)
CALLFN(runtime·call512, 512)
CALLFN(runtime·call1024, 1024)
CALLFN(runtime·call2048, 2048)
CALLFN(runtime·call4096, 4096)
CALLFN(runtime·call8192, 8192)
CALLFN(runtime·call16384, 16384)
CALLFN(runtime·call32768, 32768)
CALLFN(runtime·call65536, 65536)
CALLFN(runtime·call131072, 131072)
CALLFN(runtime·call262144, 262144)
CALLFN(runtime·call524288, 524288)
CALLFN(runtime·call1048576, 1048576)
CALLFN(runtime·call2097152, 2097152)
CALLFN(runtime·call4194304, 4194304)
CALLFN(runtime·call8388608, 8388608)
CALLFN(runtime·call16777216, 16777216)
CALLFN(runtime·call33554432, 33554432)
CALLFN(runtime·call67108864, 67108864)
CALLFN(runtime·call134217728, 134217728)
CALLFN(runtime·call268435456, 268435456)
CALLFN(runtime·call536870912, 536870912)
CALLFN(runtime·call1073741824, 1073741824)

// callRet packages register values into the ABI0 argument area for the
// nosplit runtime.reflectcallmove helper.
TEXT runtime·callRet(SB), NOSPLIT, $20-0
	NO_LOCAL_POINTERS
	STW R7, PPC32_FIXED_FRAME+0(R1)
	STW R3, PPC32_FIXED_FRAME+4(R1)
	STW R5, PPC32_FIXED_FRAME+8(R1)
	STW R6, PPC32_FIXED_FRAME+12(R1)
	STW R0, PPC32_FIXED_FRAME+16(R1)
	BL runtime·reflectcallmove(SB)
	RET
