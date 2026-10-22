// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

// Port-local (goppc-jic6x.3). Identifiers that vendored generic x/sys code or
// syscall_darwin_ppc.go references but that Mac OS X 10.5 Leopard's headers do
// not define (checked read-only against MacOSX10.5.sdk/usr/include on the
// iBook, 2026-10-22; see src/syscall/testdata/darwin_ppc32/xsys/README).
// None of these is an Apple-provided Leopard value. Values and shapes are
// copied from zerrors/ztypes_darwin_arm64.go and laid out for ppc32, so calls
// fail in the Leopard kernel or in the ENOSYS/ENOTSUP wrappers instead of
// silently succeeding.

package unix

// ABSENT on Leopard: no AT_* constants and no *at system calls. These are
// sentinels matching the other darwin ports and the port's
// internal/syscall/unix/at_sysnum_darwin.go; only syscall_darwin_ppc.go
// interprets them.
const (
	AT_FDCWD            = -0x2
	AT_REMOVEDIR        = 0x80
	AT_SYMLINK_FOLLOW   = 0x40
	AT_SYMLINK_NOFOLLOW = 0x20
	AT_EACCESS          = 0x10
)

// ABSENT on Leopard: no AF_VSOCK. AF_MAX is 37 (sys/socket.h:288), so a
// socket of this family fails with EAFNOSUPPORT in the kernel.
const AF_VSOCK = 0x28

// ABSENT on Leopard: no sys/vsock.h.
type RawSockaddrVM struct {
	Len       uint8
	Family    uint8
	Reserved1 uint16
	Port      uint32
	Cid       uint32
}

const SizeofSockaddrVM = 0xc

// ABSENT on Leopard: no connectx(2) or sae_*_t; connectx returns ENOSYS.
type SaeAssocID uint32

type SaeConnID uint32

type SaEndpoints struct {
	Srcif      uint32
	Srcaddr    *RawSockaddr
	Srcaddrlen uint32
	Dstaddr    *RawSockaddr
	Dstaddrlen uint32
}

// ABSENT on Leopard: no struct ip_mreqn.
type IPMreqn struct {
	Multiaddr [4]byte /* in_addr */
	Address   [4]byte /* in_addr */
	Ifindex   int32
}

const SizeofIPMreqn = 0xc

// ABSENT on Leopard: no struct ip6_mtuinfo. Same shape as the port's
// UNVERIFIED compatibility type in src/syscall/ztypes_darwin_ppc.go.
type IPv6MTUInfo struct {
	Addr RawSockaddrInet6
	Mtu  uint32
}

const SizeofIPv6MTUInfo = 0x20

// ABSENT on Leopard: no struct tcp_connection_info or TCP_CONNECTION_INFO.
// Sizeof is the ppc32 Go size (uint64 aligned to 4), so getsockopt can never
// write past the Go value.
type TCPConnectionInfo struct {
	State               uint8
	Snd_wscale          uint8
	Rcv_wscale          uint8
	_                   uint8
	Options             uint32
	Flags               uint32
	Rto                 uint32
	Maxseg              uint32
	Snd_ssthresh        uint32
	Snd_cwnd            uint32
	Snd_wnd             uint32
	Snd_sbbytes         uint32
	Rcv_wnd             uint32
	Rttcur              uint32
	Srtt                uint32
	Rttvar              uint32
	Txpackets           uint64
	Txbytes             uint64
	Txretransmitbytes   uint64
	Rxpackets           uint64
	Rxbytes             uint64
	Rxoutoforderbytes   uint64
	Txretransmitpackets uint64
}

const SizeofTCPConnectionInfo = 0x6c
