// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg_test

import (
	"testing"
	"unsafe"
)

//go:linkname memequal runtime.memequal
func memequal(a, b unsafe.Pointer, size uintptr) bool

func TestMemequalBoolResult(t *testing.T) {
	a := [3]byte{1, 2, 3}
	b := a
	// Call the assembly entry directly, not an optimized constant comparison.
	// A word store of 1 instead of a byte store returns false on big-endian PPC.
	if !memequal(unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), 3) {
		t.Fatal("equal distinct arrays returned false")
	}
	b[2] = 4
	if memequal(unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), 3) {
		t.Fatal("unequal arrays returned true")
	}
	if !memequal(unsafe.Pointer(&a[0]), unsafe.Pointer(&b[0]), 0) {
		t.Fatal("zero-byte comparison returned false")
	}
}
