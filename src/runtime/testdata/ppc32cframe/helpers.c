/* C helpers for the ppc32 C-call frame probe (see main.go).
 *
 * Built with the Leopard system gcc at -O0 on purpose: at -O0 a non-variadic
 * function stores every incoming register argument into its HOME SLOT, the
 * word the Darwin PowerPC ABI reserves for it in the caller's frame at
 * caller_SP+24+4*i (reference/xnu/EXTERNAL_HEADERS/architecture/ppc/asm_help.h:
 * ARG(n) = ((n-1)*4)+24, and reg_help.h for the GPR assignment). Words nine
 * and up are NOT in that area: ARG(9) = 56, so they sit above the eight home
 * slots.
 */

#include <stdio.h>

/* Words 9 and 10 only: returns 9*100+10 = 910 when the caller put the ninth
 * and tenth argument words at caller_SP+56 and +60. Anything else gives a
 * different number, so a placement at +24 cannot return 910 by accident. */
int n10(int a, int b, int c, int d, int e, int f, int g, int h, int i, int j)
{
	return i * 100 + j;
}

/* Words 9..12 folded into one decimal number, 9101112 when all four are
 * right; -1 when one of the eight register words is wrong. */
int n12(int a, int b, int c, int d, int e, int f, int g, int h,
        int i, int j, int k, int l)
{
	if (a != 1 || b != 2 || c != 3 || d != 4 || e != 5 || f != 6 || g != 7 || h != 8)
		return -1;
	return ((i * 100 + j) * 100 + k) * 100 + l;
}

/* Homes all eight register words and then proves WHERE they were homed.
 *
 * __builtin_frame_address(0) is this function's frame, whose first word is
 * the backchain, i.e. the caller's SP. The eight home slots are therefore at
 * backchain+24 .. backchain+52. After the -O0 prologue has stored r3..r10
 * there, read the slots back and return a bitmask of the ones that do not
 * hold the argument (0 = every register argument is homed exactly in the
 * caller's parameter area). Those stores are what the caller's frame must
 * survive: home slot i is caller_SP+24+4*i, and a Go trampoline with a
 * 32-byte frame keeps its saved link register at caller_SP+40, which is
 * home slot 4 (the r7 word).
 */
int home8(int a, int b, int c, int d, int e, int f, int g, int h)
{
	volatile unsigned *slot = (volatile unsigned *)((char *)*(void **)__builtin_frame_address(0) + 24);
	unsigned want[8];
	int bad = 0, n;

	want[0] = a; want[1] = b; want[2] = c; want[3] = d;
	want[4] = e; want[5] = f; want[6] = g; want[7] = h;
	for (n = 0; n < 8; n++)
		if (slot[n] != want[n])
			bad |= 1 << n;
	return bad;
}

/* ------------------------------------------------------------------------
 * Handlers for runtime.sigfwd, the darwin/ppc32 signal forwarder.
 *
 * sigfwd calls a C signal handler with the Darwin C ABI: sig in r3, info in r4,
 * ctx in r5. The ABI puts a callee's register arguments in the CALLER's frame,
 * at ARG(n) = 24 + 4*(n-1) -- so a handler, which at -O0 homes r3-r10 for the
 * whole call, writes caller_SP+24 .. caller_SP+55 -- and BUILD_FRAME floors
 * that parameter area at 32 bytes:
 * reference/xnu/EXTERNAL_HEADERS/architecture/ppc/asm_help.h:116-125 (ARG),
 * :161-172 (__argoutsize >= 32, __framesize = ROUND_TO_STACK(24 + argout +
 * locals)), :186-188 (SAVED_LR = __framesize + 8). sigfwd's frame must
 * therefore be at least 24 + 32 = 56 bytes, or the handler's homing stores run
 * past the frame's top: the third word lands on the caller's backchain, and the
 * fifth (r7 at caller_SP+40) on the saved LR that sigfwd's own prologue wrote at
 * entry_SP+8, so sigfwd returns to the handler's r7 argument.
 *
 * These two handlers measure that contract from inside the real call. Each one
 * is an ordinary -O0 function, so its own prologue does the homing; all of its
 * parameters are used, so the stores cannot be dropped. Everything it reports is
 * read at CALLER-frame addresses rather than from its own parameters:
 *
 *   caller_sp = *(unsigned *)__builtin_frame_address(0)
 *               our frame's backchain, i.e. sigfwd's SP at the call
 *   frame     = *(unsigned *)caller_sp - caller_sp
 *               sigfwd's frame size: its prologue is stwu r1,-frame(r1), which
 *               stores its entry SP in the word at its own SP
 *   linkage0  = *(unsigned *)(caller_sp + frame)
 *               the caller-of-sigfwd frame's word zero: the backchain THAT frame
 *               planted, and the first word a homing handler overwrites when
 *               sigfwd's frame is too small
 *   argword i = *(unsigned *)(caller_sp + 24 + 4*i)
 *               the parameter area: what our prologue homed there
 *   savedlr   = *(unsigned *)(caller_sp + frame + 8)
 *               the word sigfwd itself saved as its return address
 *
 * home_report[3] is the verdict, 0 when the caller's frame honoured the ABI:
 *   1  frame < 56, so the 32-byte parameter area does not fit in that frame
 *   2  a word this handler homed lies at or above caller_sp+frame, i.e. outside
 *      the frame that called us -- written over its caller's linkage area
 *   4  the word at caller_sp+frame is not a stack pointer above caller_sp: the
 *      overwrite in (2) got the backchain
 * Bit 2 is the contract violation and bit 4 its consequence; bit 1 is the same
 * fact as a size. All three are 0 on a frame of 56 bytes or more.
 */

