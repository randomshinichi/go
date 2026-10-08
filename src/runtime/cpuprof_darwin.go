// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && !ppc64

package runtime

// mOSProf is empty: these kernels deliver each ITIMER_PROF tick to the thread
// that used the CPU, so the profiler keeps no per-thread state.
// darwin/ppc64 (Mac OS X 10.5) differs; see cpuprof_darwin_ppc64.go.
type mOSProf struct{}

func setProcessCPUProfiler(hz int32) {
	setProcessCPUProfilerTimer(hz)
}

func setThreadCPUProfiler(hz int32) {
	setThreadCPUProfilerHz(hz)
}
