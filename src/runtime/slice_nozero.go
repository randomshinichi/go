// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc

package runtime

import (
	"unsafe"
)

//go:linkname bytealg_MakeNoZero internal/bytealg.MakeNoZero
func bytealg_MakeNoZero(len int) []byte {
	if uintptr(len) > maxAlloc {
		panicmakeslicelen()
	}
	cap := roundupsize(uintptr(len), true)
	return unsafe.Slice((*byte)(mallocgc(cap, nil, false)), cap)[:len]
}
