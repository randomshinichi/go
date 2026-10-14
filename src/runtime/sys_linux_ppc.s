// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

#include "go_asm.h"
#include "textflag.h"

#define SYS_exit 1
#define SYS_nanosleep 162
#define SYS_read 3
#define SYS_write 4
#define SYS_open 5
#define SYS_close 6
#define SYS_getpid 20
#define SYS_kill 37
#define SYS_brk 45
#define SYS_mmap 90
#define SYS_munmap 91
#define SYS_setitimer 104
#define SYS_clone 120
#define SYS_sched_yield 158
#define SYS_rt_sigreturn 172
#define SYS_rt_sigaction 173
#define SYS_rt_sigprocmask 174
#define SYS_sigaltstack 185
#define SYS_madvise 205
#define SYS_mincore 206
#define SYS_gettid 207
#define SYS_sched_getaffinity 223
#define SYS_exit_group 234
#define SYS_timer_create 240
#define SYS_timer_settime 241
#define SYS_timer_delete 244
#define SYS_tgkill 250
#define SYS_pipe2 317
#define SYS_clock_gettime32 246
#define SYS_clock_gettime64 403
#define SYS_timer_settime64 409
#define SYS_futex_time64 422
#define SYS_futex 221

// sigtramp is the handler the kernel calls for every signal the runtime installs.
//
// Entry state, from the kernel's ppc32 rt_sigframe setup (qemu linux-user
// linux-user/ppc/signal.c setup_rt_frame, which mirrors Linux's
// arch/powerpc/kernel/signal_32.c):
//
//	R1 = rt_sf - 80, i.e. rt_sf - (SIGNAL_FRAMESIZE 64 + RT_SIGFRAME_ADJUST 16);
//	     *(R1) = the interrupted SP; the whole of [R1, R1+80) is otherwise unused
//	     scratch, which is where this function saves LR.
//	R3 = signal number, R4 = &siginfo, R5 = &ucontext.
//	LR = the kernel's own rt_sigreturn trampoline (`li r0,172; sc`), so the
//	     handler returns by branching to LR.
//
// The kernel therefore -- not this function -- restores the interrupted machine
// context: it does so in do_rt_sigreturn, which finds the frame at R1 + 64 + 16
// again. Two things are consequently load-bearing here, and nothing else is:
// assert R1 back to its entry value before returning, and leave the signal frame
// above R1 untouched. Returning with either of those wrong resumes the wrong
// context (proved by the mutant in the report).
//
// Go's ppc32 ABI is stack-only with a 24-byte fixed frame (goarch_ppc.go
// MinFrameSize 24) whose argument words start at 24(R1), so the call to
// sigtrampgo reserves 24+12 bytes rounded to 16. LR is saved at 8(entry R1):
// that is inside the kernel's 64-byte dummy frame, and it is where ppc32 Go
// prologues save LR (see main.withLocal's prologue in a -S listing, which stores
// LR at 8(SP) before `stwu r1,-48(r1)` and reloads it from 48+8(SP)).
//
// The call is indirect, as on ppc64 and darwin/ppc: sigtrampgo is far larger
// than this frame and the linker's static nosplit check cannot see that this
// runs on an OS-provided signal stack, not a goroutine stack.
TEXT runtime·sigtramp(SB), NOSPLIT|NOFRAME|TOPFRAME, $0-0
	MFLR R0
	STW R0, 8(R1)		// LR = kernel rt_sigreturn trampoline
	STWU R1, -48(R1)
	STW R3, 24(R1)		// sigtrampgo(sig, info, ctx) at 24(R1), Go's stack-only ABI
	STW R4, 28(R1)
	STW R5, 32(R1)
	ADDI $runtime·sigtrampgo<ABIInternal>(SB), R12
	MTCTR R12
	BCTRL
	LWZ R0, 56(R1)
	MTLR R0
	ADDI R1, R1, $48
	RET

// cgo is disabled for this target (cmd/dist/build.go) and never installs a
// handler, so this is only the symbol the runtime asks for. Branch, so LR still
// holds the kernel's rt_sigreturn trampoline.
TEXT runtime·cgoSigtramp(SB), NOSPLIT|NOFRAME, $0-0
	B runtime·sigtramp(SB)

