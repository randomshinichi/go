// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "testing"

// These joins are intentionally kept out of line: PPC64's branch-elimination
// pass lowers them to ISEL, and the ppc970 backend must preserve their meaning.
//go:noinline
func ppc970SelectBool(cond bool, yes, no uint64) uint64 {
	v := no
	if cond {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectLT(a, b, yes, no uint64) uint64 {
	v := no
	if a < b {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectLE(a, b, yes, no uint64) uint64 {
	v := no
	if a <= b {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectGT(a, b, yes, no uint64) uint64 {
	v := no
	if a > b {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectGE(a, b, yes, no uint64) uint64 {
	v := no
	if a >= b {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectEQ(a, b, yes, no uint64) uint64 {
	v := no
	if a == b {
		v = yes
	}
	return v
}

//go:noinline
func ppc970SelectNE(a, b, yes, no uint64) uint64 {
	v := no
	if a != b {
		v = yes
	}
	return v
}

func TestPPC970ISELMaskSelect(t *testing.T) {
	const yes = 0x1122334455667788
	const no = 0x8877665544332211
	check := func(name string, got, want uint64) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %#016x, want %#016x", name, got, want)
		}
	}
	check("bool true", ppc970SelectBool(true, yes, no), yes)
	check("bool false", ppc970SelectBool(false, yes, no), no)
	for _, tc := range []struct {
		a, b uint64
		lt, le, gt, ge, eq, ne uint64
	}{
		{a: 1, b: 2, lt: yes, le: yes, gt: no, ge: no, eq: no, ne: yes},
		{a: 2, b: 1, lt: no, le: no, gt: yes, ge: yes, eq: no, ne: yes},
		{a: 2, b: 2, lt: no, le: yes, gt: no, ge: yes, eq: yes, ne: no},
	} {
		check("<", ppc970SelectLT(tc.a, tc.b, yes, no), tc.lt)
		check("<=", ppc970SelectLE(tc.a, tc.b, yes, no), tc.le)
		check(">", ppc970SelectGT(tc.a, tc.b, yes, no), tc.gt)
		check(">=", ppc970SelectGE(tc.a, tc.b, yes, no), tc.ge)
		check("==", ppc970SelectEQ(tc.a, tc.b, yes, no), tc.eq)
		check("!=", ppc970SelectNE(tc.a, tc.b, yes, no), tc.ne)
	}
}
