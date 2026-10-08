// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"cmd/compile/internal/types2"
	"testing"
)

func TestPPCSizes(t *testing.T) {
	for _, arch := range []string{"ppc", "ppc64", "ppc64le"} {
		s := types2.SizesFor("gc", arch)
		if s == nil {
			t.Fatalf("missing gc sizes for %s", arch)
		}
		word := int64(8)
		if arch == "ppc" {
			word = 4
		}
		for _, kind := range []types2.BasicKind{types2.Int, types2.Uint, types2.Uintptr, types2.UnsafePointer} {
			if got := s.Sizeof(types2.Typ[kind]); got != word {
				t.Errorf("%s Sizeof(%s) = %d, want %d", arch, types2.Typ[kind], got, word)
			}
		}
		for _, kind := range []types2.BasicKind{types2.Int64, types2.Float64} {
			if got := s.Alignof(types2.Typ[kind]); got != word {
				t.Errorf("%s Alignof(%s) = %d, want %d", arch, types2.Typ[kind], got, word)
			}
		}
	}
}
