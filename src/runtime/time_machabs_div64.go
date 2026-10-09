// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(386 || arm || mips || mipsle || ppc)

package runtime

// machDivmod returns n/d and n%d. A 64-bit target divides in hardware, so
// this makes no call. A zero d panics with an integer divide error.
//
//go:nosplit
func machDivmod(n uint64, d uint32) (q, r uint64) {
	return n / uint64(d), n % uint64(d)
}
