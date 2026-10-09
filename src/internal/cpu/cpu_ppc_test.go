// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package cpu_test

import (
	"internal/cpu"
	"testing"
)

func TestPPC32CPUFoundation(t *testing.T) {
	if cpu.CacheLinePadSize != 128 {
		t.Fatalf("CacheLinePadSize = %d, want 128", cpu.CacheLinePadSize)
	}
	cpu.Initialize("")
}
