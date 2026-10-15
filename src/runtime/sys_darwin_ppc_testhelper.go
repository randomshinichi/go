//go:build darwin && ppc && arc4randomtest

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"unsafe"
)

// Arc4randomBufTrampolineTestCall invokes the raw libc trampoline with a caller-
// controlled argument block. It is included only in discriminator test builds.
func Arc4randomBufTrampolineTestCall(arg unsafe.Pointer) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(arc4random_buf_trampoline)), arg)
}
