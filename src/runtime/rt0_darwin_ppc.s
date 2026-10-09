// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "go_asm.h"
#include "textflag.h"

// Darwin process entry is not a C function call: argc and argv are on the
// initial stack. The measured C ABI's R3-R10 argument registers apply to calls,
// not to the LC_UNIXTHREAD state that transfers control to this entry.
TEXT _rt0_ppc_darwin(SB), NOSPLIT|NOFRAME, $0-0
	ADDI R0, R0, $0
	ADDI R1, R1, $-32
	// Preserve the kernel's initial stack block: argc is at entry SP+0, and
	// argv starts at entry SP+8. The 32-byte adjustment above creates the
	// measured Go linkage area without moving that block.
	LWZ R3, 32(R1)
	ADDI R4, R1, $36
	STW R3, 24(R1)
	STW R4, 28(R1)
	B runtime·rt0_go(SB)
