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

// These recipes are shared with the 32-bit targets that formerly defined
// them in vlrt.go. Keeping the single-precision path separate is essential:
// rounding a wide integer to binary64 first can lose a binary32 sticky bit.
func int64tofloat64(y int64) float64 {
	if y < 0 {
		return -uint64tofloat64(-uint64(y))
	}
	return uint64tofloat64(uint64(y))
}

func uint64tofloat64(y uint64) float64 {
	hi := float64(uint32(y >> 32))
	lo := float64(uint32(y))
	d := hi*(1<<32) + lo
	return d
}

func int64tofloat32(y int64) float32 {
	if y < 0 {
		return -uint64tofloat32(-uint64(y))
	}
	return uint64tofloat32(uint64(y))
}

func uint64tofloat32(y uint64) float32 {
	// divide into top 18, mid 23, and bottom 23 bits.
	// (23-bit integers fit into a float32 without loss.)
	top := uint32(y >> 46)
	mid := uint32(y >> 23 & (1<<23 - 1))
	bot := uint32(y & (1<<23 - 1))
	if top == 0 {
		return float32(mid)*(1<<23) + float32(bot)
	}
	if bot != 0 {
		// Top is not zero, so the bits in bot
		// won't make it into the final mantissa.
		// In fact, the bottom bit of mid won't
		// make it into the mantissa either.
		// We only need to make sure that if top+mid
		// is about to round down in a round-to-even
		// scenario, and bot is not zero, we make it
		// round up instead.
		mid |= 1
	}
	return float32(top)*(1<<46) + float32(mid)*(1<<23)
}
