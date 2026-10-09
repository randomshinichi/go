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

// stackcheck is a private pre-prologue helper. The caller passes its fully
// allocated frame size in R3; it returns nonzero when SP-frame crosses g's
// stack guard. It is a leaf so it can run before the caller has a frame.
TEXT runtime·stackcheck(SB), NOSPLIT|NOFRAME, $0-0
	CMPLW R1, R3
	BC 12, 0, grow // SP-frame would wrap around the 32-bit address space.
	LWZ R4, g_stackguard0(R30)
	SUBF R3, R3, R1
	CMPLW R3, R4
	BC 12, 0, grow
	ADDI R3, R0, $0
	RET
grow:
	ADDI R3, R0, $1
	RET

// gogo restores a saved goroutine context. PPC32 has stack-only public Go
// arguments and keeps g in R30; all values needed after changing SP are first
// loaded into reserved scratch registers.
TEXT runtime·gogo(SB), NOSPLIT|NOFRAME, $0-4
	LWZ R5, buf+0(FP)
	LWZ R4, gobuf_sp(R5)
	LWZ R31, gobuf_lr(R5)
	LWZ R12, gobuf_pc(R5)
	LWZ R30, gobuf_g(R5)
	STW R0, gobuf_sp(R5)
	STW R0, gobuf_lr(R5)
	STW R0, gobuf_ctxt(R5)
	OR R1, R4, R4
	MTLR R31
	MTCTR R12
	JMP (CTR)

// morestack records the interrupted goroutine, switches to g0, and enters
// the generic stack copier. R5 is the interrupted function's saved LR.
TEXT runtime·morestack(SB), NOSPLIT|NOFRAME, $0-0
	MFLR R8
	STW R1, (g_sched+gobuf_sp)(R30)
	STW R8, (g_sched+gobuf_pc)(R30)
	STW R5, (g_sched+gobuf_lr)(R30)
	STW R0, (g_sched+gobuf_ctxt)(R30)
	LWZ R7, g_m(R30)
	LWZ R8, m_g0(R7)
	CMPW R30, R8
	BC 4, 2, checksignal
	BL runtime·badmorestackg0(SB)
	BL runtime·abort(SB)
checksignal:
	LWZ R8, m_gsignal(R7)
	CMPW R30, R8
	BC 4, 2, switchstack
	BL runtime·badmorestackgsignal(SB)
	BL runtime·abort(SB)
switchstack:
	STW R5, (m_morebuf+gobuf_pc)(R7)
	STW R1, (m_morebuf+gobuf_sp)(R7)
	STW R30, (m_morebuf+gobuf_g)(R7)
	LWZ R30, m_g0(R7)
	LWZ R1, (g_sched+gobuf_sp)(R30)
	// Reserve the 24-byte ABI0 linkage frame plus 8 bytes of padding so SP
	// remains 16-byte aligned at the call to newstack.
	STWU R1, -32(R1)
	BL runtime·newstack(SB)
	BL runtime·abort(SB)

TEXT runtime·morestack_noctxt(SB), NOSPLIT|NOFRAME, $0-0
	B runtime·morestack(SB)

// memmove handles overlapping ranges with byte copies. The target's baseline
// has no PPC64 helper, and the generic runtime requires this symbol at link time.
TEXT runtime·memmove(SB), NOSPLIT|NOFRAME, $0-12
	LWZ R3, to+0(FP)
	LWZ R4, from+4(FP)
	LWZ R5, n+8(FP)
	CMPWI R5, $0
	BC 12, 2, memdone
	CMPLW R3, R4
	BC 12, 0, memforward
	ADD R7, R4, R5
	CMPLW R3, R7
	BC 12, 0, membackward
memforward:
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R3, R3, $1
	ADDI R4, R4, $1
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, memforward
	RET
membackward:
	ADD R3, R3, R5
	ADD R4, R4, R5
membackloop:
	ADDI R3, R3, $-1
	ADDI R4, R4, $-1
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, membackloop
memdone:
	RET

// The generic 32-bit hash fallbacks are Go ABI0 functions. These small stubs
// provide the assembly names expected by compiler builtins and forward their
// arguments through the measured stack-only ABI.
TEXT runtime·memhash32(SB), NOSPLIT, $12-12
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	BL runtime·memhash32Fallback(SB)
	STW R3, ret+8(FP)
	RET

TEXT runtime·memhash64(SB), NOSPLIT, $12-12
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	BL runtime·memhash64Fallback(SB)
	STW R3, ret+8(FP)
	RET

TEXT runtime·memhash(SB), NOSPLIT, $12-16
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	LWZ R5, s+8(FP)
	STW R5, 32(R1)
	BL runtime·memhashFallback(SB)
	STW R3, ret+12(FP)
	RET

TEXT runtime·strhash(SB), NOSPLIT, $8-12
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	BL runtime·strhashFallback(SB)
	STW R3, ret+8(FP)
	RET
