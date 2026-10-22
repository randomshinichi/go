// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

/*
Input to cgo -godefs.  See README.md
*/

// +godefs map struct_in_addr [4]byte /* in_addr */
// +godefs map struct_in6_addr [16]byte /* in6_addr */

package unix

/*
#define __DARWIN_UNIX03 0
#define KERNEL 1
#define _DARWIN_USE_64_BIT_INODE
#include <dirent.h>
#include <fcntl.h>
#include <poll.h>
#include <stdint.h>
#include <signal.h>
#include <termios.h>
#include <unistd.h>
#include <mach/mach.h>
#include <mach/message.h>
#include <sys/event.h>
#include <sys/ipc.h>
#include <sys/types.h>
#include <sys/kern_control.h>
#include <sys/mman.h>
#include <sys/mount.h>
#include <sys/param.h>
#include <sys/ptrace.h>
#include <sys/resource.h>
#include <sys/select.h>
#include <sys/shm.h>
#include <sys/signal.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/sysctl.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/ucred.h>
#include <sys/uio.h>
#include <sys/un.h>
#include <sys/utsname.h>
#include <sys/wait.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/if_var.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/icmp6.h>
#include <netinet/tcp.h>

enum {
	sizeofPtr = sizeof(void*),
};

union sockaddr_all {
	struct sockaddr s1;	// this one gets used for fields
	struct sockaddr_in s2;	// these pad it out
	struct sockaddr_in6 s3;
	struct sockaddr_un s4;
	struct sockaddr_dl s5;
};

struct sockaddr_any {
	struct sockaddr addr;
	char pad[sizeof(union sockaddr_all) - sizeof(struct sockaddr)];
};

#if defined(__x86_64__) || defined(__ppc__)
typedef struct stat64 stat_t;
typedef struct statfs64 statfs_t;
#else // __arm__
typedef struct stat stat_t;
typedef struct statfs statfs_t;
#endif
*/
import "C"

// Machine characteristics

const (
	SizeofPtr      = C.sizeofPtr
	SizeofShort    = C.sizeof_short
	SizeofInt      = C.sizeof_int
	SizeofLong     = C.sizeof_long
	SizeofLongLong = C.sizeof_longlong
)

// Basic types

type (
	_C_short     int16
	_C_int       int32
	_C_long      int32
	_C_long_long int64
)

// Time

type Timespec C.struct_timespec

type Timeval C.struct_timeval

// Processes

type Rusage C.struct_rusage

type Rlimit C.struct_rlimit

type _Gid_t uint32

// Files

type Stat_t C.struct_stat64

type Statfs_t C.struct_statfs64

type Flock_t C.struct_flock

type Fstore_t C.struct_fstore

type Radvisory_t C.struct_radvisory

type Fbootstraptransfer_t C.struct_fbootstraptransfer

type Log2phys_t C.struct_log2phys

type Fsid C.struct_fsid

type Dirent C.struct_dirent

type Attrlist C.struct_attrlist

// File system limits

const (
	PathMax = C.PATH_MAX
)

// Sockets

type RawSockaddrInet4 C.struct_sockaddr_in

type RawSockaddrInet6 C.struct_sockaddr_in6

type RawSockaddrUnix C.struct_sockaddr_un

type RawSockaddrDatalink C.struct_sockaddr_dl

type RawSockaddr C.struct_sockaddr

type RawSockaddrAny C.struct_sockaddr_any

type RawSockaddrCtl C.struct_sockaddr_ctl

type _Socklen uint32

type Xucred C.struct_xucred

type Linger C.struct_linger

type Iovec C.struct_iovec

type IPMreq C.struct_ip_mreq

type IPv6Mreq C.struct_ipv6_mreq

type Msghdr C.struct_msghdr

type Cmsghdr C.struct_cmsghdr

type Inet6Pktinfo C.struct_in6_pktinfo

type ICMPv6Filter C.struct_icmp6_filter

const (
	SizeofSockaddrInet4     = C.sizeof_struct_sockaddr_in
	SizeofSockaddrInet6     = C.sizeof_struct_sockaddr_in6
	SizeofSockaddrAny       = C.sizeof_struct_sockaddr_any
	SizeofSockaddrUnix      = C.sizeof_struct_sockaddr_un
	SizeofSockaddrDatalink  = C.sizeof_struct_sockaddr_dl
	SizeofSockaddrCtl       = C.sizeof_struct_sockaddr_ctl
	SizeofXucred            = C.sizeof_struct_xucred
	SizeofLinger            = C.sizeof_struct_linger
	SizeofIovec             = C.sizeof_struct_iovec
	SizeofIPMreq            = C.sizeof_struct_ip_mreq
	SizeofIPv6Mreq          = C.sizeof_struct_ipv6_mreq
	SizeofMsghdr            = C.sizeof_struct_msghdr
	SizeofCmsghdr           = C.sizeof_struct_cmsghdr
	SizeofInet6Pktinfo      = C.sizeof_struct_in6_pktinfo
	SizeofICMPv6Filter      = C.sizeof_struct_icmp6_filter
)

