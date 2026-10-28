// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package syscall_test

import (
	"syscall"
	"unsafe"
)

// Compile-time layout control for Fbootstraptransfer_t: a size other than 24
// makes one of the two array lengths negative and the package fail to build.
// ppc is the size measured on 10.5.8 (testdata/darwin_ppc32/xsys/xsys_sizes.txt);
// ppc64 is the LP64 shape shared with darwin/amd64.
var (
	_ [int(unsafe.Sizeof(syscall.Fbootstraptransfer_t{})) - 24]int
	_ [24 - int(unsafe.Sizeof(syscall.Fbootstraptransfer_t{}))]int
)
