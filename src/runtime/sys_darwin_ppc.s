// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "go_asm.h"
#include "textflag.h"

// Darwin ppc C-call frame contract (reference/xnu/EXTERNAL_HEADERS/architecture/
// ppc/asm_help.h: ARG(n) = ((n-1)*4)+24 and BUILD_FRAME; reg_help.h for the GPR
// assignment). A C call needs from its caller's frame:
//
//	SP+0..23   linkage: backchain, saved CR, saved LR, reserved
//	SP+24..55  home area, eight words. The C callee stores its register
//	           arguments r3-r10 here (a variadic function, or any function
//	           whose parameters are address-taken or unoptimised, does), so
//	           these bytes belong to the callee for the whole call
//	SP+56..    argument words nine and up
//
// A Go frame is 24 linkage bytes plus its locals rounded up to 16, and the
// prologue saves LR in the CALLER's linkage area (entry SP+8), which with a
// 32-byte frame is SP+40 here: exactly home word five. That is why every
// trampoline that calls C declares at least $32 of locals, which gives a
// 64-byte frame with the full 32-byte home area above the linkage, and keeps
// its own scratch (the C call's output temporaries) at SP+56, above the home
// area, never inside it.
//
// libc_error_trampoline stores the address of the thread's libc errno slot.
TEXT runtime·libc_error_trampoline(SB), NOSPLIT, $32-0
	LWZ	R14, 0(R3)
	BL	libc_error(SB)
	STW	R3, 0(R14)
	RET

// syscallN_trampoline invokes a C function with uintptr arguments in a
// libcCallInfo. Darwin ppc passes eight words in R3-R10; the ninth and later
// words are at SP+56, above the eight-word home area. All stack argument words
// are copied in host (big-endian) word order.
//
// R20 is asmcgocall's stack-switch flag and is still live when this function
// returns to it, so it must not be written here. R14-R20 are asmcgocall's.
TEXT runtime·syscallN_trampoline(SB), NOSPLIT, $32-0
	OR	R16, R3, R3			// libcCallInfo
	OR	R15, R1, R1			// frame SP for restoration
	LWZ	R17, libcCallInfo_args(R16)
	LWZ	R18, libcCallInfo_n(R16)
	CMPWI	R18, $8
	BC	12, 2, loadregs			// exactly eight arguments
	BC	12, 0, loadregs			// fewer than eight

	// Our own frame already carries linkage plus the 32-byte home area, which
	// is all eight or fewer words need. Beyond that, reserve a new frame of
	// 24 + 4*n bytes (linkage, home area, words nine and up) rounded up to
	// 16 bytes, below our frame.
	ADDI	R19, R18, $-8			// stack argument count
	ADD	R19, R19, R19			// words to bytes
	ADD	R19, R19, R19
	ADDI	R19, R19, $71			// 24 linkage + 32 home + args + 15 alignment
	ADDI	R26, R0, $-16
	AND	R19, R19, R26			// round down to 16-byte boundary
	SUBF	R1, R19, R1
	STW	R15, 0(R1)

	// Copy args[8:n] to SP+56. Both source and destination are 4-byte words.
	ADDI	R27, R18, $-8
	ADDI	R21, R0, $0
copyargs:
	CMPW	R21, R27
	BC	12, 1, loadregs			// index >= count
	BC	12, 2, loadregs
	ADD	R22, R21, R21			// index * 4
	ADD	R22, R22, R22
	ADDI	R23, R17, $32			// &args[8]
	ADD	R23, R23, R22
	LWZ	R24, 0(R23)
	ADDI	R25, R22, $56
	ADD	R25, R1, R25
	STW	R24, 0(R25)
	ADDI	R21, R21, $1
	B	copyargs

