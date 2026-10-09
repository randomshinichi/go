#define _DARWIN_C_SOURCE
#define KERNEL
#define _DARWIN_USE_64_BIT_INODE
#include <stdio.h>
#include <sys/types.h>
#include <sys/event.h>
#include <sys/ptrace.h>
#include <sys/socket.h>
#include <sys/sockio.h>
#include <sys/sysctl.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_types.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <netinet/ip_mroute.h>
#include <termios.h>
#include <limits.h>
#include <sys/file.h>
#include <fcntl.h>
#include <dirent.h>
#include <netinet/ip6.h>
#include <netinet/tcp.h>
#include <errno.h>
#include <sys/signal.h>
#include <signal.h>
#include <sys/resource.h>
#include <sys/ioctl.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/event.h>
#include <sys/ptrace.h>
#include <sys/socket.h>
#include <sys/sockio.h>
#include <sys/sysctl.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <net/bpf.h>
#include <net/if.h>
#include <net/if_types.h>
#include <net/route.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <netinet/ip_mroute.h>
#include <termios.h>
int main(void) {
#ifdef AF_APPLETALK
printf("AF_APPLETALK=%lld\n", (long long)(AF_APPLETALK));
#endif
#ifdef AF_CCITT
printf("AF_CCITT=%lld\n", (long long)(AF_CCITT));
#endif
#ifdef AF_CHAOS
printf("AF_CHAOS=%lld\n", (long long)(AF_CHAOS));
#endif
#ifdef AF_CNT
printf("AF_CNT=%lld\n", (long long)(AF_CNT));
#endif
#ifdef AF_COIP
printf("AF_COIP=%lld\n", (long long)(AF_COIP));
#endif
#ifdef AF_DATAKIT
printf("AF_DATAKIT=%lld\n", (long long)(AF_DATAKIT));
#endif
#ifdef AF_DLI
printf("AF_DLI=%lld\n", (long long)(AF_DLI));
#endif
#ifdef AF_E164
printf("AF_E164=%lld\n", (long long)(AF_E164));
#endif
#ifdef AF_ECMA
printf("AF_ECMA=%lld\n", (long long)(AF_ECMA));
#endif
#ifdef AF_HYLINK
printf("AF_HYLINK=%lld\n", (long long)(AF_HYLINK));
#endif
#ifdef AF_IMPLINK
printf("AF_IMPLINK=%lld\n", (long long)(AF_IMPLINK));
#endif
#ifdef AF_INET
printf("AF_INET=%lld\n", (long long)(AF_INET));
#endif
#ifdef AF_INET6
printf("AF_INET6=%lld\n", (long long)(AF_INET6));
#endif
#ifdef AF_IPX
printf("AF_IPX=%lld\n", (long long)(AF_IPX));
#endif
#ifdef AF_ISDN
printf("AF_ISDN=%lld\n", (long long)(AF_ISDN));
#endif
#ifdef AF_ISO
printf("AF_ISO=%lld\n", (long long)(AF_ISO));
#endif
#ifdef AF_LAT
printf("AF_LAT=%lld\n", (long long)(AF_LAT));
#endif
#ifdef AF_LINK
printf("AF_LINK=%lld\n", (long long)(AF_LINK));
#endif
#ifdef AF_LOCAL
printf("AF_LOCAL=%lld\n", (long long)(AF_LOCAL));
#endif
#ifdef AF_MAX
printf("AF_MAX=%lld\n", (long long)(AF_MAX));
#endif
#ifdef AF_NATM
printf("AF_NATM=%lld\n", (long long)(AF_NATM));
#endif
#ifdef AF_NDRV
printf("AF_NDRV=%lld\n", (long long)(AF_NDRV));
#endif
#ifdef AF_NETBIOS
printf("AF_NETBIOS=%lld\n", (long long)(AF_NETBIOS));
#endif
#ifdef AF_NS
printf("AF_NS=%lld\n", (long long)(AF_NS));
#endif
#ifdef AF_OSI
printf("AF_OSI=%lld\n", (long long)(AF_OSI));
#endif
#ifdef AF_PPP
printf("AF_PPP=%lld\n", (long long)(AF_PPP));
#endif
#ifdef AF_PUP
printf("AF_PUP=%lld\n", (long long)(AF_PUP));
#endif
#ifdef AF_RESERVED_36
printf("AF_RESERVED_36=%lld\n", (long long)(AF_RESERVED_36));
#endif
#ifdef AF_ROUTE
printf("AF_ROUTE=%lld\n", (long long)(AF_ROUTE));
#endif
#ifdef AF_SIP
printf("AF_SIP=%lld\n", (long long)(AF_SIP));
#endif
#ifdef AF_SNA
printf("AF_SNA=%lld\n", (long long)(AF_SNA));
#endif
#ifdef AF_SYSTEM
printf("AF_SYSTEM=%lld\n", (long long)(AF_SYSTEM));
#endif
#ifdef AF_UNIX
printf("AF_UNIX=%lld\n", (long long)(AF_UNIX));
#endif
#ifdef AF_UNSPEC
printf("AF_UNSPEC=%lld\n", (long long)(AF_UNSPEC));
#endif
#ifdef B0
printf("B0=%lld\n", (long long)(B0));
#endif
#ifdef B110
printf("B110=%lld\n", (long long)(B110));
#endif
#ifdef B115200
printf("B115200=%lld\n", (long long)(B115200));
#endif
#ifdef B1200
printf("B1200=%lld\n", (long long)(B1200));
#endif
#ifdef B134
printf("B134=%lld\n", (long long)(B134));
#endif
#ifdef B14400
printf("B14400=%lld\n", (long long)(B14400));
#endif
#ifdef B150
printf("B150=%lld\n", (long long)(B150));
#endif
#ifdef B1800
printf("B1800=%lld\n", (long long)(B1800));
#endif
#ifdef B19200
printf("B19200=%lld\n", (long long)(B19200));
#endif
#ifdef B200
printf("B200=%lld\n", (long long)(B200));
#endif
#ifdef B230400
printf("B230400=%lld\n", (long long)(B230400));
#endif
#ifdef B2400
printf("B2400=%lld\n", (long long)(B2400));
#endif
#ifdef B28800
printf("B28800=%lld\n", (long long)(B28800));
#endif
#ifdef B300
printf("B300=%lld\n", (long long)(B300));
#endif
#ifdef B38400
printf("B38400=%lld\n", (long long)(B38400));
#endif
#ifdef B4800
printf("B4800=%lld\n", (long long)(B4800));
#endif
#ifdef B50
printf("B50=%lld\n", (long long)(B50));
#endif
#ifdef B57600
printf("B57600=%lld\n", (long long)(B57600));
#endif
#ifdef B600
printf("B600=%lld\n", (long long)(B600));
#endif
#ifdef B7200
printf("B7200=%lld\n", (long long)(B7200));
#endif
#ifdef B75
printf("B75=%lld\n", (long long)(B75));
#endif
#ifdef B76800
printf("B76800=%lld\n", (long long)(B76800));
#endif
#ifdef B9600
printf("B9600=%lld\n", (long long)(B9600));
#endif
#ifdef BIOCFLUSH
printf("BIOCFLUSH=%lld\n", (long long)(BIOCFLUSH));
#endif
#ifdef BIOCGBLEN
printf("BIOCGBLEN=%lld\n", (long long)(BIOCGBLEN));
#endif
#ifdef BIOCGDLT
printf("BIOCGDLT=%lld\n", (long long)(BIOCGDLT));
#endif
#ifdef BIOCGDLTLIST
printf("BIOCGDLTLIST=%lld\n", (long long)(BIOCGDLTLIST));
#endif
#ifdef BIOCGETIF
printf("BIOCGETIF=%lld\n", (long long)(BIOCGETIF));
#endif
#ifdef BIOCGHDRCMPLT
printf("BIOCGHDRCMPLT=%lld\n", (long long)(BIOCGHDRCMPLT));
#endif
#ifdef BIOCGRSIG
printf("BIOCGRSIG=%lld\n", (long long)(BIOCGRSIG));
#endif
#ifdef BIOCGRTIMEOUT
printf("BIOCGRTIMEOUT=%lld\n", (long long)(BIOCGRTIMEOUT));
#endif
#ifdef BIOCGSEESENT
printf("BIOCGSEESENT=%lld\n", (long long)(BIOCGSEESENT));
#endif
#ifdef BIOCGSTATS
printf("BIOCGSTATS=%lld\n", (long long)(BIOCGSTATS));
#endif
#ifdef BIOCIMMEDIATE
printf("BIOCIMMEDIATE=%lld\n", (long long)(BIOCIMMEDIATE));
#endif
#ifdef BIOCPROMISC
printf("BIOCPROMISC=%lld\n", (long long)(BIOCPROMISC));
#endif
#ifdef BIOCSBLEN
printf("BIOCSBLEN=%lld\n", (long long)(BIOCSBLEN));
#endif
#ifdef BIOCSDLT
printf("BIOCSDLT=%lld\n", (long long)(BIOCSDLT));
#endif
#ifdef BIOCSETF
printf("BIOCSETF=%lld\n", (long long)(BIOCSETF));
#endif
#ifdef BIOCSETIF
printf("BIOCSETIF=%lld\n", (long long)(BIOCSETIF));
#endif
#ifdef BIOCSHDRCMPLT
printf("BIOCSHDRCMPLT=%lld\n", (long long)(BIOCSHDRCMPLT));
#endif
#ifdef BIOCSRSIG
printf("BIOCSRSIG=%lld\n", (long long)(BIOCSRSIG));
#endif
#ifdef BIOCSRTIMEOUT
printf("BIOCSRTIMEOUT=%lld\n", (long long)(BIOCSRTIMEOUT));
#endif
#ifdef BIOCSSEESENT
printf("BIOCSSEESENT=%lld\n", (long long)(BIOCSSEESENT));
#endif
#ifdef BIOCVERSION
printf("BIOCVERSION=%lld\n", (long long)(BIOCVERSION));
#endif
#ifdef BPF_A
printf("BPF_A=%lld\n", (long long)(BPF_A));
#endif
#ifdef BPF_ABS
printf("BPF_ABS=%lld\n", (long long)(BPF_ABS));
#endif
#ifdef BPF_ADD
printf("BPF_ADD=%lld\n", (long long)(BPF_ADD));
#endif
#ifdef BPF_ALIGNMENT
printf("BPF_ALIGNMENT=%lld\n", (long long)(BPF_ALIGNMENT));
#endif
#ifdef BPF_ALU
printf("BPF_ALU=%lld\n", (long long)(BPF_ALU));
#endif
#ifdef BPF_AND
printf("BPF_AND=%lld\n", (long long)(BPF_AND));
#endif
#ifdef BPF_B
printf("BPF_B=%lld\n", (long long)(BPF_B));
#endif
#ifdef BPF_DIV
printf("BPF_DIV=%lld\n", (long long)(BPF_DIV));
#endif
#ifdef BPF_H
printf("BPF_H=%lld\n", (long long)(BPF_H));
#endif
#ifdef BPF_IMM
printf("BPF_IMM=%lld\n", (long long)(BPF_IMM));
#endif
#ifdef BPF_IND
printf("BPF_IND=%lld\n", (long long)(BPF_IND));
#endif
#ifdef BPF_JA
printf("BPF_JA=%lld\n", (long long)(BPF_JA));
#endif
#ifdef BPF_JEQ
printf("BPF_JEQ=%lld\n", (long long)(BPF_JEQ));
#endif
#ifdef BPF_JGE
printf("BPF_JGE=%lld\n", (long long)(BPF_JGE));
#endif
#ifdef BPF_JGT
printf("BPF_JGT=%lld\n", (long long)(BPF_JGT));
#endif
#ifdef BPF_JMP
printf("BPF_JMP=%lld\n", (long long)(BPF_JMP));
#endif
#ifdef BPF_JSET
printf("BPF_JSET=%lld\n", (long long)(BPF_JSET));
#endif
#ifdef BPF_K
printf("BPF_K=%lld\n", (long long)(BPF_K));
#endif
#ifdef BPF_LD
printf("BPF_LD=%lld\n", (long long)(BPF_LD));
#endif
#ifdef BPF_LDX
printf("BPF_LDX=%lld\n", (long long)(BPF_LDX));
#endif
#ifdef BPF_LEN
printf("BPF_LEN=%lld\n", (long long)(BPF_LEN));
#endif
#ifdef BPF_LSH
printf("BPF_LSH=%lld\n", (long long)(BPF_LSH));
#endif
#ifdef BPF_MAJOR_VERSION
printf("BPF_MAJOR_VERSION=%lld\n", (long long)(BPF_MAJOR_VERSION));
#endif
#ifdef BPF_MAXBUFSIZE
printf("BPF_MAXBUFSIZE=%lld\n", (long long)(BPF_MAXBUFSIZE));
#endif
#ifdef BPF_MAXINSNS
printf("BPF_MAXINSNS=%lld\n", (long long)(BPF_MAXINSNS));
#endif
#ifdef BPF_MEM
printf("BPF_MEM=%lld\n", (long long)(BPF_MEM));
#endif
#ifdef BPF_MEMWORDS
printf("BPF_MEMWORDS=%lld\n", (long long)(BPF_MEMWORDS));
#endif
#ifdef BPF_MINBUFSIZE
printf("BPF_MINBUFSIZE=%lld\n", (long long)(BPF_MINBUFSIZE));
#endif
#ifdef BPF_MINOR_VERSION
printf("BPF_MINOR_VERSION=%lld\n", (long long)(BPF_MINOR_VERSION));
#endif
#ifdef BPF_MISC
printf("BPF_MISC=%lld\n", (long long)(BPF_MISC));
#endif
#ifdef BPF_MSH
printf("BPF_MSH=%lld\n", (long long)(BPF_MSH));
#endif
#ifdef BPF_MUL
printf("BPF_MUL=%lld\n", (long long)(BPF_MUL));
#endif
#ifdef BPF_NEG
printf("BPF_NEG=%lld\n", (long long)(BPF_NEG));
#endif
#ifdef BPF_OR
printf("BPF_OR=%lld\n", (long long)(BPF_OR));
#endif
#ifdef BPF_RELEASE
printf("BPF_RELEASE=%lld\n", (long long)(BPF_RELEASE));
#endif
#ifdef BPF_RET
printf("BPF_RET=%lld\n", (long long)(BPF_RET));
#endif
#ifdef BPF_RSH
printf("BPF_RSH=%lld\n", (long long)(BPF_RSH));
#endif
#ifdef BPF_ST
printf("BPF_ST=%lld\n", (long long)(BPF_ST));
#endif
#ifdef BPF_STX
printf("BPF_STX=%lld\n", (long long)(BPF_STX));
#endif
#ifdef BPF_SUB
printf("BPF_SUB=%lld\n", (long long)(BPF_SUB));
#endif
#ifdef BPF_TAX
printf("BPF_TAX=%lld\n", (long long)(BPF_TAX));
#endif
#ifdef BPF_TXA
printf("BPF_TXA=%lld\n", (long long)(BPF_TXA));
#endif
#ifdef BPF_W
printf("BPF_W=%lld\n", (long long)(BPF_W));
#endif
#ifdef BPF_X
printf("BPF_X=%lld\n", (long long)(BPF_X));
#endif
#ifdef BRKINT
printf("BRKINT=%lld\n", (long long)(BRKINT));
#endif
#ifdef CFLUSH
printf("CFLUSH=%lld\n", (long long)(CFLUSH));
#endif
#ifdef CLOCAL
printf("CLOCAL=%lld\n", (long long)(CLOCAL));
#endif
#ifdef CREAD
printf("CREAD=%lld\n", (long long)(CREAD));
#endif
#ifdef CS5
printf("CS5=%lld\n", (long long)(CS5));
#endif
#ifdef CS6
printf("CS6=%lld\n", (long long)(CS6));
#endif
#ifdef CS7
printf("CS7=%lld\n", (long long)(CS7));
#endif
#ifdef CS8
printf("CS8=%lld\n", (long long)(CS8));
#endif
#ifdef CSIZE
printf("CSIZE=%lld\n", (long long)(CSIZE));
#endif
#ifdef CSTART
printf("CSTART=%lld\n", (long long)(CSTART));
#endif
#ifdef CSTATUS
printf("CSTATUS=%lld\n", (long long)(CSTATUS));
#endif
#ifdef CSTOP
printf("CSTOP=%lld\n", (long long)(CSTOP));
#endif
#ifdef CSTOPB
printf("CSTOPB=%lld\n", (long long)(CSTOPB));
#endif
#ifdef CSUSP
printf("CSUSP=%lld\n", (long long)(CSUSP));
#endif
#ifdef CTL_MAXNAME
printf("CTL_MAXNAME=%lld\n", (long long)(CTL_MAXNAME));
#endif
#ifdef CTL_NET
printf("CTL_NET=%lld\n", (long long)(CTL_NET));
#endif
#ifdef DLT_APPLE_IP_OVER_IEEE1394
printf("DLT_APPLE_IP_OVER_IEEE1394=%lld\n", (long long)(DLT_APPLE_IP_OVER_IEEE1394));
#endif
#ifdef DLT_ARCNET
printf("DLT_ARCNET=%lld\n", (long long)(DLT_ARCNET));
#endif
#ifdef DLT_ATM_CLIP
printf("DLT_ATM_CLIP=%lld\n", (long long)(DLT_ATM_CLIP));
#endif
#ifdef DLT_ATM_RFC1483
printf("DLT_ATM_RFC1483=%lld\n", (long long)(DLT_ATM_RFC1483));
#endif
#ifdef DLT_AX25
printf("DLT_AX25=%lld\n", (long long)(DLT_AX25));
#endif
#ifdef DLT_CHAOS
printf("DLT_CHAOS=%lld\n", (long long)(DLT_CHAOS));
#endif
#ifdef DLT_CHDLC
printf("DLT_CHDLC=%lld\n", (long long)(DLT_CHDLC));
#endif
#ifdef DLT_C_HDLC
printf("DLT_C_HDLC=%lld\n", (long long)(DLT_C_HDLC));
#endif
#ifdef DLT_EN10MB
printf("DLT_EN10MB=%lld\n", (long long)(DLT_EN10MB));
#endif
#ifdef DLT_EN3MB
printf("DLT_EN3MB=%lld\n", (long long)(DLT_EN3MB));
#endif
#ifdef DLT_FDDI
printf("DLT_FDDI=%lld\n", (long long)(DLT_FDDI));
#endif
#ifdef DLT_IEEE802
printf("DLT_IEEE802=%lld\n", (long long)(DLT_IEEE802));
#endif
#ifdef DLT_IEEE802_11
printf("DLT_IEEE802_11=%lld\n", (long long)(DLT_IEEE802_11));
#endif
#ifdef DLT_LINUX_SLL
printf("DLT_LINUX_SLL=%lld\n", (long long)(DLT_LINUX_SLL));
#endif
#ifdef DLT_LOOP
printf("DLT_LOOP=%lld\n", (long long)(DLT_LOOP));
#endif
#ifdef DLT_NULL
printf("DLT_NULL=%lld\n", (long long)(DLT_NULL));
#endif
#ifdef DLT_PPP
printf("DLT_PPP=%lld\n", (long long)(DLT_PPP));
#endif
#ifdef DLT_PPP_BSDOS
printf("DLT_PPP_BSDOS=%lld\n", (long long)(DLT_PPP_BSDOS));
#endif
#ifdef DLT_PPP_SERIAL
printf("DLT_PPP_SERIAL=%lld\n", (long long)(DLT_PPP_SERIAL));
#endif
#ifdef DLT_PRONET
printf("DLT_PRONET=%lld\n", (long long)(DLT_PRONET));
#endif
#ifdef DLT_RAW
printf("DLT_RAW=%lld\n", (long long)(DLT_RAW));
#endif
#ifdef DLT_SLIP
printf("DLT_SLIP=%lld\n", (long long)(DLT_SLIP));
#endif
#ifdef DLT_SLIP_BSDOS
printf("DLT_SLIP_BSDOS=%lld\n", (long long)(DLT_SLIP_BSDOS));
#endif
#ifdef DT_BLK
printf("DT_BLK=%lld\n", (long long)(DT_BLK));
#endif
#ifdef DT_CHR
printf("DT_CHR=%lld\n", (long long)(DT_CHR));
#endif
#ifdef DT_DIR
printf("DT_DIR=%lld\n", (long long)(DT_DIR));
#endif
#ifdef DT_FIFO
printf("DT_FIFO=%lld\n", (long long)(DT_FIFO));
#endif
#ifdef DT_LNK
printf("DT_LNK=%lld\n", (long long)(DT_LNK));
#endif
#ifdef DT_REG
printf("DT_REG=%lld\n", (long long)(DT_REG));
#endif
#ifdef DT_SOCK
printf("DT_SOCK=%lld\n", (long long)(DT_SOCK));
#endif
#ifdef DT_UNKNOWN
printf("DT_UNKNOWN=%lld\n", (long long)(DT_UNKNOWN));
#endif
#ifdef DT_WHT
printf("DT_WHT=%lld\n", (long long)(DT_WHT));
#endif
#ifdef E2BIG
printf("E2BIG=%lld\n", (long long)(E2BIG));
#endif
#ifdef EACCES
printf("EACCES=%lld\n", (long long)(EACCES));
#endif
#ifdef EADDRINUSE
printf("EADDRINUSE=%lld\n", (long long)(EADDRINUSE));
#endif
#ifdef EADDRNOTAVAIL
printf("EADDRNOTAVAIL=%lld\n", (long long)(EADDRNOTAVAIL));
#endif
#ifdef EAFNOSUPPORT
printf("EAFNOSUPPORT=%lld\n", (long long)(EAFNOSUPPORT));
#endif
#ifdef EAGAIN
printf("EAGAIN=%lld\n", (long long)(EAGAIN));
#endif
#ifdef EALREADY
printf("EALREADY=%lld\n", (long long)(EALREADY));
#endif
#ifdef EAUTH
printf("EAUTH=%lld\n", (long long)(EAUTH));
#endif
#ifdef EBADARCH
printf("EBADARCH=%lld\n", (long long)(EBADARCH));
#endif
#ifdef EBADEXEC
printf("EBADEXEC=%lld\n", (long long)(EBADEXEC));
#endif
#ifdef EBADF
printf("EBADF=%lld\n", (long long)(EBADF));
#endif
#ifdef EBADMACHO
printf("EBADMACHO=%lld\n", (long long)(EBADMACHO));
#endif
#ifdef EBADMSG
printf("EBADMSG=%lld\n", (long long)(EBADMSG));
#endif
#ifdef EBADRPC
printf("EBADRPC=%lld\n", (long long)(EBADRPC));
#endif
#ifdef EBUSY
printf("EBUSY=%lld\n", (long long)(EBUSY));
#endif
#ifdef ECANCELED
printf("ECANCELED=%lld\n", (long long)(ECANCELED));
#endif
#ifdef ECHILD
printf("ECHILD=%lld\n", (long long)(ECHILD));
#endif
#ifdef ECHO
printf("ECHO=%lld\n", (long long)(ECHO));
#endif
#ifdef ECHOCTL
printf("ECHOCTL=%lld\n", (long long)(ECHOCTL));
#endif
#ifdef ECHOE
printf("ECHOE=%lld\n", (long long)(ECHOE));
#endif
#ifdef ECHOK
printf("ECHOK=%lld\n", (long long)(ECHOK));
#endif
#ifdef ECHOKE
printf("ECHOKE=%lld\n", (long long)(ECHOKE));
#endif
#ifdef ECHONL
printf("ECHONL=%lld\n", (long long)(ECHONL));
#endif
#ifdef ECHOPRT
printf("ECHOPRT=%lld\n", (long long)(ECHOPRT));
#endif
#ifdef ECONNABORTED
printf("ECONNABORTED=%lld\n", (long long)(ECONNABORTED));
#endif
#ifdef ECONNREFUSED
printf("ECONNREFUSED=%lld\n", (long long)(ECONNREFUSED));
#endif
#ifdef ECONNRESET
printf("ECONNRESET=%lld\n", (long long)(ECONNRESET));
#endif
#ifdef EDEADLK
printf("EDEADLK=%lld\n", (long long)(EDEADLK));
#endif
#ifdef EDESTADDRREQ
printf("EDESTADDRREQ=%lld\n", (long long)(EDESTADDRREQ));
#endif
#ifdef EDEVERR
printf("EDEVERR=%lld\n", (long long)(EDEVERR));
#endif
#ifdef EDOM
printf("EDOM=%lld\n", (long long)(EDOM));
#endif
#ifdef EDQUOT
printf("EDQUOT=%lld\n", (long long)(EDQUOT));
#endif
#ifdef EEXIST
printf("EEXIST=%lld\n", (long long)(EEXIST));
#endif
#ifdef EFAULT
printf("EFAULT=%lld\n", (long long)(EFAULT));
#endif
#ifdef EFBIG
printf("EFBIG=%lld\n", (long long)(EFBIG));
#endif
#ifdef EFTYPE
printf("EFTYPE=%lld\n", (long long)(EFTYPE));
#endif
#ifdef EHOSTDOWN
printf("EHOSTDOWN=%lld\n", (long long)(EHOSTDOWN));
#endif
#ifdef EHOSTUNREACH
printf("EHOSTUNREACH=%lld\n", (long long)(EHOSTUNREACH));
#endif
#ifdef EIDRM
printf("EIDRM=%lld\n", (long long)(EIDRM));
#endif
#ifdef EILSEQ
printf("EILSEQ=%lld\n", (long long)(EILSEQ));
#endif
#ifdef EINPROGRESS
printf("EINPROGRESS=%lld\n", (long long)(EINPROGRESS));
#endif
#ifdef EINTR
printf("EINTR=%lld\n", (long long)(EINTR));
#endif
#ifdef EINVAL
printf("EINVAL=%lld\n", (long long)(EINVAL));
#endif
#ifdef EIO
printf("EIO=%lld\n", (long long)(EIO));
#endif
#ifdef EISCONN
printf("EISCONN=%lld\n", (long long)(EISCONN));
#endif
#ifdef EISDIR
printf("EISDIR=%lld\n", (long long)(EISDIR));
#endif
#ifdef ELAST
printf("ELAST=%lld\n", (long long)(ELAST));
#endif
#ifdef ELOOP
printf("ELOOP=%lld\n", (long long)(ELOOP));
#endif
#ifdef EMFILE
printf("EMFILE=%lld\n", (long long)(EMFILE));
#endif
#ifdef EMLINK
printf("EMLINK=%lld\n", (long long)(EMLINK));
#endif
#ifdef EMSGSIZE
printf("EMSGSIZE=%lld\n", (long long)(EMSGSIZE));
#endif
#ifdef EMULTIHOP
printf("EMULTIHOP=%lld\n", (long long)(EMULTIHOP));
#endif
#ifdef ENAMETOOLONG
printf("ENAMETOOLONG=%lld\n", (long long)(ENAMETOOLONG));
#endif
#ifdef ENEEDAUTH
printf("ENEEDAUTH=%lld\n", (long long)(ENEEDAUTH));
#endif
#ifdef ENETDOWN
printf("ENETDOWN=%lld\n", (long long)(ENETDOWN));
#endif
#ifdef ENETRESET
printf("ENETRESET=%lld\n", (long long)(ENETRESET));
#endif
#ifdef ENETUNREACH
printf("ENETUNREACH=%lld\n", (long long)(ENETUNREACH));
#endif
#ifdef ENFILE
printf("ENFILE=%lld\n", (long long)(ENFILE));
#endif
#ifdef ENOATTR
printf("ENOATTR=%lld\n", (long long)(ENOATTR));
#endif
#ifdef ENOBUFS
printf("ENOBUFS=%lld\n", (long long)(ENOBUFS));
#endif
#ifdef ENODATA
printf("ENODATA=%lld\n", (long long)(ENODATA));
#endif
#ifdef ENODEV
printf("ENODEV=%lld\n", (long long)(ENODEV));
#endif
#ifdef ENOENT
printf("ENOENT=%lld\n", (long long)(ENOENT));
#endif
#ifdef ENOEXEC
printf("ENOEXEC=%lld\n", (long long)(ENOEXEC));
#endif
#ifdef ENOLCK
printf("ENOLCK=%lld\n", (long long)(ENOLCK));
#endif
#ifdef ENOLINK
printf("ENOLINK=%lld\n", (long long)(ENOLINK));
#endif
#ifdef ENOMEM
printf("ENOMEM=%lld\n", (long long)(ENOMEM));
#endif
#ifdef ENOMSG
printf("ENOMSG=%lld\n", (long long)(ENOMSG));
#endif
#ifdef ENOPOLICY
printf("ENOPOLICY=%lld\n", (long long)(ENOPOLICY));
#endif
#ifdef ENOPROTOOPT
printf("ENOPROTOOPT=%lld\n", (long long)(ENOPROTOOPT));
#endif
#ifdef ENOSPC
printf("ENOSPC=%lld\n", (long long)(ENOSPC));
#endif
#ifdef ENOSR
printf("ENOSR=%lld\n", (long long)(ENOSR));
#endif
#ifdef ENOSTR
printf("ENOSTR=%lld\n", (long long)(ENOSTR));
#endif
#ifdef ENOSYS
printf("ENOSYS=%lld\n", (long long)(ENOSYS));
#endif
#ifdef ENOTBLK
printf("ENOTBLK=%lld\n", (long long)(ENOTBLK));
#endif
#ifdef ENOTCONN
printf("ENOTCONN=%lld\n", (long long)(ENOTCONN));
#endif
#ifdef ENOTDIR
printf("ENOTDIR=%lld\n", (long long)(ENOTDIR));
#endif
#ifdef ENOTEMPTY
printf("ENOTEMPTY=%lld\n", (long long)(ENOTEMPTY));
#endif
#ifdef ENOTSOCK
printf("ENOTSOCK=%lld\n", (long long)(ENOTSOCK));
#endif
#ifdef ENOTSUP
printf("ENOTSUP=%lld\n", (long long)(ENOTSUP));
#endif
#ifdef ENOTTY
printf("ENOTTY=%lld\n", (long long)(ENOTTY));
#endif
#ifdef ENXIO
printf("ENXIO=%lld\n", (long long)(ENXIO));
#endif
#ifdef EOPNOTSUPP
printf("EOPNOTSUPP=%lld\n", (long long)(EOPNOTSUPP));
#endif
#ifdef EOVERFLOW
printf("EOVERFLOW=%lld\n", (long long)(EOVERFLOW));
#endif
#ifdef EPERM
printf("EPERM=%lld\n", (long long)(EPERM));
#endif
#ifdef EPFNOSUPPORT
printf("EPFNOSUPPORT=%lld\n", (long long)(EPFNOSUPPORT));
#endif
#ifdef EPIPE
printf("EPIPE=%lld\n", (long long)(EPIPE));
#endif
#ifdef EPROCLIM
printf("EPROCLIM=%lld\n", (long long)(EPROCLIM));
#endif
#ifdef EPROCUNAVAIL
printf("EPROCUNAVAIL=%lld\n", (long long)(EPROCUNAVAIL));
#endif
#ifdef EPROGMISMATCH
printf("EPROGMISMATCH=%lld\n", (long long)(EPROGMISMATCH));
#endif
#ifdef EPROGUNAVAIL
printf("EPROGUNAVAIL=%lld\n", (long long)(EPROGUNAVAIL));
#endif
#ifdef EPROTO
printf("EPROTO=%lld\n", (long long)(EPROTO));
#endif
#ifdef EPROTONOSUPPORT
printf("EPROTONOSUPPORT=%lld\n", (long long)(EPROTONOSUPPORT));
#endif
#ifdef EPROTOTYPE
printf("EPROTOTYPE=%lld\n", (long long)(EPROTOTYPE));
#endif
#ifdef EPWROFF
printf("EPWROFF=%lld\n", (long long)(EPWROFF));
#endif
#ifdef ERANGE
printf("ERANGE=%lld\n", (long long)(ERANGE));
#endif
#ifdef EREMOTE
printf("EREMOTE=%lld\n", (long long)(EREMOTE));
#endif
#ifdef EROFS
printf("EROFS=%lld\n", (long long)(EROFS));
#endif
#ifdef ERPCMISMATCH
printf("ERPCMISMATCH=%lld\n", (long long)(ERPCMISMATCH));
#endif
#ifdef ESHLIBVERS
printf("ESHLIBVERS=%lld\n", (long long)(ESHLIBVERS));
#endif
#ifdef ESHUTDOWN
printf("ESHUTDOWN=%lld\n", (long long)(ESHUTDOWN));
#endif
#ifdef ESOCKTNOSUPPORT
printf("ESOCKTNOSUPPORT=%lld\n", (long long)(ESOCKTNOSUPPORT));
#endif
#ifdef ESPIPE
printf("ESPIPE=%lld\n", (long long)(ESPIPE));
#endif
#ifdef ESRCH
printf("ESRCH=%lld\n", (long long)(ESRCH));
#endif
#ifdef ESTALE
printf("ESTALE=%lld\n", (long long)(ESTALE));
#endif
#ifdef ETIME
printf("ETIME=%lld\n", (long long)(ETIME));
#endif
#ifdef ETIMEDOUT
printf("ETIMEDOUT=%lld\n", (long long)(ETIMEDOUT));
#endif
#ifdef ETOOMANYREFS
printf("ETOOMANYREFS=%lld\n", (long long)(ETOOMANYREFS));
#endif
#ifdef ETXTBSY
printf("ETXTBSY=%lld\n", (long long)(ETXTBSY));
#endif
#ifdef EUSERS
printf("EUSERS=%lld\n", (long long)(EUSERS));
#endif
#ifdef EVFILT_AIO
printf("EVFILT_AIO=%lld\n", (long long)(EVFILT_AIO));
#endif
#ifdef EVFILT_FS
printf("EVFILT_FS=%lld\n", (long long)(EVFILT_FS));
#endif
#ifdef EVFILT_MACHPORT
printf("EVFILT_MACHPORT=%lld\n", (long long)(EVFILT_MACHPORT));
#endif
#ifdef EVFILT_PROC
printf("EVFILT_PROC=%lld\n", (long long)(EVFILT_PROC));
#endif
#ifdef EVFILT_READ
printf("EVFILT_READ=%lld\n", (long long)(EVFILT_READ));
#endif
#ifdef EVFILT_SIGNAL
printf("EVFILT_SIGNAL=%lld\n", (long long)(EVFILT_SIGNAL));
#endif
#ifdef EVFILT_SYSCOUNT
printf("EVFILT_SYSCOUNT=%lld\n", (long long)(EVFILT_SYSCOUNT));
#endif
#ifdef EVFILT_THREADMARKER
printf("EVFILT_THREADMARKER=%lld\n", (long long)(EVFILT_THREADMARKER));
#endif
#ifdef EVFILT_TIMER
printf("EVFILT_TIMER=%lld\n", (long long)(EVFILT_TIMER));
#endif
#ifdef EVFILT_VNODE
printf("EVFILT_VNODE=%lld\n", (long long)(EVFILT_VNODE));
#endif
#ifdef EVFILT_WRITE
printf("EVFILT_WRITE=%lld\n", (long long)(EVFILT_WRITE));
#endif
#ifdef EV_ADD
printf("EV_ADD=%lld\n", (long long)(EV_ADD));
#endif
#ifdef EV_CLEAR
printf("EV_CLEAR=%lld\n", (long long)(EV_CLEAR));
#endif
#ifdef EV_DELETE
printf("EV_DELETE=%lld\n", (long long)(EV_DELETE));
#endif
#ifdef EV_DISABLE
printf("EV_DISABLE=%lld\n", (long long)(EV_DISABLE));
#endif
#ifdef EV_ENABLE
printf("EV_ENABLE=%lld\n", (long long)(EV_ENABLE));
#endif
#ifdef EV_EOF
printf("EV_EOF=%lld\n", (long long)(EV_EOF));
#endif
#ifdef EV_ERROR
printf("EV_ERROR=%lld\n", (long long)(EV_ERROR));
#endif
#ifdef EV_FLAG0
printf("EV_FLAG0=%lld\n", (long long)(EV_FLAG0));
#endif
#ifdef EV_FLAG1
printf("EV_FLAG1=%lld\n", (long long)(EV_FLAG1));
#endif
#ifdef EV_ONESHOT
printf("EV_ONESHOT=%lld\n", (long long)(EV_ONESHOT));
#endif
#ifdef EV_OOBAND
printf("EV_OOBAND=%lld\n", (long long)(EV_OOBAND));
#endif
#ifdef EV_POLL
printf("EV_POLL=%lld\n", (long long)(EV_POLL));
#endif
#ifdef EV_RECEIPT
printf("EV_RECEIPT=%lld\n", (long long)(EV_RECEIPT));
#endif
#ifdef EV_SYSFLAGS
printf("EV_SYSFLAGS=%lld\n", (long long)(EV_SYSFLAGS));
#endif
#ifdef EWOULDBLOCK
printf("EWOULDBLOCK=%lld\n", (long long)(EWOULDBLOCK));
#endif
#ifdef EXDEV
printf("EXDEV=%lld\n", (long long)(EXDEV));
#endif
#ifdef EXTA
printf("EXTA=%lld\n", (long long)(EXTA));
#endif
#ifdef EXTB
printf("EXTB=%lld\n", (long long)(EXTB));
#endif
#ifdef EXTPROC
printf("EXTPROC=%lld\n", (long long)(EXTPROC));
#endif
#ifdef FD_CLOEXEC
printf("FD_CLOEXEC=%lld\n", (long long)(FD_CLOEXEC));
#endif
#ifdef FD_SETSIZE
printf("FD_SETSIZE=%lld\n", (long long)(FD_SETSIZE));
#endif
#ifdef FLUSHO
printf("FLUSHO=%lld\n", (long long)(FLUSHO));
#endif
#ifdef F_ADDSIGS
printf("F_ADDSIGS=%lld\n", (long long)(F_ADDSIGS));
#endif
#ifdef F_ALLOCATEALL
printf("F_ALLOCATEALL=%lld\n", (long long)(F_ALLOCATEALL));
#endif
#ifdef F_ALLOCATECONTIG
printf("F_ALLOCATECONTIG=%lld\n", (long long)(F_ALLOCATECONTIG));
#endif
#ifdef F_CHKCLEAN
printf("F_CHKCLEAN=%lld\n", (long long)(F_CHKCLEAN));
#endif
#ifdef F_DUPFD
printf("F_DUPFD=%lld\n", (long long)(F_DUPFD));
#endif
#ifdef F_FREEZE_FS
printf("F_FREEZE_FS=%lld\n", (long long)(F_FREEZE_FS));
#endif
#ifdef F_FULLFSYNC
printf("F_FULLFSYNC=%lld\n", (long long)(F_FULLFSYNC));
#endif
#ifdef F_GETFD
printf("F_GETFD=%lld\n", (long long)(F_GETFD));
#endif
#ifdef F_GETFL
printf("F_GETFL=%lld\n", (long long)(F_GETFL));
#endif
#ifdef F_GETLK
printf("F_GETLK=%lld\n", (long long)(F_GETLK));
#endif
#ifdef F_GETOWN
printf("F_GETOWN=%lld\n", (long long)(F_GETOWN));
#endif
#ifdef F_GETPATH
printf("F_GETPATH=%lld\n", (long long)(F_GETPATH));
#endif
#ifdef F_GLOBAL_NOCACHE
printf("F_GLOBAL_NOCACHE=%lld\n", (long long)(F_GLOBAL_NOCACHE));
#endif
#ifdef F_LOG2PHYS
printf("F_LOG2PHYS=%lld\n", (long long)(F_LOG2PHYS));
#endif
#ifdef F_MARKDEPENDENCY
printf("F_MARKDEPENDENCY=%lld\n", (long long)(F_MARKDEPENDENCY));
#endif
#ifdef F_NOCACHE
printf("F_NOCACHE=%lld\n", (long long)(F_NOCACHE));
#endif
#ifdef F_OK
printf("F_OK=%lld\n", (long long)(F_OK));
#endif
#ifdef F_PATHPKG_CHECK
printf("F_PATHPKG_CHECK=%lld\n", (long long)(F_PATHPKG_CHECK));
#endif
#ifdef F_PEOFPOSMODE
printf("F_PEOFPOSMODE=%lld\n", (long long)(F_PEOFPOSMODE));
#endif
#ifdef F_PREALLOCATE
printf("F_PREALLOCATE=%lld\n", (long long)(F_PREALLOCATE));
#endif
#ifdef F_RDADVISE
printf("F_RDADVISE=%lld\n", (long long)(F_RDADVISE));
#endif
#ifdef F_RDAHEAD
printf("F_RDAHEAD=%lld\n", (long long)(F_RDAHEAD));
#endif
#ifdef F_RDLCK
printf("F_RDLCK=%lld\n", (long long)(F_RDLCK));
#endif
#ifdef F_READBOOTSTRAP
printf("F_READBOOTSTRAP=%lld\n", (long long)(F_READBOOTSTRAP));
#endif
#ifdef F_SETFD
printf("F_SETFD=%lld\n", (long long)(F_SETFD));
#endif
#ifdef F_SETFL
printf("F_SETFL=%lld\n", (long long)(F_SETFL));
#endif
#ifdef F_SETLK
printf("F_SETLK=%lld\n", (long long)(F_SETLK));
#endif
#ifdef F_SETLKW
printf("F_SETLKW=%lld\n", (long long)(F_SETLKW));
#endif
#ifdef F_SETOWN
printf("F_SETOWN=%lld\n", (long long)(F_SETOWN));
#endif
#ifdef F_SETSIZE
printf("F_SETSIZE=%lld\n", (long long)(F_SETSIZE));
#endif
#ifdef F_THAW_FS
printf("F_THAW_FS=%lld\n", (long long)(F_THAW_FS));
#endif
#ifdef F_UNLCK
printf("F_UNLCK=%lld\n", (long long)(F_UNLCK));
#endif
#ifdef F_VOLPOSMODE
printf("F_VOLPOSMODE=%lld\n", (long long)(F_VOLPOSMODE));
#endif
#ifdef F_WRITEBOOTSTRAP
printf("F_WRITEBOOTSTRAP=%lld\n", (long long)(F_WRITEBOOTSTRAP));
#endif
#ifdef F_WRLCK
printf("F_WRLCK=%lld\n", (long long)(F_WRLCK));
#endif
#ifdef HUPCL
printf("HUPCL=%lld\n", (long long)(HUPCL));
#endif
#ifdef ICANON
printf("ICANON=%lld\n", (long long)(ICANON));
#endif
#ifdef ICMP6_FILTER
printf("ICMP6_FILTER=%lld\n", (long long)(ICMP6_FILTER));
#endif
#ifdef ICRNL
printf("ICRNL=%lld\n", (long long)(ICRNL));
#endif
#ifdef IEXTEN
printf("IEXTEN=%lld\n", (long long)(IEXTEN));
#endif
#ifdef IFF_ALLMULTI
printf("IFF_ALLMULTI=%lld\n", (long long)(IFF_ALLMULTI));
#endif
#ifdef IFF_ALTPHYS
printf("IFF_ALTPHYS=%lld\n", (long long)(IFF_ALTPHYS));
#endif
#ifdef IFF_BROADCAST
printf("IFF_BROADCAST=%lld\n", (long long)(IFF_BROADCAST));
#endif
#ifdef IFF_DEBUG
printf("IFF_DEBUG=%lld\n", (long long)(IFF_DEBUG));
#endif
#ifdef IFF_LINK0
printf("IFF_LINK0=%lld\n", (long long)(IFF_LINK0));
#endif
#ifdef IFF_LINK1
printf("IFF_LINK1=%lld\n", (long long)(IFF_LINK1));
#endif
#ifdef IFF_LINK2
printf("IFF_LINK2=%lld\n", (long long)(IFF_LINK2));
#endif
#ifdef IFF_LOOPBACK
printf("IFF_LOOPBACK=%lld\n", (long long)(IFF_LOOPBACK));
#endif
#ifdef IFF_MULTICAST
printf("IFF_MULTICAST=%lld\n", (long long)(IFF_MULTICAST));
#endif
#ifdef IFF_NOARP
printf("IFF_NOARP=%lld\n", (long long)(IFF_NOARP));
#endif
#ifdef IFF_NOTRAILERS
printf("IFF_NOTRAILERS=%lld\n", (long long)(IFF_NOTRAILERS));
#endif
#ifdef IFF_OACTIVE
printf("IFF_OACTIVE=%lld\n", (long long)(IFF_OACTIVE));
#endif
#ifdef IFF_POINTOPOINT
printf("IFF_POINTOPOINT=%lld\n", (long long)(IFF_POINTOPOINT));
#endif
#ifdef IFF_PROMISC
printf("IFF_PROMISC=%lld\n", (long long)(IFF_PROMISC));
#endif
#ifdef IFF_RUNNING
printf("IFF_RUNNING=%lld\n", (long long)(IFF_RUNNING));
#endif
#ifdef IFF_SIMPLEX
printf("IFF_SIMPLEX=%lld\n", (long long)(IFF_SIMPLEX));
#endif
#ifdef IFF_UP
printf("IFF_UP=%lld\n", (long long)(IFF_UP));
#endif
#ifdef IFNAMSIZ
printf("IFNAMSIZ=%lld\n", (long long)(IFNAMSIZ));
#endif
#ifdef IFT_1822
printf("IFT_1822=%lld\n", (long long)(IFT_1822));
#endif
#ifdef IFT_AAL5
printf("IFT_AAL5=%lld\n", (long long)(IFT_AAL5));
#endif
#ifdef IFT_ARCNET
printf("IFT_ARCNET=%lld\n", (long long)(IFT_ARCNET));
#endif
#ifdef IFT_ARCNETPLUS
printf("IFT_ARCNETPLUS=%lld\n", (long long)(IFT_ARCNETPLUS));
#endif
#ifdef IFT_ATM
printf("IFT_ATM=%lld\n", (long long)(IFT_ATM));
#endif
#ifdef IFT_CEPT
printf("IFT_CEPT=%lld\n", (long long)(IFT_CEPT));
#endif
#ifdef IFT_DS3
printf("IFT_DS3=%lld\n", (long long)(IFT_DS3));
#endif
#ifdef IFT_EON
printf("IFT_EON=%lld\n", (long long)(IFT_EON));
#endif
#ifdef IFT_ETHER
printf("IFT_ETHER=%lld\n", (long long)(IFT_ETHER));
#endif
#ifdef IFT_FAITH
printf("IFT_FAITH=%lld\n", (long long)(IFT_FAITH));
#endif
#ifdef IFT_FDDI
printf("IFT_FDDI=%lld\n", (long long)(IFT_FDDI));
#endif
#ifdef IFT_FRELAY
printf("IFT_FRELAY=%lld\n", (long long)(IFT_FRELAY));
#endif
#ifdef IFT_FRELAYDCE
printf("IFT_FRELAYDCE=%lld\n", (long long)(IFT_FRELAYDCE));
#endif
#ifdef IFT_GIF
printf("IFT_GIF=%lld\n", (long long)(IFT_GIF));
#endif
#ifdef IFT_HDH1822
printf("IFT_HDH1822=%lld\n", (long long)(IFT_HDH1822));
#endif
#ifdef IFT_HIPPI
printf("IFT_HIPPI=%lld\n", (long long)(IFT_HIPPI));
#endif
#ifdef IFT_HSSI
printf("IFT_HSSI=%lld\n", (long long)(IFT_HSSI));
#endif
#ifdef IFT_HY
printf("IFT_HY=%lld\n", (long long)(IFT_HY));
#endif
#ifdef IFT_IEEE1394
printf("IFT_IEEE1394=%lld\n", (long long)(IFT_IEEE1394));
#endif
#ifdef IFT_IEEE8023ADLAG
printf("IFT_IEEE8023ADLAG=%lld\n", (long long)(IFT_IEEE8023ADLAG));
#endif
#ifdef IFT_ISDNBASIC
printf("IFT_ISDNBASIC=%lld\n", (long long)(IFT_ISDNBASIC));
#endif
#ifdef IFT_ISDNPRIMARY
printf("IFT_ISDNPRIMARY=%lld\n", (long long)(IFT_ISDNPRIMARY));
#endif
#ifdef IFT_ISO88022LLC
printf("IFT_ISO88022LLC=%lld\n", (long long)(IFT_ISO88022LLC));
#endif
#ifdef IFT_ISO88023
printf("IFT_ISO88023=%lld\n", (long long)(IFT_ISO88023));
#endif
#ifdef IFT_ISO88024
printf("IFT_ISO88024=%lld\n", (long long)(IFT_ISO88024));
#endif
#ifdef IFT_ISO88025
printf("IFT_ISO88025=%lld\n", (long long)(IFT_ISO88025));
#endif
#ifdef IFT_ISO88026
printf("IFT_ISO88026=%lld\n", (long long)(IFT_ISO88026));
#endif
#ifdef IFT_L2VLAN
printf("IFT_L2VLAN=%lld\n", (long long)(IFT_L2VLAN));
#endif
#ifdef IFT_LAPB
printf("IFT_LAPB=%lld\n", (long long)(IFT_LAPB));
#endif
#ifdef IFT_LOCALTALK
printf("IFT_LOCALTALK=%lld\n", (long long)(IFT_LOCALTALK));
#endif
#ifdef IFT_LOOP
printf("IFT_LOOP=%lld\n", (long long)(IFT_LOOP));
#endif
#ifdef IFT_MIOX25
printf("IFT_MIOX25=%lld\n", (long long)(IFT_MIOX25));
#endif
#ifdef IFT_MODEM
printf("IFT_MODEM=%lld\n", (long long)(IFT_MODEM));
#endif
#ifdef IFT_NSIP
printf("IFT_NSIP=%lld\n", (long long)(IFT_NSIP));
#endif
#ifdef IFT_OTHER
printf("IFT_OTHER=%lld\n", (long long)(IFT_OTHER));
#endif
#ifdef IFT_P10
printf("IFT_P10=%lld\n", (long long)(IFT_P10));
#endif
#ifdef IFT_P80
printf("IFT_P80=%lld\n", (long long)(IFT_P80));
#endif
#ifdef IFT_PARA
printf("IFT_PARA=%lld\n", (long long)(IFT_PARA));
#endif
#ifdef IFT_PPP
printf("IFT_PPP=%lld\n", (long long)(IFT_PPP));
#endif
#ifdef IFT_PROPMUX
printf("IFT_PROPMUX=%lld\n", (long long)(IFT_PROPMUX));
#endif
#ifdef IFT_PROPVIRTUAL
printf("IFT_PROPVIRTUAL=%lld\n", (long long)(IFT_PROPVIRTUAL));
#endif
#ifdef IFT_PTPSERIAL
printf("IFT_PTPSERIAL=%lld\n", (long long)(IFT_PTPSERIAL));
#endif
#ifdef IFT_RS232
printf("IFT_RS232=%lld\n", (long long)(IFT_RS232));
#endif
#ifdef IFT_SDLC
printf("IFT_SDLC=%lld\n", (long long)(IFT_SDLC));
#endif
#ifdef IFT_SIP
printf("IFT_SIP=%lld\n", (long long)(IFT_SIP));
#endif
#ifdef IFT_SLIP
printf("IFT_SLIP=%lld\n", (long long)(IFT_SLIP));
#endif
#ifdef IFT_SMDSDXI
printf("IFT_SMDSDXI=%lld\n", (long long)(IFT_SMDSDXI));
#endif
#ifdef IFT_SMDSICIP
printf("IFT_SMDSICIP=%lld\n", (long long)(IFT_SMDSICIP));
#endif
#ifdef IFT_SONET
printf("IFT_SONET=%lld\n", (long long)(IFT_SONET));
#endif
#ifdef IFT_SONETPATH
printf("IFT_SONETPATH=%lld\n", (long long)(IFT_SONETPATH));
#endif
#ifdef IFT_SONETVT
printf("IFT_SONETVT=%lld\n", (long long)(IFT_SONETVT));
#endif
#ifdef IFT_STARLAN
printf("IFT_STARLAN=%lld\n", (long long)(IFT_STARLAN));
#endif
#ifdef IFT_STF
printf("IFT_STF=%lld\n", (long long)(IFT_STF));
#endif
#ifdef IFT_T1
printf("IFT_T1=%lld\n", (long long)(IFT_T1));
#endif
#ifdef IFT_ULTRA
printf("IFT_ULTRA=%lld\n", (long long)(IFT_ULTRA));
#endif
#ifdef IFT_V35
printf("IFT_V35=%lld\n", (long long)(IFT_V35));
#endif
#ifdef IFT_X25
printf("IFT_X25=%lld\n", (long long)(IFT_X25));
#endif
#ifdef IFT_X25DDN
printf("IFT_X25DDN=%lld\n", (long long)(IFT_X25DDN));
#endif
#ifdef IFT_X25PLE
printf("IFT_X25PLE=%lld\n", (long long)(IFT_X25PLE));
#endif
#ifdef IFT_XETHER
printf("IFT_XETHER=%lld\n", (long long)(IFT_XETHER));
#endif
#ifdef IGNBRK
printf("IGNBRK=%lld\n", (long long)(IGNBRK));
#endif
#ifdef IGNCR
printf("IGNCR=%lld\n", (long long)(IGNCR));
#endif
#ifdef IGNPAR
printf("IGNPAR=%lld\n", (long long)(IGNPAR));
#endif
#ifdef IMAXBEL
printf("IMAXBEL=%lld\n", (long long)(IMAXBEL));
#endif
#ifdef INLCR
printf("INLCR=%lld\n", (long long)(INLCR));
#endif
#ifdef INPCK
printf("INPCK=%lld\n", (long long)(INPCK));
#endif
#ifdef IN_CLASSA_HOST
printf("IN_CLASSA_HOST=%lld\n", (long long)(IN_CLASSA_HOST));
#endif
#ifdef IN_CLASSA_MAX
printf("IN_CLASSA_MAX=%lld\n", (long long)(IN_CLASSA_MAX));
#endif
#ifdef IN_CLASSA_NET
printf("IN_CLASSA_NET=%lld\n", (long long)(IN_CLASSA_NET));
#endif
#ifdef IN_CLASSA_NSHIFT
printf("IN_CLASSA_NSHIFT=%lld\n", (long long)(IN_CLASSA_NSHIFT));
#endif
#ifdef IN_CLASSB_HOST
printf("IN_CLASSB_HOST=%lld\n", (long long)(IN_CLASSB_HOST));
#endif
#ifdef IN_CLASSB_MAX
printf("IN_CLASSB_MAX=%lld\n", (long long)(IN_CLASSB_MAX));
#endif
#ifdef IN_CLASSB_NET
printf("IN_CLASSB_NET=%lld\n", (long long)(IN_CLASSB_NET));
#endif
#ifdef IN_CLASSB_NSHIFT
printf("IN_CLASSB_NSHIFT=%lld\n", (long long)(IN_CLASSB_NSHIFT));
#endif
#ifdef IN_CLASSC_HOST
printf("IN_CLASSC_HOST=%lld\n", (long long)(IN_CLASSC_HOST));
#endif
#ifdef IN_CLASSC_NET
printf("IN_CLASSC_NET=%lld\n", (long long)(IN_CLASSC_NET));
#endif
#ifdef IN_CLASSC_NSHIFT
printf("IN_CLASSC_NSHIFT=%lld\n", (long long)(IN_CLASSC_NSHIFT));
#endif
#ifdef IN_CLASSD_HOST
printf("IN_CLASSD_HOST=%lld\n", (long long)(IN_CLASSD_HOST));
#endif
#ifdef IN_CLASSD_NET
printf("IN_CLASSD_NET=%lld\n", (long long)(IN_CLASSD_NET));
#endif
#ifdef IN_CLASSD_NSHIFT
printf("IN_CLASSD_NSHIFT=%lld\n", (long long)(IN_CLASSD_NSHIFT));
#endif
#ifdef IN_LINKLOCALNETNUM
printf("IN_LINKLOCALNETNUM=%lld\n", (long long)(IN_LINKLOCALNETNUM));
#endif
#ifdef IN_LOOPBACKNET
printf("IN_LOOPBACKNET=%lld\n", (long long)(IN_LOOPBACKNET));
#endif
#ifdef IPPROTO_3PC
printf("IPPROTO_3PC=%lld\n", (long long)(IPPROTO_3PC));
#endif
#ifdef IPPROTO_ADFS
printf("IPPROTO_ADFS=%lld\n", (long long)(IPPROTO_ADFS));
#endif
#ifdef IPPROTO_AH
printf("IPPROTO_AH=%lld\n", (long long)(IPPROTO_AH));
#endif
#ifdef IPPROTO_AHIP
printf("IPPROTO_AHIP=%lld\n", (long long)(IPPROTO_AHIP));
#endif
#ifdef IPPROTO_APES
printf("IPPROTO_APES=%lld\n", (long long)(IPPROTO_APES));
#endif
#ifdef IPPROTO_ARGUS
printf("IPPROTO_ARGUS=%lld\n", (long long)(IPPROTO_ARGUS));
#endif
#ifdef IPPROTO_AX25
printf("IPPROTO_AX25=%lld\n", (long long)(IPPROTO_AX25));
#endif
#ifdef IPPROTO_BHA
printf("IPPROTO_BHA=%lld\n", (long long)(IPPROTO_BHA));
#endif
#ifdef IPPROTO_BLT
printf("IPPROTO_BLT=%lld\n", (long long)(IPPROTO_BLT));
#endif
#ifdef IPPROTO_BRSATMON
printf("IPPROTO_BRSATMON=%lld\n", (long long)(IPPROTO_BRSATMON));
#endif
#ifdef IPPROTO_CFTP
printf("IPPROTO_CFTP=%lld\n", (long long)(IPPROTO_CFTP));
#endif
#ifdef IPPROTO_CHAOS
printf("IPPROTO_CHAOS=%lld\n", (long long)(IPPROTO_CHAOS));
#endif
#ifdef IPPROTO_CMTP
printf("IPPROTO_CMTP=%lld\n", (long long)(IPPROTO_CMTP));
#endif
#ifdef IPPROTO_CPHB
printf("IPPROTO_CPHB=%lld\n", (long long)(IPPROTO_CPHB));
#endif
#ifdef IPPROTO_CPNX
printf("IPPROTO_CPNX=%lld\n", (long long)(IPPROTO_CPNX));
#endif
#ifdef IPPROTO_DDP
printf("IPPROTO_DDP=%lld\n", (long long)(IPPROTO_DDP));
#endif
#ifdef IPPROTO_DGP
printf("IPPROTO_DGP=%lld\n", (long long)(IPPROTO_DGP));
#endif
#ifdef IPPROTO_DIVERT
printf("IPPROTO_DIVERT=%lld\n", (long long)(IPPROTO_DIVERT));
#endif
#ifdef IPPROTO_DONE
printf("IPPROTO_DONE=%lld\n", (long long)(IPPROTO_DONE));
#endif
#ifdef IPPROTO_DSTOPTS
printf("IPPROTO_DSTOPTS=%lld\n", (long long)(IPPROTO_DSTOPTS));
#endif
#ifdef IPPROTO_EGP
printf("IPPROTO_EGP=%lld\n", (long long)(IPPROTO_EGP));
#endif
#ifdef IPPROTO_EMCON
printf("IPPROTO_EMCON=%lld\n", (long long)(IPPROTO_EMCON));
#endif
#ifdef IPPROTO_ENCAP
printf("IPPROTO_ENCAP=%lld\n", (long long)(IPPROTO_ENCAP));
#endif
#ifdef IPPROTO_EON
printf("IPPROTO_EON=%lld\n", (long long)(IPPROTO_EON));
#endif
#ifdef IPPROTO_ESP
printf("IPPROTO_ESP=%lld\n", (long long)(IPPROTO_ESP));
#endif
#ifdef IPPROTO_ETHERIP
printf("IPPROTO_ETHERIP=%lld\n", (long long)(IPPROTO_ETHERIP));
#endif
#ifdef IPPROTO_FRAGMENT
printf("IPPROTO_FRAGMENT=%lld\n", (long long)(IPPROTO_FRAGMENT));
#endif
#ifdef IPPROTO_GGP
printf("IPPROTO_GGP=%lld\n", (long long)(IPPROTO_GGP));
#endif
#ifdef IPPROTO_GMTP
printf("IPPROTO_GMTP=%lld\n", (long long)(IPPROTO_GMTP));
#endif
#ifdef IPPROTO_GRE
printf("IPPROTO_GRE=%lld\n", (long long)(IPPROTO_GRE));
#endif
#ifdef IPPROTO_HELLO
printf("IPPROTO_HELLO=%lld\n", (long long)(IPPROTO_HELLO));
#endif
#ifdef IPPROTO_HMP
printf("IPPROTO_HMP=%lld\n", (long long)(IPPROTO_HMP));
#endif
#ifdef IPPROTO_HOPOPTS
printf("IPPROTO_HOPOPTS=%lld\n", (long long)(IPPROTO_HOPOPTS));
#endif
#ifdef IPPROTO_ICMP
printf("IPPROTO_ICMP=%lld\n", (long long)(IPPROTO_ICMP));
#endif
#ifdef IPPROTO_ICMPV6
printf("IPPROTO_ICMPV6=%lld\n", (long long)(IPPROTO_ICMPV6));
#endif
#ifdef IPPROTO_IDP
printf("IPPROTO_IDP=%lld\n", (long long)(IPPROTO_IDP));
#endif
#ifdef IPPROTO_IDPR
printf("IPPROTO_IDPR=%lld\n", (long long)(IPPROTO_IDPR));
#endif
#ifdef IPPROTO_IDRP
printf("IPPROTO_IDRP=%lld\n", (long long)(IPPROTO_IDRP));
#endif
#ifdef IPPROTO_IGMP
printf("IPPROTO_IGMP=%lld\n", (long long)(IPPROTO_IGMP));
#endif
#ifdef IPPROTO_IGP
printf("IPPROTO_IGP=%lld\n", (long long)(IPPROTO_IGP));
#endif
#ifdef IPPROTO_IGRP
printf("IPPROTO_IGRP=%lld\n", (long long)(IPPROTO_IGRP));
#endif
#ifdef IPPROTO_IL
printf("IPPROTO_IL=%lld\n", (long long)(IPPROTO_IL));
#endif
#ifdef IPPROTO_INLSP
printf("IPPROTO_INLSP=%lld\n", (long long)(IPPROTO_INLSP));
#endif
#ifdef IPPROTO_INP
printf("IPPROTO_INP=%lld\n", (long long)(IPPROTO_INP));
#endif
#ifdef IPPROTO_IP
printf("IPPROTO_IP=%lld\n", (long long)(IPPROTO_IP));
#endif
#ifdef IPPROTO_IPCOMP
printf("IPPROTO_IPCOMP=%lld\n", (long long)(IPPROTO_IPCOMP));
#endif
#ifdef IPPROTO_IPCV
printf("IPPROTO_IPCV=%lld\n", (long long)(IPPROTO_IPCV));
#endif
#ifdef IPPROTO_IPEIP
printf("IPPROTO_IPEIP=%lld\n", (long long)(IPPROTO_IPEIP));
#endif
#ifdef IPPROTO_IPIP
printf("IPPROTO_IPIP=%lld\n", (long long)(IPPROTO_IPIP));
#endif
#ifdef IPPROTO_IPPC
printf("IPPROTO_IPPC=%lld\n", (long long)(IPPROTO_IPPC));
#endif
#ifdef IPPROTO_IPV4
printf("IPPROTO_IPV4=%lld\n", (long long)(IPPROTO_IPV4));
#endif
#ifdef IPPROTO_IPV6
printf("IPPROTO_IPV6=%lld\n", (long long)(IPPROTO_IPV6));
#endif
#ifdef IPPROTO_IRTP
printf("IPPROTO_IRTP=%lld\n", (long long)(IPPROTO_IRTP));
#endif
#ifdef IPPROTO_KRYPTOLAN
printf("IPPROTO_KRYPTOLAN=%lld\n", (long long)(IPPROTO_KRYPTOLAN));
#endif
#ifdef IPPROTO_LARP
printf("IPPROTO_LARP=%lld\n", (long long)(IPPROTO_LARP));
#endif
#ifdef IPPROTO_LEAF1
printf("IPPROTO_LEAF1=%lld\n", (long long)(IPPROTO_LEAF1));
#endif
#ifdef IPPROTO_LEAF2
printf("IPPROTO_LEAF2=%lld\n", (long long)(IPPROTO_LEAF2));
#endif
#ifdef IPPROTO_MAX
printf("IPPROTO_MAX=%lld\n", (long long)(IPPROTO_MAX));
#endif
#ifdef IPPROTO_MAXID
printf("IPPROTO_MAXID=%lld\n", (long long)(IPPROTO_MAXID));
#endif
#ifdef IPPROTO_MEAS
printf("IPPROTO_MEAS=%lld\n", (long long)(IPPROTO_MEAS));
#endif
#ifdef IPPROTO_MHRP
printf("IPPROTO_MHRP=%lld\n", (long long)(IPPROTO_MHRP));
#endif
#ifdef IPPROTO_MICP
printf("IPPROTO_MICP=%lld\n", (long long)(IPPROTO_MICP));
#endif
#ifdef IPPROTO_MTP
printf("IPPROTO_MTP=%lld\n", (long long)(IPPROTO_MTP));
#endif
#ifdef IPPROTO_MUX
printf("IPPROTO_MUX=%lld\n", (long long)(IPPROTO_MUX));
#endif
#ifdef IPPROTO_ND
printf("IPPROTO_ND=%lld\n", (long long)(IPPROTO_ND));
#endif
#ifdef IPPROTO_NHRP
printf("IPPROTO_NHRP=%lld\n", (long long)(IPPROTO_NHRP));
#endif
#ifdef IPPROTO_NONE
printf("IPPROTO_NONE=%lld\n", (long long)(IPPROTO_NONE));
#endif
#ifdef IPPROTO_NSP
printf("IPPROTO_NSP=%lld\n", (long long)(IPPROTO_NSP));
#endif
#ifdef IPPROTO_NVPII
printf("IPPROTO_NVPII=%lld\n", (long long)(IPPROTO_NVPII));
#endif
#ifdef IPPROTO_OSPFIGP
printf("IPPROTO_OSPFIGP=%lld\n", (long long)(IPPROTO_OSPFIGP));
#endif
#ifdef IPPROTO_PGM
printf("IPPROTO_PGM=%lld\n", (long long)(IPPROTO_PGM));
#endif
#ifdef IPPROTO_PIGP
printf("IPPROTO_PIGP=%lld\n", (long long)(IPPROTO_PIGP));
#endif
#ifdef IPPROTO_PIM
printf("IPPROTO_PIM=%lld\n", (long long)(IPPROTO_PIM));
#endif
#ifdef IPPROTO_PRM
printf("IPPROTO_PRM=%lld\n", (long long)(IPPROTO_PRM));
#endif
#ifdef IPPROTO_PUP
printf("IPPROTO_PUP=%lld\n", (long long)(IPPROTO_PUP));
#endif
#ifdef IPPROTO_PVP
printf("IPPROTO_PVP=%lld\n", (long long)(IPPROTO_PVP));
#endif
#ifdef IPPROTO_RAW
printf("IPPROTO_RAW=%lld\n", (long long)(IPPROTO_RAW));
#endif
#ifdef IPPROTO_RCCMON
printf("IPPROTO_RCCMON=%lld\n", (long long)(IPPROTO_RCCMON));
#endif
#ifdef IPPROTO_RDP
printf("IPPROTO_RDP=%lld\n", (long long)(IPPROTO_RDP));
#endif
#ifdef IPPROTO_ROUTING
printf("IPPROTO_ROUTING=%lld\n", (long long)(IPPROTO_ROUTING));
#endif
#ifdef IPPROTO_RSVP
printf("IPPROTO_RSVP=%lld\n", (long long)(IPPROTO_RSVP));
#endif
#ifdef IPPROTO_RVD
printf("IPPROTO_RVD=%lld\n", (long long)(IPPROTO_RVD));
#endif
#ifdef IPPROTO_SATEXPAK
printf("IPPROTO_SATEXPAK=%lld\n", (long long)(IPPROTO_SATEXPAK));
#endif
#ifdef IPPROTO_SATMON
printf("IPPROTO_SATMON=%lld\n", (long long)(IPPROTO_SATMON));
#endif
#ifdef IPPROTO_SCCSP
printf("IPPROTO_SCCSP=%lld\n", (long long)(IPPROTO_SCCSP));
#endif
#ifdef IPPROTO_SDRP
printf("IPPROTO_SDRP=%lld\n", (long long)(IPPROTO_SDRP));
#endif
#ifdef IPPROTO_SEP
printf("IPPROTO_SEP=%lld\n", (long long)(IPPROTO_SEP));
#endif
#ifdef IPPROTO_SRPC
printf("IPPROTO_SRPC=%lld\n", (long long)(IPPROTO_SRPC));
#endif
#ifdef IPPROTO_ST
printf("IPPROTO_ST=%lld\n", (long long)(IPPROTO_ST));
#endif
#ifdef IPPROTO_SVMTP
printf("IPPROTO_SVMTP=%lld\n", (long long)(IPPROTO_SVMTP));
#endif
#ifdef IPPROTO_SWIPE
printf("IPPROTO_SWIPE=%lld\n", (long long)(IPPROTO_SWIPE));
#endif
#ifdef IPPROTO_TCF
printf("IPPROTO_TCF=%lld\n", (long long)(IPPROTO_TCF));
#endif
#ifdef IPPROTO_TCP
printf("IPPROTO_TCP=%lld\n", (long long)(IPPROTO_TCP));
#endif
#ifdef IPPROTO_TP
printf("IPPROTO_TP=%lld\n", (long long)(IPPROTO_TP));
#endif
#ifdef IPPROTO_TPXX
printf("IPPROTO_TPXX=%lld\n", (long long)(IPPROTO_TPXX));
#endif
#ifdef IPPROTO_TRUNK1
printf("IPPROTO_TRUNK1=%lld\n", (long long)(IPPROTO_TRUNK1));
#endif
#ifdef IPPROTO_TRUNK2
printf("IPPROTO_TRUNK2=%lld\n", (long long)(IPPROTO_TRUNK2));
#endif
#ifdef IPPROTO_TTP
printf("IPPROTO_TTP=%lld\n", (long long)(IPPROTO_TTP));
#endif
#ifdef IPPROTO_UDP
printf("IPPROTO_UDP=%lld\n", (long long)(IPPROTO_UDP));
#endif
#ifdef IPPROTO_VINES
printf("IPPROTO_VINES=%lld\n", (long long)(IPPROTO_VINES));
#endif
#ifdef IPPROTO_VISA
printf("IPPROTO_VISA=%lld\n", (long long)(IPPROTO_VISA));
#endif
#ifdef IPPROTO_VMTP
printf("IPPROTO_VMTP=%lld\n", (long long)(IPPROTO_VMTP));
#endif
#ifdef IPPROTO_WBEXPAK
printf("IPPROTO_WBEXPAK=%lld\n", (long long)(IPPROTO_WBEXPAK));
#endif
#ifdef IPPROTO_WBMON
printf("IPPROTO_WBMON=%lld\n", (long long)(IPPROTO_WBMON));
#endif
#ifdef IPPROTO_WSN
printf("IPPROTO_WSN=%lld\n", (long long)(IPPROTO_WSN));
#endif
#ifdef IPPROTO_XNET
printf("IPPROTO_XNET=%lld\n", (long long)(IPPROTO_XNET));
#endif
#ifdef IPPROTO_XTP
printf("IPPROTO_XTP=%lld\n", (long long)(IPPROTO_XTP));
#endif
#ifdef IPV6_BINDV6ONLY
printf("IPV6_BINDV6ONLY=%lld\n", (long long)(IPV6_BINDV6ONLY));
#endif
#ifdef IPV6_CHECKSUM
printf("IPV6_CHECKSUM=%lld\n", (long long)(IPV6_CHECKSUM));
#endif
#ifdef IPV6_DEFAULT_MULTICAST_HOPS
printf("IPV6_DEFAULT_MULTICAST_HOPS=%lld\n", (long long)(IPV6_DEFAULT_MULTICAST_HOPS));
#endif
#ifdef IPV6_DEFAULT_MULTICAST_LOOP
printf("IPV6_DEFAULT_MULTICAST_LOOP=%lld\n", (long long)(IPV6_DEFAULT_MULTICAST_LOOP));
#endif
#ifdef IPV6_DEFHLIM
printf("IPV6_DEFHLIM=%lld\n", (long long)(IPV6_DEFHLIM));
#endif
#ifdef IPV6_DSTOPTS
printf("IPV6_DSTOPTS=%lld\n", (long long)(IPV6_DSTOPTS));
#endif
#ifdef IPV6_FAITH
printf("IPV6_FAITH=%lld\n", (long long)(IPV6_FAITH));
#endif
#ifdef IPV6_FLOWINFO_MASK
printf("IPV6_FLOWINFO_MASK=%lld\n", (long long)(IPV6_FLOWINFO_MASK));
#endif
#ifdef IPV6_FLOWLABEL_MASK
printf("IPV6_FLOWLABEL_MASK=%lld\n", (long long)(IPV6_FLOWLABEL_MASK));
#endif
#ifdef IPV6_FRAGTTL
printf("IPV6_FRAGTTL=%lld\n", (long long)(IPV6_FRAGTTL));
#endif
#ifdef IPV6_FW_ADD
printf("IPV6_FW_ADD=%lld\n", (long long)(IPV6_FW_ADD));
#endif
#ifdef IPV6_FW_DEL
printf("IPV6_FW_DEL=%lld\n", (long long)(IPV6_FW_DEL));
#endif
#ifdef IPV6_FW_FLUSH
printf("IPV6_FW_FLUSH=%lld\n", (long long)(IPV6_FW_FLUSH));
#endif
#ifdef IPV6_FW_GET
printf("IPV6_FW_GET=%lld\n", (long long)(IPV6_FW_GET));
#endif
#ifdef IPV6_FW_ZERO
printf("IPV6_FW_ZERO=%lld\n", (long long)(IPV6_FW_ZERO));
#endif
#ifdef IPV6_HLIMDEC
printf("IPV6_HLIMDEC=%lld\n", (long long)(IPV6_HLIMDEC));
#endif
#ifdef IPV6_HOPLIMIT
printf("IPV6_HOPLIMIT=%lld\n", (long long)(IPV6_HOPLIMIT));
#endif
#ifdef IPV6_HOPOPTS
printf("IPV6_HOPOPTS=%lld\n", (long long)(IPV6_HOPOPTS));
#endif
#ifdef IPV6_IPSEC_POLICY
printf("IPV6_IPSEC_POLICY=%lld\n", (long long)(IPV6_IPSEC_POLICY));
#endif
#ifdef IPV6_JOIN_GROUP
printf("IPV6_JOIN_GROUP=%lld\n", (long long)(IPV6_JOIN_GROUP));
#endif
#ifdef IPV6_LEAVE_GROUP
printf("IPV6_LEAVE_GROUP=%lld\n", (long long)(IPV6_LEAVE_GROUP));
#endif
#ifdef IPV6_MAXHLIM
printf("IPV6_MAXHLIM=%lld\n", (long long)(IPV6_MAXHLIM));
#endif
#ifdef IPV6_MAXPACKET
printf("IPV6_MAXPACKET=%lld\n", (long long)(IPV6_MAXPACKET));
#endif
#ifdef IPV6_MMTU
printf("IPV6_MMTU=%lld\n", (long long)(IPV6_MMTU));
#endif
#ifdef IPV6_MULTICAST_HOPS
printf("IPV6_MULTICAST_HOPS=%lld\n", (long long)(IPV6_MULTICAST_HOPS));
#endif
#ifdef IPV6_MULTICAST_IF
printf("IPV6_MULTICAST_IF=%lld\n", (long long)(IPV6_MULTICAST_IF));
#endif
#ifdef IPV6_MULTICAST_LOOP
printf("IPV6_MULTICAST_LOOP=%lld\n", (long long)(IPV6_MULTICAST_LOOP));
#endif
#ifdef IPV6_NEXTHOP
printf("IPV6_NEXTHOP=%lld\n", (long long)(IPV6_NEXTHOP));
#endif
#ifdef IPV6_PKTINFO
printf("IPV6_PKTINFO=%lld\n", (long long)(IPV6_PKTINFO));
#endif
#ifdef IPV6_PKTOPTIONS
printf("IPV6_PKTOPTIONS=%lld\n", (long long)(IPV6_PKTOPTIONS));
#endif
#ifdef IPV6_PORTRANGE
printf("IPV6_PORTRANGE=%lld\n", (long long)(IPV6_PORTRANGE));
#endif
#ifdef IPV6_PORTRANGE_DEFAULT
printf("IPV6_PORTRANGE_DEFAULT=%lld\n", (long long)(IPV6_PORTRANGE_DEFAULT));
#endif
#ifdef IPV6_PORTRANGE_HIGH
printf("IPV6_PORTRANGE_HIGH=%lld\n", (long long)(IPV6_PORTRANGE_HIGH));
#endif
#ifdef IPV6_PORTRANGE_LOW
printf("IPV6_PORTRANGE_LOW=%lld\n", (long long)(IPV6_PORTRANGE_LOW));
#endif
#ifdef IPV6_RTHDR
printf("IPV6_RTHDR=%lld\n", (long long)(IPV6_RTHDR));
#endif
#ifdef IPV6_RTHDR_LOOSE
printf("IPV6_RTHDR_LOOSE=%lld\n", (long long)(IPV6_RTHDR_LOOSE));
#endif
#ifdef IPV6_RTHDR_STRICT
printf("IPV6_RTHDR_STRICT=%lld\n", (long long)(IPV6_RTHDR_STRICT));
#endif
#ifdef IPV6_RTHDR_TYPE_0
printf("IPV6_RTHDR_TYPE_0=%lld\n", (long long)(IPV6_RTHDR_TYPE_0));
#endif
#ifdef IPV6_SOCKOPT_RESERVED1
printf("IPV6_SOCKOPT_RESERVED1=%lld\n", (long long)(IPV6_SOCKOPT_RESERVED1));
#endif
#ifdef IPV6_UNICAST_HOPS
printf("IPV6_UNICAST_HOPS=%lld\n", (long long)(IPV6_UNICAST_HOPS));
#endif
#ifdef IPV6_V6ONLY
printf("IPV6_V6ONLY=%lld\n", (long long)(IPV6_V6ONLY));
#endif
#ifdef IPV6_VERSION
printf("IPV6_VERSION=%lld\n", (long long)(IPV6_VERSION));
#endif
#ifdef IPV6_VERSION_MASK
printf("IPV6_VERSION_MASK=%lld\n", (long long)(IPV6_VERSION_MASK));
#endif
#ifdef IP_ADD_MEMBERSHIP
printf("IP_ADD_MEMBERSHIP=%lld\n", (long long)(IP_ADD_MEMBERSHIP));
#endif
#ifdef IP_DEFAULT_MULTICAST_LOOP
printf("IP_DEFAULT_MULTICAST_LOOP=%lld\n", (long long)(IP_DEFAULT_MULTICAST_LOOP));
#endif
#ifdef IP_DEFAULT_MULTICAST_TTL
printf("IP_DEFAULT_MULTICAST_TTL=%lld\n", (long long)(IP_DEFAULT_MULTICAST_TTL));
#endif
#ifdef IP_DF
printf("IP_DF=%lld\n", (long long)(IP_DF));
#endif
#ifdef IP_DROP_MEMBERSHIP
printf("IP_DROP_MEMBERSHIP=%lld\n", (long long)(IP_DROP_MEMBERSHIP));
#endif
#ifdef IP_DUMMYNET_CONFIGURE
printf("IP_DUMMYNET_CONFIGURE=%lld\n", (long long)(IP_DUMMYNET_CONFIGURE));
#endif
#ifdef IP_DUMMYNET_DEL
printf("IP_DUMMYNET_DEL=%lld\n", (long long)(IP_DUMMYNET_DEL));
#endif
#ifdef IP_DUMMYNET_FLUSH
printf("IP_DUMMYNET_FLUSH=%lld\n", (long long)(IP_DUMMYNET_FLUSH));
#endif
#ifdef IP_DUMMYNET_GET
printf("IP_DUMMYNET_GET=%lld\n", (long long)(IP_DUMMYNET_GET));
#endif
#ifdef IP_FAITH
printf("IP_FAITH=%lld\n", (long long)(IP_FAITH));
#endif
#ifdef IP_FW_ADD
printf("IP_FW_ADD=%lld\n", (long long)(IP_FW_ADD));
#endif
#ifdef IP_FW_DEL
printf("IP_FW_DEL=%lld\n", (long long)(IP_FW_DEL));
#endif
#ifdef IP_FW_FLUSH
printf("IP_FW_FLUSH=%lld\n", (long long)(IP_FW_FLUSH));
#endif
#ifdef IP_FW_GET
printf("IP_FW_GET=%lld\n", (long long)(IP_FW_GET));
#endif
#ifdef IP_FW_RESETLOG
printf("IP_FW_RESETLOG=%lld\n", (long long)(IP_FW_RESETLOG));
#endif
#ifdef IP_FW_ZERO
printf("IP_FW_ZERO=%lld\n", (long long)(IP_FW_ZERO));
#endif
#ifdef IP_HDRINCL
printf("IP_HDRINCL=%lld\n", (long long)(IP_HDRINCL));
#endif
#ifdef IP_IPSEC_POLICY
printf("IP_IPSEC_POLICY=%lld\n", (long long)(IP_IPSEC_POLICY));
#endif
#ifdef IP_MAXPACKET
printf("IP_MAXPACKET=%lld\n", (long long)(IP_MAXPACKET));
#endif
#ifdef IP_MAX_MEMBERSHIPS
printf("IP_MAX_MEMBERSHIPS=%lld\n", (long long)(IP_MAX_MEMBERSHIPS));
#endif
#ifdef IP_MF
printf("IP_MF=%lld\n", (long long)(IP_MF));
#endif
#ifdef IP_MSS
printf("IP_MSS=%lld\n", (long long)(IP_MSS));
#endif
#ifdef IP_MULTICAST_IF
printf("IP_MULTICAST_IF=%lld\n", (long long)(IP_MULTICAST_IF));
#endif
#ifdef IP_MULTICAST_LOOP
printf("IP_MULTICAST_LOOP=%lld\n", (long long)(IP_MULTICAST_LOOP));
#endif
#ifdef IP_MULTICAST_TTL
printf("IP_MULTICAST_TTL=%lld\n", (long long)(IP_MULTICAST_TTL));
#endif
#ifdef IP_MULTICAST_VIF
printf("IP_MULTICAST_VIF=%lld\n", (long long)(IP_MULTICAST_VIF));
#endif
#ifdef IP_NAT__XXX
printf("IP_NAT__XXX=%lld\n", (long long)(IP_NAT__XXX));
#endif
#ifdef IP_OFFMASK
printf("IP_OFFMASK=%lld\n", (long long)(IP_OFFMASK));
#endif
#ifdef IP_OLD_FW_ADD
printf("IP_OLD_FW_ADD=%lld\n", (long long)(IP_OLD_FW_ADD));
#endif
#ifdef IP_OLD_FW_DEL
printf("IP_OLD_FW_DEL=%lld\n", (long long)(IP_OLD_FW_DEL));
#endif
#ifdef IP_OLD_FW_FLUSH
printf("IP_OLD_FW_FLUSH=%lld\n", (long long)(IP_OLD_FW_FLUSH));
#endif
#ifdef IP_OLD_FW_GET
printf("IP_OLD_FW_GET=%lld\n", (long long)(IP_OLD_FW_GET));
#endif
#ifdef IP_OLD_FW_RESETLOG
printf("IP_OLD_FW_RESETLOG=%lld\n", (long long)(IP_OLD_FW_RESETLOG));
#endif
#ifdef IP_OLD_FW_ZERO
printf("IP_OLD_FW_ZERO=%lld\n", (long long)(IP_OLD_FW_ZERO));
#endif
#ifdef IP_OPTIONS
printf("IP_OPTIONS=%lld\n", (long long)(IP_OPTIONS));
#endif
#ifdef IP_PORTRANGE
printf("IP_PORTRANGE=%lld\n", (long long)(IP_PORTRANGE));
#endif
#ifdef IP_PORTRANGE_DEFAULT
printf("IP_PORTRANGE_DEFAULT=%lld\n", (long long)(IP_PORTRANGE_DEFAULT));
#endif
#ifdef IP_PORTRANGE_HIGH
printf("IP_PORTRANGE_HIGH=%lld\n", (long long)(IP_PORTRANGE_HIGH));
#endif
#ifdef IP_PORTRANGE_LOW
printf("IP_PORTRANGE_LOW=%lld\n", (long long)(IP_PORTRANGE_LOW));
#endif
#ifdef IP_RECVDSTADDR
printf("IP_RECVDSTADDR=%lld\n", (long long)(IP_RECVDSTADDR));
#endif
#ifdef IP_RECVIF
printf("IP_RECVIF=%lld\n", (long long)(IP_RECVIF));
#endif
#ifdef IP_RECVOPTS
printf("IP_RECVOPTS=%lld\n", (long long)(IP_RECVOPTS));
#endif
#ifdef IP_RECVRETOPTS
printf("IP_RECVRETOPTS=%lld\n", (long long)(IP_RECVRETOPTS));
#endif
#ifdef IP_RECVTTL
printf("IP_RECVTTL=%lld\n", (long long)(IP_RECVTTL));
#endif
#ifdef IP_RETOPTS
printf("IP_RETOPTS=%lld\n", (long long)(IP_RETOPTS));
#endif
#ifdef IP_RF
printf("IP_RF=%lld\n", (long long)(IP_RF));
#endif
#ifdef IP_RSVP_OFF
printf("IP_RSVP_OFF=%lld\n", (long long)(IP_RSVP_OFF));
#endif
#ifdef IP_RSVP_ON
printf("IP_RSVP_ON=%lld\n", (long long)(IP_RSVP_ON));
#endif
#ifdef IP_RSVP_VIF_OFF
printf("IP_RSVP_VIF_OFF=%lld\n", (long long)(IP_RSVP_VIF_OFF));
#endif
#ifdef IP_RSVP_VIF_ON
printf("IP_RSVP_VIF_ON=%lld\n", (long long)(IP_RSVP_VIF_ON));
#endif
#ifdef IP_STRIPHDR
printf("IP_STRIPHDR=%lld\n", (long long)(IP_STRIPHDR));
#endif
#ifdef IP_TOS
printf("IP_TOS=%lld\n", (long long)(IP_TOS));
#endif
#ifdef IP_TRAFFIC_MGT_BACKGROUND
printf("IP_TRAFFIC_MGT_BACKGROUND=%lld\n", (long long)(IP_TRAFFIC_MGT_BACKGROUND));
#endif
#ifdef IP_TTL
printf("IP_TTL=%lld\n", (long long)(IP_TTL));
#endif
#ifdef ISIG
printf("ISIG=%lld\n", (long long)(ISIG));
#endif
#ifdef ISTRIP
printf("ISTRIP=%lld\n", (long long)(ISTRIP));
#endif
#ifdef IUTF8
printf("IUTF8=%lld\n", (long long)(IUTF8));
#endif
#ifdef IXANY
printf("IXANY=%lld\n", (long long)(IXANY));
#endif
#ifdef IXOFF
printf("IXOFF=%lld\n", (long long)(IXOFF));
#endif
#ifdef IXON
printf("IXON=%lld\n", (long long)(IXON));
#endif
#ifdef LOCK_EX
printf("LOCK_EX=%lld\n", (long long)(LOCK_EX));
#endif
#ifdef LOCK_NB
printf("LOCK_NB=%lld\n", (long long)(LOCK_NB));
#endif
#ifdef LOCK_SH
printf("LOCK_SH=%lld\n", (long long)(LOCK_SH));
#endif
#ifdef LOCK_UN
printf("LOCK_UN=%lld\n", (long long)(LOCK_UN));
#endif
#ifdef MADV_DONTNEED
printf("MADV_DONTNEED=%lld\n", (long long)(MADV_DONTNEED));
#endif
#ifdef MADV_FREE
printf("MADV_FREE=%lld\n", (long long)(MADV_FREE));
#endif
#ifdef MADV_NORMAL
printf("MADV_NORMAL=%lld\n", (long long)(MADV_NORMAL));
#endif
#ifdef MADV_RANDOM
printf("MADV_RANDOM=%lld\n", (long long)(MADV_RANDOM));
#endif
#ifdef MADV_SEQUENTIAL
printf("MADV_SEQUENTIAL=%lld\n", (long long)(MADV_SEQUENTIAL));
#endif
#ifdef MADV_WILLNEED
printf("MADV_WILLNEED=%lld\n", (long long)(MADV_WILLNEED));
#endif
#ifdef MAP_ANON
printf("MAP_ANON=%lld\n", (long long)(MAP_ANON));
#endif
#ifdef MAP_COPY
printf("MAP_COPY=%lld\n", (long long)(MAP_COPY));
#endif
#ifdef MAP_FILE
printf("MAP_FILE=%lld\n", (long long)(MAP_FILE));
#endif
#ifdef MAP_FIXED
printf("MAP_FIXED=%lld\n", (long long)(MAP_FIXED));
#endif
#ifdef MAP_HASSEMAPHORE
printf("MAP_HASSEMAPHORE=%lld\n", (long long)(MAP_HASSEMAPHORE));
#endif
#ifdef MAP_NOCACHE
printf("MAP_NOCACHE=%lld\n", (long long)(MAP_NOCACHE));
#endif
#ifdef MAP_NOEXTEND
printf("MAP_NOEXTEND=%lld\n", (long long)(MAP_NOEXTEND));
#endif
#ifdef MAP_NORESERVE
printf("MAP_NORESERVE=%lld\n", (long long)(MAP_NORESERVE));
#endif
#ifdef MAP_PRIVATE
printf("MAP_PRIVATE=%lld\n", (long long)(MAP_PRIVATE));
#endif
#ifdef MAP_RENAME
printf("MAP_RENAME=%lld\n", (long long)(MAP_RENAME));
#endif
#ifdef MAP_RESERVED0080
printf("MAP_RESERVED0080=%lld\n", (long long)(MAP_RESERVED0080));
#endif
#ifdef MAP_SHARED
printf("MAP_SHARED=%lld\n", (long long)(MAP_SHARED));
#endif
#ifdef MCL_CURRENT
printf("MCL_CURRENT=%lld\n", (long long)(MCL_CURRENT));
#endif
#ifdef MCL_FUTURE
printf("MCL_FUTURE=%lld\n", (long long)(MCL_FUTURE));
#endif
#ifdef MSG_CTRUNC
printf("MSG_CTRUNC=%lld\n", (long long)(MSG_CTRUNC));
#endif
#ifdef MSG_DONTROUTE
printf("MSG_DONTROUTE=%lld\n", (long long)(MSG_DONTROUTE));
#endif
#ifdef MSG_DONTWAIT
printf("MSG_DONTWAIT=%lld\n", (long long)(MSG_DONTWAIT));
#endif
#ifdef MSG_EOF
printf("MSG_EOF=%lld\n", (long long)(MSG_EOF));
#endif
#ifdef MSG_EOR
printf("MSG_EOR=%lld\n", (long long)(MSG_EOR));
#endif
#ifdef MSG_FLUSH
printf("MSG_FLUSH=%lld\n", (long long)(MSG_FLUSH));
#endif
#ifdef MSG_HAVEMORE
printf("MSG_HAVEMORE=%lld\n", (long long)(MSG_HAVEMORE));
#endif
#ifdef MSG_HOLD
printf("MSG_HOLD=%lld\n", (long long)(MSG_HOLD));
#endif
#ifdef MSG_NEEDSA
printf("MSG_NEEDSA=%lld\n", (long long)(MSG_NEEDSA));
#endif
#ifdef MSG_OOB
printf("MSG_OOB=%lld\n", (long long)(MSG_OOB));
#endif
#ifdef MSG_PEEK
printf("MSG_PEEK=%lld\n", (long long)(MSG_PEEK));
#endif
#ifdef MSG_RCVMORE
printf("MSG_RCVMORE=%lld\n", (long long)(MSG_RCVMORE));
#endif
#ifdef MSG_SEND
printf("MSG_SEND=%lld\n", (long long)(MSG_SEND));
#endif
#ifdef MSG_TRUNC
printf("MSG_TRUNC=%lld\n", (long long)(MSG_TRUNC));
#endif
#ifdef MSG_WAITALL
printf("MSG_WAITALL=%lld\n", (long long)(MSG_WAITALL));
#endif
#ifdef MSG_WAITSTREAM
printf("MSG_WAITSTREAM=%lld\n", (long long)(MSG_WAITSTREAM));
#endif
#ifdef MS_ASYNC
printf("MS_ASYNC=%lld\n", (long long)(MS_ASYNC));
#endif
#ifdef MS_DEACTIVATE
printf("MS_DEACTIVATE=%lld\n", (long long)(MS_DEACTIVATE));
#endif
#ifdef MS_INVALIDATE
printf("MS_INVALIDATE=%lld\n", (long long)(MS_INVALIDATE));
#endif
#ifdef MS_KILLPAGES
printf("MS_KILLPAGES=%lld\n", (long long)(MS_KILLPAGES));
#endif
#ifdef MS_SYNC
printf("MS_SYNC=%lld\n", (long long)(MS_SYNC));
#endif
#ifdef NAME_MAX
printf("NAME_MAX=%lld\n", (long long)(NAME_MAX));
#endif
#ifdef NET_RT_DUMP
printf("NET_RT_DUMP=%lld\n", (long long)(NET_RT_DUMP));
#endif
#ifdef NET_RT_DUMP2
printf("NET_RT_DUMP2=%lld\n", (long long)(NET_RT_DUMP2));
#endif
#ifdef NET_RT_FLAGS
printf("NET_RT_FLAGS=%lld\n", (long long)(NET_RT_FLAGS));
#endif
#ifdef NET_RT_IFLIST
printf("NET_RT_IFLIST=%lld\n", (long long)(NET_RT_IFLIST));
#endif
#ifdef NET_RT_IFLIST2
printf("NET_RT_IFLIST2=%lld\n", (long long)(NET_RT_IFLIST2));
#endif
#ifdef NET_RT_MAXID
printf("NET_RT_MAXID=%lld\n", (long long)(NET_RT_MAXID));
#endif
#ifdef NET_RT_STAT
printf("NET_RT_STAT=%lld\n", (long long)(NET_RT_STAT));
#endif
#ifdef NET_RT_TRASH
printf("NET_RT_TRASH=%lld\n", (long long)(NET_RT_TRASH));
#endif
#ifdef NOFLSH
printf("NOFLSH=%lld\n", (long long)(NOFLSH));
#endif
#ifdef NOTE_ABSOLUTE
printf("NOTE_ABSOLUTE=%lld\n", (long long)(NOTE_ABSOLUTE));
#endif
#ifdef NOTE_ATTRIB
printf("NOTE_ATTRIB=%lld\n", (long long)(NOTE_ATTRIB));
#endif
#ifdef NOTE_CHILD
printf("NOTE_CHILD=%lld\n", (long long)(NOTE_CHILD));
#endif
#ifdef NOTE_DELETE
printf("NOTE_DELETE=%lld\n", (long long)(NOTE_DELETE));
#endif
#ifdef NOTE_EXEC
printf("NOTE_EXEC=%lld\n", (long long)(NOTE_EXEC));
#endif
#ifdef NOTE_EXIT
printf("NOTE_EXIT=%lld\n", (long long)(NOTE_EXIT));
#endif
#ifdef NOTE_EXTEND
printf("NOTE_EXTEND=%lld\n", (long long)(NOTE_EXTEND));
#endif
#ifdef NOTE_FORK
printf("NOTE_FORK=%lld\n", (long long)(NOTE_FORK));
#endif
#ifdef NOTE_LINK
printf("NOTE_LINK=%lld\n", (long long)(NOTE_LINK));
#endif
#ifdef NOTE_LOWAT
printf("NOTE_LOWAT=%lld\n", (long long)(NOTE_LOWAT));
#endif
#ifdef NOTE_NSECONDS
printf("NOTE_NSECONDS=%lld\n", (long long)(NOTE_NSECONDS));
#endif
#ifdef NOTE_PCTRLMASK
printf("NOTE_PCTRLMASK=%lld\n", (long long)(NOTE_PCTRLMASK));
#endif
#ifdef NOTE_PDATAMASK
printf("NOTE_PDATAMASK=%lld\n", (long long)(NOTE_PDATAMASK));
#endif
#ifdef NOTE_REAP
printf("NOTE_REAP=%lld\n", (long long)(NOTE_REAP));
#endif
#ifdef NOTE_RENAME
printf("NOTE_RENAME=%lld\n", (long long)(NOTE_RENAME));
#endif
#ifdef NOTE_REVOKE
printf("NOTE_REVOKE=%lld\n", (long long)(NOTE_REVOKE));
#endif
#ifdef NOTE_SECONDS
printf("NOTE_SECONDS=%lld\n", (long long)(NOTE_SECONDS));
#endif
#ifdef NOTE_SIGNAL
printf("NOTE_SIGNAL=%lld\n", (long long)(NOTE_SIGNAL));
#endif
#ifdef NOTE_TRACK
printf("NOTE_TRACK=%lld\n", (long long)(NOTE_TRACK));
#endif
#ifdef NOTE_TRACKERR
printf("NOTE_TRACKERR=%lld\n", (long long)(NOTE_TRACKERR));
#endif
#ifdef NOTE_USECONDS
printf("NOTE_USECONDS=%lld\n", (long long)(NOTE_USECONDS));
#endif
#ifdef NOTE_WRITE
printf("NOTE_WRITE=%lld\n", (long long)(NOTE_WRITE));
#endif
#ifdef OCRNL
printf("OCRNL=%lld\n", (long long)(OCRNL));
#endif
#ifdef OFDEL
printf("OFDEL=%lld\n", (long long)(OFDEL));
#endif
#ifdef OFILL
printf("OFILL=%lld\n", (long long)(OFILL));
#endif
#ifdef ONLCR
printf("ONLCR=%lld\n", (long long)(ONLCR));
#endif
#ifdef ONLRET
printf("ONLRET=%lld\n", (long long)(ONLRET));
#endif
#ifdef ONOCR
printf("ONOCR=%lld\n", (long long)(ONOCR));
#endif
#ifdef ONOEOT
printf("ONOEOT=%lld\n", (long long)(ONOEOT));
#endif
#ifdef OPOST
printf("OPOST=%lld\n", (long long)(OPOST));
#endif
#ifdef O_ACCMODE
printf("O_ACCMODE=%lld\n", (long long)(O_ACCMODE));
#endif
#ifdef O_ALERT
printf("O_ALERT=%lld\n", (long long)(O_ALERT));
#endif
#ifdef O_APPEND
printf("O_APPEND=%lld\n", (long long)(O_APPEND));
#endif
#ifdef O_ASYNC
printf("O_ASYNC=%lld\n", (long long)(O_ASYNC));
#endif
#ifdef O_CREAT
printf("O_CREAT=%lld\n", (long long)(O_CREAT));
#endif
#ifdef O_DIRECTORY
printf("O_DIRECTORY=%lld\n", (long long)(O_DIRECTORY));
#endif
#ifdef O_EVTONLY
printf("O_EVTONLY=%lld\n", (long long)(O_EVTONLY));
#endif
#ifdef O_EXCL
printf("O_EXCL=%lld\n", (long long)(O_EXCL));
#endif
#ifdef O_EXLOCK
printf("O_EXLOCK=%lld\n", (long long)(O_EXLOCK));
#endif
#ifdef O_FSYNC
printf("O_FSYNC=%lld\n", (long long)(O_FSYNC));
#endif
#ifdef O_NDELAY
printf("O_NDELAY=%lld\n", (long long)(O_NDELAY));
#endif
#ifdef O_NOCTTY
printf("O_NOCTTY=%lld\n", (long long)(O_NOCTTY));
#endif
#ifdef O_NOFOLLOW
printf("O_NOFOLLOW=%lld\n", (long long)(O_NOFOLLOW));
#endif
#ifdef O_NONBLOCK
printf("O_NONBLOCK=%lld\n", (long long)(O_NONBLOCK));
#endif
#ifdef O_POPUP
printf("O_POPUP=%lld\n", (long long)(O_POPUP));
#endif
#ifdef O_RDONLY
printf("O_RDONLY=%lld\n", (long long)(O_RDONLY));
#endif
#ifdef O_RDWR
printf("O_RDWR=%lld\n", (long long)(O_RDWR));
#endif
#ifdef O_SHLOCK
printf("O_SHLOCK=%lld\n", (long long)(O_SHLOCK));
#endif
#ifdef O_SYMLINK
printf("O_SYMLINK=%lld\n", (long long)(O_SYMLINK));
#endif
#ifdef O_SYNC
printf("O_SYNC=%lld\n", (long long)(O_SYNC));
#endif
#ifdef O_TRUNC
printf("O_TRUNC=%lld\n", (long long)(O_TRUNC));
#endif
#ifdef O_WRONLY
printf("O_WRONLY=%lld\n", (long long)(O_WRONLY));
#endif
#ifdef PARENB
printf("PARENB=%lld\n", (long long)(PARENB));
#endif
#ifdef PARMRK
printf("PARMRK=%lld\n", (long long)(PARMRK));
#endif
#ifdef PARODD
printf("PARODD=%lld\n", (long long)(PARODD));
#endif
#ifdef PENDIN
printf("PENDIN=%lld\n", (long long)(PENDIN));
#endif
#ifdef PRIO_PGRP
printf("PRIO_PGRP=%lld\n", (long long)(PRIO_PGRP));
#endif
#ifdef PRIO_PROCESS
printf("PRIO_PROCESS=%lld\n", (long long)(PRIO_PROCESS));
#endif
#ifdef PRIO_USER
printf("PRIO_USER=%lld\n", (long long)(PRIO_USER));
#endif
#ifdef PROT_EXEC
printf("PROT_EXEC=%lld\n", (long long)(PROT_EXEC));
#endif
#ifdef PROT_NONE
printf("PROT_NONE=%lld\n", (long long)(PROT_NONE));
#endif
#ifdef PROT_READ
printf("PROT_READ=%lld\n", (long long)(PROT_READ));
#endif
#ifdef PROT_WRITE
printf("PROT_WRITE=%lld\n", (long long)(PROT_WRITE));
#endif
#ifdef PT_ATTACH
printf("PT_ATTACH=%lld\n", (long long)(PT_ATTACH));
#endif
#ifdef PT_ATTACHEXC
printf("PT_ATTACHEXC=%lld\n", (long long)(PT_ATTACHEXC));
#endif
#ifdef PT_CONTINUE
printf("PT_CONTINUE=%lld\n", (long long)(PT_CONTINUE));
#endif
#ifdef PT_DENY_ATTACH
printf("PT_DENY_ATTACH=%lld\n", (long long)(PT_DENY_ATTACH));
#endif
#ifdef PT_DETACH
printf("PT_DETACH=%lld\n", (long long)(PT_DETACH));
#endif
#ifdef PT_FIRSTMACH
printf("PT_FIRSTMACH=%lld\n", (long long)(PT_FIRSTMACH));
#endif
#ifdef PT_FORCEQUOTA
printf("PT_FORCEQUOTA=%lld\n", (long long)(PT_FORCEQUOTA));
#endif
#ifdef PT_KILL
printf("PT_KILL=%lld\n", (long long)(PT_KILL));
#endif
#ifdef PT_READ_D
printf("PT_READ_D=%lld\n", (long long)(PT_READ_D));
#endif
#ifdef PT_READ_I
printf("PT_READ_I=%lld\n", (long long)(PT_READ_I));
#endif
#ifdef PT_READ_U
printf("PT_READ_U=%lld\n", (long long)(PT_READ_U));
#endif
#ifdef PT_SIGEXC
printf("PT_SIGEXC=%lld\n", (long long)(PT_SIGEXC));
#endif
#ifdef PT_STEP
printf("PT_STEP=%lld\n", (long long)(PT_STEP));
#endif
#ifdef PT_THUPDATE
printf("PT_THUPDATE=%lld\n", (long long)(PT_THUPDATE));
#endif
#ifdef PT_TRACE_ME
printf("PT_TRACE_ME=%lld\n", (long long)(PT_TRACE_ME));
#endif
#ifdef PT_WRITE_D
printf("PT_WRITE_D=%lld\n", (long long)(PT_WRITE_D));
#endif
#ifdef PT_WRITE_I
printf("PT_WRITE_I=%lld\n", (long long)(PT_WRITE_I));
#endif
#ifdef PT_WRITE_U
printf("PT_WRITE_U=%lld\n", (long long)(PT_WRITE_U));
#endif
#ifdef RLIMIT_AS
printf("RLIMIT_AS=%lld\n", (long long)(RLIMIT_AS));
#endif
#ifdef RLIMIT_CORE
printf("RLIMIT_CORE=%lld\n", (long long)(RLIMIT_CORE));
#endif
#ifdef RLIMIT_CPU
printf("RLIMIT_CPU=%lld\n", (long long)(RLIMIT_CPU));
#endif
#ifdef RLIMIT_DATA
printf("RLIMIT_DATA=%lld\n", (long long)(RLIMIT_DATA));
#endif
#ifdef RLIMIT_FSIZE
printf("RLIMIT_FSIZE=%lld\n", (long long)(RLIMIT_FSIZE));
#endif
#ifdef RLIMIT_NOFILE
printf("RLIMIT_NOFILE=%lld\n", (long long)(RLIMIT_NOFILE));
#endif
#ifdef RLIMIT_STACK
printf("RLIMIT_STACK=%lld\n", (long long)(RLIMIT_STACK));
#endif
#ifdef RLIM_INFINITY
printf("RLIM_INFINITY=%lld\n", (long long)(RLIM_INFINITY));
#endif
#ifdef RTAX_AUTHOR
printf("RTAX_AUTHOR=%lld\n", (long long)(RTAX_AUTHOR));
#endif
#ifdef RTAX_BRD
printf("RTAX_BRD=%lld\n", (long long)(RTAX_BRD));
#endif
#ifdef RTAX_DST
printf("RTAX_DST=%lld\n", (long long)(RTAX_DST));
#endif
#ifdef RTAX_GATEWAY
printf("RTAX_GATEWAY=%lld\n", (long long)(RTAX_GATEWAY));
#endif
#ifdef RTAX_GENMASK
printf("RTAX_GENMASK=%lld\n", (long long)(RTAX_GENMASK));
#endif
#ifdef RTAX_IFA
printf("RTAX_IFA=%lld\n", (long long)(RTAX_IFA));
#endif
#ifdef RTAX_IFP
printf("RTAX_IFP=%lld\n", (long long)(RTAX_IFP));
#endif
#ifdef RTAX_MAX
printf("RTAX_MAX=%lld\n", (long long)(RTAX_MAX));
#endif
#ifdef RTAX_NETMASK
printf("RTAX_NETMASK=%lld\n", (long long)(RTAX_NETMASK));
#endif
#ifdef RTA_AUTHOR
printf("RTA_AUTHOR=%lld\n", (long long)(RTA_AUTHOR));
#endif
#ifdef RTA_BRD
printf("RTA_BRD=%lld\n", (long long)(RTA_BRD));
#endif
#ifdef RTA_DST
printf("RTA_DST=%lld\n", (long long)(RTA_DST));
#endif
#ifdef RTA_GATEWAY
printf("RTA_GATEWAY=%lld\n", (long long)(RTA_GATEWAY));
#endif
#ifdef RTA_GENMASK
printf("RTA_GENMASK=%lld\n", (long long)(RTA_GENMASK));
#endif
#ifdef RTA_IFA
printf("RTA_IFA=%lld\n", (long long)(RTA_IFA));
#endif
#ifdef RTA_IFP
printf("RTA_IFP=%lld\n", (long long)(RTA_IFP));
#endif
#ifdef RTA_NETMASK
printf("RTA_NETMASK=%lld\n", (long long)(RTA_NETMASK));
#endif
#ifdef RTF_BLACKHOLE
printf("RTF_BLACKHOLE=%lld\n", (long long)(RTF_BLACKHOLE));
#endif
#ifdef RTF_BROADCAST
printf("RTF_BROADCAST=%lld\n", (long long)(RTF_BROADCAST));
#endif
#ifdef RTF_CLONING
printf("RTF_CLONING=%lld\n", (long long)(RTF_CLONING));
#endif
#ifdef RTF_DELCLONE
printf("RTF_DELCLONE=%lld\n", (long long)(RTF_DELCLONE));
#endif
#ifdef RTF_DONE
printf("RTF_DONE=%lld\n", (long long)(RTF_DONE));
#endif
#ifdef RTF_DYNAMIC
printf("RTF_DYNAMIC=%lld\n", (long long)(RTF_DYNAMIC));
#endif
#ifdef RTF_GATEWAY
printf("RTF_GATEWAY=%lld\n", (long long)(RTF_GATEWAY));
#endif
#ifdef RTF_HOST
printf("RTF_HOST=%lld\n", (long long)(RTF_HOST));
#endif
#ifdef RTF_LLINFO
printf("RTF_LLINFO=%lld\n", (long long)(RTF_LLINFO));
#endif
#ifdef RTF_LOCAL
printf("RTF_LOCAL=%lld\n", (long long)(RTF_LOCAL));
#endif
#ifdef RTF_MODIFIED
printf("RTF_MODIFIED=%lld\n", (long long)(RTF_MODIFIED));
#endif
#ifdef RTF_MULTICAST
printf("RTF_MULTICAST=%lld\n", (long long)(RTF_MULTICAST));
#endif
#ifdef RTF_PINNED
printf("RTF_PINNED=%lld\n", (long long)(RTF_PINNED));
#endif
#ifdef RTF_PRCLONING
printf("RTF_PRCLONING=%lld\n", (long long)(RTF_PRCLONING));
#endif
#ifdef RTF_PROTO1
printf("RTF_PROTO1=%lld\n", (long long)(RTF_PROTO1));
#endif
#ifdef RTF_PROTO2
printf("RTF_PROTO2=%lld\n", (long long)(RTF_PROTO2));
#endif
#ifdef RTF_PROTO3
printf("RTF_PROTO3=%lld\n", (long long)(RTF_PROTO3));
#endif
#ifdef RTF_REJECT
printf("RTF_REJECT=%lld\n", (long long)(RTF_REJECT));
#endif
#ifdef RTF_STATIC
printf("RTF_STATIC=%lld\n", (long long)(RTF_STATIC));
#endif
#ifdef RTF_TRACKREFS
printf("RTF_TRACKREFS=%lld\n", (long long)(RTF_TRACKREFS));
#endif
#ifdef RTF_UP
printf("RTF_UP=%lld\n", (long long)(RTF_UP));
#endif
#ifdef RTF_WASCLONED
printf("RTF_WASCLONED=%lld\n", (long long)(RTF_WASCLONED));
#endif
#ifdef RTF_XRESOLVE
printf("RTF_XRESOLVE=%lld\n", (long long)(RTF_XRESOLVE));
#endif
#ifdef RTM_ADD
printf("RTM_ADD=%lld\n", (long long)(RTM_ADD));
#endif
#ifdef RTM_CHANGE
printf("RTM_CHANGE=%lld\n", (long long)(RTM_CHANGE));
#endif
#ifdef RTM_DELADDR
printf("RTM_DELADDR=%lld\n", (long long)(RTM_DELADDR));
#endif
#ifdef RTM_DELETE
printf("RTM_DELETE=%lld\n", (long long)(RTM_DELETE));
#endif
#ifdef RTM_DELMADDR
printf("RTM_DELMADDR=%lld\n", (long long)(RTM_DELMADDR));
#endif
#ifdef RTM_GET
printf("RTM_GET=%lld\n", (long long)(RTM_GET));
#endif
#ifdef RTM_GET2
printf("RTM_GET2=%lld\n", (long long)(RTM_GET2));
#endif
#ifdef RTM_IFINFO
printf("RTM_IFINFO=%lld\n", (long long)(RTM_IFINFO));
#endif
#ifdef RTM_IFINFO2
printf("RTM_IFINFO2=%lld\n", (long long)(RTM_IFINFO2));
#endif
#ifdef RTM_LOCK
printf("RTM_LOCK=%lld\n", (long long)(RTM_LOCK));
#endif
#ifdef RTM_LOSING
printf("RTM_LOSING=%lld\n", (long long)(RTM_LOSING));
#endif
#ifdef RTM_MISS
printf("RTM_MISS=%lld\n", (long long)(RTM_MISS));
#endif
#ifdef RTM_NEWADDR
printf("RTM_NEWADDR=%lld\n", (long long)(RTM_NEWADDR));
#endif
#ifdef RTM_NEWMADDR
printf("RTM_NEWMADDR=%lld\n", (long long)(RTM_NEWMADDR));
#endif
#ifdef RTM_NEWMADDR2
printf("RTM_NEWMADDR2=%lld\n", (long long)(RTM_NEWMADDR2));
#endif
#ifdef RTM_OLDADD
printf("RTM_OLDADD=%lld\n", (long long)(RTM_OLDADD));
#endif
#ifdef RTM_OLDDEL
printf("RTM_OLDDEL=%lld\n", (long long)(RTM_OLDDEL));
#endif
#ifdef RTM_REDIRECT
printf("RTM_REDIRECT=%lld\n", (long long)(RTM_REDIRECT));
#endif
#ifdef RTM_RESOLVE
printf("RTM_RESOLVE=%lld\n", (long long)(RTM_RESOLVE));
#endif
#ifdef RTM_RTTUNIT
printf("RTM_RTTUNIT=%lld\n", (long long)(RTM_RTTUNIT));
#endif
#ifdef RTM_VERSION
printf("RTM_VERSION=%lld\n", (long long)(RTM_VERSION));
#endif
#ifdef RTV_EXPIRE
printf("RTV_EXPIRE=%lld\n", (long long)(RTV_EXPIRE));
#endif
#ifdef RTV_HOPCOUNT
printf("RTV_HOPCOUNT=%lld\n", (long long)(RTV_HOPCOUNT));
#endif
#ifdef RTV_MTU
printf("RTV_MTU=%lld\n", (long long)(RTV_MTU));
#endif
#ifdef RTV_RPIPE
printf("RTV_RPIPE=%lld\n", (long long)(RTV_RPIPE));
#endif
#ifdef RTV_RTT
printf("RTV_RTT=%lld\n", (long long)(RTV_RTT));
#endif
#ifdef RTV_RTTVAR
printf("RTV_RTTVAR=%lld\n", (long long)(RTV_RTTVAR));
#endif
#ifdef RTV_SPIPE
printf("RTV_SPIPE=%lld\n", (long long)(RTV_SPIPE));
#endif
#ifdef RTV_SSTHRESH
printf("RTV_SSTHRESH=%lld\n", (long long)(RTV_SSTHRESH));
#endif
#ifdef RUSAGE_CHILDREN
printf("RUSAGE_CHILDREN=%lld\n", (long long)(RUSAGE_CHILDREN));
#endif
#ifdef RUSAGE_SELF
printf("RUSAGE_SELF=%lld\n", (long long)(RUSAGE_SELF));
#endif
#ifdef SCM_CREDS
printf("SCM_CREDS=%lld\n", (long long)(SCM_CREDS));
#endif
#ifdef SCM_RIGHTS
printf("SCM_RIGHTS=%lld\n", (long long)(SCM_RIGHTS));
#endif
#ifdef SCM_TIMESTAMP
printf("SCM_TIMESTAMP=%lld\n", (long long)(SCM_TIMESTAMP));
#endif
#ifdef SHUT_RD
printf("SHUT_RD=%lld\n", (long long)(SHUT_RD));
#endif
#ifdef SHUT_RDWR
printf("SHUT_RDWR=%lld\n", (long long)(SHUT_RDWR));
#endif
#ifdef SHUT_WR
printf("SHUT_WR=%lld\n", (long long)(SHUT_WR));
#endif
#ifdef SIGABRT
printf("SIGABRT=%lld\n", (long long)(SIGABRT));
#endif
#ifdef SIGALRM
printf("SIGALRM=%lld\n", (long long)(SIGALRM));
#endif
#ifdef SIGBUS
printf("SIGBUS=%lld\n", (long long)(SIGBUS));
#endif
#ifdef SIGCHLD
printf("SIGCHLD=%lld\n", (long long)(SIGCHLD));
#endif
#ifdef SIGCONT
printf("SIGCONT=%lld\n", (long long)(SIGCONT));
#endif
#ifdef SIGEMT
printf("SIGEMT=%lld\n", (long long)(SIGEMT));
#endif
#ifdef SIGFPE
printf("SIGFPE=%lld\n", (long long)(SIGFPE));
#endif
#ifdef SIGHUP
printf("SIGHUP=%lld\n", (long long)(SIGHUP));
#endif
#ifdef SIGILL
printf("SIGILL=%lld\n", (long long)(SIGILL));
#endif
#ifdef SIGINFO
printf("SIGINFO=%lld\n", (long long)(SIGINFO));
#endif
#ifdef SIGINT
printf("SIGINT=%lld\n", (long long)(SIGINT));
#endif
#ifdef SIGIO
printf("SIGIO=%lld\n", (long long)(SIGIO));
#endif
#ifdef SIGIOT
printf("SIGIOT=%lld\n", (long long)(SIGIOT));
#endif
#ifdef SIGKILL
printf("SIGKILL=%lld\n", (long long)(SIGKILL));
#endif
#ifdef SIGPIPE
printf("SIGPIPE=%lld\n", (long long)(SIGPIPE));
#endif
#ifdef SIGPROF
printf("SIGPROF=%lld\n", (long long)(SIGPROF));
#endif
#ifdef SIGQUIT
printf("SIGQUIT=%lld\n", (long long)(SIGQUIT));
#endif
#ifdef SIGSEGV
printf("SIGSEGV=%lld\n", (long long)(SIGSEGV));
#endif
#ifdef SIGSTOP
printf("SIGSTOP=%lld\n", (long long)(SIGSTOP));
#endif
#ifdef SIGSYS
printf("SIGSYS=%lld\n", (long long)(SIGSYS));
#endif
#ifdef SIGTERM
printf("SIGTERM=%lld\n", (long long)(SIGTERM));
#endif
#ifdef SIGTRAP
printf("SIGTRAP=%lld\n", (long long)(SIGTRAP));
#endif
#ifdef SIGTSTP
printf("SIGTSTP=%lld\n", (long long)(SIGTSTP));
#endif
#ifdef SIGTTIN
printf("SIGTTIN=%lld\n", (long long)(SIGTTIN));
#endif
#ifdef SIGTTOU
printf("SIGTTOU=%lld\n", (long long)(SIGTTOU));
#endif
#ifdef SIGURG
printf("SIGURG=%lld\n", (long long)(SIGURG));
#endif
#ifdef SIGUSR1
printf("SIGUSR1=%lld\n", (long long)(SIGUSR1));
#endif
#ifdef SIGUSR2
printf("SIGUSR2=%lld\n", (long long)(SIGUSR2));
#endif
#ifdef SIGVTALRM
printf("SIGVTALRM=%lld\n", (long long)(SIGVTALRM));
#endif
#ifdef SIGWINCH
printf("SIGWINCH=%lld\n", (long long)(SIGWINCH));
#endif
#ifdef SIGXCPU
printf("SIGXCPU=%lld\n", (long long)(SIGXCPU));
#endif
#ifdef SIGXFSZ
printf("SIGXFSZ=%lld\n", (long long)(SIGXFSZ));
#endif
#ifdef SIOCADDMULTI
printf("SIOCADDMULTI=%lld\n", (long long)(SIOCADDMULTI));
#endif
#ifdef SIOCAIFADDR
printf("SIOCAIFADDR=%lld\n", (long long)(SIOCAIFADDR));
#endif
#ifdef SIOCALIFADDR
printf("SIOCALIFADDR=%lld\n", (long long)(SIOCALIFADDR));
#endif
#ifdef SIOCARPIPLL
printf("SIOCARPIPLL=%lld\n", (long long)(SIOCARPIPLL));
#endif
#ifdef SIOCATMARK
printf("SIOCATMARK=%lld\n", (long long)(SIOCATMARK));
#endif
#ifdef SIOCAUTOADDR
printf("SIOCAUTOADDR=%lld\n", (long long)(SIOCAUTOADDR));
#endif
#ifdef SIOCAUTONETMASK
printf("SIOCAUTONETMASK=%lld\n", (long long)(SIOCAUTONETMASK));
#endif
#ifdef SIOCDELMULTI
printf("SIOCDELMULTI=%lld\n", (long long)(SIOCDELMULTI));
#endif
#ifdef SIOCDIFADDR
printf("SIOCDIFADDR=%lld\n", (long long)(SIOCDIFADDR));
#endif
#ifdef SIOCDIFPHYADDR
printf("SIOCDIFPHYADDR=%lld\n", (long long)(SIOCDIFPHYADDR));
#endif
#ifdef SIOCDLIFADDR
printf("SIOCDLIFADDR=%lld\n", (long long)(SIOCDLIFADDR));
#endif
#ifdef SIOCGETSGCNT
printf("SIOCGETSGCNT=%lld\n", (long long)(SIOCGETSGCNT));
#endif
#ifdef SIOCGETVIFCNT
printf("SIOCGETVIFCNT=%lld\n", (long long)(SIOCGETVIFCNT));
#endif
#ifdef SIOCGETVLAN
printf("SIOCGETVLAN=%lld\n", (long long)(SIOCGETVLAN));
#endif
#ifdef SIOCGHIWAT
printf("SIOCGHIWAT=%lld\n", (long long)(SIOCGHIWAT));
#endif
#ifdef SIOCGIFADDR
printf("SIOCGIFADDR=%lld\n", (long long)(SIOCGIFADDR));
#endif
#ifdef SIOCGIFALTMTU
printf("SIOCGIFALTMTU=%lld\n", (long long)(SIOCGIFALTMTU));
#endif
#ifdef SIOCGIFASYNCMAP
printf("SIOCGIFASYNCMAP=%lld\n", (long long)(SIOCGIFASYNCMAP));
#endif
#ifdef SIOCGIFBOND
printf("SIOCGIFBOND=%lld\n", (long long)(SIOCGIFBOND));
#endif
#ifdef SIOCGIFBRDADDR
printf("SIOCGIFBRDADDR=%lld\n", (long long)(SIOCGIFBRDADDR));
#endif
#ifdef SIOCGIFCONF
printf("SIOCGIFCONF=%lld\n", (long long)(SIOCGIFCONF));
#endif
#ifdef SIOCGIFDEVMTU
printf("SIOCGIFDEVMTU=%lld\n", (long long)(SIOCGIFDEVMTU));
#endif
#ifdef SIOCGIFDSTADDR
printf("SIOCGIFDSTADDR=%lld\n", (long long)(SIOCGIFDSTADDR));
#endif
#ifdef SIOCGIFFLAGS
printf("SIOCGIFFLAGS=%lld\n", (long long)(SIOCGIFFLAGS));
#endif
#ifdef SIOCGIFGENERIC
printf("SIOCGIFGENERIC=%lld\n", (long long)(SIOCGIFGENERIC));
#endif
#ifdef SIOCGIFKPI
printf("SIOCGIFKPI=%lld\n", (long long)(SIOCGIFKPI));
#endif
#ifdef SIOCGIFMAC
printf("SIOCGIFMAC=%lld\n", (long long)(SIOCGIFMAC));
#endif
#ifdef SIOCGIFMEDIA
printf("SIOCGIFMEDIA=%lld\n", (long long)(SIOCGIFMEDIA));
#endif
#ifdef SIOCGIFMETRIC
printf("SIOCGIFMETRIC=%lld\n", (long long)(SIOCGIFMETRIC));
#endif
#ifdef SIOCGIFMTU
printf("SIOCGIFMTU=%lld\n", (long long)(SIOCGIFMTU));
#endif
#ifdef SIOCGIFNETMASK
printf("SIOCGIFNETMASK=%lld\n", (long long)(SIOCGIFNETMASK));
#endif
#ifdef SIOCGIFPDSTADDR
printf("SIOCGIFPDSTADDR=%lld\n", (long long)(SIOCGIFPDSTADDR));
#endif
#ifdef SIOCGIFPHYS
printf("SIOCGIFPHYS=%lld\n", (long long)(SIOCGIFPHYS));
#endif
#ifdef SIOCGIFPSRCADDR
printf("SIOCGIFPSRCADDR=%lld\n", (long long)(SIOCGIFPSRCADDR));
#endif
#ifdef SIOCGIFSTATUS
printf("SIOCGIFSTATUS=%lld\n", (long long)(SIOCGIFSTATUS));
#endif
#ifdef SIOCGIFVLAN
printf("SIOCGIFVLAN=%lld\n", (long long)(SIOCGIFVLAN));
#endif
#ifdef SIOCGLIFADDR
printf("SIOCGLIFADDR=%lld\n", (long long)(SIOCGLIFADDR));
#endif
#ifdef SIOCGLIFPHYADDR
printf("SIOCGLIFPHYADDR=%lld\n", (long long)(SIOCGLIFPHYADDR));
#endif
#ifdef SIOCGLOWAT
printf("SIOCGLOWAT=%lld\n", (long long)(SIOCGLOWAT));
#endif
#ifdef SIOCGPGRP
printf("SIOCGPGRP=%lld\n", (long long)(SIOCGPGRP));
#endif
#ifdef SIOCIFCREATE
printf("SIOCIFCREATE=%lld\n", (long long)(SIOCIFCREATE));
#endif
#ifdef SIOCIFDESTROY
printf("SIOCIFDESTROY=%lld\n", (long long)(SIOCIFDESTROY));
#endif
#ifdef SIOCRSLVMULTI
printf("SIOCRSLVMULTI=%lld\n", (long long)(SIOCRSLVMULTI));
#endif
#ifdef SIOCSETVLAN
printf("SIOCSETVLAN=%lld\n", (long long)(SIOCSETVLAN));
#endif
#ifdef SIOCSHIWAT
printf("SIOCSHIWAT=%lld\n", (long long)(SIOCSHIWAT));
#endif
#ifdef SIOCSIFADDR
printf("SIOCSIFADDR=%lld\n", (long long)(SIOCSIFADDR));
#endif
#ifdef SIOCSIFALTMTU
printf("SIOCSIFALTMTU=%lld\n", (long long)(SIOCSIFALTMTU));
#endif
#ifdef SIOCSIFASYNCMAP
printf("SIOCSIFASYNCMAP=%lld\n", (long long)(SIOCSIFASYNCMAP));
#endif
#ifdef SIOCSIFBOND
printf("SIOCSIFBOND=%lld\n", (long long)(SIOCSIFBOND));
#endif
#ifdef SIOCSIFBRDADDR
printf("SIOCSIFBRDADDR=%lld\n", (long long)(SIOCSIFBRDADDR));
#endif
#ifdef SIOCSIFDSTADDR
printf("SIOCSIFDSTADDR=%lld\n", (long long)(SIOCSIFDSTADDR));
#endif
#ifdef SIOCSIFFLAGS
printf("SIOCSIFFLAGS=%lld\n", (long long)(SIOCSIFFLAGS));
#endif
#ifdef SIOCSIFGENERIC
printf("SIOCSIFGENERIC=%lld\n", (long long)(SIOCSIFGENERIC));
#endif
#ifdef SIOCSIFKPI
printf("SIOCSIFKPI=%lld\n", (long long)(SIOCSIFKPI));
#endif
#ifdef SIOCSIFLLADDR
printf("SIOCSIFLLADDR=%lld\n", (long long)(SIOCSIFLLADDR));
#endif
#ifdef SIOCSIFMAC
printf("SIOCSIFMAC=%lld\n", (long long)(SIOCSIFMAC));
#endif
#ifdef SIOCSIFMEDIA
printf("SIOCSIFMEDIA=%lld\n", (long long)(SIOCSIFMEDIA));
#endif
#ifdef SIOCSIFMETRIC
printf("SIOCSIFMETRIC=%lld\n", (long long)(SIOCSIFMETRIC));
#endif
#ifdef SIOCSIFMTU
printf("SIOCSIFMTU=%lld\n", (long long)(SIOCSIFMTU));
#endif
#ifdef SIOCSIFNETMASK
printf("SIOCSIFNETMASK=%lld\n", (long long)(SIOCSIFNETMASK));
#endif
#ifdef SIOCSIFPHYADDR
printf("SIOCSIFPHYADDR=%lld\n", (long long)(SIOCSIFPHYADDR));
#endif
#ifdef SIOCSIFPHYS
printf("SIOCSIFPHYS=%lld\n", (long long)(SIOCSIFPHYS));
#endif
#ifdef SIOCSIFVLAN
printf("SIOCSIFVLAN=%lld\n", (long long)(SIOCSIFVLAN));
#endif
#ifdef SIOCSLIFPHYADDR
printf("SIOCSLIFPHYADDR=%lld\n", (long long)(SIOCSLIFPHYADDR));
#endif
#ifdef SIOCSLOWAT
printf("SIOCSLOWAT=%lld\n", (long long)(SIOCSLOWAT));
#endif
#ifdef SIOCSPGRP
printf("SIOCSPGRP=%lld\n", (long long)(SIOCSPGRP));
#endif
#ifdef SOCK_DGRAM
printf("SOCK_DGRAM=%lld\n", (long long)(SOCK_DGRAM));
#endif
#ifdef SOCK_MAXADDRLEN
printf("SOCK_MAXADDRLEN=%lld\n", (long long)(SOCK_MAXADDRLEN));
#endif
#ifdef SOCK_RAW
printf("SOCK_RAW=%lld\n", (long long)(SOCK_RAW));
#endif
#ifdef SOCK_RDM
printf("SOCK_RDM=%lld\n", (long long)(SOCK_RDM));
#endif
#ifdef SOCK_SEQPACKET
printf("SOCK_SEQPACKET=%lld\n", (long long)(SOCK_SEQPACKET));
#endif
#ifdef SOCK_STREAM
printf("SOCK_STREAM=%lld\n", (long long)(SOCK_STREAM));
#endif
#ifdef SOL_SOCKET
printf("SOL_SOCKET=%lld\n", (long long)(SOL_SOCKET));
#endif
#ifdef SOMAXCONN
printf("SOMAXCONN=%lld\n", (long long)(SOMAXCONN));
#endif
#ifdef SO_ACCEPTCONN
printf("SO_ACCEPTCONN=%lld\n", (long long)(SO_ACCEPTCONN));
#endif
#ifdef SO_BROADCAST
printf("SO_BROADCAST=%lld\n", (long long)(SO_BROADCAST));
#endif
#ifdef SO_DEBUG
printf("SO_DEBUG=%lld\n", (long long)(SO_DEBUG));
#endif
#ifdef SO_DONTROUTE
printf("SO_DONTROUTE=%lld\n", (long long)(SO_DONTROUTE));
#endif
#ifdef SO_DONTTRUNC
printf("SO_DONTTRUNC=%lld\n", (long long)(SO_DONTTRUNC));
#endif
#ifdef SO_ERROR
printf("SO_ERROR=%lld\n", (long long)(SO_ERROR));
#endif
#ifdef SO_KEEPALIVE
printf("SO_KEEPALIVE=%lld\n", (long long)(SO_KEEPALIVE));
#endif
#ifdef SO_LABEL
printf("SO_LABEL=%lld\n", (long long)(SO_LABEL));
#endif
#ifdef SO_LINGER
printf("SO_LINGER=%lld\n", (long long)(SO_LINGER));
#endif
#ifdef SO_LINGER_SEC
printf("SO_LINGER_SEC=%lld\n", (long long)(SO_LINGER_SEC));
#endif
#ifdef SO_NKE
printf("SO_NKE=%lld\n", (long long)(SO_NKE));
#endif
#ifdef SO_NOADDRERR
printf("SO_NOADDRERR=%lld\n", (long long)(SO_NOADDRERR));
#endif
#ifdef SO_NOSIGPIPE
printf("SO_NOSIGPIPE=%lld\n", (long long)(SO_NOSIGPIPE));
#endif
#ifdef SO_NOTIFYCONFLICT
printf("SO_NOTIFYCONFLICT=%lld\n", (long long)(SO_NOTIFYCONFLICT));
#endif
#ifdef SO_NREAD
printf("SO_NREAD=%lld\n", (long long)(SO_NREAD));
#endif
#ifdef SO_NWRITE
printf("SO_NWRITE=%lld\n", (long long)(SO_NWRITE));
#endif
#ifdef SO_OOBINLINE
printf("SO_OOBINLINE=%lld\n", (long long)(SO_OOBINLINE));
#endif
#ifdef SO_PEERLABEL
printf("SO_PEERLABEL=%lld\n", (long long)(SO_PEERLABEL));
#endif
#ifdef SO_RCVBUF
printf("SO_RCVBUF=%lld\n", (long long)(SO_RCVBUF));
#endif
#ifdef SO_RCVLOWAT
printf("SO_RCVLOWAT=%lld\n", (long long)(SO_RCVLOWAT));
#endif
#ifdef SO_RCVTIMEO
printf("SO_RCVTIMEO=%lld\n", (long long)(SO_RCVTIMEO));
#endif
#ifdef SO_RESTRICTIONS
printf("SO_RESTRICTIONS=%lld\n", (long long)(SO_RESTRICTIONS));
#endif
#ifdef SO_RESTRICT_DENYIN
printf("SO_RESTRICT_DENYIN=%lld\n", (long long)(SO_RESTRICT_DENYIN));
#endif
#ifdef SO_RESTRICT_DENYOUT
printf("SO_RESTRICT_DENYOUT=%lld\n", (long long)(SO_RESTRICT_DENYOUT));
#endif
#ifdef SO_RESTRICT_DENYSET
printf("SO_RESTRICT_DENYSET=%lld\n", (long long)(SO_RESTRICT_DENYSET));
#endif
#ifdef SO_REUSEADDR
printf("SO_REUSEADDR=%lld\n", (long long)(SO_REUSEADDR));
#endif
#ifdef SO_REUSEPORT
printf("SO_REUSEPORT=%lld\n", (long long)(SO_REUSEPORT));
#endif
#ifdef SO_REUSESHAREUID
printf("SO_REUSESHAREUID=%lld\n", (long long)(SO_REUSESHAREUID));
#endif
#ifdef SO_SNDBUF
printf("SO_SNDBUF=%lld\n", (long long)(SO_SNDBUF));
#endif
#ifdef SO_SNDLOWAT
printf("SO_SNDLOWAT=%lld\n", (long long)(SO_SNDLOWAT));
#endif
#ifdef SO_SNDTIMEO
printf("SO_SNDTIMEO=%lld\n", (long long)(SO_SNDTIMEO));
#endif
#ifdef SO_TIMESTAMP
printf("SO_TIMESTAMP=%lld\n", (long long)(SO_TIMESTAMP));
#endif
#ifdef SO_TYPE
printf("SO_TYPE=%lld\n", (long long)(SO_TYPE));
#endif
#ifdef SO_USELOOPBACK
printf("SO_USELOOPBACK=%lld\n", (long long)(SO_USELOOPBACK));
#endif
#ifdef SO_WANTMORE
printf("SO_WANTMORE=%lld\n", (long long)(SO_WANTMORE));
#endif
#ifdef SO_WANTOOBFLAG
printf("SO_WANTOOBFLAG=%lld\n", (long long)(SO_WANTOOBFLAG));
#endif
#ifdef S_IEXEC
printf("S_IEXEC=%lld\n", (long long)(S_IEXEC));
#endif
#ifdef S_IFBLK
printf("S_IFBLK=%lld\n", (long long)(S_IFBLK));
#endif
#ifdef S_IFCHR
printf("S_IFCHR=%lld\n", (long long)(S_IFCHR));
#endif
#ifdef S_IFDIR
printf("S_IFDIR=%lld\n", (long long)(S_IFDIR));
#endif
#ifdef S_IFIFO
printf("S_IFIFO=%lld\n", (long long)(S_IFIFO));
#endif
#ifdef S_IFLNK
printf("S_IFLNK=%lld\n", (long long)(S_IFLNK));
#endif
#ifdef S_IFMT
printf("S_IFMT=%lld\n", (long long)(S_IFMT));
#endif
#ifdef S_IFREG
printf("S_IFREG=%lld\n", (long long)(S_IFREG));
#endif
#ifdef S_IFSOCK
printf("S_IFSOCK=%lld\n", (long long)(S_IFSOCK));
#endif
#ifdef S_IFWHT
printf("S_IFWHT=%lld\n", (long long)(S_IFWHT));
#endif
#ifdef S_IFXATTR
printf("S_IFXATTR=%lld\n", (long long)(S_IFXATTR));
#endif
#ifdef S_IREAD
printf("S_IREAD=%lld\n", (long long)(S_IREAD));
#endif
#ifdef S_IRGRP
printf("S_IRGRP=%lld\n", (long long)(S_IRGRP));
#endif
#ifdef S_IROTH
printf("S_IROTH=%lld\n", (long long)(S_IROTH));
#endif
#ifdef S_IRUSR
printf("S_IRUSR=%lld\n", (long long)(S_IRUSR));
#endif
#ifdef S_IRWXG
printf("S_IRWXG=%lld\n", (long long)(S_IRWXG));
#endif
#ifdef S_IRWXO
printf("S_IRWXO=%lld\n", (long long)(S_IRWXO));
#endif
#ifdef S_IRWXU
printf("S_IRWXU=%lld\n", (long long)(S_IRWXU));
#endif
#ifdef S_ISGID
printf("S_ISGID=%lld\n", (long long)(S_ISGID));
#endif
#ifdef S_ISTXT
printf("S_ISTXT=%lld\n", (long long)(S_ISTXT));
#endif
#ifdef S_ISUID
printf("S_ISUID=%lld\n", (long long)(S_ISUID));
#endif
#ifdef S_ISVTX
printf("S_ISVTX=%lld\n", (long long)(S_ISVTX));
#endif
#ifdef S_IWGRP
printf("S_IWGRP=%lld\n", (long long)(S_IWGRP));
#endif
#ifdef S_IWOTH
printf("S_IWOTH=%lld\n", (long long)(S_IWOTH));
#endif
#ifdef S_IWRITE
printf("S_IWRITE=%lld\n", (long long)(S_IWRITE));
#endif
#ifdef S_IWUSR
printf("S_IWUSR=%lld\n", (long long)(S_IWUSR));
#endif
#ifdef S_IXGRP
printf("S_IXGRP=%lld\n", (long long)(S_IXGRP));
#endif
#ifdef S_IXOTH
printf("S_IXOTH=%lld\n", (long long)(S_IXOTH));
#endif
#ifdef S_IXUSR
printf("S_IXUSR=%lld\n", (long long)(S_IXUSR));
#endif
#ifdef TCIFLUSH
printf("TCIFLUSH=%lld\n", (long long)(TCIFLUSH));
#endif
#ifdef TCIOFLUSH
printf("TCIOFLUSH=%lld\n", (long long)(TCIOFLUSH));
#endif
#ifdef TCOFLUSH
printf("TCOFLUSH=%lld\n", (long long)(TCOFLUSH));
#endif
#ifdef TCP_KEEPALIVE
printf("TCP_KEEPALIVE=%lld\n", (long long)(TCP_KEEPALIVE));
#endif
#ifdef TCP_MAXBURST
printf("TCP_MAXBURST=%lld\n", (long long)(TCP_MAXBURST));
#endif
#ifdef TCP_MAXHLEN
printf("TCP_MAXHLEN=%lld\n", (long long)(TCP_MAXHLEN));
#endif
#ifdef TCP_MAXOLEN
printf("TCP_MAXOLEN=%lld\n", (long long)(TCP_MAXOLEN));
#endif
#ifdef TCP_MAXSEG
printf("TCP_MAXSEG=%lld\n", (long long)(TCP_MAXSEG));
#endif
#ifdef TCP_MAXWIN
printf("TCP_MAXWIN=%lld\n", (long long)(TCP_MAXWIN));
#endif
#ifdef TCP_MAX_SACK
printf("TCP_MAX_SACK=%lld\n", (long long)(TCP_MAX_SACK));
#endif
#ifdef TCP_MAX_WINSHIFT
printf("TCP_MAX_WINSHIFT=%lld\n", (long long)(TCP_MAX_WINSHIFT));
#endif
#ifdef TCP_MINMSS
printf("TCP_MINMSS=%lld\n", (long long)(TCP_MINMSS));
#endif
#ifdef TCP_MINMSSOVERLOAD
printf("TCP_MINMSSOVERLOAD=%lld\n", (long long)(TCP_MINMSSOVERLOAD));
#endif
#ifdef TCP_MSS
printf("TCP_MSS=%lld\n", (long long)(TCP_MSS));
#endif
#ifdef TCP_NODELAY
printf("TCP_NODELAY=%lld\n", (long long)(TCP_NODELAY));
#endif
#ifdef TCP_NOOPT
printf("TCP_NOOPT=%lld\n", (long long)(TCP_NOOPT));
#endif
#ifdef TCP_NOPUSH
printf("TCP_NOPUSH=%lld\n", (long long)(TCP_NOPUSH));
#endif
#ifdef TCSAFLUSH
printf("TCSAFLUSH=%lld\n", (long long)(TCSAFLUSH));
#endif
#ifdef TIOCCBRK
printf("TIOCCBRK=%lld\n", (long long)(TIOCCBRK));
#endif
#ifdef TIOCCDTR
printf("TIOCCDTR=%lld\n", (long long)(TIOCCDTR));
#endif
#ifdef TIOCCONS
printf("TIOCCONS=%lld\n", (long long)(TIOCCONS));
#endif
#ifdef TIOCDCDTIMESTAMP
printf("TIOCDCDTIMESTAMP=%lld\n", (long long)(TIOCDCDTIMESTAMP));
#endif
#ifdef TIOCDRAIN
printf("TIOCDRAIN=%lld\n", (long long)(TIOCDRAIN));
#endif
#ifdef TIOCDSIMICROCODE
printf("TIOCDSIMICROCODE=%lld\n", (long long)(TIOCDSIMICROCODE));
#endif
#ifdef TIOCEXCL
printf("TIOCEXCL=%lld\n", (long long)(TIOCEXCL));
#endif
#ifdef TIOCEXT
printf("TIOCEXT=%lld\n", (long long)(TIOCEXT));
#endif
#ifdef TIOCFLUSH
printf("TIOCFLUSH=%lld\n", (long long)(TIOCFLUSH));
#endif
#ifdef TIOCGDRAINWAIT
printf("TIOCGDRAINWAIT=%lld\n", (long long)(TIOCGDRAINWAIT));
#endif
#ifdef TIOCGETA
printf("TIOCGETA=%lld\n", (long long)(TIOCGETA));
#endif
#ifdef TIOCGETD
printf("TIOCGETD=%lld\n", (long long)(TIOCGETD));
#endif
#ifdef TIOCGPGRP
printf("TIOCGPGRP=%lld\n", (long long)(TIOCGPGRP));
#endif
#ifdef TIOCGWINSZ
printf("TIOCGWINSZ=%lld\n", (long long)(TIOCGWINSZ));
#endif
#ifdef TIOCIXOFF
printf("TIOCIXOFF=%lld\n", (long long)(TIOCIXOFF));
#endif
#ifdef TIOCIXON
printf("TIOCIXON=%lld\n", (long long)(TIOCIXON));
#endif
#ifdef TIOCMBIC
printf("TIOCMBIC=%lld\n", (long long)(TIOCMBIC));
#endif
#ifdef TIOCMBIS
printf("TIOCMBIS=%lld\n", (long long)(TIOCMBIS));
#endif
#ifdef TIOCMGDTRWAIT
printf("TIOCMGDTRWAIT=%lld\n", (long long)(TIOCMGDTRWAIT));
#endif
#ifdef TIOCMGET
printf("TIOCMGET=%lld\n", (long long)(TIOCMGET));
#endif
#ifdef TIOCMODG
printf("TIOCMODG=%lld\n", (long long)(TIOCMODG));
#endif
#ifdef TIOCMODS
printf("TIOCMODS=%lld\n", (long long)(TIOCMODS));
#endif
#ifdef TIOCMSDTRWAIT
printf("TIOCMSDTRWAIT=%lld\n", (long long)(TIOCMSDTRWAIT));
#endif
#ifdef TIOCMSET
printf("TIOCMSET=%lld\n", (long long)(TIOCMSET));
#endif
#ifdef TIOCM_CAR
printf("TIOCM_CAR=%lld\n", (long long)(TIOCM_CAR));
#endif
#ifdef TIOCM_CD
printf("TIOCM_CD=%lld\n", (long long)(TIOCM_CD));
#endif
#ifdef TIOCM_CTS
printf("TIOCM_CTS=%lld\n", (long long)(TIOCM_CTS));
#endif
#ifdef TIOCM_DSR
printf("TIOCM_DSR=%lld\n", (long long)(TIOCM_DSR));
#endif
#ifdef TIOCM_DTR
printf("TIOCM_DTR=%lld\n", (long long)(TIOCM_DTR));
#endif
#ifdef TIOCM_LE
printf("TIOCM_LE=%lld\n", (long long)(TIOCM_LE));
#endif
#ifdef TIOCM_RI
printf("TIOCM_RI=%lld\n", (long long)(TIOCM_RI));
#endif
#ifdef TIOCM_RNG
printf("TIOCM_RNG=%lld\n", (long long)(TIOCM_RNG));
#endif
#ifdef TIOCM_RTS
printf("TIOCM_RTS=%lld\n", (long long)(TIOCM_RTS));
#endif
#ifdef TIOCM_SR
printf("TIOCM_SR=%lld\n", (long long)(TIOCM_SR));
#endif
#ifdef TIOCM_ST
printf("TIOCM_ST=%lld\n", (long long)(TIOCM_ST));
#endif
#ifdef TIOCNOTTY
printf("TIOCNOTTY=%lld\n", (long long)(TIOCNOTTY));
#endif
#ifdef TIOCNXCL
printf("TIOCNXCL=%lld\n", (long long)(TIOCNXCL));
#endif
#ifdef TIOCOUTQ
printf("TIOCOUTQ=%lld\n", (long long)(TIOCOUTQ));
#endif
#ifdef TIOCPKT
printf("TIOCPKT=%lld\n", (long long)(TIOCPKT));
#endif
#ifdef TIOCPKT_DATA
printf("TIOCPKT_DATA=%lld\n", (long long)(TIOCPKT_DATA));
#endif
#ifdef TIOCPKT_DOSTOP
printf("TIOCPKT_DOSTOP=%lld\n", (long long)(TIOCPKT_DOSTOP));
#endif
#ifdef TIOCPKT_FLUSHREAD
printf("TIOCPKT_FLUSHREAD=%lld\n", (long long)(TIOCPKT_FLUSHREAD));
#endif
#ifdef TIOCPKT_FLUSHWRITE
printf("TIOCPKT_FLUSHWRITE=%lld\n", (long long)(TIOCPKT_FLUSHWRITE));
#endif
#ifdef TIOCPKT_IOCTL
printf("TIOCPKT_IOCTL=%lld\n", (long long)(TIOCPKT_IOCTL));
#endif
#ifdef TIOCPKT_NOSTOP
printf("TIOCPKT_NOSTOP=%lld\n", (long long)(TIOCPKT_NOSTOP));
#endif
#ifdef TIOCPKT_START
printf("TIOCPKT_START=%lld\n", (long long)(TIOCPKT_START));
#endif
#ifdef TIOCPKT_STOP
printf("TIOCPKT_STOP=%lld\n", (long long)(TIOCPKT_STOP));
#endif
#ifdef TIOCPTYGNAME
printf("TIOCPTYGNAME=%lld\n", (long long)(TIOCPTYGNAME));
#endif
#ifdef TIOCPTYGRANT
printf("TIOCPTYGRANT=%lld\n", (long long)(TIOCPTYGRANT));
#endif
#ifdef TIOCPTYUNLK
printf("TIOCPTYUNLK=%lld\n", (long long)(TIOCPTYUNLK));
#endif
#ifdef TIOCREMOTE
printf("TIOCREMOTE=%lld\n", (long long)(TIOCREMOTE));
#endif
#ifdef TIOCSBRK
printf("TIOCSBRK=%lld\n", (long long)(TIOCSBRK));
#endif
#ifdef TIOCSCONS
printf("TIOCSCONS=%lld\n", (long long)(TIOCSCONS));
#endif
#ifdef TIOCSCTTY
printf("TIOCSCTTY=%lld\n", (long long)(TIOCSCTTY));
#endif
#ifdef TIOCSDRAINWAIT
printf("TIOCSDRAINWAIT=%lld\n", (long long)(TIOCSDRAINWAIT));
#endif
#ifdef TIOCSDTR
printf("TIOCSDTR=%lld\n", (long long)(TIOCSDTR));
#endif
#ifdef TIOCSETA
printf("TIOCSETA=%lld\n", (long long)(TIOCSETA));
#endif
#ifdef TIOCSETAF
printf("TIOCSETAF=%lld\n", (long long)(TIOCSETAF));
#endif
#ifdef TIOCSETAW
printf("TIOCSETAW=%lld\n", (long long)(TIOCSETAW));
#endif
#ifdef TIOCSETD
printf("TIOCSETD=%lld\n", (long long)(TIOCSETD));
#endif
#ifdef TIOCSIG
printf("TIOCSIG=%lld\n", (long long)(TIOCSIG));
#endif
#ifdef TIOCSPGRP
printf("TIOCSPGRP=%lld\n", (long long)(TIOCSPGRP));
#endif
#ifdef TIOCSTART
printf("TIOCSTART=%lld\n", (long long)(TIOCSTART));
#endif
#ifdef TIOCSTAT
printf("TIOCSTAT=%lld\n", (long long)(TIOCSTAT));
#endif
#ifdef TIOCSTI
printf("TIOCSTI=%lld\n", (long long)(TIOCSTI));
#endif
#ifdef TIOCSTOP
printf("TIOCSTOP=%lld\n", (long long)(TIOCSTOP));
#endif
#ifdef TIOCSWINSZ
printf("TIOCSWINSZ=%lld\n", (long long)(TIOCSWINSZ));
#endif
#ifdef TIOCTIMESTAMP
printf("TIOCTIMESTAMP=%lld\n", (long long)(TIOCTIMESTAMP));
#endif
#ifdef TIOCUCNTL
printf("TIOCUCNTL=%lld\n", (long long)(TIOCUCNTL));
#endif
#ifdef TOSTOP
printf("TOSTOP=%lld\n", (long long)(TOSTOP));
#endif
#ifdef VDISCARD
printf("VDISCARD=%lld\n", (long long)(VDISCARD));
#endif
#ifdef VDSUSP
printf("VDSUSP=%lld\n", (long long)(VDSUSP));
#endif
#ifdef VEOF
printf("VEOF=%lld\n", (long long)(VEOF));
#endif
#ifdef VEOL
printf("VEOL=%lld\n", (long long)(VEOL));
#endif
#ifdef VEOL2
printf("VEOL2=%lld\n", (long long)(VEOL2));
#endif
#ifdef VERASE
printf("VERASE=%lld\n", (long long)(VERASE));
#endif
#ifdef VINTR
printf("VINTR=%lld\n", (long long)(VINTR));
#endif
#ifdef VKILL
printf("VKILL=%lld\n", (long long)(VKILL));
#endif
#ifdef VLNEXT
printf("VLNEXT=%lld\n", (long long)(VLNEXT));
#endif
#ifdef VMIN
printf("VMIN=%lld\n", (long long)(VMIN));
#endif
#ifdef VQUIT
printf("VQUIT=%lld\n", (long long)(VQUIT));
#endif
#ifdef VREPRINT
printf("VREPRINT=%lld\n", (long long)(VREPRINT));
#endif
#ifdef VSTART
printf("VSTART=%lld\n", (long long)(VSTART));
#endif
#ifdef VSTATUS
printf("VSTATUS=%lld\n", (long long)(VSTATUS));
#endif
#ifdef VSTOP
printf("VSTOP=%lld\n", (long long)(VSTOP));
#endif
#ifdef VSUSP
printf("VSUSP=%lld\n", (long long)(VSUSP));
#endif
#ifdef VT0
printf("VT0=%lld\n", (long long)(VT0));
#endif
#ifdef VT1
printf("VT1=%lld\n", (long long)(VT1));
#endif
#ifdef VTDLY
printf("VTDLY=%lld\n", (long long)(VTDLY));
#endif
#ifdef VTIME
printf("VTIME=%lld\n", (long long)(VTIME));
#endif
#ifdef VWERASE
printf("VWERASE=%lld\n", (long long)(VWERASE));
#endif
#ifdef WCONTINUED
printf("WCONTINUED=%lld\n", (long long)(WCONTINUED));
#endif
#ifdef WCOREFLAG
printf("WCOREFLAG=%lld\n", (long long)(WCOREFLAG));
#endif
#ifdef WEXITED
printf("WEXITED=%lld\n", (long long)(WEXITED));
#endif
#ifdef WNOHANG
printf("WNOHANG=%lld\n", (long long)(WNOHANG));
#endif
#ifdef WNOWAIT
printf("WNOWAIT=%lld\n", (long long)(WNOWAIT));
#endif
#ifdef __WORDSIZE
printf("WORDSIZE=%lld\n", (long long)(__WORDSIZE));
#endif
#ifdef WSTOPPED
printf("WSTOPPED=%lld\n", (long long)(WSTOPPED));
#endif
#ifdef WUNTRACED
printf("WUNTRACED=%lld\n", (long long)(WUNTRACED));
#endif
return 0;
}
