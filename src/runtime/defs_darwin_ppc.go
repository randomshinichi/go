// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

import "unsafe"

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

	_MADV_DONTNEED      = 4
	_MADV_FREE          = 5
	_MADV_FREE_REUSABLE = 7
	_MADV_FREE_REUSE    = 8

	_MS_KILLPAGES = 0x4

	_F_GETFL = 3
	_F_SETFL = 4

	_O_WRONLY   = 0x1
	_O_NONBLOCK = 0x4
	_O_CREAT    = 0x200
	_O_TRUNC    = 0x400

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
	_EVFILT_USER  = -10
	_NOTE_TRIGGER = 0x1000000
)

type usigactiont struct {
	__sigaction_u [4]byte
	sa_mask       uint32
	sa_flags      int32
}

type timeval struct {
	tv_sec  int32
	tv_usec int32
}

func (tv *timeval) set_usec(x int32) { tv.tv_usec = x }

type timespec struct {
	tv_sec  int32
	tv_nsec int32
}

//go:nosplit
func (ts *timespec) setNsec(ns int64) {
	ts.tv_sec = int32(ns / 1e9)
	ts.tv_nsec = int32(ns % 1e9)
}

type itimerval struct {
	it_interval timeval
	it_value    timeval
}

type stackt struct {
	ss_sp    *byte
	ss_size  uintptr
	ss_flags int32
}

type keventt struct {
	ident  uint32
	filter int16
	flags  uint16
	fflags uint32
	data   int32
	udata  *byte
}

type pthread uintptr

type pthreadattr struct {
	_ [0]uint32
	_ [40]byte
}

type pthreadmutexattr struct {
	_ [0]uint32
	_ [12]byte
}

type pthreadcondattr struct {
	_ [0]uint32
	_ [8]byte
}

type pthreadmutex struct {
	_ [0]uint32
	_ [44]byte
}

type pthreadcond struct {
	_ [0]uint32
	_ [28]byte
}

type machPort uint32
type machMsgTypeNumber uint32
type machVMRegionFlavour int32

type siginfo struct {
	si_signo  int32
	si_errno  int32
	si_code   int32
	si_pid    int32
	si_uid    uint32
	si_status int32
	si_addr   *byte
	_         [36]byte
}

type regs32 struct {
	srr0, srr1 uint32
	gpr        [32]uint32
	cr, xer    uint32
	lr, ctr    uint32
	mq, vrsave uint32
}

type mcontext32 struct {
	_  [0]uint32
	es [32]byte
	ss regs32
	fs [264]byte
	vs [576]byte
}

type ucontext struct {
	uc_onstack  int32
	uc_sigmask  uint32
	uc_stack    stackt
	uc_link     *ucontext
	uc_mcsize   uint32
	uc_mcontext *mcontext32
}

// Keep the alignment and sizes measured from the Leopard ppc headers.
var _ [12 - unsafe.Sizeof(stackt{})]byte
var _ [20 - unsafe.Sizeof(keventt{})]byte
var _ [44 - unsafe.Sizeof(pthreadmutex{})]byte
var _ [28 - unsafe.Sizeof(pthreadcond{})]byte
var _ [64 - unsafe.Sizeof(siginfo{})]byte
var _ [1032 - unsafe.Sizeof(mcontext32{})]byte
var _ [32 - unsafe.Offsetof(mcontext32{}.ss)]byte
var _ [192 - unsafe.Offsetof(mcontext32{}.fs)]byte
var _ [456 - unsafe.Offsetof(mcontext32{}.vs)]byte
var _ [32 - unsafe.Sizeof(ucontext{})]byte
