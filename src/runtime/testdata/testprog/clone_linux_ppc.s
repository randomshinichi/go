// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·cloneBoundaryStart(SB),NOSPLIT|NOFRAME,$0-24
	B runtime·clone(SB)

TEXT ·cloneBoundaryChildPC(SB),NOSPLIT|NOFRAME,$0-4
	ADDI $cloneBoundaryChild<>(SB), R3
	STW R3, ret+0(FP)
	RET

TEXT cloneBoundaryChild<>(SB),NOSPLIT|NOFRAME,$0-0
	// Model the first write in a Go non-leaf prologue, such as mstart.
	MFLR R0
	STW R0, 8(R1)
	SYNC
	ADDI R3, R0, $1
	STW R3, 0(R8) // Release completion after the linkage write.
	ADDI R3, R0, $0
	ADDI R0, R0, $1 // Linux SYS_exit, not exit_group.
	SC
	B 0(PC)
