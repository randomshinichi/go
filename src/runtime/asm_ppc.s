// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

// mstart is the entry point for a Go machine's scheduler loop.
TEXT runtime·mstart(SB), NOSPLIT|TOPFRAME, $0-0
	BL	runtime·mstart0(SB)
	RET // not reached

// goexit marks the end of a goroutine stack and never returns.
//
// goexit must not allocate a frame. newproc1 records gp.stktopsp as the SP at the
// fabricated call to goexit, and finishInternal (traceback.go) requires the
// unwound outermost frame's SP to equal it, so SP movement here fails every GC
// stack walk of a goroutine that has already returned. The instruction at
// goexit+0 is what makes the fabricated entry PC, goexit+PCQuantum (newproc1
// and gostartcall), the goexit1 call, as on every other link-register
// architecture.
TEXT runtime·goexit(SB), NOSPLIT|NOFRAME|TOPFRAME, $0-0
	ADDI	R0, R0, $0		// Writes 0 to R0; never executed. Keep goexit+PCQuantum on the call below.
	BL	runtime·goexit1(SB)	// does not return
	BL	runtime·abort(SB)	// not reached

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

// mstart_stub is called with the Darwin PPC32 C ABI on a new pthread. Go's
// internal ABI may clobber the C nonvolatile GPRs/FPRs and CR across mstart.
// Save R13..R31 (19*4), CR (4), and F14..F31 (18*8): 224 local bytes. The
// PPC32 fixed linkage area is 24 bytes, so (224+24) rounded to 16-byte stack
// alignment reserves 256 bytes. The explicit NOSPLIT frame makes that whole
// reservation visible to the ppc32 linker's nosplit stack-bound check; this
// stub does not take a morestack_noctxt path. The PPC32 backend does not emit
// VMX operations, so vector nonvolatiles remain untouched.
// Use a manual return: the generated PPC32 RET epilogue uses R31 to restore LR,
// which would overwrite the C caller's R31 after it had been restored.
TEXT runtime·mstart_stub(SB), NOSPLIT, $224-0
	STW	R13, 32(R1)
	STW	R14, 36(R1)
	STW	R15, 40(R1)
	STW	R16, 44(R1)
	STW	R17, 48(R1)
	STW	R18, 52(R1)
	STW	R19, 56(R1)
	STW	R20, 60(R1)
	STW	R21, 64(R1)
	STW	R22, 68(R1)
	STW	R23, 72(R1)
	STW	R24, 76(R1)
	STW	R25, 80(R1)
	STW	R26, 84(R1)
	STW	R27, 88(R1)
	STW	R28, 92(R1)
	STW	R29, 96(R1)
	STW	R30, 100(R1)
	STW	R31, 104(R1)
	MFCR	R0
	STW	R0, 108(R1)
	STFD	F14, 112(R1)
	STFD	F15, 120(R1)
	STFD	F16, 128(R1)
	STFD	F17, 136(R1)
	STFD	F18, 144(R1)
	STFD	F19, 152(R1)
	STFD	F20, 160(R1)
	STFD	F21, 168(R1)
	STFD	F22, 176(R1)
	STFD	F23, 184(R1)
	STFD	F24, 192(R1)
	STFD	F25, 200(R1)
	STFD	F26, 208(R1)
	STFD	F27, 216(R1)
	STFD	F28, 224(R1)
	STFD	F29, 232(R1)
	STFD	F30, 240(R1)
	STFD	F31, 248(R1)
	LWZ	R30, m_g0(R3)
	BL	runtime·mstart(SB)
	ADDI	R3, R0, $0
	LFD	112(R1), F14
	LFD	120(R1), F15
	LFD	128(R1), F16
	LFD	136(R1), F17
	LFD	144(R1), F18
	LFD	152(R1), F19
	LFD	160(R1), F20
	LFD	168(R1), F21
	LFD	176(R1), F22
	LFD	184(R1), F23
	LFD	192(R1), F24
	LFD	200(R1), F25
	LFD	208(R1), F26
	LFD	216(R1), F27
	LFD	224(R1), F28
	LFD	232(R1), F29
	LFD	240(R1), F30
	LFD	248(R1), F31
	LWZ	R13, 32(R1)
	LWZ	R14, 36(R1)
	LWZ	R15, 40(R1)
	LWZ	R16, 44(R1)
	LWZ	R17, 48(R1)
	LWZ	R18, 52(R1)
	LWZ	R19, 56(R1)
	LWZ	R20, 60(R1)
	LWZ	R21, 64(R1)
	LWZ	R22, 68(R1)
	LWZ	R23, 72(R1)
	LWZ	R24, 76(R1)
	LWZ	R25, 80(R1)
	LWZ	R26, 84(R1)
	LWZ	R27, 88(R1)
	LWZ	R28, 92(R1)
	LWZ	R29, 96(R1)
	LWZ	R30, 100(R1)
	LWZ	R31, 104(R1)
	LWZ	R0, 108(R1)
	MTCRF	$0xff, R0
	LWZ	R0, 264(R1)
	ADDI	R1, R1, $256
	MTLR	R0
	JMP	(LR)

