/*
 * Deterministic DYLD_INSERT_LIBRARIES shim for sys_darwin_ppc_test.go.
 * Build on the Leopard target with:
 *   gcc -dynamiclib -o arc4random-shim.dylib arc4random-shim.c
 */
#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>

static unsigned int calls;

uint32_t arc4random(void) {
	const char *path = getenv("ARC4RANDOM_SHIM_COUNT");
	unsigned int word = 0x10203040U + calls * 0x01010101U;
	char count[32];
	int fd;
	int len;

	calls++;
	if (path != NULL) {
		fd = open(path, O_WRONLY | O_CREAT | O_TRUNC, 0600);
		if (fd >= 0) {
			len = snprintf(count, sizeof(count), "%u", calls);
			if (len > 0)
				(void)write(fd, count, (size_t)len);
			(void)close(fd);
		}
	}
	return word;
}
