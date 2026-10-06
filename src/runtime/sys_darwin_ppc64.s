// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// System calls and other sys.stuff for PPC64, Darwin (Mac OS X 10.5 Leopard).
// System calls are implemented in libSystem; this file contains
// trampolines that convert from Go to the Darwin C calling convention.
//
// Darwin/ppc64 C ABI facts used here, each measured on the G5
// (out/g5-leopard-abi-20261004, docs/darwin-abi-reference.md):
//   - integer/pointer arguments in R3..R10, result in R3 (batch 5 d4abi-O2.s)
//   - arguments beyond the eighth go on the stack starting at 112(R1):
//     a 48-byte linkage area plus a 64-byte parameter save area (batch 5)
//   - a callee saves its LR at 16(R1) of its caller's frame (batch 5)
//   - R2 is a volatile scratch register: there is no TOC (t_ppc64 disassembly)
//   - function pointers are plain code addresses, with no descriptors (batch 5)
//   - variadic arguments travel in the same registers as fixed ones: open and
//     fcntl pass their third argument in R5 (batch 5 callopen)
//   - R13 is the thread pointer and is never touched by Go code (batch 5/6)
//   - R14..R31 and F14..F31 are callee-saved (batch 5 hungry/fhungry)
//
// The trampolines below are called through asmcgocall, which is itself Go
// code that keeps nothing live in R14..R29 across the call, so they may use
// those registers as scratch without saving them. The two functions that C
// code calls directly, mstart_stub and sigtramp, are different: they save and
// restore the complete callee-saved set.
//
// Every function that calls into libc is declared with a Go frame of $80. The
// assembler adds the 32-byte Go frame header, giving 112 bytes at 0(R1): the
// linkage area plus parameter save area that a C callee may write into. Local
// storage of our own must live above 112(R1).

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"
#include "asm_ppc64x.h"
#include "cgo/abi_ppc64x.h"

// Leopard has no clock_gettime, arc4random_buf, notify_is_valid_token or
// xpc_date_create_from_current (dlsym probe, out/g5-leopard-abi-20261004/batch5
// d4syms.out). The affected trampolines below use a different mechanism and
// say so.

TEXT notok<>(SB),NOSPLIT|NOFRAME,$0
	MOVD	$0, R3
	MOVD	R3, 0(R3)	// nil store: crash
	BR	0(PC)

TEXT runtime·open_trampoline(SB),NOSPLIT,$80
	MOVW	8(R3), R4	// arg 2 flags
	MOVW	12(R3), R5	// arg 3 mode (variadic, still passed in R5)
	MOVD	0(R3), R3	// arg 1 pathname
	BL	libc_open(SB)
	RET

TEXT runtime·close_trampoline(SB),NOSPLIT,$80
	MOVW	0(R3), R3	// arg 1 fd
	BL	libc_close(SB)
	RET

TEXT runtime·write_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 buf
	MOVW	16(R3), R5	// arg 3 count
	MOVD	0(R3), R3	// arg 1 fd (uintptr)
	BL	libc_write(SB)
	CMP	R3, $-1
	BNE	noerr
	BL	libc_error(SB)
	MOVW	0(R3), R3
	NEG	R3, R3		// caller expects negative errno value
noerr:
	RET

TEXT runtime·read_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 buf
	MOVW	16(R3), R5	// arg 3 count
	MOVW	0(R3), R3	// arg 1 fd
	BL	libc_read(SB)
	CMP	R3, $-1
	BNE	noerr
	BL	libc_error(SB)
	MOVW	0(R3), R3
	NEG	R3, R3		// caller expects negative errno value
noerr:
	RET

TEXT runtime·pipe_trampoline(SB),NOSPLIT,$80
	BL	libc_pipe(SB)	// pointer already in R3
	CMPW	R3, $0
	BEQ	done
	BL	libc_error(SB)
	MOVW	0(R3), R3
	NEG	R3, R3		// return negative errno value
done:
	RET