// Ptrace requests

const (
	PTRACE_TRACEME = C.PT_TRACE_ME
	PTRACE_CONT    = C.PT_CONTINUE
	PTRACE_KILL    = C.PT_KILL
)

// Events (kqueue, kevent)

type Kevent_t C.struct_kevent

// Select

type FdSet C.struct_fd_set

// Routing and interface messages

const (
	SizeofIfMsghdr    = C.sizeof_struct_if_msghdr
	SizeofIfMsghdr2   = C.sizeof_struct_if_msghdr2
	SizeofIfData      = C.sizeof_struct_if_data
	SizeofIfData64    = C.sizeof_struct_if_data64
	SizeofIfaMsghdr   = C.sizeof_struct_ifa_msghdr
	SizeofIfmaMsghdr  = C.sizeof_struct_ifma_msghdr
	SizeofIfmaMsghdr2 = C.sizeof_struct_ifma_msghdr2
	SizeofRtMsghdr    = C.sizeof_struct_rt_msghdr
	SizeofRtMsghdr2   = C.sizeof_struct_rt_msghdr2
	SizeofRtMetrics   = C.sizeof_struct_rt_metrics
)

type IfMsghdr C.struct_if_msghdr

type IfMsghdr2 C.struct_if_msghdr2

type IfData C.struct_if_data

type IfData64 C.struct_if_data64

type IfaMsghdr C.struct_ifa_msghdr

type IfmaMsghdr C.struct_ifma_msghdr

type IfmaMsghdr2 C.struct_ifma_msghdr2

type RtMsghdr C.struct_rt_msghdr

type RtMsghdr2 C.struct_rt_msghdr2

type RtMetrics C.struct_rt_metrics

// Berkeley packet filter

const (
	SizeofBpfVersion = C.sizeof_struct_bpf_version
	SizeofBpfStat    = C.sizeof_struct_bpf_stat
	SizeofBpfProgram = C.sizeof_struct_bpf_program
	SizeofBpfInsn    = C.sizeof_struct_bpf_insn
	SizeofBpfHdr     = C.sizeof_struct_bpf_hdr
)

type BpfVersion C.struct_bpf_version

type BpfStat C.struct_bpf_stat

type BpfProgram C.struct_bpf_program

type BpfInsn C.struct_bpf_insn

type BpfHdr C.struct_bpf_hdr

// Terminal handling

type Termios C.struct_termios

type Winsize C.struct_winsize

// poll

type PollFd C.struct_pollfd

const (
	POLLERR    = C.POLLERR
	POLLHUP    = C.POLLHUP
	POLLIN     = C.POLLIN
	POLLNVAL   = C.POLLNVAL
	POLLOUT    = C.POLLOUT
	POLLPRI    = C.POLLPRI
	POLLRDBAND = C.POLLRDBAND
	POLLRDNORM = C.POLLRDNORM
	POLLWRBAND = C.POLLWRBAND
	POLLWRNORM = C.POLLWRNORM
)

// uname

type Utsname C.struct_utsname

// Clockinfo

const SizeofClockinfo = C.sizeof_struct_clockinfo

type Clockinfo C.struct_clockinfo

// ctl_info

type CtlInfo C.struct_ctl_info

// KinfoProc

const SizeofKinfoProc = C.sizeof_struct_kinfo_proc

type Eproc C.struct_eproc

type ExternProc C.struct_extern_proc

type Itimerval C.struct_itimerval

type KinfoProc C.struct_kinfo_proc

type Vmspace C.struct_vmspace

type Pcred C.struct__pcred

type Ucred C.struct__ucred

// shm

type SysvIpcPerm C.struct_ipc_perm
type SysvShmDesc C.struct_shmid_ds

const (
	IPC_CREAT   = C.IPC_CREAT
	IPC_EXCL    = C.IPC_EXCL
	IPC_NOWAIT  = C.IPC_NOWAIT
	IPC_PRIVATE = C.IPC_PRIVATE
)

const (
	IPC_RMID = C.IPC_RMID
	IPC_SET  = C.IPC_SET
	IPC_STAT = C.IPC_STAT
)

const (
	SHM_RDONLY = C.SHM_RDONLY
	SHM_RND    = C.SHM_RND
)
