// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "textflag.h"

// dlopen and dlsym are reached the way the runtime reaches every libc entry:
// a Go-ABI0 stub that tail-branches to the dynamic import, whose ADDRESS
// (stored in a data word, since a Go func value would point at an ABIInternal
// wrapper instead) is handed to syscall.rawsyscalln.
TEXT dlopen_tramp<>(SB), NOSPLIT, $0-0
	JMP	libc_dlopen(SB)

TEXT dlsym_tramp<>(SB), NOSPLIT, $0-0
	JMP	libc_dlsym(SB)

DATA ·dlopenAddr+0(SB)/4, $dlopen_tramp<>(SB)
GLOBL ·dlopenAddr(SB), RODATA, $4
DATA ·dlsymAddr+0(SB)/4, $dlsym_tramp<>(SB)
GLOBL ·dlsymAddr(SB), RODATA, $4

// func gsp() (gptr, stk uintptr)
// The current goroutine pointer (R30) and stack pointer (R1), read without
// a frame so R1 is the caller's SP exactly.
TEXT ·gsp(SB), NOSPLIT|NOFRAME, $0-8
	STW	R30, gptr+0(FP)
	STW	R1, stk+4(FP)
	RET
