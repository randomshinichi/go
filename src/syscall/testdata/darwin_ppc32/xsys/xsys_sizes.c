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

#include <stdio.h>

#define P(N,T) printf("%s=%lu\n", N, (unsigned long)sizeof(T))

int main(void) {
	P("Timespec", struct timespec);
	P("Timeval", struct timeval);
	P("Rusage", struct rusage);
	P("Rlimit", struct rlimit);
	P("Stat_t", stat_t);
	P("Statfs_t", statfs_t);
	P("Flock_t", struct flock);
	P("Fstore_t", struct fstore);
	P("Radvisory_t", struct radvisory);
	P("Fbootstraptransfer_t", struct fbootstraptransfer);
	P("Log2phys_t", struct log2phys);
	P("Fsid", struct fsid);
	P("Dirent", struct dirent);
	P("Attrlist", struct attrlist);
	P("RawSockaddrInet4", struct sockaddr_in);
	P("RawSockaddrInet6", struct sockaddr_in6);
	P("RawSockaddrUnix", struct sockaddr_un);
	P("RawSockaddrDatalink", struct sockaddr_dl);
	P("RawSockaddr", struct sockaddr);
	P("RawSockaddrAny", struct sockaddr_any);
	P("RawSockaddrCtl", struct sockaddr_ctl);
	P("Xucred", struct xucred);
	P("Linger", struct linger);
	P("Iovec", struct iovec);
	P("IPMreq", struct ip_mreq);
	P("IPv6Mreq", struct ipv6_mreq);
	P("Msghdr", struct msghdr);
	P("Cmsghdr", struct cmsghdr);
	P("Inet6Pktinfo", struct in6_pktinfo);
	P("ICMPv6Filter", struct icmp6_filter);
	P("Kevent_t", struct kevent);
	P("FdSet", fd_set);
	P("IfMsghdr", struct if_msghdr);
	P("IfMsghdr2", struct if_msghdr2);
	P("IfData", struct if_data);
	P("IfData64", struct if_data64);
	P("IfaMsghdr", struct ifa_msghdr);
	P("IfmaMsghdr", struct ifma_msghdr);
	P("IfmaMsghdr2", struct ifma_msghdr2);
	P("RtMsghdr", struct rt_msghdr);
	P("RtMsghdr2", struct rt_msghdr2);
	P("RtMetrics", struct rt_metrics);
	P("BpfVersion", struct bpf_version);
	P("BpfStat", struct bpf_stat);
	P("BpfProgram", struct bpf_program);
	P("BpfInsn", struct bpf_insn);
	P("BpfHdr", struct bpf_hdr);
	P("Termios", struct termios);
	P("Winsize", struct winsize);
	P("PollFd", struct pollfd);
	P("Utsname", struct utsname);
	P("Clockinfo", struct clockinfo);
	P("CtlInfo", struct ctl_info);
	P("Eproc", struct eproc);
	P("ExternProc", struct extern_proc);
	P("Itimerval", struct itimerval);
	P("KinfoProc", struct kinfo_proc);
	P("Vmspace", struct vmspace);
	P("Pcred", struct _pcred);
	P("Ucred", struct _ucred);
	P("SysvIpcPerm", struct ipc_perm);
	P("SysvShmDesc", struct shmid_ds);
	return 0;
}
