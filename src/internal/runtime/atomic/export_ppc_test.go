// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic

import "unsafe"

// Test hooks for the ppc32 64-bit atomic stripes (goppc-nzqrh).

const Stripes = atomic64Stripes

var (
	StripeIndex      = atomic64StripeIndex
	StripeAcquire    = atomic64StripeAcquire
	StripeRelease    = atomic64StripeRelease
	StripeAcquireAll = atomic64StripeAcquireAll
	StripeReleaseAll = atomic64StripeReleaseAll
	AnyStripeHeld    = Atomic64AnyStripeHeld
)

// StripeWord is the lock word of stripe i: 1 while held.
func StripeWord(i int) uint32 { return atomic64Table[i].l }

// StripeAddr is the address of the lock word of stripe i.
func StripeAddr(i int) uintptr { return uintptr(unsafe.Pointer(&atomic64Table[i].l)) }