loadregs:
	CMPWI	R18, $1
	BC	12, 0, call
	LWZ	R3, 0(R17)
	CMPWI	R18, $2
	BC	12, 0, call
	LWZ	R4, 4(R17)
	CMPWI	R18, $3
	BC	12, 0, call
	LWZ	R5, 8(R17)
	CMPWI	R18, $4
	BC	12, 0, call
	LWZ	R6, 12(R17)
	CMPWI	R18, $5
	BC	12, 0, call
	LWZ	R7, 16(R17)
	CMPWI	R18, $6
	BC	12, 0, call
	LWZ	R8, 20(R17)
	CMPWI	R18, $7
	BC	12, 0, call
	LWZ	R9, 24(R17)
	CMPWI	R18, $8
	BC	12, 0, call
	LWZ	R10, 28(R17)

call:
	LWZ	R12, libcCallInfo_fn(R16)
	MTCTR	R12
	BCTRL
	OR	R1, R15, R15			// restore Go frame SP
	STW	R3, libcCallInfo_r1(R16)
	STW	R4, libcCallInfo_r2(R16)
	RET

// syscall_x509 calls a C function with five integer and one floating argument.
TEXT runtime·syscall_x509(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R12, 0(R14)			// fn
	LWZ	R3, 4(R14)			// a1
	LWZ	R4, 8(R14)			// a2
	LWZ	R5, 12(R14)			// a3
	LWZ	R6, 16(R14)			// a4
	LWZ	R7, 20(R14)			// a5
	LFD	24(R14), F1			// f1
	MTCTR	R12
	BCTRL
	STW	R3, 32(R14)			// r1
	RET

TEXT runtime·pthread_attr_init_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_pthread_attr_init(SB)
	RET

TEXT runtime·pthread_attr_getstacksize_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_attr_getstacksize(SB)
	RET

TEXT runtime·pthread_attr_setdetachstate_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_attr_setdetachstate(SB)
	RET

TEXT runtime·pthread_create_trampoline(SB), NOSPLIT, $40-0
	OR	R14, R3, R3
	LWZ	R4, 0(R14)			// attr
	LWZ	R5, 4(R14)			// start
	LWZ	R6, 8(R14)			// arg
	ADDI	R3, R1, $56			// pthread_t temporary, above the home area
	BL	libc_pthread_create(SB)
	RET

TEXT runtime·raise_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_raise(SB)
	RET

TEXT runtime·pthread_self_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	BL	libc_pthread_self(SB)
	STW	R3, 0(R14)
	RET

TEXT runtime·pthread_kill_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)			// signal
	LWZ	R3, 0(R14)			// pthread
	BL	libc_pthread_kill(SB)
	RET

// osinit_hack is a no-op on Leopard, which predates the affected libc bug.
TEXT runtime·osinit_hack_trampoline(SB), NOSPLIT, $0-0
	RET

TEXT runtime·mmap_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R3, 0(R14)			// addr
	LWZ	R4, 4(R14)			// length
	LWZ	R5, 8(R14)			// prot
	LWZ	R6, 12(R14)			// flags
	LWZ	R7, 16(R14)			// fd
	// Go's offset is uint32; libc's off_t is the 64-bit pair R8:R9.
	ADDI	R8, R0, $0			// offset high word (zero extension)
	LWZ	R9, 20(R14)			// offset low word
	BL	libc_mmap(SB)
	CMPWI	R3, $-1
	BC	12, 2, mmaperr
	STW	R3, 24(R14)
	STW	R4, 28(R14)
	RET
mmaperr:
	BL	libc_error(SB)
	LWZ	R4, 0(R3)
	ADDI	R3, R0, $0
	STW	R3, 24(R14)
	STW	R4, 28(R14)
	RET

TEXT runtime·munmap_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_munmap(SB)
	RET

TEXT runtime·madvise_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_madvise(SB)
	RET

TEXT runtime·msync_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_msync(SB)
	RET

TEXT runtime·mlock_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_mlock(SB)
	RET

