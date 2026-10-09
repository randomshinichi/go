// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"internal/testenv"
	"os"
	"runtime"
	"testing"
	"time"
)

const idleMainThreadEnv = "GO_PPROF_TEST_IDLE_MAIN_THREAD"

// With idleMainThreadEnv set, the test binary's main goroutine keeps the
// process's first thread for the whole run, so that thread sits idle while
// the tests run on other threads.
func init() {
	if os.Getenv(idleMainThreadEnv) == "1" {
		runtime.LockOSThread()
	}
}

// TestCPUProfileIdleMainThread checks that the CPU profile samples the thread
// that uses the CPU while the process's first thread is idle. Mac OS X 10.5
// sends every ITIMER_PROF SIGPROF to the first thread that does not block it,
// whatever that thread is doing, so a profiler that relies on the kernel to
// pick the running thread samples only the idle main thread there
// (go.dev/issue/6047).
func TestCPUProfileIdleMainThread(t *testing.T) {
	if os.Getenv(idleMainThreadEnv) != "1" {
		testenv.MustHaveExec(t)
		args := []string{"-test.run=^TestCPUProfileIdleMainThread$", "-test.v", "-test.timeout=0"}
		if testing.Short() {
			args = append(args, "-test.short")
		}
		cmd := testenv.Command(t, testenv.Executable(t), args...)
		cmd.Env = append(cmd.Environ(), idleMainThreadEnv+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatalf("subprocess failed: %v", err)
		}
		return
	}

	matches := matchAndAvoidStacks(stackContains, []string{"runtime/pprof.cpuHog1"}, avoidFunctions())
	testCPUProfile(t, matches, func(dur time.Duration) {
		cpuHogger(cpuHog1, &salt1, dur)
	})
}
