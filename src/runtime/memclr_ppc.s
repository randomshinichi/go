// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// memclrNoHeapPointers clears exactly n bytes. Aligned 4-byte words use
// word stores; byte stores in the head and tail avoid unaligned words and
// never touch the byte immediately following the range.
TEXT runtime·memclrNoHeapPointers(SB), NOSPLIT|NOFRAME, $0-8
	LWZ	R3, ptr+0(FP)
	LWZ	R4, n+4(FP)
	XOR	R5, R5, R5
	ADDI	R7, R0, $3
	CMPWI	R4, $0
	BC	12, 2, memclr_done
	AND	R6, R3, R7
	CMPWI	R6, $0
	BC	12, 2, memclr_words
memclr_head:
	STB	R5, 0(R3)
	ADDI	R3, R3, $1
	ADDI	R4, R4, $-1
	CMPWI	R4, $0
	BC	12, 2, memclr_done
	AND	R6, R3, R7
	CMPWI	R6, $0
	BC	4, 2, memclr_head
memclr_words:
	CMPWI	R4, $4
	BC	12, 0, memclr_tail
	STW	R5, 0(R3)
	ADDI	R3, R3, $4
	ADDI	R4, R4, $-4
	B memclr_words
memclr_tail:
	CMPWI	R4, $0
	BC	12, 2, memclr_done
	STB	R5, 0(R3)
	ADDI	R3, R3, $1
	ADDI	R4, R4, $-1
	CMPWI	R4, $0
	BC	4, 2, memclr_tail
memclr_done:
	RET
