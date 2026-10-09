// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || arm || mips || mipsle || ppc

package runtime

// machDivmod returns n/d and n%d without calling a function that has a stack
// check: here a 64-bit / or % would call runtime.uint64div or uint64mod, while
// dodiv is nosplit. A zero d panics with an integer divide error.
//
//go:nosplit
func machDivmod(n uint64, d uint32) (q, r uint64) {
	return dodiv(n, uint64(d))
}
