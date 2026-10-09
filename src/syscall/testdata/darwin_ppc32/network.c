#include <stdio.h>
#include <stddef.h>
#include <sys/types.h>
#include <sys/time.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/uio.h>
#include <sys/resource.h>
#include <sys/select.h>
#include <sys/event.h>
#include <sys/stat.h>
#include <sys/mount.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/if_var.h>
#include <net/route.h>
#include <net/bpf.h>
#include <netinet/in.h>
#include <netinet/icmp6.h>
#define S(T) printf("sizeof(%s)=%lu\n", #T, (unsigned long)sizeof(struct T))
#define O(T,F) printf("offsetof(%s.%s)=%lu\n", #T, #F, (unsigned long)offsetof(struct T,F))
int main(void) {
 printf("sizeof(void*)=%lu sizeof(short)=%lu sizeof(int)=%lu sizeof(long)=%lu sizeof(long long)=%lu IFNAMSIZ=%d\n", (unsigned long)sizeof(void*),(unsigned long)sizeof(short),(unsigned long)sizeof(int),(unsigned long)sizeof(long),(unsigned long)sizeof(long long),IFNAMSIZ);
 S(sockaddr); S(sockaddr_in); S(sockaddr_in6); S(sockaddr_un); S(sockaddr_dl); S(msghdr); S(cmsghdr); S(iovec); S(linger); S(ip_mreq); S(ipv6_mreq); S(kevent); S(if_msghdr); S(if_data); S(ifa_msghdr); S(ifma_msghdr); S(ifma_msghdr2); S(rt_msghdr); S(rt_metrics); S(bpf_version); S(bpf_stat); S(bpf_program); S(bpf_insn); S(bpf_hdr); S(timeval); S(fd_set); S(ifreq);
 O(sockaddr_in,sin_len); O(sockaddr_in,sin_family); O(sockaddr_in,sin_port); O(sockaddr_in,sin_addr);
 O(sockaddr_in6,sin6_len); O(sockaddr_in6,sin6_family); O(sockaddr_in6,sin6_port); O(sockaddr_in6,sin6_flowinfo); O(sockaddr_in6,sin6_addr); O(sockaddr_in6,sin6_scope_id);
 O(sockaddr_dl,sdl_len); O(sockaddr_dl,sdl_family); O(sockaddr_dl,sdl_index); O(sockaddr_dl,sdl_type); O(sockaddr_dl,sdl_nlen); O(sockaddr_dl,sdl_alen); O(sockaddr_dl,sdl_slen); O(sockaddr_dl,sdl_data);
 O(msghdr,msg_name); O(msghdr,msg_namelen); O(msghdr,msg_iov); O(msghdr,msg_iovlen); O(msghdr,msg_control); O(msghdr,msg_controllen); O(msghdr,msg_flags);
 O(cmsghdr,cmsg_len); O(cmsghdr,cmsg_level); O(cmsghdr,cmsg_type);
 O(if_msghdr,ifm_msglen); O(if_msghdr,ifm_version); O(if_msghdr,ifm_type); O(if_msghdr,ifm_addrs); O(if_msghdr,ifm_flags); O(if_msghdr,ifm_index); O(if_msghdr,ifm_data);
 O(if_data,ifi_type); O(if_data,ifi_mtu); O(if_data,ifi_lastchange);
 O(ifa_msghdr,ifam_msglen); O(ifa_msghdr,ifam_addrs); O(ifa_msghdr,ifam_flags); O(ifa_msghdr,ifam_index); O(ifa_msghdr,ifam_metric);
 O(ifma_msghdr2,ifmam_msglen); O(ifma_msghdr2,ifmam_addrs); O(ifma_msghdr2,ifmam_flags); O(ifma_msghdr2,ifmam_index); O(ifma_msghdr2,ifmam_refcount);
 O(rt_msghdr,rtm_msglen); O(rt_msghdr,rtm_version); O(rt_msghdr,rtm_type); O(rt_msghdr,rtm_index); O(rt_msghdr,rtm_flags); O(rt_msghdr,rtm_addrs); O(rt_msghdr,rtm_pid); O(rt_msghdr,rtm_seq); O(rt_msghdr,rtm_errno); O(rt_msghdr,rtm_use); O(rt_msghdr,rtm_inits); O(rt_msghdr,rtm_rmx);
 O(rt_metrics,rmx_mtu); O(rt_metrics,rmx_expire);
 O(bpf_program,bf_len); O(bpf_program,bf_insns);
 return 0;
}
