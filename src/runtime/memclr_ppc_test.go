// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestMemclrNoHeapPointersExactRange(t *testing.T) {
	for offset := 0; offset < 4; offset++ {
		for n := 0; n <= 17; n++ {
			b := [24]byte{}
			for i := range b {
				b[i] = 0xa5
			}
			runtime.MemclrNoHeapPointers(unsafe.Pointer(&b[offset]), uintptr(n))
			for i, got := range b {
				want := byte(0xa5)
				if i >= offset && i < offset+n {
					want = 0
				}
				if got != want {
					t.Fatalf("offset=%d n=%d byte[%d]=%#x want %#x", offset, n, i, got, want)
				}
			}
		}
	}
}
