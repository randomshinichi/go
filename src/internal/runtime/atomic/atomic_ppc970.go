// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64 && ppc64.ppc970

package atomic

import "unsafe"

// The PPC970 has no byte-wide load-reserve/store-conditional (LBAR/STBCCC are
// Power ISA v2.06). The compiler lowers calls to these three functions to
// word-masked sequences (cmd/compile/internal/ssa/_gen/PPC64.rules,
// LoweredAtomic*8PPC970); these bodies serve any reference that is not
// intrinsified, using the 32-bit Cas, which is built from LWAR/STWCCC.

// byteLane returns the aligned word containing ptr and the bit shift of ptr's
// byte inside it. This file is big-endian only (ppc64), so byte 0 is the most
// significant byte of the word.
//
//go:nosplit
func byteLane(ptr *uint8) (*uint32, uint) {
	a := uintptr(unsafe.Pointer(ptr))
	return (*uint32)(unsafe.Pointer(a &^ 3)), uint((a&3)^3) * 8
}

//go:nosplit
func Xchg8(ptr *uint8, new uint8) uint8 {
	return goXchg8(ptr, new)
}

//go:nosplit
func And8(ptr *uint8, val uint8) {
	w, shift := byteLane(ptr)
	keep := ^(uint32(0xFF) << shift) | uint32(val)<<shift
	for {
		old := *w
		if Cas(w, old, old&keep) {
			return
		}
	}
}

//go:nosplit
func Or8(ptr *uint8, val uint8) {
	w, shift := byteLane(ptr)
	set := uint32(val) << shift
	for {
		old := *w
		if Cas(w, old, old|set) {
			return
		}
	}
}
