// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// swtch_pri is Mach trap -59 on Leopard (reference/xnu/osfmk/mach/syscall_sw.h).
TEXT runtime·sched_yield(SB),NOSPLIT,$0-0
	ADDI R3, R0, $0 // swtch_pri takes a priority argument; yield at normal priority
	ADDI R0, R0, $-59
	SC
	RET
