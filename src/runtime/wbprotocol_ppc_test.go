// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
)

// TestPPC32WriteBarrierProtocol checks the compiler-private contract between
// PPCOps.go's LoweredWB (result in R29; clobbers only R18, R19, R31, FPRs and
// flags) and runtime/wb_ppc.s, against the real runtime layout and the real
// wbBufFlush. A register the compiler believes survives the call but the
// runtime does not restore is silent corruption in whatever value the
// register allocator kept there, and only on the rare flush path.
func TestPPC32WriteBarrierProtocol(t *testing.T) {
	// 8 entries per call: two full buffers plus one call forces >= 2 flushes.
	iters := uint32(2*runtime.PPC32WBBufEntries/8 + 1)
	bad, flushes := runtime.PPC32WBProtocolCheck(iters)
	switch {
	case bad == 29:
		t.Fatalf("gcWriteBarrier8 returned R29 that is not 8 free entries ending at p.wbBuf.next (after %d flushes)", flushes)
	case bad == 30:
		t.Fatalf("gcWriteBarrier8 did not preserve g (R30) (after %d flushes)", flushes)
	case bad != 0:
		t.Fatalf("gcWriteBarrier8 did not preserve R%d, which LoweredWB promises the register allocator (after %d flushes)", bad, flushes)
	}
	if flushes == 0 {
		t.Fatalf("%d calls never took the flush path; the test did not exercise wbBufFlush", iters)
	}
	t.Logf("%d calls, %d flushes, all promised registers preserved", iters, flushes)
}
