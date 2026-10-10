// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"
#include "funcdata.h"

// Darwin ppc32 uses the measured 24-byte linkage area. ABI0 arguments start
// at SP+24, and the caller's stack is 16-byte aligned.
#define PPC32_FIXED_FRAME 24
// Physical offsets from the unmodified caller SP: ABI0 arguments start at +24.
#define CALL_ARG_TYPE 24
#define CALL_ARG_FN 28
#define CALL_ARG_STACKARGS 32
#define CALL_ARG_STACKSIZE 36
#define CALL_ARG_RETOFFSET 40

#define DISPATCH(NAME, HIGH, LOW) \
	ADDIS R4, R0, $HIGH; \
	ADDI R4, R4, $LOW; \
	CMPLW R3, R4; \
	BC 12, 1, 2(PC); \
	B NAME(SB)

// reflectcall dispatches to a fixed-size frame so the GC has a static frame
// map. The arguments already use stack-only D-2 layout; callN copies them to
// the target's incoming argument area and copies results back afterward.
TEXT runtime·reflectcall(SB), NOSPLIT|NOFRAME, $0-28
	LWZ R3, frameSize+20(FP)
	DISPATCH(runtime·call16, 0, 16)
	DISPATCH(runtime·call32, 0, 32)
	DISPATCH(runtime·call64, 0, 64)
	DISPATCH(runtime·call128, 0, 128)
	DISPATCH(runtime·call256, 0, 256)
	DISPATCH(runtime·call512, 0, 512)
	DISPATCH(runtime·call1024, 0, 1024)
	DISPATCH(runtime·call2048, 0, 2048)
	DISPATCH(runtime·call4096, 0, 4096)
	DISPATCH(runtime·call8192, 0, 8192)
	DISPATCH(runtime·call16384, 0, 16384)
	DISPATCH(runtime·call32768, 1, -32768)
	DISPATCH(runtime·call65536, 1, 0)
	DISPATCH(runtime·call131072, 2, 0)
	DISPATCH(runtime·call262144, 4, 0)
	DISPATCH(runtime·call524288, 8, 0)
	DISPATCH(runtime·call1048576, 16, 0)
	DISPATCH(runtime·call2097152, 32, 0)
	DISPATCH(runtime·call4194304, 64, 0)
	DISPATCH(runtime·call8388608, 128, 0)
	DISPATCH(runtime·call16777216, 256, 0)
	DISPATCH(runtime·call33554432, 512, 0)
	DISPATCH(runtime·call67108864, 1024, 0)
	DISPATCH(runtime·call134217728, 2048, 0)
	DISPATCH(runtime·call268435456, 4096, 0)
	DISPATCH(runtime·call536870912, 8192, 0)
	DISPATCH(runtime·call1073741824, 16384, 0)
	B runtime·badreflectcall(SB)

