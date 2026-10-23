// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && !goppc_atcompat_disabled

package unix_test

import (
	"internal/syscall/unix"
	"testing"
)

// TestAtCompatEnabledOnMeasuredKernelDarwinPPC requires the anchored backend to
// be enabled on the kernel it was designed and measured against. These Unix-layer
// tests only make sense on that regime, so a false answer here is a real failure,
// not a skip. Its counterpart — the same assertion inverted, for the build that
// forces the gate off — is in at_disabled_darwin_ppc_test.go, which is why the
// build tag on this file and that one are exact complements.
func TestAtCompatEnabledOnMeasuredKernelDarwinPPC(t *testing.T) {
	if !unix.AtCompatEnabledForTest() {
		t.Fatalf("the anchored backend is disabled on this kernel; the anchored tests only make sense on the reviewed 9.8.x regime")
	}
}