// Exit the entire program (like C exit)
TEXT runtime·exit_trampoline(SB),NOSPLIT,$80
	MOVW	0(R3), R3	// arg 1 exit status
	BL	libc_exit(SB)
	MOVD	$1234, R3
	MOVD	$1002, R4
	MOVD	R3, 0(R4)	// fail hard

TEXT runtime·raiseproc_trampoline(SB),NOSPLIT,$80
	MOVW	0(R3), R14	// signal
	BL	libc_getpid(SB)
	// arg 1 pid already in R3 from getpid
	MOVD	R14, R4		// arg 2 signal
	BL	libc_kill(SB)
	RET

TEXT runtime·mmap_trampoline(SB),NOSPLIT,$80
	MOVD	R3, R14		// save the argument block
	MOVD	0(R14), R3	// arg 1 addr
	MOVD	8(R14), R4	// arg 2 len
	MOVW	16(R14), R5	// arg 3 prot
	MOVW	20(R14), R6	// arg 4 flags
	MOVW	24(R14), R7	// arg 5 fd
	MOVWZ	28(R14), R8	// arg 6 off
	BL	libc_mmap(SB)
	MOVD	$0, R4
	CMP	R3, $-1
	BNE	ok
	BL	libc_error(SB)
	MOVW	0(R3), R4
	MOVD	$0, R3
ok:
	MOVD	R3, 32(R14)	// ret 1 p
	MOVD	R4, 40(R14)	// ret 2 err
	RET

TEXT runtime·munmap_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 len
	MOVD	0(R3), R3	// arg 1 addr
	BL	libc_munmap(SB)
	CMPW	R3, $0
	BEQ	2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·madvise_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 len
	MOVW	16(R3), R5	// arg 3 advice
	MOVD	0(R3), R3	// arg 1 addr
	BL	libc_madvise(SB)
	// ignore failure - maybe pages are locked
	RET

TEXT runtime·msync_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 len
	MOVW	16(R3), R5	// arg 3 flags
	MOVD	0(R3), R3	// arg 1 addr
	BL	libc_msync(SB)
	RET

TEXT runtime·mlock_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 len
	MOVD	0(R3), R3	// arg 1 addr
	BL	libc_mlock(SB)
	RET

TEXT runtime·setitimer_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 new
	MOVD	16(R3), R5	// arg 3 old
	MOVW	0(R3), R3	// arg 1 which
	BL	libc_setitimer(SB)
	RET

// walltime_trampoline: SUBSTITUTED MECHANISM. Leopard has no clock_gettime
// (added in macOS 10.12), so wall-clock time comes from gettimeofday, which
// has microsecond resolution. gettimeofday fills a struct timeval laid out as
// {int64 tv_sec; int32 tv_usec; 4 bytes padding} (docs/darwin-abi-reference.md
// section 10); the 16-byte buffer is the caller's struct timespec, whose
// tv_nsec is an int64 at offset 8, so tv_usec is widened and scaled in place.
// The loss against modern Darwin is resolution only: wall time here is not
// monotonic on either system, and monotonic time (nanotime) does not use this.
TEXT runtime·walltime_trampoline(SB),NOSPLIT,$80
	MOVD	R3, R14		// timespec, reused as the timeval buffer
	MOVD	$0, R4		// arg 2 tz
	BL	libc_gettimeofday(SB)
	MOVW	8(R14), R3	// tv_usec (signed int32)
	MULLD	$1000, R3, R3
	MOVD	R3, 8(R14)	// tv_nsec (int64)
	RET

// timebase<> caches mach_timebase_info: numer at 0, denom at 4. denom is
// written last, after a barrier, and is zero until initialised; a reader that
// sees a non-zero denom and then orders its read of numer after it sees both.
// Two threads racing to initialise store identical values.
GLOBL timebase<>(SB),NOPTR,$8

