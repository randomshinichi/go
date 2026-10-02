// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64 && ppc64.ppc970

package subtle

import "unsafe"

// xorBytes implements XOR bytewise because PPC970 does not support the
// unaligned vector operations used by the PPC64 assembly implementation.
func xorBytes(dstb, xb, yb *byte, n int) {
	dst := unsafe.Slice(dstb, n)
	x := unsafe.Slice(xb, n)
	y := unsafe.Slice(yb, n)
	for i := 0; i < n; i++ {
		dst[i] = x[i] ^ y[i]
	}
}
