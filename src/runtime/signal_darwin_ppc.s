// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// sigtramp is entered with the Darwin C signal-handler arguments in R3-R5.
// The OS signal context owns the interrupted machine state and restores it
// when the signal return trampoline runs, so there is no need to save all of
// the C ABI's callee-saved registers here. The 8-byte local frame is retained
// for the direct Go call's link register; its actual ppc32 cost is 32 bytes
// including the fixed linkage area.
TEXT runtime·sigtramp(SB), NOSPLIT|TOPFRAME, $8-0
	BL	runtime·sigtrampgo<ABIInternal>(SB)
	RET

// cgo is not supported by the ppc32 Darwin port; forward without a frame.
TEXT runtime·cgoSigtramp(SB), NOSPLIT|NOFRAME, $0-0
	B	runtime·sigtramp(SB)
