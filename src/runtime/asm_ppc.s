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

// mcall saves the current goroutine and runs fn(g) on g0. fn must not return.
TEXT runtime·mcall(SB), NOSPLIT|NOFRAME, $0-4
	OR R5, R30, R30
	STW R1, (g_sched+gobuf_sp)(R5)
	MFLR R6
	STW R6, (g_sched+gobuf_pc)(R5)
	STW R0, (g_sched+gobuf_lr)(R5)
	LWZ R7, g_m(R5)
	LWZ R8, m_g0(R7)
	CMPW R8, R5
	BC 4, 2, mcall_switch
	BL runtime·badmcall(SB)
	BL runtime·abort(SB)
mcall_switch:
	OR R30, R8, R8
	LWZ R1, (g_sched+gobuf_sp)(R30)
	ADDI R1, R1, $-32
	STW R5, 24(R1)
	LWZ R11, fn+0(FP)
	LWZ R12, 0(R11)
	MTCTR R12
	BCTRL
	BL runtime·badmcall2(SB)
	BL runtime·abort(SB)

// systemstack_switch is the synthetic frame used when a goroutine stack is
// scanned while systemstack is running on g0.
TEXT runtime·systemstack_switch(SB), NOSPLIT, $0-0
	BL runtime·abort(SB) // Non-leaf, with the same linkage frame as systemstack.
	RET

TEXT runtime·gosave_systemstack_switch(SB), NOSPLIT|NOFRAME, $0-0
	ADDI $runtime·systemstack_switch(SB), R5
	ADDI R5, R5, $16 // Past the non-leaf prologue, as on PPC64.
	STW R5, (g_sched+gobuf_pc)(R30)
	STW R1, (g_sched+gobuf_sp)(R30)
	STW R0, (g_sched+gobuf_lr)(R30)
	RET

// systemstack runs fn on the system stack, restoring the caller goroutine after.
TEXT runtime·systemstack(SB), NOSPLIT, $0-4
	OR R18, R1, R1
	LWZ R14, fn+0(FP)
	OR R15, R30, R30
	LWZ R16, g_m(R30)
	LWZ R17, m_gsignal(R16)
	CMPW R17, R30
	BC 4, 2, systemstack_checkg0
	B systemstack_noswitch
systemstack_checkg0:
	LWZ R17, m_g0(R16)
	CMPW R17, R30
	BC 4, 2, systemstack_checkcurg
	B systemstack_noswitch
systemstack_checkcurg:
	LWZ R17, m_curg(R16)
	CMPW R17, R30
	BC 4, 2, systemstack_bad
	BL runtime·gosave_systemstack_switch(SB)
	LWZ R30, m_g0(R16)
	LWZ R1, (g_sched+gobuf_sp)(R30)
systemstack_noswitch:
	ADDI R1, R1, $-32
	STW R15, 24(R1)
	STW R18, 28(R1)
	OR R11, R14, R14
	LWZ R12, 0(R14)
	MTCTR R12
	BCTRL
	LWZ R15, 24(R1)
	CMPW R15, R30
	BC 12, 2, systemstack_restore_noswitch
	OR R30, R15, R15
	// A stack copy may have moved the suspended goroutine's stack.
	LWZ R1, (g_sched+gobuf_sp)(R30)
	STW R0, (g_sched+gobuf_sp)(R30)
	RET
systemstack_restore_noswitch:
	LWZ R1, 28(R1)
	RET
systemstack_bad:
	BL runtime·badsystemstack(SB)
	BL runtime·abort(SB)

// asmcgocall_no_g invokes a C function on an aligned system stack without g.
TEXT runtime·asmcgocall_no_g(SB), NOSPLIT, $0-8
	LWZ R14, fn+0(FP)
	LWZ R15, arg+4(FP)
	OR R16, R1, R1
	ADDI R1, R1, $-32
	STW R16, 24(R1)
	OR R3, R15, R15
	MTCTR R14
	BCTRL
	LWZ R1, 24(R1)
	RET

// asmcgocall marshals one C argument and switches from a user goroutine to g0.
TEXT runtime·asmcgocall(SB), NOSPLIT, $0-12
	LWZ R14, fn+0(FP)
	LWZ R15, arg+4(FP)
	OR R16, R1, R1
	OR R17, R30, R30
	ADDI R20, R0, $0
	CMPW R0, R30
	BC 12, 2, asmcgocall_stack
	B asmcgocall_call
asmcgocall_stack:
	LWZ R18, g_m(R30)
	LWZ R19, m_gsignal(R18)
	CMPW R19, R30
	BC 12, 2, asmcgocall_checkg0
	B asmcgocall_call
asmcgocall_checkg0:
	LWZ R19, m_g0(R18)
	CMPW R19, R30
	BC 12, 2, asmcgocall_switch
	B asmcgocall_call
asmcgocall_switch:
	BL runtime·gosave_systemstack_switch(SB)
	LWZ R30, m_g0(R18)
	LWZ R1, (g_sched+gobuf_sp)(R30)
	ADDI R20, R0, $1