TEXT runtime·exit(SB), NOSPLIT|NOFRAME, $0-4
	LWZ R3, code+0(FP)
	ADDI R0, R0, $SYS_exit_group
	SC
	B 0(PC)

TEXT runtime·exitThread(SB), NOSPLIT|NOFRAME, $0-4
	LWZ R4, wait+0(FP)
	ADDI R5, R0, $0
	SYNC
	STW R5, (R4)
	ADDI R3, R0, $0
	ADDI R0, R0, $SYS_exit
	SC
	B 0(PC)

TEXT runtime·open(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, name+0(FP)
	LWZ R4, mode+4(FP)
	LWZ R5, perm+8(FP)
	ADDI R0, R0, $SYS_open
	SC
	BC 4, 3, open_ok
	NEG R3, R3
open_ok:
	STW R3, ret+12(FP)
	RET

TEXT runtime·closefd(SB), NOSPLIT|NOFRAME, $0-8
	LWZ R3, fd+0(FP)
	ADDI R0, R0, $SYS_close
	SC
	BC 4, 3, close_ok
	ADDI R3, R0, $-1
close_ok:
	STW R3, ret+4(FP)
	RET

TEXT runtime·write1(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, fd+0(FP)
	LWZ R4, p+4(FP)
	LWZ R5, n+8(FP)
	ADDI R0, R0, $SYS_write
	SC
	BC 4, 3, write_ok
	NEG R3, R3
write_ok:
	STW R3, ret+12(FP)
	RET

TEXT runtime·read(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, fd+0(FP)
	LWZ R4, p+4(FP)
	LWZ R5, n+8(FP)
	ADDI R0, R0, $SYS_read
	SC
	BC 4, 3, read_ok
	NEG R3, R3
read_ok:
	STW R3, ret+12(FP)
	RET

TEXT runtime·pipe2(SB), NOSPLIT|NOFRAME, $0-16
	ADDI R3, R1, $28
	LWZ R4, flags+0(FP)
	ADDI R0, R0, $SYS_pipe2
	SC
	BC 4, 3, pipe2_ok
	STW R3, errno+12(FP)
	ADDI R3, R0, $-1
	STW R3, r+4(FP)
	STW R3, w+8(FP)
	RET
pipe2_ok:
	ADDI R3, R0, $0
	STW R3, errno+12(FP)
	RET

TEXT runtime·gettid(SB), NOSPLIT|NOFRAME, $0-4
	ADDI R0, R0, $SYS_gettid
	SC
	STW R3, ret+0(FP)
	RET

TEXT runtime·getpid(SB), NOSPLIT|NOFRAME, $0-4
	ADDI R0, R0, $SYS_getpid
	SC
	STW R3, ret+0(FP)
	RET

TEXT runtime·tgkill(SB), NOSPLIT|NOFRAME, $0-12
	LWZ R3, tgid+0(FP)
	LWZ R4, tid+4(FP)
	LWZ R5, sig+8(FP)
	ADDI R0, R0, $SYS_tgkill
	SC
	RET

TEXT runtime·raise(SB), NOSPLIT|NOFRAME, $0-4
	ADDI R0, R0, $SYS_getpid
	SC
	OR R6, R3, R3
	ADDI R0, R0, $SYS_gettid
	SC
	OR R4, R3, R3
	OR R3, R6, R6
	LWZ R5, sig+0(FP)
	ADDI R0, R0, $SYS_tgkill
	SC
	RET

TEXT runtime·raiseproc(SB), NOSPLIT|NOFRAME, $0-4
	ADDI R0, R0, $SYS_getpid
	SC
	LWZ R4, sig+0(FP)
	ADDI R0, R0, $SYS_kill
	SC
	RET

TEXT runtime·setitimer(SB), NOSPLIT|NOFRAME, $0-12
	LWZ R3, mode+0(FP)
	LWZ R4, new+4(FP)
	LWZ R5, old+8(FP)
	ADDI R0, R0, $SYS_setitimer
	SC
	RET

TEXT runtime·timer_create(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, clockid+0(FP)
	LWZ R4, sevp+4(FP)
	LWZ R5, timerid+8(FP)
	ADDI R0, R0, $SYS_timer_create
	SC
	STW R3, ret+12(FP)
	RET

TEXT runtime·timer_delete(SB), NOSPLIT|NOFRAME, $0-8
	LWZ R3, timerid+0(FP)
	ADDI R0, R0, $SYS_timer_delete
	SC
	STW R3, ret+4(FP)
	RET

TEXT runtime·timer_settime32(SB), NOSPLIT|NOFRAME, $0-20
	LWZ R3, timerid+0(FP)
	LWZ R4, flags+4(FP)
	LWZ R5, new+8(FP)
	LWZ R6, old+12(FP)
	ADDI R0, R0, $SYS_timer_settime
	SC
	STW R3, ret+16(FP)
	RET

TEXT runtime·timer_settime64(SB), NOSPLIT|NOFRAME, $0-20
	LWZ R3, timerid+0(FP)
	LWZ R4, flags+4(FP)
	LWZ R5, new+8(FP)
	LWZ R6, old+12(FP)
	ADDI R0, R0, $SYS_timer_settime64
	SC
	STW R3, ret+16(FP)
	RET

TEXT runtime·futex_time32(SB), NOSPLIT|NOFRAME, $0-28
	LWZ R3, addr+0(FP)
	LWZ R4, op+4(FP)
	LWZ R5, val+8(FP)
	LWZ R6, ts+12(FP)
	LWZ R7, addr2+16(FP)
	LWZ R8, val3+20(FP)
	ADDI R0, R0, $SYS_futex
	SC
	BC 4, 3, futex32_ok
	NEG R3, R3
futex32_ok:
	STW R3, ret+24(FP)
	RET

TEXT runtime·futex_time64(SB), NOSPLIT|NOFRAME, $0-28
	LWZ R3, addr+0(FP)
	LWZ R4, op+4(FP)
	LWZ R5, val+8(FP)
	LWZ R6, ts+12(FP)
	LWZ R7, addr2+16(FP)
	LWZ R8, val3+20(FP)
	ADDI R0, R0, $SYS_futex_time64
	SC
	BC 4, 3, futex64_ok
	NEG R3, R3
futex64_ok:
	STW R3, ret+24(FP)
	RET

TEXT runtime·mmap(SB), NOSPLIT|NOFRAME, $0-32
	LWZ R3, addr+0(FP)
	LWZ R4, n+4(FP)
	LWZ R5, prot+8(FP)
	LWZ R6, flags+12(FP)
	LWZ R7, fd+16(FP)
	LWZ R8, off+20(FP)
	ADDI R0, R0, $SYS_mmap
	SC
	BC 4, 3, mmap_ok
	STW R3, err+28(FP)
	ADDI R3, R0, $0
	STW R3, p+24(FP)
	RET
mmap_ok:
	ADDI R5, R0, $0
	STW R5, err+28(FP)
	STW R3, p+24(FP)
	RET

TEXT runtime·munmap(SB), NOSPLIT|NOFRAME, $0-8
	LWZ R3, addr+0(FP)
	LWZ R4, n+4(FP)
	ADDI R0, R0, $SYS_munmap
	SC
	BC 4, 3, munmap_ok
	STW R0, 0xf0(R0)
munmap_ok:
	RET

TEXT runtime·madvise(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, addr+0(FP)
	LWZ R4, n+4(FP)
	LWZ R5, flags+8(FP)
	ADDI R0, R0, $SYS_madvise
	SC
	STW R3, ret+12(FP)
	RET

TEXT runtime·mincore(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, addr+0(FP)
	LWZ R4, n+4(FP)
	LWZ R5, dst+8(FP)
	ADDI R0, R0, $SYS_mincore
	SC
	BC 4, 3, mincore_ok
	NEG R3, R3
mincore_ok:
	STW R3, ret+12(FP)
	RET

TEXT runtime·sched_getaffinity(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, pid+0(FP)
	LWZ R4, len+4(FP)
	LWZ R5, buf+8(FP)
	ADDI R0, R0, $SYS_sched_getaffinity
	SC
	BC 4, 3, affinity_ok
	NEG R3, R3
affinity_ok:
	STW R3, ret+12(FP)
	RET

TEXT runtime·sigaltstack(SB), NOSPLIT|NOFRAME, $0-8
	LWZ R3, new+0(FP)
	LWZ R4, old+4(FP)
	ADDI R0, R0, $SYS_sigaltstack
	SC
	BC 4, 3, sigaltstack_ok
	STW R0, 0xf0(R0)
sigaltstack_ok:
	RET

TEXT runtime·rtsigprocmask(SB), NOSPLIT|NOFRAME, $0-16
	LWZ R3, how+0(FP)
	LWZ R4, new+4(FP)
	LWZ R5, old+8(FP)
	LWZ R6, size+12(FP)
	ADDI R0, R0, $SYS_rt_sigprocmask
	SC
	BC 4, 3, sigprocmask_ok
	STW R0, 0xf0(R0)
sigprocmask_ok:
	RET

TEXT runtime·rt_sigaction(SB), NOSPLIT|NOFRAME, $0-20
	LWZ R3, sig+0(FP)
	LWZ R4, new+4(FP)
	LWZ R5, old+8(FP)
	LWZ R6, size+12(FP)
	ADDI R0, R0, $SYS_rt_sigaction
	SC
	BC 4, 3, sigaction_ok
	NEG R3, R3
sigaction_ok:
	STW R3, ret+16(FP)
	RET

TEXT runtime·osyield(SB), NOSPLIT|NOFRAME, $0-0
	ADDI R0, R0, $SYS_sched_yield
	SC
	RET

TEXT runtime·usleep(SB), NOSPLIT, $8-4
	LWZ R3, usec+0(FP)
	ADDI R5, R0, $1000000
	DIVWU R3, R5, R4
	MULLW R4, R5, R6
	SUBF R6, R3, R7
	ADDI R8, R0, $1000
	MULLW R7, R8, R8
	STW R4, 24(R1)
	STW R8, 28(R1)
	ADDI R3, R1, $24
	ADDI R4, R0, $0
	ADDI R0, R0, $SYS_nanosleep
	SC
	RET

TEXT runtime·sbrk0(SB), NOSPLIT|NOFRAME, $0-4
	ADDI R3, R0, $0
	ADDI R0, R0, $SYS_brk
	SC
	STW R3, ret+0(FP)
	RET

TEXT runtime·clone(SB), NOSPLIT|NOFRAME, $0-24
	LWZ R3, flags+0(FP)
	LWZ R4, stk+4(FP)
	LWZ R7, mp+8(FP)
	LWZ R8, gp+12(FP)
	LWZ R12, fn+16(FP)
	STW R7, -4(R4)
	STW R8, -8(R4)
	STW R12, -12(R4)
	ADDI R7, R0, $1234
	STW R7, -16(R4)
	ADDI R0, R0, $SYS_clone
	SC
	BC 4, 3, clone_ok
	NEG R3, R3
clone_ok:
	ADDI R0, R0, $0
	CMPW R3, R0
	BC 12, 2, clone_child
	STW R3, ret+20(FP)
	RET
clone_child:
	LWZ R7, -16(R1)
	ADDI R0, R0, $1234
	CMPW R7, R0
	BC 12, 2, clone_marker_ok
	STW R0, 0(R0)
clone_marker_ok:
	ADDI R0, R0, $SYS_gettid
	SC
	ADDI R0, R0, $0
	LWZ R7, -4(R1)
	LWZ R8, -8(R1)
	LWZ R12, -12(R1)
	CMPW R7, R0
	BC 12, 2, clone_nog
	CMPW R8, R0
	BC 12, 2, clone_nog
	STW R3, m_procid(R7)
	STW R7, g_m(R8)
	OR R30, R8, R8
clone_nog:
	MTCTR R12
	BCTRL
	ADDI R3, R0, $111
	ADDI R0, R0, $SYS_exit
	SC
	B 0(PC)

