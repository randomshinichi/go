// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

#include "textflag.h"

// Linux enters a statically linked PPC32 program with argc and argv at the
// initial stack pointer. Build the stack arguments expected by rt0_go.
TEXT _rt0_ppc_linux(SB), NOSPLIT|NOFRAME, $0-0
	LWZ R3, 0(R1)
	ADDI R4, R1, $4
	ADDI R6, R0, $-16
	AND R1, R1, R6
	ADDI R1, R1, $-32
	STW R3, 24(R1)
	STW R4, 28(R1)
	B runtime·rt0_go(SB)