asmcgocall_call:
	ADDI R1, R1, $-32
	STW R17, 24(R1)
	CMPWI R20, $0
	BC 12, 2, asmcgocall_save_sp
	LWZ R18, g_stack+stack_hi(R17)
	SUBF R19, R16, R18
	STW R19, 28(R1)
	B asmcgocall_call_c
asmcgocall_save_sp:
	STW R16, 28(R1)
asmcgocall_call_c:
	OR R3, R15, R15
	MTCTR R14
	BCTRL
	OR R14, R3, R3
	CMPWI R20, $0
	BC 12, 2, asmcgocall_restore_direct
	LWZ R17, 24(R1)
	LWZ R18, g_stack+stack_hi(R17)
	LWZ R19, 28(R1)
	SUBF R1, R19, R18
	OR R30, R17, R17
	B asmcgocall_result
asmcgocall_restore_direct:
	LWZ R17, 24(R1)
	OR R30, R17, R17
	LWZ R1, 28(R1)
asmcgocall_result:
	OR R3, R14, R14
	STW R3, ret+8(FP)
	RET

// PPC32 has no pause instruction; use a bounded decrement loop.
TEXT runtime·procyieldAsm(SB), NOSPLIT|NOFRAME, $0-4
	LWZ R5, cycles+0(FP)
procyield_loop:
	CMPWI R5, $0
	BC 12, 2, procyield_done
	ADDI R5, R5, $-1
	B procyield_loop
procyield_done:
	RET

// PowerPC's sync orders preceding stores before subsequent stores.
TEXT runtime·publicationBarrier(SB), NOSPLIT|NOFRAME, $0-0
	SYNC
	RET

TEXT runtime·rt0_go(SB), NOSPLIT|NOFRAME|TOPFRAME, $0-0
	// Save argv and argc before creating the initial g0 stack frame.
	LWZ R14, 24(R1)
	LWZ R15, 28(R1)
	ADDI R1, R1, $-128
	STW R14, 120(R1)
	STW R15, 124(R1)

	// g lives in R30; initialize g0's stack before any stack-checked calls.
	ADDI $runtime·g0(SB), R30
	ADDIS R6, R1, $-1
	STW R6, (g_stack+stack_lo)(R30)
	ADDI R7, R6, $const_stackGuard
	STW R7, g_stackguard0(R30)
	STW R7, g_stackguard1(R30)
	ADDI R7, R1, $128
	STW R7, (g_stack+stack_hi)(R30)

	// Publish the initial machine and goroutine.
	ADDI $runtime·m0(SB), R3
	STW R30, m_g0(R3)
	STW R3, g_m(R30)

	BL runtime·check(SB)
	LWZ R3, 120(R1)
	STW R3, 24(R1)
	LWZ R3, 124(R1)
	STW R3, 28(R1)
	BL runtime·args(SB)
	BL runtime·osinit(SB)
	BL runtime·schedinit(SB)

	// Start runtime.main in its initial goroutine.
	ADDI $runtime·mainPC(SB), R3
	STW R3, 24(R1)
	BL runtime·newproc(SB)
	ADDI R1, R1, $128
	BL runtime·mstart(SB)
	BL runtime·abort(SB)

TEXT runtime·asminit(SB), NOSPLIT|NOFRAME, $0-0
	RET

// breakpoint raises SIGTRAP through the Darwin libc wrapper.
TEXT runtime·breakpoint(SB), NOSPLIT, $0-0
	ADDI R3, R0, $5
	STW R3, 24(R1)
	BL runtime·raise(SB)
	RET

TEXT ·checkASM(SB), NOSPLIT|NOFRAME, $0-1
	ADDI R3, R0, $1
	STB R3, ret+0(FP)
	RET

TEXT runtime·abort(SB), NOSPLIT|NOFRAME, $0-0
	LWZ R1, 0(R0) // deliberate fault; abort must never return

// switchToCrashStack0 runs fn on gcrash's dedicated stack and never returns.
TEXT runtime·switchToCrashStack0(SB), NOSPLIT, $0-4
	LWZ R14, fn+0(FP)
	LWZ R15, g_m(R30)
	ADDI $runtime·gcrash(SB), R16
	STW R15, g_m(R16)
	STW R16, m_g0(R15)
	OR R30, R16, R16
	LWZ R1, g_stack+stack_hi(R30)
	ADDI R1, R1, $-32
	OR R11, R14, R14
	LWZ R12, 0(R14)
	MTCTR R12
	BCTRL
	BL runtime·abort(SB)

// sigfwd calls a signal handler with the Darwin C ABI argument registers.
TEXT runtime·sigfwd(SB), NOSPLIT, $0-16
	LWZ R12, fn+0(FP)
	LWZ R3, sig+4(FP)
	LWZ R4, info+8(FP)
	LWZ R5, ctx+12(FP)
	MTCTR R12
	BCTRL
	RET

DATA runtime·mainPC+0(SB)/4, $runtime·main(SB)
GLOBL runtime·mainPC(SB), RODATA, $4