TEXT runtime·nanotime_trampoline(SB),NOSPLIT,$96
	MOVD	R3, R14
	BL	libc_mach_absolute_time(SB)
	MOVD	R3, 0(R14)
	MOVD	$timebase<>(SB), R16
	MOVWZ	4(R16), R15		// denom
	CMPW	R15, $0
	BNE	initialized

	ADD	$112, R1, R3		// scratch mach_timebase_info in our frame
	BL	libc_mach_timebase_info(SB)
	MOVWZ	112(R1), R17		// numer
	MOVWZ	116(R1), R15		// denom
	MOVW	R17, 0(R16)
	LWSYNC
	MOVW	R15, 4(R16)
	BR	publish

initialized:
	LWSYNC
	MOVWZ	0(R16), R17		// numer
publish:
	MOVW	R17, 8(R14)		// numer
	MOVW	R15, 12(R14)		// denom
	RET

// sigfwd calls a C signal handler fn(sig, info, ctx).
TEXT runtime·sigfwd(SB),NOSPLIT,$80-32
	MOVWZ	sig+8(FP), R3
	MOVD	info+16(FP), R4
	MOVD	ctx+24(FP), R5
	MOVD	fn+0(FP), R12
	MOVD	R12, CTR
	BL	(CTR)
	XOR	R0, R0		// C code does not preserve R0
	RET

// sigtramp is the callback from libc when a signal is received. It is called
// with the C calling convention (sig in R3, siginfo_t* in R4, ucontext_t* in
// R5) on the signal stack. Go clobbers every callee-saved register, so
// STACK_AND_SAVE_HOST_TO_GO_ABI saves and restores R14..R31, F14..F31 and
// V20..V31, as the other PPC64 hosts do.
//
// Not measured on Leopard: the CR save word at 8(R1) of the caller's linkage
// area and vector register save are inherited from the shared macro.
TEXT runtime·sigtramp(SB),NOSPLIT|NOFRAME|TOPFRAME,$0
	STACK_AND_SAVE_HOST_TO_GO_ABI(32)

	// This might be called in external code context, where g is not set,
	// or where R30 holds a C value: always reload g from thread storage.
	BL	runtime·load_g(SB)

	// R3, R4, R5 already hold the arguments. Call sigtrampgo through a
	// register so the linker's nosplit stack check does not follow it: we are
	// on the signal stack, not on a small Go stack.
	MOVD	$runtime·sigtrampgo<ABIInternal>(SB), R12
	MOVD	R12, CTR
	BL	(CTR)

	UNSTACK_AND_RESTORE_GO_TO_HOST_ABI(32)
	RET

// cgoSigtramp: this port builds with cgo disabled and does not support
// runtime/cgo's traceback hooks, so it is the plain signal trampoline.
TEXT runtime·cgoSigtramp(SB),NOSPLIT|NOFRAME,$0
	BR	runtime·sigtramp(SB)

TEXT runtime·sigprocmask_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 new
	MOVD	16(R3), R5	// arg 3 old
	MOVWZ	0(R3), R3	// arg 1 how
	BL	libc_pthread_sigmask(SB)
	CMPW	R3, $0
	BEQ	2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·sigaction_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 new
	MOVD	16(R3), R5	// arg 3 old
	MOVWZ	0(R3), R3	// arg 1 sig
	BL	libc_sigaction(SB)
	CMPW	R3, $0
	BEQ	2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·usleep_trampoline(SB),NOSPLIT,$80
	MOVWZ	0(R3), R3	// arg 1 usec
	BL	libc_usleep(SB)
	RET

TEXT runtime·sysctl_trampoline(SB),NOSPLIT,$80
	MOVWZ	8(R3), R4	// arg 2 miblen
	MOVD	16(R3), R5	// arg 3 oldp
	MOVD	24(R3), R6	// arg 4 oldlenp
	MOVD	32(R3), R7	// arg 5 newp
	MOVD	40(R3), R8	// arg 6 newlen
	MOVD	0(R3), R3	// arg 1 mib
	BL	libc_sysctl(SB)
	RET

TEXT runtime·sysctlbyname_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 oldp
	MOVD	16(R3), R5	// arg 3 oldlenp
	MOVD	24(R3), R6	// arg 4 newp
	MOVD	32(R3), R7	// arg 5 newlen
	MOVD	0(R3), R3	// arg 1 name
	BL	libc_sysctlbyname(SB)
	RET

