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

	// Four pointer-sized words: the link register save, plus the three extra
	// reserved words described in goarch.go (compiler, link editor, TOC save).
	// The Darwin ppc32 linkage-area layout is an open question (ppc32 design
	// Q-3, "FixedFrameSize / linkage-area layout"); this value is the ppc64
	// derivation expressed in 32-bit words and must be replaced by the measured
	// one before a ppc32 runtime or backend consumes it.
	_MinFrameSize = 4 * PtrSize

	// Minimum stack alignment 4: the Darwin PPC32 convention recorded in
	// docs/darwin-abi-reference.md ("4-byte slots; minimum stack alignment 4;
	// vectors 16"), matching the other 32-bit ports.
	_StackAlign = PtrSize
)
