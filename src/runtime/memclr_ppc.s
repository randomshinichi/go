// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// memclrNoHeapPointers clears exactly n bytes. Byte stores avoid alignment
// assumptions and never touch the byte immediately following the range.
TEXT runtime·memclrNoHeapPointers(SB), NOSPLIT|NOFRAME, $0-8
	LWZ	R3, ptr+0(FP)
	LWZ	R4, n+4(FP)
	XOR	R5, R5, R5
	CMPWI	R4, $0
	BC	12, 2, memclr_done
memclr_loop:
	STB	R5, 0(R3)
	ADDI	R3, R3, $1
	ADDI	R4, R4, $-1
	CMPWI	R4, $0
	BC	4, 2, memclr_loop
memclr_done:
	RET
