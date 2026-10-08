// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goarch

const (
	_ArchFamily = PPC

	// Darwin allocates in 4 KiB pages on PowerPC. This default is not
	// load-bearing: runtime.getPageSize (os_darwin.go) replaces physPageSize
	// with sysctl hw.pagesize before the runtime allocates. The value has not
	// been measured on this project's target machine.
	_DefaultPhysPageSize = 4096

	_PCQuantum = 4

	// Keep in sync with sys.ArchPPC.FixedFrameSize: the measured Darwin ppc32
	// linkage area ends at SP+24 (docs/darwin-ppc32-abi-measured.md).
	// This is Go's stack ABI, not a promise of Linux SysV C interoperability.
	_MinFrameSize = 24

	// Total frames, including the linkage area, are aligned at call boundaries.
	_StackAlign = 16
)
