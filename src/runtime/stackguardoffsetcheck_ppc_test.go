// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
	"unsafe"
)

// TestPPC32Stackguard0Offset checks the field offset the ppc32
// linker-inserted prologue hardcodes for its inline stack guard.
//
// If the head of runtime.g ever changes shape, the prologue compares SP-frame
// against a garbage word in the wrong slot: stacks would grow uselessly, or not
// at all, in every non-leaf function -- silent corruption rather than a build
// failure. That is the drift this test stops. It runs on the target only (both
// files are ppc-only by their names), so a pass here is hardware evidence, not a
// host-only claim.
func TestPPC32Stackguard0Offset(t *testing.T) {
	const ptrSize = unsafe.Sizeof(uintptr(0))
	off := runtime.PPC32StackguardOffsets
	if off.Stackguard0 != 2*ptrSize {
		t.Fatalf("g.stackguard0 is at offset %d, but the ppc32 prologue in src/cmd/internal/obj/ppc/link.go compares against %d: the two must move together",
			off.Stackguard0, 2*ptrSize)
	}
	// The derivation, named, so a rename or reorder points at what moved.
	if off.Stack != 0 {
		t.Fatalf("g.stack is at offset %d, want 0 (stackguard0 is derived from it)", off.Stack)
	}
	if off.StackHi != ptrSize {
		t.Fatalf("g.stack.hi is at offset %d, want %d", off.StackHi, ptrSize)
	}
	if off.Stackguard1 != 3*ptrSize {
		t.Fatalf("g.stackguard1 is at offset %d, want %d (ppc64's prologue uses this expression for C functions)", off.Stackguard1, 3*ptrSize)
	}
}
