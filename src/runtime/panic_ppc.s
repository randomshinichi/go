// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "funcdata.h"

// Compiler-private bounds protocol, independent of the C ABI:
// PCDATA_PanicBounds numbers R3..R9 as slots 0..6. panicExtend packs
// hi/lo slots from R3..R6 into two two-bit indices (panicBounds32X).
// Both Go ABIs are stack-only: outgoing pc/regs are at SP+24/+28.
// 72 locals + 24 linkage = 96 bytes, including a full [16]int32 save area.
// obj/ppc saves the incoming LR at caller SP+8 before allocating the frame.
TEXT runtime·panicBounds<ABIInternal>(SB),NOSPLIT,$72-0
	NO_LOCAL_POINTERS
	STW R3, 32(R1)
	STW R4, 36(R1)
	STW R5, 40(R1)
	STW R6, 44(R1)
	STW R7, 48(R1)
	STW R8, 52(R1)
	STW R9, 56(R1)
	LWZ R3, 104(R1)
	STW R3, 24(R1)
	ADDI R3, R1, $32
	STW R3, 28(R1)
	BL runtime·panicBounds32<ABIInternal>(SB)
	RET

TEXT runtime·panicExtend<ABIInternal>(SB),NOSPLIT,$72-0
	NO_LOCAL_POINTERS
	STW R3, 32(R1)
	STW R4, 36(R1)
	STW R5, 40(R1)
	STW R6, 44(R1)
	STW R7, 48(R1)
	STW R8, 52(R1)
	STW R9, 56(R1)
	LWZ R3, 104(R1)
	STW R3, 24(R1)
	ADDI R3, R1, $32
	STW R3, 28(R1)
	BL runtime·panicBounds32X<ABIInternal>(SB)
	RET
