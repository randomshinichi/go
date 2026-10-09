#include <stdio.h>
#include <sys/syscall.h>
static void raw(unsigned int nr, unsigned int a1, unsigned int a2, unsigned int a3, int *result, unsigned int *error) {
	register unsigned int r0 __asm__("r0") = nr;
	register unsigned int r3 __asm__("r3") = a1;
	register unsigned int r4 __asm__("r4") = a2;
	register unsigned int r5 __asm__("r5") = a3;
	__asm__ volatile(
		"sc\n\t"
		"b 1f\n\t"       /* Darwin resumes here only when the syscall failed. */
		"b 2f\n\t"       /* Success resumes after the failure branch. */
		"1: mr r4,r3\n\t"  /* Preserve errno before returning -1. */
		"li r3,-1\n\t"
		"b 3f\n\t"
		"2: nop\n\t"
		"3:"
		: "+r"(r0), "+r"(r3), "+r"(r4), "+r"(r5)
		:
		: "cr0", "memory");
	*result = (int)r3;
	*error = r4;
}
int main(void) {
	int result;
	unsigned int error;
	raw(SYS_getpid, 0, 0, 0, &result, &error);
	printf("SYS_getpid=%d result=%d error=%u\n", SYS_getpid, result, error);
	raw(SYS_open, (unsigned int)"/tmp/ppc32-syscall-no-such-file", 0, 0, &result, &error);
	printf("SYS_open_missing=%d result=%d error=%u\n", SYS_open, result, error);
	return 0;
}
