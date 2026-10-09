// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"testing"

	"cmd/compile/internal/types"
	"cmd/internal/obj/ppc"
)

// GetG stays generic only when regalloc can bind it to the reserved g
// register. A false config flag or an allocatable R30 contradicts that ABI.
func TestPPCGetGConvention(t *testing.T) {
	c := testConfigPPC32(t)
	if !c.config.hasGReg {
		t.Fatal("ppc32 must use the dedicated R30 g register")
	}
	found := false
	for _, r := range c.config.registers {
		if r.name != "g" {
			continue
		}
		found = true
		if r.ObjNum() != ppc.REGG || c.config.gpRegMask&(1<<uint(r.num)) != 0 {
			t.Fatalf("g register = %v, obj=%d, gp mask=%x", r, r.ObjNum(), c.config.gpRegMask)
		}
	}
	if !found {
		t.Fatal("ppc32 has no named g register")
	}
	f := c.Fun("entry", Bloc("entry",
		Valu("mem", OpInitMem, types.TypeMem, 0, nil),
		Valu("g", OpGetG, c.config.Types.UInt32, 0, nil, "mem"),
		Valu("sp", OpSP, c.config.Types.UInt32, 0, nil),
		Valu("store", OpStore, types.TypeMem, 0, c.config.Types.UInt32, "sp", "g", "mem"),
		Exit("store")))
	lower(f.f)
	checkLower(f.f)
	if f.values["g"].Op != OpGetG {
		t.Fatalf("GetG became %v rather than a regalloc binding", f.values["g"].Op)
	}
}

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