volatile unsigned home_report[16];

/* caller_sp is the backchain of the HANDLER's own frame, i.e. the word at
 * __builtin_frame_address(0): a handler is sigfwd's direct callee, so that word
 * is sigfwd's SP. It is passed in rather than read here, because this function's
 * own backchain would name the handler, not sigfwd. */
static void home_measure(unsigned nwords, unsigned caller_sp)
{
	unsigned frame = *(unsigned *)caller_sp - caller_sp;
	unsigned bad = 0, i;

	home_report[0] = frame;
	home_report[1] = caller_sp;
	home_report[2] = nwords;
	home_report[4] = *(unsigned *)(caller_sp + frame);
	home_report[13] = *(unsigned *)(caller_sp + frame + 8);
	for (i = 0; i < 8; i++)
		home_report[5 + i] = *(unsigned *)(caller_sp + 24 + 4 * i);

	if (frame < 56)
		bad |= 1;
	for (i = 0; i < nwords; i++)
		if (24 + 4 * i + 4 > frame)
			bad |= 2;
	if (home_report[4] <= caller_sp + frame)
		bad |= 4;
	home_report[3] = bad;

	/* Report before returning: when the frame is too small the homing above has
	 * already overwritten the saved LR, and the return through it never reaches
	 * the caller. stderr is unbuffered, so this line is in the log either way. */
	fprintf(stderr,
	        "fwd handler: frame=%u caller_SP=%#x nwords=%u bad=%u linkage0=%#x savedLR=%#x\n",
	        frame, caller_sp, nwords, bad, home_report[4], home_report[13]);
	for (i = 0; i < 8; i++)
		fprintf(stderr, "fwd handler: argword%u(caller_SP+%u)=%#x\n",
		        i + 1, 24 + 4 * i, home_report[5 + i]);
}

/* The three-argument signal-handler shape: sig, info, ctx in r3, r4, r5. Its
 * -O0 prologue homes those three words at caller_SP+24, +28, +32, so the third
 * one is the word that leaves a 32-byte caller frame. */
int home3(int sig, void *info, void *ctx)
{
	home_report[14] = (unsigned)sig ^ (unsigned)(unsigned long)info ^ (unsigned)(unsigned long)ctx;
	home_measure(3, *(unsigned *)__builtin_frame_address(0));
	return (int)home_report[3];
}

/* Eight argument words, so -O0 homes r3-r10: the full 32-byte parameter area,
 * including word five (r7 at caller_SP+40), the slot a caller with a 32-byte
 * frame keeps its saved LR in. Called through sigfwd it receives three real
 * arguments and five leftovers, like any handler compiled to home: the ABI is
 * the same either way. */
int home8fwd(int a, int b, int c, int d, int e, int f, int g, int h)
{
	home_report[14] = (unsigned)(a ^ b ^ c ^ d ^ e ^ f ^ g ^ h);
	home_report[15] = (unsigned)(a ^ b ^ c);
	home_measure(8, *(unsigned *)__builtin_frame_address(0));
	return (int)home_report[3];
}
