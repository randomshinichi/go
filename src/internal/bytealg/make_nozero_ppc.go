// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package bytealg

// MakeNoZero is a safe fallback for ppc32. The runtime linkname allocator
// used by other architectures is not selected because its package declaration
// has no ppc32 assembly body; make still satisfies the API contract, though it
// initializes the returned bytes.
func MakeNoZero(n int) []byte {
	return make([]byte, n)
}