TEXT runtime·kqueue_trampoline(SB),NOSPLIT,$80
	BL	libc_kqueue(SB)
	RET

TEXT runtime·kevent_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 keventt
	MOVW	16(R3), R5	// arg 3 nch
	MOVD	24(R3), R6	// arg 4 ev
	MOVW	32(R3), R7	// arg 5 nev
	MOVD	40(R3), R8	// arg 6 ts
	MOVW	0(R3), R3	// arg 1 kq
	BL	libc_kevent(SB)
	CMPW	R3, $-1
	BNE	ok
	BL	libc_error(SB)
	MOVW	0(R3), R3	// errno
	NEG	R3, R3		// caller wants it as a negative error code
ok:
	RET

TEXT runtime·fcntl_trampoline(SB),NOSPLIT,$80
	MOVD	R3, R14
	MOVW	0(R14), R3	// arg 1 fd
	MOVW	4(R14), R4	// arg 2 cmd
	MOVW	8(R14), R5	// arg 3 arg (variadic, still passed in R5)
	BL	libc_fcntl(SB)
	MOVD	$0, R4
	CMPW	R3, $-1
	BNE	noerr
	BL	libc_error(SB)
	MOVW	0(R3), R4
	MOVW	$-1, R3
noerr:
	MOVW	R3, 12(R14)	// ret
	MOVW	R4, 16(R14)	// errno
	RET

TEXT runtime·sigaltstack_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 old
	MOVD	0(R3), R3	// arg 1 new
	BL	libc_sigaltstack(SB)
	CMPW	R3, $0
	BEQ	2(PC)
	BL	notok<>(SB)
	RET

// Thread related functions

// mstart_stub is the first function executed on a new thread started by
// pthread_create. It just does some low-level setup and then calls mstart.
// Note: called with the C calling convention.
TEXT runtime·mstart_stub(SB),NOSPLIT|NOFRAME,$0
	// R3 points to the m.
	// We are already on m's g0 stack.
	STACK_AND_SAVE_HOST_TO_GO_ABI(32)

	MOVD	m_g0(R3), g
	BL	·save_g(SB)

	BL	runtime·mstart(SB)

	UNSTACK_AND_RESTORE_GO_TO_HOST_ABI(32)

	// Go is all done with this OS thread.
	// Tell pthread everything is ok (we never join with this thread, so
	// the value here doesn't really matter).
	MOVD	$0, R3

	RET

TEXT runtime·pthread_attr_init_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R3	// arg 1 attr
	BL	libc_pthread_attr_init(SB)
	RET

TEXT runtime·pthread_attr_getstacksize_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 size
	MOVD	0(R3), R3	// arg 1 attr
	BL	libc_pthread_attr_getstacksize(SB)
	RET

TEXT runtime·pthread_attr_setdetachstate_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 state
	MOVD	0(R3), R3	// arg 1 attr
	BL	libc_pthread_attr_setdetachstate(SB)
	RET

TEXT runtime·pthread_create_trampoline(SB),NOSPLIT,$96
	MOVD	0(R3), R4	// arg 2 attr
	MOVD	8(R3), R5	// arg 3 start
	MOVD	16(R3), R6	// arg 4 arg
	ADD	$112, R1, R3	// arg 1 &threadid (which we throw away)
	BL	libc_pthread_create(SB)
	RET

TEXT runtime·raise_trampoline(SB),NOSPLIT,$80
	MOVW	0(R3), R3	// arg 1 sig
	BL	libc_raise(SB)
	RET

TEXT runtime·pthread_mutex_init_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 attr
	MOVD	0(R3), R3	// arg 1 mutex
	BL	libc_pthread_mutex_init(SB)
	RET

TEXT runtime·pthread_mutex_lock_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R3	// arg 1 mutex
	BL	libc_pthread_mutex_lock(SB)
	RET

TEXT runtime·pthread_mutex_unlock_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R3	// arg 1 mutex
	BL	libc_pthread_mutex_unlock(SB)
	RET

