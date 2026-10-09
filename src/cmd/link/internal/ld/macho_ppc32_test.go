// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"testing"
)

func TestPPC32UnixThreadState(t *testing.T) {
	const entry = uint32(0x12345678)
	ml := newPPC32ThreadLoad(entry)
	if !validPPC32ThreadLoad(ml, entry) {
		t.Fatal("measured thread state rejected")
	}
	// Negative control: imitate the otool-display-order mistake by moving
	// srr0 from struct word 0 to word 38. The structural check must reject it.
	mutant := *ml
	mutant.data = append([]uint32(nil), ml.data...)
	mutant.data[2], mutant.data[40] = mutant.data[40], entry
	if validPPC32ThreadLoad(&mutant, entry) {
		t.Fatal("wrong srr0-at-word-38 control was accepted")
	}
	if got, want := len(ml.data), 42; got != want {
		t.Fatalf("load data words = %d, want %d", got, want)
	}
	if got := 8 + len(ml.data)*4; got != 176 {
		t.Fatalf("LC_UNIXTHREAD cmdsize = %d, want 176", got)
	}
	if ml.data[0] != 1 {
		t.Errorf("thread flavor = %d, want PPC_THREAD_STATE (1)", ml.data[0])
	}
	if ml.data[1] != 40 {
		t.Errorf("thread count = %d, want PPC_THREAD_STATE_COUNT (40)", ml.data[1])
	}
	if ml.data[2] != entry {
		t.Errorf("srr0/data[2] = %#x, want %#x", ml.data[2], entry)
	}
	if ml.data[40] != 0 || ml.data[41] != 0 {
		t.Errorf("unexpected nonzero state tail: %#x %#x", ml.data[40], ml.data[41])
	}
}

func validPPC32ThreadLoad(ml *MachoLoad, entry uint32) bool {
	return len(ml.data) == 42 &&
		8+len(ml.data)*4 == 176 &&
		ml.data[0] == 1 &&
		ml.data[1] == 40 &&
		ml.data[2] == entry
}
