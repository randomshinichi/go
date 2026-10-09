#include <stdio.h>
#include <stddef.h>
#include <netinet/in.h>
int main(void) {
	printf("sizeof(struct in6_pktinfo)=%lu\n", (unsigned long)sizeof(struct in6_pktinfo));
	printf("offsetof(struct in6_pktinfo, ipi6_addr)=%lu\n", (unsigned long)offsetof(struct in6_pktinfo, ipi6_addr));
	printf("offsetof(struct in6_pktinfo, ipi6_ifindex)=%lu\n", (unsigned long)offsetof(struct in6_pktinfo, ipi6_ifindex));
	return 0;
}