TEXT runtime·pthread_cond_init_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 attr
	MOVD	0(R3), R3	// arg 1 cond
	BL	libc_pthread_cond_init(SB)
	RET

TEXT runtime·pthread_cond_wait_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 mutex
	MOVD	0(R3), R3	// arg 1 cond
	BL	libc_pthread_cond_wait(SB)
	RET

TEXT runtime·pthread_cond_timedwait_relative_np_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 mutex
	MOVD	16(R3), R5	// arg 3 timeout
	MOVD	0(R3), R3	// arg 1 cond
	BL	libc_pthread_cond_timedwait_relative_np(SB)
	RET

TEXT runtime·pthread_cond_signal_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R3	// arg 1 cond
	BL	libc_pthread_cond_signal(SB)
	RET

TEXT runtime·pthread_self_trampoline(SB),NOSPLIT,$80
	MOVD	R3, R14		// R14 is callee-save in C
	BL	libc_pthread_self(SB)
	MOVD	R3, 0(R14)	// return value
	RET

TEXT runtime·pthread_kill_trampoline(SB),NOSPLIT,$80
	MOVWZ	8(R3), R4	// arg 2 sig
	MOVD	0(R3), R3	// arg 1 thread
	BL	libc_pthread_kill(SB)
	RET

TEXT runtime·pthread_key_create_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 destructor
	MOVD	0(R3), R3	// arg 1 *key
	BL	libc_pthread_key_create(SB)
	RET

TEXT runtime·pthread_setspecific_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// arg 2 value
	MOVD	0(R3), R3	// arg 1 key
	BL	libc_pthread_setspecific(SB)
	RET

// osinit_hack_trampoline: SUBSTITUTED MECHANISM (no-op).
// The modern-Darwin version calls notify_is_valid_token(0) and
// xpc_date_create_from_current() to work around an Apple libc bug in which a
// fork+exec could hang or crash (sys_darwin.go, osinit_hack). Neither function
// exists on Leopard (dlsym probe, batch5 d4syms.out), and the bug lives in the
// modern libnotify/libxpc fork handlers that Leopard does not have. What this
// loses: that workaround, on a libc that does not have the problem it
// addresses. A cgo-disabled pure-Go runtime on Leopard has no code path that
// can reach those handlers. This is a justified no-op, not a stub.
TEXT runtime·osinit_hack_trampoline(SB),NOSPLIT|NOFRAME,$0
	RET

// arc4random_buf_trampoline: SUBSTITUTED MECHANISM. Leopard has arc4random
// but neither arc4random_buf nor arc4random_uniform (batch5 d4syms.out), so
// the buffer is filled four bytes at a time from arc4random(), least
// significant byte first, one byte per store so that p needs no alignment.
// No modulo reduction is involved, so there is no bias to account for; the
// generator is the same libc arc4random stream.
TEXT runtime·arc4random_buf_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R14	// buf
	MOVW	8(R3), R15	// nbytes
	MOVD	$0, R16		// bytes left in R17
nextbyte:
	CMP	R15, $0
	BLE	done
	CMP	R16, $0
	BNE	haveword
	BL	libc_arc4random(SB)
	MOVWZ	R3, R17
	MOVD	$4, R16
haveword:
	MOVB	R17, 0(R14)
	SRD	$8, R17, R17
	ADD	$1, R14
	SUB	$1, R16
	SUB	$1, R15
	BR	nextbyte
done:
	RET

