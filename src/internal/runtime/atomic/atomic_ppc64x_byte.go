// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (ppc64 && !ppc64.ppc970) || ppc64le

package atomic

//go:noescape
func Xchg8(ptr *uint8, new uint8) uint8

//go:noescape
func And8(ptr *uint8, val uint8)

//go:noescape
func Or8(ptr *uint8, val uint8)
