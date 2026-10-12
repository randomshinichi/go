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
