// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import "testing"

var s int

func BenchmarkCallClosure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s += func(ii int) int { return 2 * ii }(i)
	}
}

func BenchmarkCallClosure1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i
		s += func(ii int) int { return 2*ii + j }(i)
	}
}

var ss *int

func BenchmarkCallClosure2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		j := i
		s += func() int {
			ss = &j
			return 2
		}()
	}
}

func addr1(x int) *int {
	return func() *int { return &x }()
}

func BenchmarkCallClosure3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ss = addr1(i)
	}
}

func addr2() (x int, p *int) {
	return 0, func() *int { return &x }()
}

func BenchmarkCallClosure4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, ss = addr2()
	}
}

// closureCtxClobber is called from inside a closure before that closure first
// reads one of its captured variables. It allocates, so it is a real call that
// is free to use (and thereby destroy) whatever register the target ABI uses to
// pass the closure context: on darwin/ppc32 that is R11, which
// OpPPCLoweredGetClosurePtr reads with `OR R11,R11,R11`.
//
//go:noinline
func closureCtxClobber(x int32) int32 {
	closureCtxSink = make([]int32, x+1)
	return int32(len(closureCtxSink))
}

var closureCtxSink []int32
var closureCtxFn func(int32) int32

// TestClosureCtxReadAfterCall pins the placement of the closure-context value.
// The first use of a captured variable must be scheduled where the context is
// still live, not in the block after a call. When the closure-context value was
// not recognised as one that must stay in the entry block (ppc32's
// OpPPCLoweredGetClosurePtr was missing from ssa.(*Op).isLoweredGetClosurePtr
// before CL 4-figure a73b980439-equivalent), the read of a[3] was moved behind
// the call and loaded from the callee's leftover context register, a fault or a
// wrong value at the small capture offset. See goppc-if2l.
func TestClosureCtxReadAfterCall(t *testing.T) {
	var a [4]int32
	a[0], a[1], a[2], a[3] = 1, 2, 3, 0x33333333
	closureCtxFn = func(k int32) int32 {
		closureCtxClobber(k)
		return a[3]
	}
	if got := closureCtxFn(1); got != 0x33333333 {
		t.Fatalf("a[3] = %#x, want %#x (closure context read after a call)", got, int32(0x33333333))
	}
}
