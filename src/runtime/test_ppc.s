// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"
#include "funcdata.h"

// func wbProtocolCheck(iters uint32) (bad, flushes uint32)
//
// Calls gcWriteBarrier8 iters times. Before each call every GPR the protocol
// promises to preserve (wb_ppc.s: all allocatable GPRs except R18/R19, plus g)
// holds 0x100+n; after it, each must still hold its sentinel, R30 must still be
// g, and R29 must point at 8 free 4-byte entries ending at p.wbBuf.next. The
// reserved entries are zeroed, so a flush ignores them (wbBufFlush1 skips
// values below minLegalPointer). 512 entries / 8 per call means iters > 64
// forces the flush path; flushes counts how often R29 moved backwards.
// bad is 0, or the number of the first broken register (29: reservation).
// R18 carries that number to fail; it is WB scratch, so it is free here.
TEXT ·wbProtocolCheck(SB), 0, $16-12
	NO_LOCAL_POINTERS
	LWZ	R3, iters+0(FP)
	STW	R3, 24(R1)	// iterations left
	ADDI	R3, R0, $0
	STW	R3, 28(R1)	// flushes seen
	STW	R30, 32(R1)	// g
	STW	R3, 36(R1)	// previous reservation
loop:
	ADDI	R3, R0, $0x103
	ADDI	R4, R0, $0x104
	ADDI	R5, R0, $0x105
	ADDI	R6, R0, $0x106
	ADDI	R7, R0, $0x107
	ADDI	R8, R0, $0x108
	ADDI	R9, R0, $0x109
	ADDI	R10, R0, $0x10a
	ADDI	R11, R0, $0x10b
	ADDI	R12, R0, $0x10c
	ADDI	R14, R0, $0x10e
	ADDI	R15, R0, $0x10f
	ADDI	R16, R0, $0x110
	ADDI	R17, R0, $0x111
	ADDI	R20, R0, $0x114
	ADDI	R21, R0, $0x115
	ADDI	R22, R0, $0x116
	ADDI	R23, R0, $0x117
	ADDI	R24, R0, $0x118
	ADDI	R25, R0, $0x119
	ADDI	R26, R0, $0x11a
	ADDI	R27, R0, $0x11b
	ADDI	R28, R0, $0x11c
	BL	runtime·gcWriteBarrier8<ABIInternal>(SB)
	ADDI	R18, R0, $3
	CMPWI	R3, $0x103
	BC	4, 2, fail
	ADDI	R18, R0, $4
	CMPWI	R4, $0x104
	BC	4, 2, fail
	ADDI	R18, R0, $5
	CMPWI	R5, $0x105
	BC	4, 2, fail
	ADDI	R18, R0, $6
	CMPWI	R6, $0x106
	BC	4, 2, fail
	ADDI	R18, R0, $7
	CMPWI	R7, $0x107
	BC	4, 2, fail
	ADDI	R18, R0, $8
	CMPWI	R8, $0x108
	BC	4, 2, fail
	ADDI	R18, R0, $9
	CMPWI	R9, $0x109
	BC	4, 2, fail
	ADDI	R18, R0, $10
	CMPWI	R10, $0x10a
	BC	4, 2, fail
	ADDI	R18, R0, $11
	CMPWI	R11, $0x10b
	BC	4, 2, fail
	ADDI	R18, R0, $12
	CMPWI	R12, $0x10c
	BC	4, 2, fail
	ADDI	R18, R0, $14
	CMPWI	R14, $0x10e
	BC	4, 2, fail
	ADDI	R18, R0, $15
	CMPWI	R15, $0x10f
	BC	4, 2, fail
	ADDI	R18, R0, $16
	CMPWI	R16, $0x110
	BC	4, 2, fail
	ADDI	R18, R0, $17
	CMPWI	R17, $0x111
	BC	4, 2, fail
	ADDI	R18, R0, $20
	CMPWI	R20, $0x114
	BC	4, 2, fail
	ADDI	R18, R0, $21
	CMPWI	R21, $0x115
	BC	4, 2, fail
	ADDI	R18, R0, $22
	CMPWI	R22, $0x116
	BC	4, 2, fail
	ADDI	R18, R0, $23
	CMPWI	R23, $0x117
	BC	4, 2, fail
	ADDI	R18, R0, $24
	CMPWI	R24, $0x118
	BC	4, 2, fail
	ADDI	R18, R0, $25
	CMPWI	R25, $0x119
	BC	4, 2, fail
	ADDI	R18, R0, $26
	CMPWI	R26, $0x11a
	BC	4, 2, fail
	ADDI	R18, R0, $27
	CMPWI	R27, $0x11b
	BC	4, 2, fail
	ADDI	R18, R0, $28
	CMPWI	R28, $0x11c
	BC	4, 2, fail
	// g survived.
	ADDI	R18, R0, $30
	LWZ	R3, 32(R1)
	CMPW	R30, R3
	BC	4, 2, fail
	// The reservation is [R29, R29+32) inside p.wbBuf.buf, and next is R29+32.
	ADDI	R18, R0, $29
	LWZ	R4, g_m(R30)
	LWZ	R4, m_p(R4)
	LWZ	R5, (p_wbBuf+wbBuf_next)(R4)
	ADDI	R6, R29, $32
	CMPW	R5, R6
	BC	4, 2, fail
	ADDI	R7, R4, $(p_wbBuf+wbBuf_buf)
	CMPLW	R29, R7
	BC	12, 0, fail	// R29 below buf[0]
	LWZ	R8, (p_wbBuf+wbBuf_end)(R4)
	CMPLW	R8, R6
	BC	12, 0, fail	// R29+32 past end
	ADDI	R3, R0, $0
	STW	R3, 0(R29)
	STW	R3, 4(R29)
	STW	R3, 8(R29)
	STW	R3, 12(R29)
	STW	R3, 16(R29)
	STW	R3, 20(R29)
	STW	R3, 24(R29)
	STW	R3, 28(R29)
	// A reservation below the previous one means the buffer was flushed.
	LWZ	R5, 36(R1)
	STW	R29, 36(R1)
	CMPLW	R29, R5
	BC	4, 0, noflush
	LWZ	R5, 28(R1)
	ADDI	R5, R5, $1
	STW	R5, 28(R1)
noflush:
	LWZ	R3, 24(R1)
	ADDI	R3, R3, $-1
	STW	R3, 24(R1)
	CMPWI	R3, $0
	BC	4, 2, loop
	ADDI	R18, R0, $0
fail:
	STW	R18, bad+4(FP)
	LWZ	R3, 28(R1)
	STW	R3, flushes+8(FP)
	CMPWI	R18, $0
	BC	12, 2, done
	// Failed with entries possibly unzeroed: discard the buffer so no flush
	// reads them. A failing run is not trusted past this test.
	LWZ	R30, 32(R1)
	LWZ	R4, g_m(R30)
	LWZ	R4, m_p(R4)
	ADDI	R5, R4, $(p_wbBuf+wbBuf_buf)
	STW	R5, (p_wbBuf+wbBuf_next)(R4)
done:
	RET
