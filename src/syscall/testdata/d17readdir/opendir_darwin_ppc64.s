// Assembly trampoline for opendir$INODE64, plus its address in a GLOBL so Go
// code can pass the address to syscall.syscallPtr. This is the x/sys/unix
// pattern (src/cmd/vendor/golang.org/x/sys/unix/zsyscall_darwin_amd64.s: the
// pair TEXT libc_x_trampoline<>(SB) ... / GLOBL ·libc_x_trampoline_addr(SB)).
//
// The test needs a DIR* because readdir_r takes one, and darwin/ppc64 has no
// fdopendir to hand out a DIR* from a descriptor (that is why
// syscall.Getdirentries is a raw SYS_getdirentries64 there). The DIR must come
// from the 64-bit-inode member of the family, opendir$INODE64: C's
// __DARWIN_INODE64(opendir) macro selects it together with readdir_r, and the
// two variants keep different DIR structures, so opening with the legacy
// opendir and reading with readdir_r$INODE64 would be a second mispairing.
//
// This binding lives in the test module, not in the shipped syscall package:
// nothing in the standard library calls opendir on darwin/ppc64, so adding an
// entry point for it there would be API surface added for a test.

#include "textflag.h"

TEXT libc_opendir_inode64_trampoline<>(SB),NOSPLIT,$0-0
	JMP	libc_opendir_inode64(SB)
GLOBL	·libc_opendir_inode64_trampoline_addr(SB), RODATA, $8
DATA	·libc_opendir_inode64_trampoline_addr(SB)/8, $libc_opendir_inode64_trampoline<>(SB)
