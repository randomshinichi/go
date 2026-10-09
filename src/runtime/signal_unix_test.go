// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package runtime_test

import (
	"runtime"
	"testing"
)

// TestSigFetchGInC checks how the signal handler finds g on darwin/ppc when
// the signal interrupted C code, where R30 does not hold g: the signal stack
// names the m, and the interrupted SP picks its g0 or gsignal. An SP that
// fits neither must give nil, so the handler reports a bad g instead of
// trusting a guess.
func TestSigFetchGInC(t *testing.T) {
	onG0, onGsignal, onNeither, offSignalStack := runtime.SigFetchGInCResults()
	if !onG0 {
		t.Error("interrupted SP on g0: did not return the m's g0")
	}
	if !onGsignal {
		t.Error("interrupted SP on gsignal: did not return the m's gsignal")
	}
	if !onNeither {
		t.Error("interrupted SP on a goroutine stack: did not return nil")
	}
	if !offSignalStack {
		t.Error("handler SP on no signal stack: did not return nil")
	}
}
