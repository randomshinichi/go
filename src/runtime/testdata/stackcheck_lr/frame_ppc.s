// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

#include "textflag.h"

// bigFrame has a frame large enough that the assembler emits the split-stack
// prologue (MFLR; STW; BL runtime.stackcheck; ...) but a body of one instruction,
// so most samples that land in this function land in that prologue.
TEXT ·bigFrame(SB), 0, $2432-0
	RET
