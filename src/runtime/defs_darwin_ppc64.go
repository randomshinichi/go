// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Leopard/ppc64 layouts and constants measured on the G5; see the port's
// docs/darwin-abi-reference.md and out/g5-leopard-abi-20261004/.
const (
	_EINTR     = 4
	_EAGAIN    = 35
	_ETIMEDOUT = 60

	_PROT_NONE  = 0
	_PROT_READ  = 1
	_PROT_WRITE = 2
	_PROT_EXEC  = 4

	_MAP_PRIVATE = 0x2
	_MAP_FIXED   = 0x10
	_MAP_ANON    = 0x1000

	_MADV_DONTNEED = 4
	_MADV_FREE     = 5

	_F_GETFL = 3
	_F_SETFL = 4

	_O_WRONLY   = 0x1
	_O_NONBLOCK = 0x4
	_O_CREAT    = 0x200
	_O_TRUNC    = 0x400

	// Go's internal name for C's PTHREAD_CREATE_DETACHED.
	_PTHREAD_CREATE_DETACHED = 2

	_VM_REGION_BASIC_INFO_64       = 9
	_VM_REGION_BASIC_INFO_COUNT_64 = 9

	_SA_ONSTACK = 0x1
	_SA_RESTART = 0x2
	_SA_SIGINFO = 0x40

	_SIGHUP  = 1
	_SIGINT  = 2
	_SIGQUIT = 3
	_SIGILL  = 4
	_SIGTRAP = 5
	_SIGABRT = 6
	_SIGFPE  = 8
	_SIGBUS  = 10
	_SIGSEGV = 11
	_SIGSYS  = 12
	_SIGPIPE = 13
	_SIGURG  = 16
	_SIGCHLD = 20
	_SIGPROF = 27
	_SIGUSR1 = 30
	_SIGUSR2 = 31

	_BUS_ADRERR  = 2
	_SEGV_MAPERR = 1
	_SEGV_ACCERR = 2
	_FPE_INTDIV  = 7
	_FPE_INTOVF  = 8

	_ITIMER_REAL    = 0
	_ITIMER_VIRTUAL = 1
	_ITIMER_PROF    = 2

	_EV_ADD       = 0x1
	_EV_DELETE    = 0x2
	_EV_ENABLE    = 0x4
	_EV_DISABLE   = 0x8
	_EV_ONESHOT   = 0x10
	_EV_CLEAR     = 0x20
	_EV_RECEIPT   = 0x40
	_EV_ERROR     = 0x4000
	_EV_EOF       = 0x8000
	_EVFILT_READ  = -1
	_EVFILT_WRITE = -2
)

type machPort uint32
type machMsgTypeNumber uint32
type machVMRegionFlavour int32

type usigactiont struct {
	_             [0]uint64
	__sigaction_u [8]byte
	sa_mask       uint32
	sa_flags      int32
}

type siginfo struct {
	si_signo int32
	si_errno int32
	si_code  int32
	_        [12]byte
	si_addr  *byte
	_        [72]byte
}

type timeval struct {
	tv_sec  int64
	tv_usec int32
	_       [4]byte
}

func (tv *timeval) set_usec(x int32) { tv.tv_usec = x }

type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

//go:nosplit
func (ts *timespec) setNsec(ns int64) {
	ts.tv_sec = ns / 1e9
	ts.tv_nsec = ns % 1e9
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

type stackt struct {
	ss_sp    *byte
	ss_size  uintptr
	ss_flags int32
	_        [4]byte
}

// Go over-aligns keventt to 8 rather than C's packed alignment of 4.
// All field offsets and the 32-byte array stride still match the C layout;
// runtime-owned storage is consequently also sufficiently aligned for C.
type keventt struct {
	ident  uint64
	filter int16
	flags  uint16
	fflags uint32
	data   int64
	udata  *byte
}

// pthread objects are opaque: only their measured allocation size/alignment
// is known. Initialization and access belong to libSystem, not Go.
type pthread uintptr

type pthreadmutexattr struct {
	_ [0]uint64
	_ [16]byte
}

type pthreadcondattr struct {
	_ [0]uint64
	_ [16]byte
}

type pthreadmutex struct {
	_ [0]uint64
	_ [64]byte
}

type pthreadcond struct {
	_ [0]uint64
	_ [48]byte
}

// _PTHREAD_TSD_OFFSET is the byte offset from the thread pointer (R13) of the
// first pthread-specific-data slot. libSystem's pthread_getspecific computes
// r13 + 0x60 + key*8 (disassembly of the ppc64 slice supplied 2026-10-04); the
// d4tls probe agreed (key 256, slot found at byte offset 2144 = 0x60+256*8).
// tlsinit validates the formula at startup instead of trusting it.
const _PTHREAD_TSD_OFFSET = 0x60

// pthreadkey is pthread_key_t. UNMEASURED: its width and signedness are not
// recorded in docs/darwin-abi-reference.md (the header only says
// "typedef __darwin_pthread_key_t pthread_key_t"). It is modelled as an 8-byte
// unsigned integer, as on darwin/arm64, and the slot check in tlsinit aborts
// the program if that is wrong in a way that matters. Pending measurement.
type pthreadkey uint64

type pthreadattr struct {
	_ [0]uint64
	_ [64]byte
}

// The thread-state ABI aligns 64-bit registers to 4, including LR at 284
// and CTR at 292. Two big-endian words avoid Go's 8-byte uint64 alignment
// and unaligned doubleword accesses when reading or updating those registers.
type register64 [2]uint32

// Keep the word accesses out of line: SSA memcombine would otherwise fuse
// adjacent halves into an unaligned doubleword load/store on PPC64.
//
//go:nosplit
//go:noinline
func readRegisterWord(p *uint32) uint32 { return *p }

//go:nosplit
//go:noinline
func writeRegisterWord(p *uint32, x uint32) { *p = x }

//go:nosplit
func (r *register64) get() uint64 {
	return uint64(readRegisterWord(&r[0]))<<32 | uint64(readRegisterWord(&r[1]))
}

//go:nosplit
func (r *register64) set(x uint64) {
	writeRegisterWord(&r[0], uint32(x>>32))
	writeRegisterWord(&r[1], uint32(x))
}

type regs64 struct {
	srr0   register64
	srr1   register64
	gpr    [32]register64
	cr     uint32
	xer    register64
	lr     register64
	ctr    register64
	vrsave uint32
}

type mcontext64 struct {
	_  [0]uint64
	es [32]byte
	ss regs64
	fs [840]byte // floating/vector state, not accessed by sigctxt
}

type ucontext struct {
	uc_onstack  int32
	uc_sigmask  sigset // os_darwin.go defines the measured 4-byte sigset
	uc_stack    stackt
	uc_link     *ucontext
	uc_mcsize   uint64
	uc_mcontext *mcontext64
}
