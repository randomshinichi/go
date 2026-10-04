// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64 && ppc64.ppc970

package atomic_test

import (
	"internal/runtime/atomic"
	"testing"
	"unsafe"
)

// Taking the functions as values defeats intrinsification, so these calls reach
// the Go bodies in atomic_ppc970.go, which the compiler's own LoweredAtomic*8PPC970
// sequences otherwise shadow.
var (
	xchg8Func = atomic.Xchg8
	and8Func  = atomic.And8
	or8Func   = atomic.Or8
)

func TestPPC970ByteBodies(t *testing.T) {
	var backing [2]uint64
	b := (*[16]byte)(unsafe.Pointer(&backing))
	vals := []uint8{0x00, 0x01, 0x7f, 0x80, 0xa5, 0xff}
	for off := 0; off < 16; off++ {
		for _, v := range vals {
			for _, op := range []string{"xchg8", "and8", "or8"} {
				for i := range b {
					b[i] = byte(0x11*i + 3)
				}
				want := *b
				old := want[off]
				p := &b[off]
				switch op {
				case "xchg8":
					if got := xchg8Func(p, v); got != old {
						t.Fatalf("Xchg8 off %d: returned %#x, want old %#x", off, got, old)
					}
					want[off] = v
				case "and8":
					and8Func(p, v)
					want[off] = old & v
				case "or8":
					or8Func(p, v)
					want[off] = old | v
				}
				if *b != want {
					t.Fatalf("%s off %d val %#x: memory %x, want %x (neighbours must be untouched)", op, off, v, *b, want)
				}
			}
		}
	}
}
