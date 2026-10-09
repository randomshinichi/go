// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package bytealg_test

import (
	"internal/bytealg"
	"testing"
)

func TestMakeNoZero(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, 15, 16, 31, 32, 33, 63, 64, 65} {
		b := bytealg.MakeNoZero(n)
		if len(b) != n || cap(b) < n {
			t.Fatalf("MakeNoZero(%d) = len %d cap %d", n, len(b), cap(b))
		}
		if n != 0 {
			b[n-1] = 0xa5
			if b[n-1] != 0xa5 {
				t.Fatalf("MakeNoZero(%d) returned unwritable memory", n)
			}
		}
	}
}