TEXT runtime·read_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)			// buffer
	LWZ	R5, 8(R14)			// count
	LWZ	R3, 0(R14)			// fd
	BL	libc_read(SB)
	CMPWI	R3, $-1
	BC	12, 2, readerr
	RET
readerr:
	BL	libc_error(SB)
	LWZ	R3, 0(R3)
	NEG	R3, R3
	RET

TEXT runtime·pipe_trampoline(SB), NOSPLIT, $32-0
	BL	libc_pipe(SB)
	CMPWI	R3, $0
	BC	12, 2, pipeok
	BL	libc_error(SB)
	LWZ	R3, 0(R3)
	NEG	R3, R3
	RET
pipeok:
	RET

TEXT runtime·close_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_close(SB)
	RET

TEXT runtime·exit_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_exit(SB)
	ADDI	R3, R0, $0
	ADDI	R4, R0, $0
	STW	R3, 0(R4)			// must not return
	RET

TEXT runtime·usleep_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_usleep(SB)
	RET

TEXT runtime·write_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)			// buffer
	LWZ	R5, 8(R14)			// count
	LWZ	R3, 0(R14)			// fd
	BL	libc_write(SB)
	CMPWI	R3, $-1
	BC	12, 2, writeerr
	RET
writeerr:
	BL	libc_error(SB)
	LWZ	R3, 0(R3)
	NEG	R3, R3
	RET

TEXT runtime·open_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)			// flags
	LWZ	R5, 8(R14)			// mode
	LWZ	R3, 0(R14)			// pathname
	BL	libc_open(SB)
	RET

// nanotime_trampoline asks Leopard for its raw mach clock and timebase.
// The 32-bit Darwin ABI returns uint64 in R3 (high word), R4 (low word).
TEXT runtime·nanotime_trampoline(SB), NOSPLIT, $40-0
	OR	R14, R3, R3
	BL	libc_mach_absolute_time(SB)
	STW	R3, 0(R14)
	STW	R4, 4(R14)
	ADDI	R3, R1, $56			// mach_timebase_info, above the home area
	BL	libc_mach_timebase_info(SB)
	LWZ	R15, 56(R1)			// numer
	LWZ	R16, 60(R1)			// denom
	STW	R15, 8(R14)
	STW	R16, 12(R14)
	RET

// walltime uses gettimeofday on Leopard: timeval matches the 32-bit
// timespec layout here, with microseconds converted to nanoseconds in place.
TEXT runtime·walltime_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	ADDI	R4, R0, $0
	BL	libc_gettimeofday(SB)
	LWZ	R3, 4(R14)
	ADDI	R5, R0, $1000
	MULLW	R3, R3, R5
	STW	R3, 4(R14)
	RET

TEXT runtime·sigaction_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_sigaction(SB)
	RET

TEXT runtime·sigprocmask_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_sigmask(SB)
	RET

TEXT runtime·sigaltstack_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_sigaltstack(SB)
	RET

TEXT runtime·raiseproc_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_raise(SB)
	RET

TEXT runtime·setitimer_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_setitimer(SB)
	RET

TEXT runtime·sysctl_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R6, 12(R14)
	LWZ	R7, 16(R14)
	LWZ	R8, 20(R14)
	LWZ	R3, 0(R14)
	BL	libc_sysctl(SB)
	RET

TEXT runtime·sysctlbyname_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R6, 12(R14)
	LWZ	R7, 16(R14)
	LWZ	R3, 0(R14)
	BL	libc_sysctlbyname(SB)
	RET

TEXT runtime·fcntl_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R3, 0(R14)			// fd
	LWZ	R4, 4(R14)			// command
	LWZ	R5, 8(R14)			// variadic third argument
	BL	libc_fcntl(SB)
	ADDI	R4, R0, $0
	CMPWI	R3, $-1
	BC	12, 2, fcntlerr
	B	fcntlret
fcntlerr:
	BL	libc_error(SB)
	LWZ	R4, 0(R3)
	ADDI	R3, R0, $-1
