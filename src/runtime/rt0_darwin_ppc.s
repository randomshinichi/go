// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "go_asm.h"
#include "textflag.h"

// Darwin enters the PPC32 executable with argc and argv in the measured C ABI
// argument registers. The Go runtime uses stack-passed ABI0 arguments instead.
TEXT _rt0_ppc_darwin(SB), NOSPLIT|NOFRAME, $0-0
	ADDI R0, R0, $0
	ADDI R1, R1, $-32
	// Darwin enters with argc at the initial SP and argv immediately after it.
	LWZ R3, 32(R1)
	ADDI R4, R1, $36
	STW R3, 24(R1)
	STW R4, 28(R1)
	B runtime·rt0_go(SB)