// syscallN_trampoline calls fn(args...) with n integer arguments taken from
// a libcCallInfo. Arguments 1..8 go in R3..R10; argument 9 onward go on the
// stack starting at 112(R1), above the 48-byte linkage area and 64-byte
// parameter save area (batch 5 many()). The outgoing frame is built below the
// current one so that nothing of this function's own frame is overwritten.
TEXT runtime·syscallN_trampoline(SB),NOSPLIT,$80
	MOVD	R3, R14				// *libcCallInfo
	MOVD	R1, R15				// original SP
	MOVD	libcCallInfo_args(R14), R16
	MOVD	libcCallInfo_n(R14), R17
	CMP	R17, $8
	BLE	loadregs

	// More than 8 arguments: 112 + 8*(n-8) bytes rounded up to 16 below SP.
	SUB	$7, R17, R18			// n-8+1
	SRD	$1, R18, R18
	SLD	$4, R18, R18			// 16*ceil((n-8)/2)
	ADD	$112, R18, R18
	SUB	R18, R1, R19
	MOVD	R19, R1
	MOVD	R15, 0(R1)			// back chain

	// Copy args[8:n] to 112(R1).
	SUB	$8, R17, R18
	MOVD	R18, CTR
	ADD	$56, R16, R20			// &args[7]; MOVDU pre-increments
	ADD	$104, R1, R21			// 112(R1)-8
copystack:
	MOVDU	8(R20), R22
	MOVDU	R22, 8(R21)
	BDNZ	copystack

loadregs:
	CMP	R17, $8
	BLT	n7
	MOVD	56(R16), R10
n7:
	CMP	R17, $7
	BLT	n6
	MOVD	48(R16), R9
n6:
	CMP	R17, $6
	BLT	n5
	MOVD	40(R16), R8
n5:
	CMP	R17, $5
	BLT	n4
	MOVD	32(R16), R7
n4:
	CMP	R17, $4
	BLT	n3
	MOVD	24(R16), R6
n3:
	CMP	R17, $3
	BLT	n2
	MOVD	16(R16), R5
n2:
	CMP	R17, $2
	BLT	n1
	MOVD	8(R16), R4
n1:
	CMP	R17, $1
	BLT	call
	MOVD	0(R16), R3
call:
	MOVD	libcCallInfo_fn(R14), R12
	MOVD	R12, CTR
	BL	(CTR)

	MOVD	R15, R1				// free any stack arguments

	MOVD	R3, libcCallInfo_r1(R14)
	MOVD	R4, libcCallInfo_r2(R14)	// only meaningful for two-register results
	RET

TEXT runtime·libc_error_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R14
	BL	libc_error(SB)
	MOVD	R3, 0(R14)
	RET

// syscall_x509 is for crypto/x509. It is like syscall6 but does not check for
// errors, takes 5 uintptrs and 1 float64, and only returns one value, for use
// with standard C ABI functions.
TEXT runtime·syscall_x509(SB),NOSPLIT,$80
	MOVD	R3, R14		// structure pointer

	MOVD	0(R14), R12	// fn
	MOVD	16(R14), R4	// a2
	MOVD	24(R14), R5	// a3
	MOVD	32(R14), R6	// a4
	MOVD	40(R14), R7	// a5
	FMOVD	48(R14), F1	// f1
	MOVD	8(R14), R3	// a1
	MOVD	R12, CTR
	BL	(CTR)

	MOVD	R3, 56(R14)	// save r1
	RET

TEXT runtime·issetugid_trampoline(SB),NOSPLIT,$80
	BL	libc_issetugid(SB)
	RET

// mach_vm_region_trampoline calls mach_vm_region from libc.
TEXT runtime·mach_vm_region_trampoline(SB),NOSPLIT,$80
	MOVD	0(R3), R4	// address
	MOVD	8(R3), R5	// size
	MOVW	16(R3), R6	// flavor
	MOVD	24(R3), R7	// info
	MOVD	32(R3), R8	// count
	MOVD	40(R3), R9	// object_name
	MOVD	$libc_mach_task_self_(SB), R3
	MOVWZ	0(R3), R3
	BL	libc_mach_vm_region(SB)
	RET

// proc_regionfilename_trampoline calls proc_regionfilename for
// the current process.
TEXT runtime·proc_regionfilename_trampoline(SB),NOSPLIT,$80
	MOVD	8(R3), R4	// address
	MOVD	16(R3), R5	// buffer
	MOVD	24(R3), R6	// buffer_size
	MOVD	0(R3), R3	// pid
	BL	libc_proc_regionfilename(SB)
	RET
