// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
	"time"
)

// TestUsleepDuration pins the usec -> timespec arithmetic in the linux/ppc
// usleep assembly. With the operands in the wrong order the timespec was built
// from whatever the caller left in R4/R6: the sleep was 1us, failed with
// EINVAL, or lasted years. Each of those misses the lower bound or the
// deadline. nanosleep never returns early except on a signal, so an attempt
// may fall short only when interrupted; one full-length attempt is required.
func TestUsleepDuration(t *testing.T) {
	for _, usec := range []uint32{2000, 5000, 9000} {
		want := time.Duration(usec) * time.Microsecond
		done := make(chan time.Duration, 1)
		go func() {
			var best time.Duration
			for range 5 {
				start := time.Now()
				runtime.Usleep(usec)
				if d := time.Since(start); d > best {
					best = d
				}
				if best >= want {
					break
				}
			}
			done <- best
		}()
		select {
		case got := <-done:
			if got < want {
				t.Errorf("usleep(%d) slept at most %v in 5 attempts, want >= %v", usec, got, want)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("usleep(%d) did not return within 10s", usec)
		}
	}
}
