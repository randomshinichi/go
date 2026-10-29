// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build math_big_pure_go || ppc

package big

import _ "unsafe" // for linkname

// The pure-Go path must provide the same entry points as arith_decl.go:
// widely used packages access these internal details using linkname.
// See go.dev/issue/67401.

// shlVU should be an internal detail (and a stale one at that),
// but widely used packages access it using linkname.
// Notable members of the hall of shame include:
//   - github.com/remyoudompheng/bigfft
//
// Do not remove or change the type signature.
// See go.dev/issue/67401.
//
//go:linkname shlVU
func shlVU(z, x []Word, s uint) (c Word) {
	if s == 0 {
		copy(z, x)
		return 0
	}
	return lshVU(z, x, s)
}

// addMulVVW should be an internal detail (and a stale one at that),
// but widely used packages access it using linkname.
// Notable members of the hall of shame include:
//   - github.com/remyoudompheng/bigfft
//
// Do not remove or change the type signature.
// See go.dev/issue/67401.
//
//go:linkname addMulVVW
func addMulVVW(z, x []Word, y Word) (c Word) {
	return addMulVVWW(z, z, x, y, 0)
}
