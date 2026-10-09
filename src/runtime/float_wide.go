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
	if y>>38 == 0 {
		// At most 38 significant bits: binary64 represents y exactly.
		return float32(uint64tofloat64(y))
	}
	// Round to odd before converting to binary64. The shifted integer has
	// at most 52 bits, so both its conversion and power-of-two scaling are
	// exact. In the original integer, bit 12 is strictly below the binary32
	// guard bit (at least bit 14 for y >= 2^38). Jamming the discarded bits
	// there preserves the guard and sticky information for the sole rounding
	// to binary32, including values just above or below an even midpoint.
	sticky := uint64(0)
	if y&(1<<12-1) != 0 {
		sticky = 1
	}
	y = y>>12 | sticky
	return float32(uint64tofloat64(y) * (1 << 12))
}
