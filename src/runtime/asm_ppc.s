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
TEXT runtime·goexit(SB), NOSPLIT|NOFRAME|TOPFRAME, $0-0
	BL	runtime·goexit1(SB)
	UNDEF

// getg returns the current goroutine from the reserved Go register R30.
TEXT runtime·getg(SB), NOSPLIT|NOFRAME, $0-4
	MOVW	R30, ret+0(FP)
	RET

// setg changes the current goroutine in the reserved Go register R30.
TEXT runtime·setg(SB), NOSPLIT|NOFRAME, $0-4
	MOVW	gg+0(FP), R30
	RET

// mstart_stub is called with the C ABI on a newly created pthread. The
// pthread_create argument is m; its g0 is already initialized by newm.
TEXT runtime·mstart_stub(SB), NOSPLIT|NOFRAME, $0-0
	MOVW	m_g0(R3), R30
	BL	runtime·mstart(SB)
	MOVW	$0, R3
	RET
