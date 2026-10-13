// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"bytes"
	"cmd/internal/obj"
	"internal/abi"
	"testing"
)

// The marking-only gate that used to live here (TestStackCheckWindowIsUnsafe and
// its broken-shape controls, bead goppc-1ye4) scored "stale-LR PCs not marked
// Unsafe" over a prologue that called runtime.stackcheck. It was a gate on the
// async-preemption consumer only: SIGPROF never reads the marking, so a prologue
// that marked its bad PCs Unsafe satisfied it while a profiling sample at the PC
// after the BL still threw "traceback stuck". The inline guard has no BL, and the
// gate that replaces it asserts the register STATE at every frameless PC for both
// consumers: see prologuestuck_test.go (goppc-h890, goppc-5vcc), whose
// counterfeit table includes this file's old shape as the "marking-only fix".

// stripUnsafeMarkers deletes every unsafe-point marker whose value is Unsafe,
// reproducing the pre-fix emitter. It reports how many it removed.
func stripUnsafeMarkers(text *obj.Prog) int {
	n := 0
	for prev := text; prev.Link != nil; {
		if p := prev.Link; isUnsafePointPCDATA(p) && p.To.Offset == abi.UnsafePointUnsafe {
			prev.Link = p.Link
			n++
			continue
		}
		prev = prev.Link
	}
	return n
}

func assembledSplitFixture(t *testing.T, locals int64) (*obj.LSym, *obj.Prog) {
	t.Helper()
	ctxt, s, text := splitFixture(t, locals)
	assemble(ctxt, s, nil)
	return s, text
}

// TestStackCheckWindowAddsNoCodeBytes: the marker is PCDATA only. The assembled
// bytes with and without it must be identical, so frame and nosplit accounting
// are untouched.
func TestStackCheckWindowAddsNoCodeBytes(t *testing.T) {
	for _, locals := range []int64{24, 65512} {
		s, _ := assembledSplitFixture(t, locals)
		ctxt, s2, text2 := splitFixture(t, locals)
		stripUnsafeMarkers(text2)
		assemble(ctxt, s2, nil)
		if !bytes.Equal(s.P, s2.P) {
			t.Errorf("locals=%d: marker changed the emitted code (%d vs %d bytes)", locals, len(s.P), len(s2.P))
		}
	}
}