// gogo restores a saved goroutine context. PPC32 has stack-only public Go
// arguments and keeps g in R30; all values needed after changing SP are first
// loaded into reserved scratch registers.
TEXT runtime·gogo(SB), NOSPLIT|NOFRAME, $0-4
	LWZ R5, buf+0(FP)
	LWZ R4, gobuf_sp(R5)
	LWZ R31, gobuf_lr(R5)
	LWZ R12, gobuf_pc(R5)
	LWZ R30, gobuf_g(R5)
	LWZ R11, gobuf_ctxt(R5)
	// R0 is not a zero register on PPC32. R6 is unused throughout gogo,
	// so materialize zero there before clearing the saved context.
	ADDI R6, R0, $0
	STW R6, gobuf_sp(R5)
	STW R6, gobuf_lr(R5)
	STW R6, gobuf_ctxt(R5)
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
	STW R11, (g_sched+gobuf_ctxt)(R30)
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
	ADDI R11, R0, $0
	B runtime·morestack(SB)

// memmove handles overlapping ranges with aligned word copies where possible.
// Byte paths preserve support for unaligned pointers and exact range boundaries.
TEXT runtime·memmove(SB), NOSPLIT|NOFRAME, $0-12
	LWZ R3, to+0(FP)
	LWZ R4, from+4(FP)
	LWZ R5, n+8(FP)
	ADDI R8, R0, $3
	CMPWI R5, $0
	BC 12, 2, memdone
	CMPLW R3, R4
	BC 12, 0, memforward
	ADD R7, R4, R5
	CMPLW R3, R7
	BC 12, 0, membackward
memforward:
	XOR R6, R3, R4
	AND R6, R6, R8
	CMPWI R6, $0
	BC 4, 2, memforward_bytes
	AND R6, R3, R8
	CMPWI R6, $0
	BC 12, 2, memforward_words
memforward_head:
	LBZ R7, 0(R4)
	STB R7, 0(R3)
	ADDI R3, R3, $1
	ADDI R4, R4, $1
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 12, 2, memdone
	AND R6, R3, R8
	CMPWI R6, $0
	BC 4, 2, memforward_head
memforward_words:
	CMPWI R5, $4
	BC 12, 0, memforward_tail
	LWZ R6, 0(R4)
	STW R6, 0(R3)
	ADDI R3, R3, $4
	ADDI R4, R4, $4
	ADDI R5, R5, $-4
	B memforward_words
memforward_tail:
	CMPWI R5, $0
	BC 12, 2, memdone
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R3, R3, $1
	ADDI R4, R4, $1
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, memforward_tail
	B memdone
memforward_bytes:
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R3, R3, $1
	ADDI R4, R4, $1
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, memforward_bytes
	RET
membackward:
	ADD R3, R3, R5
	ADD R4, R4, R5
	XOR R6, R3, R4
	AND R6, R6, R8
	CMPWI R6, $0
	BC 4, 2, memback_bytes
	AND R6, R3, R8
	CMPWI R6, $0
	BC 12, 2, memback_words
memback_tail:
	ADDI R3, R3, $-1
	ADDI R4, R4, $-1
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 12, 2, memdone
	AND R6, R3, R8
	CMPWI R6, $0
	BC 4, 2, memback_tail
memback_words:
	CMPWI R5, $4
	BC 12, 0, memback_head
	ADDI R3, R3, $-4
	ADDI R4, R4, $-4
	LWZ R6, 0(R4)
	STW R6, 0(R3)
	ADDI R5, R5, $-4
	B memback_words
memback_head:
	CMPWI R5, $0
	BC 12, 2, memdone
memback_head_loop:
	ADDI R3, R3, $-1
	ADDI R4, R4, $-1
	LBZ R6, 0(R4)
	STB R6, 0(R3)
	ADDI R5, R5, $-1
	CMPWI R5, $0
	BC 4, 2, memback_head_loop
	B memdone
memback_bytes:
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
// arguments through the measured stack-only ABI. Their frames exceed the
// linkage area, so a stack scan that stops in the fallback needs a locals map;
// the outgoing arguments are covered by the fallback's own argument map.
TEXT runtime·memhash32(SB), NOSPLIT, $12-12
	NO_LOCAL_POINTERS
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	BL runtime·memhash32Fallback(SB)
	STW R3, ret+8(FP)
	RET

TEXT runtime·memhash64(SB), NOSPLIT, $12-12
	NO_LOCAL_POINTERS
	LWZ R3, p+0(FP)
	STW R3, 24(R1)
	LWZ R4, h+4(FP)
	STW R4, 28(R1)
	BL runtime·memhash64Fallback(SB)
	STW R3, ret+8(FP)
	RET

