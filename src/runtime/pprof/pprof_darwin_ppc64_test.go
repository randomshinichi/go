// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"io"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestCPUProfileStartStopSchedulerContention exercises the sampler's rate
// changes while other Ms contend for sched.lock. SIGPROF can interrupt one
// of those Ms with sched.lock held, so the rate setter must not acquire
// sched.lock while holding prof.signalLock.
func TestCPUProfileStartStopSchedulerContention(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))
	stop := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					runtime.Gosched()
				}
			}
		}()
	}
	defer func() {
		close(stop)
		workers.Wait()
	}()
	for i := 0; i < 20; i++ {
		if err := StartCPUProfile(io.Discard); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(5+i%20) * time.Millisecond)
		StopCPUProfile()
	}
}
