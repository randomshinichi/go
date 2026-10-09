#define __DARWIN_UNIX03 0
#define KERNEL
#define _DARWIN_USE_64_BIT_INODE
#include <stdio.h>
#include <stddef.h>
#include <dirent.h>
#include <fcntl.h>
#include <signal.h>
#include <termios.h>
#include <unistd.h>
#include <sys/event.h>
#include <sys/mman.h>
#include <sys/mount.h>
#include <sys/param.h>
#include <sys/ptrace.h>
#include <sys/resource.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/uio.h>
#include <sys/un.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/if_var.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/icmp6.h>
#define S(T) printf("sizeof(%s)=%lu\n", #T, (unsigned long)sizeof(struct T))
#define O(T,F) printf("offsetof(%s.%s)=%lu\n", #T, #F, (unsigned long)offsetof(struct T,F))
int main(void) {
S(timespec); S(timeval); S(rusage); S(rlimit); S(stat64); S(statfs64); S(flock); S(fstore); S(radvisory); S(fbootstraptransfer); S(log2phys); S(fsid); S(dirent); S(sockaddr_in); S(sockaddr_in6); S(sockaddr_un); S(sockaddr_dl); S(sockaddr); S(msghdr); S(cmsghdr); S(iovec); S(linger); S(ip_mreq); S(ipv6_mreq); S(kevent); S(if_msghdr); S(if_data); S(ifa_msghdr); S(ifma_msghdr); S(ifma_msghdr2); S(rt_msghdr); S(rt_metrics); S(bpf_version); S(bpf_stat); S(bpf_program); S(bpf_insn); S(bpf_hdr); S(fd_set); S(termios); S(in6_pktinfo); S(icmp6_filter);
O(stat64,st_dev); O(stat64,st_mode); O(stat64,st_nlink); O(stat64,st_ino); O(stat64,st_uid); O(stat64,st_gid); O(stat64,st_rdev); O(stat64,st_atimespec); O(stat64,st_mtimespec); O(stat64,st_ctimespec); O(stat64,st_birthtimespec); O(stat64,st_size); O(stat64,st_blocks); O(stat64,st_blksize); O(stat64,st_flags); O(stat64,st_gen); O(stat64,st_lspare); O(stat64,st_qspare);
O(statfs64,f_bsize); O(statfs64,f_iosize); O(statfs64,f_blocks); O(statfs64,f_fsid); O(statfs64,f_owner); O(statfs64,f_type); O(statfs64,f_flags); O(statfs64,f_fstypename); O(statfs64,f_mntonname); O(statfs64,f_mntfromname); O(statfs64,f_reserved);
O(flock,l_start); O(flock,l_len); O(flock,l_pid); O(flock,l_type); O(flock,l_whence);
O(fstore,fst_flags); O(fstore,fst_posmode); O(fstore,fst_offset); O(fstore,fst_length); O(fstore,fst_bytesalloc);
O(radvisory,ra_offset); O(radvisory,ra_count); O(fbootstraptransfer,fbt_offset); O(fbootstraptransfer,fbt_length); O(fbootstraptransfer,fbt_buffer);
O(log2phys,l2p_flags); O(log2phys,l2p_contigbytes); O(log2phys,l2p_devoffset); O(dirent,d_ino); O(dirent,d_seekoff); O(dirent,d_reclen); O(dirent,d_namlen); O(dirent,d_type); O(dirent,d_name);
O(termios,c_iflag); O(termios,c_oflag); O(termios,c_cflag); O(termios,c_lflag); O(termios,c_cc); O(termios,c_ispeed); O(termios,c_ospeed);
return 0;
}
