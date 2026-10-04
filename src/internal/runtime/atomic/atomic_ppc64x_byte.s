// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (ppc64 && !ppc64.ppc970) || ppc64le

// Byte-wide reservations (LBAR/STBCCC) are Power ISA v2.06; the PPC970 does not
// implement them. The ppc970 floor gets Go bodies from atomic_ppc970.go instead.

#include "textflag.h"

// uint8 Xchg(ptr *uint8, new uint8)
// Atomically:
//	old := *ptr;
//	*ptr = new;
//	return old;
TEXT ·Xchg8(SB), NOSPLIT, $0-17
	MOVD	ptr+0(FP), R4
	MOVB	new+8(FP), R5
	LWSYNC
	LBAR	(R4), R3
	STBCCC	R5, (R4)
	BNE	-2(PC)
	ISYNC
	MOVB	R3, ret+16(FP)
	RET

// void ·Or8(byte volatile*, byte);
TEXT ·Or8(SB), NOSPLIT, $0-9
	MOVD	ptr+0(FP), R3
	MOVBZ	val+8(FP), R4
	LWSYNC
again:
	LBAR	(R3), R6
	OR	R4, R6
	STBCCC	R6, (R3)
	BNE	again
	LWSYNC
	RET

// void ·And8(byte volatile*, byte);
TEXT ·And8(SB), NOSPLIT, $0-9
	MOVD	ptr+0(FP), R3
	MOVBZ	val+8(FP), R4
	LWSYNC
again:
	LBAR	(R3), R6
	AND	R4, R6
	STBCCC	R6, (R3)
	BNE	again
	LWSYNC
	RET
