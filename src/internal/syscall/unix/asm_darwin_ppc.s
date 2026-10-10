// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·libc_arc4random_trampoline(SB),NOSPLIT,$0-0; JMP libc_arc4random(SB)
TEXT ·libc_ptsname_trampoline(SB),NOSPLIT,$0-0; JMP libc_ptsname(SB)
