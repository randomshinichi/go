// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"runtime"
	"testing"
)

type ppcSmallFrameNode struct {
	words [64]uint32
}

var ppcSmallFrameRoot, ppcSmallFrameSink *ppcSmallFrameNode

//go:noinline
func preparePPCSmallFrameRoot() {
	ppcSmallFrameRoot = new(ppcSmallFrameNode)
	ppcSmallFrameRoot.words[3] = 0x12345678
}

//go:noinline
func churnPPCSmallFrameNodes() {
	for i := 0; i < 10000; i++ {
		p := new(ppcSmallFrameNode)
		p.words[3] = 0xdeadbeef
		ppcSmallFrameSink = p
	}
}

// retainPPCSmallFrameRoot has no pointer arguments and only one pointer local.
// Its 32-byte frame contains 24 linkage bytes and the pointer spill at SP+28.
// Once the global is cleared, the local map is the object's only GC root.
//
//go:noinline
func retainPPCSmallFrameRoot() uint32 {
	p := ppcSmallFrameRoot
	ppcSmallFrameRoot = nil
	runtime.GC()
	churnPPCSmallFrameNodes()
	return p.words[3]
}

func TestPPCSmallFrameStackRoot(t *testing.T) {
	defer func() {
		ppcSmallFrameRoot = nil
		ppcSmallFrameSink = nil
	}()
	for i := 0; i < 3; i++ {
		preparePPCSmallFrameRoot()
		if got := retainPPCSmallFrameRoot(); got != 0x12345678 {
			t.Fatalf("iteration %d: local root lost across GC: got %#x, want %#x", i, got, uint32(0x12345678))
		}
	}
}
