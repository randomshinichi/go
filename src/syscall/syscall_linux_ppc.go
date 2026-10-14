// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

package syscall

const (
	_SYS_setgroups   = SYS_SETGROUPS
	_SYS_clone3      = 435
	_SYS_faccessat2  = 439
	_SYS_fchmodat2   = 452
	SYS_NEWFSTATAT   = SYS_FSTATAT64
)

func setTimespec(sec, nsec int64) Timespec {
	return Timespec{Sec: int32(sec), Nsec: int32(nsec)}
}

func setTimeval(sec, usec int64) Timeval {
	return Timeval{Sec: int32(sec), Usec: int32(usec)}
}

func (r *PtraceRegs) PC() uint64 { return uint64(r.Nip) }

func (r *PtraceRegs) SetPC(pc uint64) { r.Nip = uint32(pc) }

func (iov *Iovec) SetLen(length int) { iov.Len = uint32(length) }

func (msghdr *Msghdr) SetControllen(length int) { msghdr.Controllen = uint32(length) }

func (cmsg *Cmsghdr) SetLen(length int) { cmsg.Len = uint32(length) }
