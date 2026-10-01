// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "testing"

func TestPPC64BswapFusion(t *testing.T) {
	for _, tc := range []struct {
		arch  string
		level int
		want  bool
	}{
		{arch: "ppc64", level: 5, want: false}, // ppc970 has no byte-reversed loads/stores.
		{arch: "ppc64", level: 8, want: true},  // Default power8 must retain its output.
		{arch: "ppc64", level: 9, want: true},
		{arch: "ppc64", level: 10, want: true},
		{arch: "ppc64le", level: 8, want: true}, // Default ppc64le is unchanged.
	} {
		if got := ppc64HaveBswapFusion(tc.arch, tc.level); got != tc.want {
			t.Errorf("ppc64HaveBswapFusion(%q, %d) = %v, want %v", tc.arch, tc.level, got, tc.want)
		}
	}
}
