// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import "testing"

func TestCloneStackBoundary(t *testing.T) {
	if got := runTestProg(t, "testprog", "CloneStackBoundary"); got != "OK\n" {
		t.Fatalf("clone stack boundary: %s", got)
	}
}
