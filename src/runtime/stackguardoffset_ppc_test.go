// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file lives in the runtime package so it can read the layout of g. The
// assertion itself is in stackguardoffsetcheck_ppc_test.go, which is in package
// runtime_test because a runtime-internal test file cannot import testing (the
// import cycle the align_runtime_test.go/align_test.go pair already splits the
// same way).

package runtime

import "unsafe"

// PPC32StackguardOffsets records the g fields the ppc32 prologue's inline stack
// guard depends on. src/cmd/internal/obj/ppc/link.go hardcodes 2*PtrSize for
// stackguard0, because the prologue is spliced by the assembler, long after
// type checking, and cannot ask for a field offset. ppc64 makes the same
// hardcoded choice (obj/ppc64/obj9.go: 2*PtrSize for stackguard0, 3*PtrSize for
// stackguard1), and runtime2.go marks both fields "offset known to liblink".
var PPC32StackguardOffsets = struct {
	Stack, StackHi, Stackguard0, Stackguard1 uintptr
}{
	unsafe.Offsetof(g{}.stack),
	unsafe.Offsetof(g{}.stack.hi),
	unsafe.Offsetof(g{}.stackguard0),
	unsafe.Offsetof(g{}.stackguard1),
}