fcntlret:
	STW	R3, 12(R14)
	STW	R4, 16(R14)
	RET

TEXT runtime·kqueue_trampoline(SB), NOSPLIT, $32-0
	BL	libc_kqueue(SB)
	RET

TEXT runtime·kevent_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R6, 12(R14)
	LWZ	R7, 16(R14)
	LWZ	R8, 20(R14)
	LWZ	R3, 0(R14)
	BL	libc_kevent(SB)
	CMPWI	R3, $-1
	BC	4, 2, keventok
	BL	libc_error(SB)
	LWZ	R3, 0(R3)			// errno
	NEG	R3, R3				// caller wants it as a negative error code
keventok:
	RET

TEXT runtime·pthread_mutex_init_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_mutex_init(SB)
	RET

TEXT runtime·pthread_mutex_lock_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_pthread_mutex_lock(SB)
	RET

TEXT runtime·pthread_mutex_unlock_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_pthread_mutex_unlock(SB)
	RET

TEXT runtime·pthread_cond_init_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_cond_init(SB)
	RET

TEXT runtime·pthread_cond_wait_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_cond_wait(SB)
	RET

TEXT runtime·pthread_cond_timedwait_relative_np_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)
	LWZ	R5, 8(R14)
	LWZ	R3, 0(R14)
	BL	libc_pthread_cond_timedwait_relative_np(SB)
	RET

TEXT runtime·pthread_cond_signal_trampoline(SB), NOSPLIT, $32-0
	LWZ	R3, 0(R3)
	BL	libc_pthread_cond_signal(SB)
	RET

// Leopard lacks arc4random_buf; assemble the same stream from arc4random words.
TEXT runtime·arc4random_buf_trampoline(SB), NOSPLIT, $32-0
	LWZ	R15, 4(R3)			// remaining bytes
	LWZ	R14, 0(R3)			// destination buffer
	ADDI	R16, R0, $0			// bytes left in current word
	ADDI	R31, R0, $8			// shift count
arc4next:
	CMPWI	R15, $0
	BC	12, 0, arc4done
	BC	12, 2, arc4done
	CMPWI	R16, $0
	BC	4, 2, arc4byte
	BL	libc_arc4random(SB)
	OR	R17, R3, R3
	ADDI	R16, R0, $4
arc4byte:
	STB	R17, 0(R14)
	SRW	R17, R17, R31
	ADDI	R14, R14, $1
	ADDI	R15, R15, $-1
	ADDI	R16, R16, $-1
	B	arc4next
arc4done:
	RET

TEXT runtime·issetugid_trampoline(SB), NOSPLIT, $32-0
	BL	libc_issetugid(SB)
	RET

TEXT runtime·mach_vm_region_trampoline(SB), NOSPLIT, $40-0
	OR	R14, R3, R3
	BL	libc_task_self_trap(SB)		// task self port
	LWZ	R4, 0(R14)			// address
	LWZ	R5, 4(R14)			// size
	LWZ	R6, 8(R14)			// flavor
	LWZ	R7, 12(R14)			// info
	LWZ	R8, 16(R14)			// count
	ADDI	R9, R1, $56			// object name output, above the home area
	ADDI	R0, R0, $0			// task_self_trap is C and clobbers R0.
	STW	R0, 56(R1)
	BL	libc_mach_vm_region(SB)
	RET

TEXT runtime·proc_regionfilename_trampoline(SB), NOSPLIT, $32-0
	OR	R14, R3, R3
	LWZ	R4, 4(R14)			// address high word
	LWZ	R5, 8(R14)			// address low word
	LWZ	R6, 12(R14)			// buffer
	// Go's bufSize is int64; libc takes uint32_t (the low word).
	LWZ	R7, 20(R14)			// buffer size
	LWZ	R3, 0(R14)			// pid
	BL	libc_proc_regionfilename(SB)
	RET
