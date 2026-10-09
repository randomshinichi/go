// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package runtime

//go:cgo_import_dynamic libc_gettimeofday gettimeofday "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_arc4random arc4random "/usr/lib/libSystem.B.dylib"

//go:nosplit
func cputicks() int64 {
	// Darwin's monotonic clock is a suitable profiler approximation. Do not
	// read the PPC time-base register: it is not part of the G3/G4 CPU floor.
	return nanotime()
}
