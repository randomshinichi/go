// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"

// Compiler-private WB convention, NOT an ordinary Go call:
// R29 = requested bytes / returned reservation pointer. Entries are 4 bytes.
// R18/R19/R31, condition codes and FPRs may be clobbered. All other
// allocatable GPRs and g are preserved, including across wbBufFlush.
// The runtime's flush forbids GC safe points for these untyped saved words.
TEXT gcWriteBarrier<>(SB),NOSPLIT,$104-0
 NO_LOCAL_POINTERS
retry:
 LWZ R18, g_m(g)
 LWZ R18, m_p(R18)
 LWZ R19, (p_wbBuf+wbBuf_next)(R18)
 LWZ R31, (p_wbBuf+wbBuf_end)(R18)
 ADD R19, R19, R29
 CMPLW R31, R19
 BC 12, 0, flush
 STW R19, (p_wbBuf+wbBuf_next)(R18)
 SUBF R29, R29, R19
 RET
flush:
 STW R3, 24(R1)
 STW R4, 28(R1)
 STW R5, 32(R1)
 STW R6, 36(R1)
 STW R7, 40(R1)
 STW R8, 44(R1)
 STW R9, 48(R1)
 STW R10, 52(R1)
 STW R11, 56(R1)
 STW R12, 60(R1)
 STW R14, 64(R1)
 STW R15, 68(R1)
 STW R16, 72(R1)
 STW R17, 76(R1)
 STW R20, 80(R1)
 STW R21, 84(R1)
 STW R22, 88(R1)
 STW R23, 92(R1)
 STW R24, 96(R1)
 STW R25, 100(R1)
 STW R26, 104(R1)
 STW R27, 108(R1)
 STW R28, 112(R1)
 STW R29, 116(R1)
 STW R30, 120(R1)
 BL runtime·wbBufFlush<ABIInternal>(SB)
 LWZ R3, 24(R1)
 LWZ R4, 28(R1)
 LWZ R5, 32(R1)
 LWZ R6, 36(R1)
 LWZ R7, 40(R1)
 LWZ R8, 44(R1)
 LWZ R9, 48(R1)
 LWZ R10, 52(R1)
 LWZ R11, 56(R1)
 LWZ R12, 60(R1)
 LWZ R14, 64(R1)
 LWZ R15, 68(R1)
 LWZ R16, 72(R1)
 LWZ R17, 76(R1)
 LWZ R20, 80(R1)
 LWZ R21, 84(R1)
 LWZ R22, 88(R1)
 LWZ R23, 92(R1)
 LWZ R24, 96(R1)
 LWZ R25, 100(R1)
 LWZ R26, 104(R1)
 LWZ R27, 108(R1)
 LWZ R28, 112(R1)
 LWZ R29, 116(R1)
 LWZ R30, 120(R1)
 B retry

TEXT runtime·gcWriteBarrier1<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $4
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier2<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $8
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier3<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $12
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier4<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $16
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier5<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $20
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier6<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $24
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier7<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $28
 B gcWriteBarrier<>(SB)

TEXT runtime·gcWriteBarrier8<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
 ADDI R29, R0, $32
 B gcWriteBarrier<>(SB)
