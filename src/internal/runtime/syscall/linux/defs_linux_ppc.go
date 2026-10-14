// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linux

const (
	SYS_CLOSE         = 6
	SYS_FCNTL         = 55
	SYS_MPROTECT      = 125
	SYS_PRCTL         = 171
	SYS_EPOLL_CTL     = 237
	SYS_EPOLL_PWAIT   = 303
	SYS_EPOLL_CREATE1 = 315
	SYS_EPOLL_PWAIT2  = 441
	SYS_EVENTFD2      = 314
	SYS_OPENAT        = 286
	SYS_PREAD64       = 179
	SYS_READ          = 3
	SYS_UNAME         = 122

	EFD_NONBLOCK = 0x800

	O_LARGEFILE = 0x10000
)

type EpollEvent struct {
	Events uint32
	_pad   uint32
	Data   [8]byte // the 64-bit kernel data field is aligned after Events
}
