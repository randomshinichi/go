// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package bytealg

// MakeNoZero is implemented by runtime.slice_nozero_ppc.go so ppc32 receives
// the same allocator-rounded capacity as other architectures.
func MakeNoZero(n int) []byte
