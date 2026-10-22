// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gc

#include "textflag.h"

// Port-local (goppc-jic6x.3): system call support for darwin/ppc32.
// Just jump to package syscall's implementation for all these functions.
// darwin/ppc has no syscall.Syscall9, so unix.Syscall9 is not provided.

TEXT	·Syscall(SB),NOSPLIT,$0-28
	JMP	syscall·Syscall(SB)

TEXT	·Syscall6(SB),NOSPLIT,$0-40
	JMP	syscall·Syscall6(SB)

TEXT	·RawSyscall(SB),NOSPLIT,$0-28
	JMP	syscall·RawSyscall(SB)

TEXT	·RawSyscall6(SB),NOSPLIT,$0-40
	JMP	syscall·RawSyscall6(SB)
