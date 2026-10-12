// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "funcdata.h"
#include "textflag.h"

// abi.BoundsEncode(abi.BoundsIndex, signed, xIsReg, yIsReg, x, y): bits
// 0|1|2 set, x at bit 3, y at bit 7, times numBoundsCodes (9).
#define BOUNDS(x, y) ((7 + ((x)<<3) + ((y)<<7)) * 9)
// panicExtend packs the high and low save slots of x as hi<<2 | lo.
#define EXTEND(hi, lo, y) BOUNDS(((hi)<<2) + (lo), y)

// func PanicSlot(site uint32)
// Sites 0-6 call panicBounds with x in slot site and y in slot (site+1)%7.
// Sites 7-10 call panicExtend with (hi, lo, y) = (0,1,2) (1,2,3) (2,3,4) (3,0,5).
// Any other site returns normally.
TEXT ·PanicSlot(SB), 0, $8-4
	NO_LOCAL_POINTERS
	LWZ	R10, site+0(FP)
	ADDI	R3, R0, $1000
	ADDI	R4, R0, $1001
	ADDI	R5, R0, $1002
	ADDI	R6, R0, $1003
	ADDI	R7, R0, $1004
	ADDI	R8, R0, $1005
	ADDI	R9, R0, $1006
	CMPWI	R10, $0
	BC	12, 2, b0
	CMPWI	R10, $1
	BC	12, 2, b1
	CMPWI	R10, $2
	BC	12, 2, b2
	CMPWI	R10, $3
	BC	12, 2, b3
	CMPWI	R10, $4
	BC	12, 2, b4
	CMPWI	R10, $5
	BC	12, 2, b5
	CMPWI	R10, $6
	BC	12, 2, b6
	CMPWI	R10, $7
	BC	12, 2, x0
	CMPWI	R10, $8
	BC	12, 2, x1
	CMPWI	R10, $9
	BC	12, 2, x2
	CMPWI	R10, $10
	BC	12, 2, x3
	RET
b0:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(0, 1)
	BL	runtime·panicBounds<ABIInternal>(SB)
b1:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(1, 2)
	BL	runtime·panicBounds<ABIInternal>(SB)
b2:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(2, 3)
	BL	runtime·panicBounds<ABIInternal>(SB)
b3:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(3, 4)
	BL	runtime·panicBounds<ABIInternal>(SB)
b4:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(4, 5)
	BL	runtime·panicBounds<ABIInternal>(SB)
b5:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(5, 6)
	BL	runtime·panicBounds<ABIInternal>(SB)
b6:
	PCDATA	$PCDATA_PanicBounds, $BOUNDS(6, 0)
	BL	runtime·panicBounds<ABIInternal>(SB)
x0:
	PCDATA	$PCDATA_PanicBounds, $EXTEND(0, 1, 2)
	BL	runtime·panicExtend<ABIInternal>(SB)
x1:
	PCDATA	$PCDATA_PanicBounds, $EXTEND(1, 2, 3)
	BL	runtime·panicExtend<ABIInternal>(SB)
x2:
	PCDATA	$PCDATA_PanicBounds, $EXTEND(2, 3, 4)
	BL	runtime·panicExtend<ABIInternal>(SB)
x3:
	PCDATA	$PCDATA_PanicBounds, $EXTEND(3, 0, 5)
	BL	runtime·panicExtend<ABIInternal>(SB)
	RET
