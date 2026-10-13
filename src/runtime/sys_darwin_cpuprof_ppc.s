// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

#include "textflag.h"

// mach_port_t pthread_mach_thread_np(pthread_t). The Go wrapper passes a
// pointer to the pthread argument in R3; libcCall provides its C stack.
TEXT runtime·pthread_mach_thread_np_trampoline(SB), NOSPLIT, $8-0
	OR R14, R3, R3
	LWZ R3, 0(R14)
	BL libc_pthread_mach_thread_np(SB)
	STW R3, 4(R14)
	RET

// kern_return_t thread_info(thread, flavor, info, count).
TEXT runtime·thread_info_trampoline(SB), NOSPLIT, $8-0
	OR R14, R3, R3
	LWZ R3, 0(R14)
	LWZ R4, 4(R14)
	LWZ R5, 8(R14)
	LWZ R6, 12(R14)
	BL libc_thread_info(SB)
	RET
