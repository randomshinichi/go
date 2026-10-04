// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// Thread-local storage for g on darwin/ppc64 (Mac OS X 10.5 Leopard).
//
// R13 is the thread pointer: libSystem's pthread_self is `or r3,r13,r13`, and
// pthread_getspecific(key) loads from r13 + 0x60 + key*8. Leopard has no
// __thread TLS, so g lives in a pthread-key slot, exactly as on darwin/arm64.
// runtime·tls_g is not a TLS symbol but an ordinary variable holding the byte
// offset of that slot from R13; tlsinit (sys_darwin_ppc64.go) allocates the key
// and stores the offset before the first save_g. Go code never writes R13.
//
// Neither function makes a call, so neither clobbers anything but g (R30) and
// R31, which is what every caller of load_g and save_g relies on.

// save_g saves the g register into thread-local memory, so that we can call
// externally compiled code that will overwrite this register. Unlike the
// shared PPC64 version it does not skip the store when !iscgo: Darwin's signal
// trampoline always reloads g from this slot.
TEXT runtime·save_g(SB),NOSPLIT|NOFRAME,$0-0
	MOVD	runtime·tls_g(SB), R31
	MOVD	g, (R31)(R13)
	RET

// load_g loads the g register from thread-local memory, for use after calling
// externally compiled code that overwrote those registers.
//
// This is never called directly from C code (it doesn't have to follow the C
// ABI), but it may be called from a C context, where the usual Go registers
// aren't set up.
TEXT runtime·load_g(SB),NOSPLIT|NOFRAME,$0-0
	MOVD	runtime·tls_g(SB), R31
	MOVD	(R31)(R13), g
	RET

GLOBL runtime·tls_g+0(SB), NOPTR, $8
