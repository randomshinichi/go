// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package platform

import "testing"

func TestDarwinDefaultPIE(t *testing.T) {
	for _, arch := range []string{"ppc", "ppc64", "amd64", "arm64"} {
		for _, race := range []bool{false, true} {
			want := arch == "amd64" || arch == "arm64"
			if got := DefaultPIE("darwin", arch, race); got != want {
				t.Errorf("DefaultPIE(darwin, %s, %t) = %t, want %t", arch, race, got, want)
			}
		}
	}
}
