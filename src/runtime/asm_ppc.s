// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "go_asm.h"
#include "textflag.h"

// mstart is the entry point for a Go machine's scheduler loop.
TEXT runtime·mstart(SB), NOSPLIT|TOPFRAME, $0-0
	BL	runtime·mstart0(SB)
	RET // not reached

// goexit marks the end of a goroutine stack and never returns.
TEXT runtime·goexit(SB), NOSPLIT|TOPFRAME, $0-0
	BL	runtime·goexit1(SB)
	BL	runtime·abort(SB)
	RET

// getg returns the current goroutine from the reserved Go register R30.
// R13 and R2 are not thread pointers on the measured Darwin ppc32 ABI; core
// Go runtime access to g needs no separate TLS lookup.
TEXT runtime·getg(SB), NOSPLIT, $0-4
	STW	R30, ret+0(FP)
	RET

// setg changes the current goroutine in the reserved Go register R30.
TEXT runtime·setg(SB), NOSPLIT, $0-4
	LWZ	R30, gg+0(FP)
	RET

// mstart_stub is called with the C ABI on a newly created pthread. The
// pthread_create argument is m; its g0 is already initialized by newm.
TEXT runtime·mstart_stub(SB), NOSPLIT, $0-0
	LWZ	R30, m_g0(R3)
	BL	runtime·mstart(SB)
	ADDI	R3, R0, $0
	RET
