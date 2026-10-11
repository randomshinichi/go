// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic_test

import (
	. "sync/atomic"
	"testing"
	"time"
)

// TestNil64DoesNotPoisonLock guards the lock-based 32-bit 64-bit-atomic
// implementations (ppc, and the other ports that serialize 64-bit operations
// through a global lock). A nil operand must fault while that lock is still
// free. If the fault happens after the lock is taken, the panic unwinds out of
// a held lock, and every subsequent 64-bit atomic operation spins forever in
// the lock acquire — a recovered nil call would then wedge the whole program,
// not just the nil path.
//
// For each primitive it recovers exactly one nil call, then performs a valid
// 64-bit operation, and requires that the valid operation complete. The valid
// operation is the assertion: it is reached only if the nil call released the
// lock. A watchdog turns a resulting deadlock into a test failure instead of a
// whole-suite timeout.
func TestNil64DoesNotPoisonLock(t *testing.T) {
	primitives := []struct {
		name string
		nil  func() // a call that must panic on the nil operand
		valid func() // a call on a real, 8-aligned operand
	}{
		{"SwapInt64", func() { SwapInt64(nil, 1) }, func() { var x int64; SwapInt64(&x, 1) }},
		{"AddInt64", func() { AddInt64(nil, 1) }, func() { var x int64; AddInt64(&x, 1) }},
		{"LoadInt64", func() { LoadInt64(nil) }, func() { var x int64; LoadInt64(&x) }},
		{"StoreInt64", func() { StoreInt64(nil, 1) }, func() { var x int64; StoreInt64(&x, 1) }},
		{"CompareAndSwapInt64", func() { CompareAndSwapInt64(nil, 0, 1) }, func() { var x int64; CompareAndSwapInt64(&x, 0, 1) }},
		{"SwapUint64", func() { SwapUint64(nil, 1) }, func() { var x uint64; SwapUint64(&x, 1) }},
		{"AddUint64", func() { AddUint64(nil, 1) }, func() { var x uint64; AddUint64(&x, 1) }},
		{"LoadUint64", func() { LoadUint64(nil) }, func() { var x uint64; LoadUint64(&x) }},
		{"StoreUint64", func() { StoreUint64(nil, 1) }, func() { var x uint64; StoreUint64(&x, 1) }},
		{"CompareAndSwapUint64", func() { CompareAndSwapUint64(nil, 0, 1) }, func() { var x uint64; CompareAndSwapUint64(&x, 0, 1) }},
	}

	for _, p := range primitives {
		t.Run(p.name, func(t *testing.T) {
			// The nil call must actually panic; recovering a call that silently
			// returns would hide a swallowed-nil counterfeit.
			panicked := func() (ok bool) {
				defer func() {
					if recover() != nil {
						ok = true
					}
				}()
				p.nil()
				return false
			}()
			if !panicked {
				t.Fatalf("%s(nil) did not panic", p.name)
			}

			// A valid operation afterwards must complete. Run it under a
			// watchdog: if the nil call leaked the global lock, this blocks.
			done := make(chan struct{})
			go func() {
				defer close(done)
				p.valid()
			}()
			select {
			case <-done:
				// Lock was free; the nil call did not poison it.
			case <-time.After(2 * time.Second):
				t.Fatalf("%s(nil) poisoned the 64-bit atomic lock: a subsequent valid operation deadlocked", p.name)
			}
		})
	}
}