// callNN allocates its frame by DECLARATION rather than by hand. K is the
// class's whole frame -- its stack-argument area plus the measured 24-byte
// linkage area, rounded up to 16 -- and it is spelled as (HIGH<<16)+LOW so the
// body can rebuild it without a second literal.
//
// Why a declared frame: the unwinder describes a frame by the PCSP table this
// assembler builds from the Progs, and a hand-rolled `STWUX R1, R1, R4` moves
// SP by a value no static table can describe. The assembler therefore marks
// the function FuncFlagSPWrite, and traceback.go treats a non-innermost
// SPWRITE frame as fatal -- so any GC scanstack that walks through a live
// reflected call dies with "unexpected SPWRITE function runtime.call16".
// Declaring the frame makes the assembler generate the prologue (LR save,
// stack-growth check, SP allocation), the epilogue and the PCSP entries for
// every one of them, which is what ppc64, amd64 and 386 do with the same
// macro shape.
//
// LOCALS is what the assembler rounds into the frame: it lays out
// round16(locals + 24), so LOCALS = K-24 gives a frame of exactly K. That
// matters because R31 is the incoming ABI0 argument pointer, R1+K, and the
// argument offsets below are measured from it; a frame that is not exactly K
// would put R31 at the wrong address silently.
//
// `-28` is the argument size. callNN is reached by a branch from reflectcall,
// which declares $0-28, so callNN's incoming arguments are those same 28 bytes
// above its frame. Declaring them matters to the stack-growth path: a
// splittable function's prologue calls morestack, and morestack/newstack
// describe the frame it is growing from this function's own funcdata.
#define CALLFN(NAME, HIGH, LOW, LOCALS) \
TEXT NAME(SB), WRAPPER, $LOCALS-28; \
	NO_LOCAL_POINTERS; \
	/* Rebuild the incoming ABI0 argument pointer: R31 = R1 + K. */ \
	ADDIS R31, R0, $HIGH; \
	ADDI R31, R31, $LOW; \
	ADD R31, R1, R31; \
	/* Copy stack arguments into the target's SP+24 argument area. */ \
	LWZ R3, CALL_ARG_STACKARGS(R31); \
	LWZ R4, CALL_ARG_STACKSIZE(R31); \
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
	LWZ R11, CALL_ARG_FN(R31); \
	LWZ R12, 0(R11); \
	MTCTR R12; \
	PCDATA $PCDATA_StackMapIndex, $0; \
	BCTRL; \
	/* Recompute the old SP after calls, which may clobber R31. */ \
	ADDIS R31, R0, $HIGH; \
	ADDI R31, R31, $LOW; \
	ADD R31, R1, R31; \
	LWZ R7, CALL_ARG_TYPE(R31); \
	LWZ R3, CALL_ARG_STACKARGS(R31); \
	LWZ R4, CALL_ARG_RETOFFSET(R31); \
	ADD R3, R3, R4; \
	ADDI R5, R1, $PPC32_FIXED_FRAME; \
	ADD R5, R5, R4; \
	LWZ R6, CALL_ARG_STACKSIZE(R31); \
	SUBF R6, R4, R6; \
	BL runtime·callRet(SB); \
	RET


CALLFN(runtime·call16, 0, 48, 24)
CALLFN(runtime·call32, 0, 64, 40)
CALLFN(runtime·call64, 0, 96, 72)
CALLFN(runtime·call128, 0, 160, 136)
CALLFN(runtime·call256, 0, 288, 264)
CALLFN(runtime·call512, 0, 544, 520)
CALLFN(runtime·call1024, 0, 1056, 1032)
CALLFN(runtime·call2048, 0, 2080, 2056)
CALLFN(runtime·call4096, 0, 4128, 4104)
CALLFN(runtime·call8192, 0, 8224, 8200)
CALLFN(runtime·call16384, 0, 16416, 16392)
CALLFN(runtime·call32768, 1, -32736, 32776)
CALLFN(runtime·call65536, 1, 32, 65544)
CALLFN(runtime·call131072, 2, 32, 131080)
CALLFN(runtime·call262144, 4, 32, 262152)
CALLFN(runtime·call524288, 8, 32, 524296)
CALLFN(runtime·call1048576, 16, 32, 1048584)
CALLFN(runtime·call2097152, 32, 32, 2097160)
CALLFN(runtime·call4194304, 64, 32, 4194312)
CALLFN(runtime·call8388608, 128, 32, 8388616)
CALLFN(runtime·call16777216, 256, 32, 16777224)
CALLFN(runtime·call33554432, 512, 32, 33554440)
CALLFN(runtime·call67108864, 1024, 32, 67108872)
CALLFN(runtime·call134217728, 2048, 32, 134217736)
CALLFN(runtime·call268435456, 4096, 32, 268435464)
CALLFN(runtime·call536870912, 8192, 32, 536870920)
CALLFN(runtime·call1073741824, 16384, 32, 1073741832)

// callRet packages arguments into the ABI0 area for reflectcallmove.
// PPC32 has no ABI register arguments or returns, so RegArgs is nil.
TEXT runtime·callRet(SB), NOSPLIT, $20-0
	NO_LOCAL_POINTERS
	STW R7, PPC32_FIXED_FRAME+0(R1)
	STW R3, PPC32_FIXED_FRAME+4(R1)
	STW R5, PPC32_FIXED_FRAME+8(R1)
	STW R6, PPC32_FIXED_FRAME+12(R1)
	ADDI R0, R0, $0 // Explicitly materialize the null register-argument area.
	STW R0, PPC32_FIXED_FRAME+16(R1)
	BL runtime·reflectcallmove(SB)
	RET
