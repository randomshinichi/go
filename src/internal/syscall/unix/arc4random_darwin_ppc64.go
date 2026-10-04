// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"internal/abi"
)

//go:cgo_import_dynamic libc_arc4random arc4random "/usr/lib/libSystem.B.dylib"

func libc_arc4random_trampoline()

// ARC4Random fills p with random bytes.
//
// Mac OS X 10.5 (Leopard) has arc4random(3) but not arc4random_buf, so p is
// filled from successive arc4random draws, four bytes per draw, least
// significant byte first. This is the same substitution the runtime makes in
// runtime·arc4random_buf_trampoline (runtime/sys_darwin_ppc64.s). arc4random is
// Leopard's own kernel-seeded CSPRNG, so the bytes are cryptographically
// equivalent in kind to what arc4random_buf returns; they are not one draw from
// a single call.
func ARC4Random(p []byte) {
	for len(p) > 0 {
		r, _, _ := syscall_syscall(abi.FuncPCABI0(libc_arc4random_trampoline), 0, 0, 0)
		// arc4random returns a 32-bit value; the upper half of R3 is not defined.
		v := uint32(r)
		for n := 0; n < 4 && len(p) > 0; n++ {
			p[0] = byte(v)
			p = p[1:]
			v >>= 8
		}
	}
}
