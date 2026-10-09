// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

TEXT ·loadRaw(SB),NOSPLIT,$0-8
	LWZ R3, ptr+0(FP)
	SYNC
	LWZ R4, 0(R3)
	SYNC
	STW R4, ret+4(FP)
	RET