TEXT runtime·memhash(SB), NOSPLIT, $12-16
	NO_LOCAL_POINTERS
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
	LWZ R11, fn+0(FP) // Read on the caller's stack, before switching to g0.
	OR R5, R30, R30
	STW R1, (g_sched+gobuf_sp)(R5)
	MFLR R6
	STW R6, (g_sched+gobuf_pc)(R5)
	ADDI R0, R0, $0
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
	ADDI R0, R0, $0
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
	B systemstack_tailcall
systemstack_checkg0:
	LWZ R17, m_g0(R16)
	CMPW R17, R30
	BC 4, 2, systemstack_checkcurg
	B systemstack_tailcall
systemstack_checkcurg:
	LWZ R17, m_curg(R16)
	CMPW R17, R30
	BC 4, 2, systemstack_bad
	BL runtime·gosave_systemstack_switch(SB)
	LWZ R30, m_g0(R16)
	LWZ R1, (g_sched+gobuf_sp)(R30)
	// This path must return here to restore the caller goroutine.
	B systemstack_noswitch
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
	ADDI R0, R0, $0
	STW R0, (g_sched+gobuf_sp)(R30)
	RET
systemstack_restore_noswitch:
	LWZ R1, 28(R1)
	RET
systemstack_tailcall:
	// On an already-system-stack call no stack switch or call has occurred,
	// so the ABI prologue frame is still at SP. RET restores LR and pops that
	// frame with unwind metadata before branching without link. A handwritten
	// pop would leave stale PC/SP metadata here because this function also
	// writes SP dynamically. R11 carries the function-value context.
	OR R11, R14, R14
	LWZ R12, 0(R14)
	MTCTR R12
	RET (CTR)
systemstack_bad:
	// Hide the call from the linker's nosplit analysis, as the other
	// architectures do (asm_ppc64x.s: "Hide call from linker nosplit
	// analysis"). badsystemstack is nosplit and belongs to the fatal
	// systemstack path, so its subtree is a real nosplit chain: 256 bytes
	// here (writeErrData, write, write1, libcCall, asmcgocall). On ppc64
	// and amd64 that subtree is larger still (304 and 224 bytes) and is
	// hidden the same way; only on ppc32 was the call direct, and it is
	// what pushes chansend1 over the 800 byte limit once the PGO inliner
	// puts chansend's body inside nosplit chansend1.
	ADDI $runtime·badsystemstack(SB), R12
	MTCTR R12
	BCTRL
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
	ADDI R0, R0, $0 // C may clobber R0; restore Go's zero-register convention.
	LWZ R1, 24(R1)
	RET

// asmcgocall marshals one C argument and switches from a user goroutine to g0.
TEXT runtime·asmcgocall(SB), NOSPLIT, $0-12
	LWZ R14, fn+0(FP)
	LWZ R15, arg+4(FP)
	OR R16, R1, R1
	OR R17, R30, R30
	ADDI R0, R0, $0 // Establish zero before comparing R0 with g below.
	ADDI R20, R0, $0
	CMPW R0, R30
	BC 4, 2, asmcgocall_stack
	B asmcgocall_call
asmcgocall_stack:
	LWZ R18, g_m(R30)
	LWZ R19, m_gsignal(R18)
	CMPW R19, R30
	BC 4, 2, asmcgocall_checkg0
	B asmcgocall_call
asmcgocall_checkg0:
	LWZ R19, m_g0(R18)
	CMPW R19, R30
	BC 4, 2, asmcgocall_switch
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
	ADDI R0, R0, $0 // C may clobber R0; restore Go's zero-register convention.
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
//
// The callee is C, so it may home its register arguments for the whole call:
// ARG(n) = 24 + 4*(n-1) counts from the SP of the frame that made the call
// (reference/xnu/EXTERNAL_HEADERS/architecture/ppc/asm_help.h:116-125), so a
// homing callee writes SP+24..SP+55 of THIS frame, and BUILD_FRAME floors that
// parameter area at 32 bytes (:161-172). Declare $32 of locals: obj/ppc/link.go
// then allocates 64 bytes, 24 of linkage plus the full 32-byte parameter area,
// and the handler's home area lies inside the frame instead of running over the
// frame's top. A $0 declaration allocates 32 bytes, leaving only SP+24..SP+31
// inside the frame: the handler's third word then lands on this frame's caller's
// backchain slot, and its fifth (r7's home slot, SP+40) on the saved LR this
// function's own prologue wrote at entry_SP+8, so the RET below returns to the
// handler's r7 argument. Every C-calling trampoline in sys_darwin_ppc.s has the
// same shape for the same reason (goppc-xprg.22); this one lives here and was
// missed by that census (goppc-3zqt).
TEXT runtime·sigfwd(SB), NOSPLIT, $32-16
	LWZ R12, fn+0(FP)
	LWZ R3, sig+4(FP)
	LWZ R4, info+8(FP)
	LWZ R5, ctx+12(FP)
	MTCTR R12
	BCTRL
	RET

DATA runtime·mainPC+0(SB)/4, $runtime·main(SB)
GLOBL runtime·mainPC(SB), RODATA, $4
