// Inferno's libkern/vlrt-arm.c
// https://bitbucket.org/inferno-os/inferno-os/src/master/libkern/vlrt-arm.c
//
//         Copyright © 1994-1999 Lucent Technologies Inc. All rights reserved.
//         Revisions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com).  All rights reserved.
//         Portions Copyright 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

//go:build arm || 386 || mips || mipsle || ppc

package runtime

const (
	sign32 = 1 << (32 - 1)
	sign64 = 1 << (64 - 1)
)

func uint64div(n, d uint64) uint64 {
	// Check for 32 bit operands
	if uint32(n>>32) == 0 && uint32(d>>32) == 0 {
		if uint32(d) == 0 {
			panicdivide()
		}
		return uint64(uint32(n) / uint32(d))
	}
	q, _ := dodiv(n, d)
	return q
}

func uint64mod(n, d uint64) uint64 {
	// Check for 32 bit operands
	if uint32(n>>32) == 0 && uint32(d>>32) == 0 {
		if uint32(d) == 0 {
			panicdivide()
		}
		return uint64(uint32(n) % uint32(d))
	}
	_, r := dodiv(n, d)
	return r
}

func int64div(n, d int64) int64 {
	// Check for 32 bit operands
	if int64(int32(n)) == n && int64(int32(d)) == d {
		if int32(n) == -0x80000000 && int32(d) == -1 {
			// special case: 32-bit -0x80000000 / -1 = -0x80000000,
			// but 64-bit -0x80000000 / -1 = 0x80000000.
			return 0x80000000
		}
		if int32(d) == 0 {
			panicdivide()
		}
		return int64(int32(n) / int32(d))
	}

	nneg := n < 0
	dneg := d < 0
	if nneg {
		n = -n
	}
	if dneg {
		d = -d
	}
	uq, _ := dodiv(uint64(n), uint64(d))
	q := int64(uq)
	if nneg != dneg {
		q = -q
	}
	return q
}

//go:nosplit
func int64mod(n, d int64) int64 {
	// Check for 32 bit operands
	if int64(int32(n)) == n && int64(int32(d)) == d {
		if int32(d) == 0 {
			panicdivide()
		}
		return int64(int32(n) % int32(d))
	}

	nneg := n < 0
	if nneg {
		n = -n
	}
	if d < 0 {
		d = -d
	}
	_, ur := dodiv(uint64(n), uint64(d))
	r := int64(ur)
	if nneg {
		r = -r
	}
	return r
}

//go:noescape
func _div64by32(a uint64, b uint32, r *uint32) (q uint32)

//go:nosplit
func dodiv(n, d uint64) (q, r uint64) {
	if GOARCH == "arm" || GOARCH == "ppc" {
		// arm lacks division instructions, and ppc32 has no helper for
		// 64-by-64 division; slowdodiv is the portable fallback.
		return slowdodiv(n, d)
	}

	if GOARCH == "mips" || GOARCH == "mipsle" {
		// No _div64by32 on mips and using only _mul64by32 doesn't bring much benefit
		return slowdodiv(n, d)
	}

	if d > n {
		return 0, n
	}

	if uint32(d>>32) != 0 {
		t := uint32(n>>32) / uint32(d>>32)
		var lo64 uint64
		hi32 := _mul64by32(&lo64, d, t)
		if hi32 != 0 || lo64 > n {
			return slowdodiv(n, d)
		}
		return uint64(t), n - lo64
	}

	// d is 32 bit
	var qhi uint32
	if uint32(n>>32) >= uint32(d) {
		if uint32(d) == 0 {
			panicdivide()
		}
		qhi = uint32(n>>32) / uint32(d)
		n -= uint64(uint32(d)*qhi) << 32
	} else {
		qhi = 0
	}

	var rlo uint32
	qlo := _div64by32(n, uint32(d), &rlo)
	return uint64(qhi)<<32 + uint64(qlo), uint64(rlo)
}

//go:nosplit
func slowdodiv(n, d uint64) (q, r uint64) {
	if d == 0 {
		panicdivide()
	}

	// Set up the divisor and find the number of iterations needed.
	capn := n
	if n >= sign64 {
		capn = sign64
	}
	i := 0
	for d < capn {
		d <<= 1
		i++
	}

	for ; i >= 0; i-- {
		q <<= 1
		if n >= d {
			n -= d
			q |= 1
		}
		d >>= 1
	}
	return q, n
}

// Floating point control word values.
// Bits 0-5 are bits to disable floating-point exceptions.
// Bits 8-9 are the precision control:
//
//	0 = single precision a.k.a. float32
//	2 = double precision a.k.a. float64
//
// Bits 10-11 are the rounding mode:
//
//	0 = round to nearest (even on a tie)
//	3 = round toward zero
var (
	controlWord64      uint16 = 0x3f + 2<<8 + 0<<10
	controlWord64trunc uint16 = 0x3f + 2<<8 + 3<<10
)
