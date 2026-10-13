// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc

package runtime

// mutexProfilingOn is constant off ppc: where 64-bit atomics are lock-free or
// have native support, mutexSampleContention's read of mutexprofilerate cannot
// block, so there is no flag to consult. See mutexprofiling_ppc.go for why ppc32
// needs one (goppc-7tqu).
func mutexProfilingOn() bool { return true }

// setMutexProfiling has nothing to maintain off ppc.
func setMutexProfiling(on bool) {}
