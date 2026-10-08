// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"go/types"
	"testing"
)

func TestPPCSizes(t *testing.T) {
	for _, arch := range []string{"ppc", "ppc64", "ppc64le"} {
		s := types.SizesFor("gc", arch)
		if s == nil {
			t.Fatalf("missing gc sizes for %s", arch)
		}
		word := int64(8)
		if arch == "ppc" {
			word = 4
		}
		for _, kind := range []types.BasicKind{types.Int, types.Uint, types.Uintptr, types.UnsafePointer} {
			if got := s.Sizeof(types.Typ[kind]); got != word {
				t.Errorf("%s Sizeof(%s) = %d, want %d", arch, types.Typ[kind], got, word)
			}
		}
		for _, kind := range []types.BasicKind{types.Int64, types.Float64} {
			if got := s.Alignof(types.Typ[kind]); got != word {
				t.Errorf("%s Alignof(%s) = %d, want %d", arch, types.Typ[kind], got, word)
			}
		}
	}
}
