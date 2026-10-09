// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

// _mul64by32 returns the high 32 bits of a 96-bit product and stores the low
// 64 bits through lo64. Keeping the word arithmetic in Go makes the operation
// independent of the target's byte order.
func _mul64by32(lo64 *uint64, a uint64, b uint32) (hi32 uint32) {
	loProduct := uint64(uint32(a)) * uint64(b)
	hiProduct := uint64(uint32(a>>32)) * uint64(b)
	lo := loProduct + hiProduct<<32
	*lo64 = lo
	hi32 = uint32(hiProduct >> 32)
	if lo < loProduct {
		hi32++
	}
	return hi32
}
