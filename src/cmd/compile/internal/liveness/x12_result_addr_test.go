// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package liveness

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/bitvec"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
	"cmd/internal/sys"
	"testing"
)

func init() {
	// The same setup that ssa's in-package tests do (ssa/export_test.go:115):
	// enough of the type universe and the linker context for named variables
	// and for the type sizes liveness asks about.
	types.PtrSize = 8
	types.RegSize = 8
	types.MaxWidth = 1 << 50
	base.Ctxt = &obj.Link{Arch: &obj.LinkArch{Arch: &sys.Arch{Alignment: 1, CanMergeLoads: true}}}
	typecheck.InitUniverse()
}

// entryUevar runs the block effects of prologue() over an entry block holding
// the given values in order, with result as the tracked variable, and reports
// whether result comes out upward exposed. An upward exposed result is what
// prologue() feeds into the entry check in solve(), which rejects it as "bad
// live variable at entry" -- the compiler error this file is about.
//
// What liveness reacts to is the sequence of effects, so the order of the
// values is the order the scheduler leaves them in.
func entryUevar(t *testing.T, values []*ssa.Value, result *ir.Name) bool {
	t.Helper()
	b := &ssa.Block{ID: 0, Values: values}
	lv := Liveness{
		f:    &ssa.Func{Blocks: []*ssa.Block{b}},
		vars: []*ir.Name{result},
		idx:  map[*ir.Name]int32{result: 0},
		be:   []blockEffects{{uevar: bitvec.New(1), varkill: bitvec.New(1)}},
	}
	lv.prologue()
	return lv.be[0].uevar.Get(0)
}

// newResult returns a named result of pointer type, which is what liveness
// counts an address of as a use of (liveness.affectedVar).
func newResult() *ir.Name {
	r := ir.NewNameAt(src.NoXPos, &types.Sym{Name: "r"}, types.NewPtr(types.Types[types.TUINT8]))
	r.Class = ir.PPARAMOUT
	r.SetUsed(true)
	return r
}

// resultGraph returns the entry block of the graph in
// TestScheduleResultAddrAfterFirstKill (ssa/schedule_test.go): a symbolic
// store initializes the result, the address of the result is taken, the result
// is read back through that address, and the result is declared by a VarDef
// chained behind the store.
func resultGraph(r *ir.Name) []*ssa.Value {
	return []*ssa.Value{
		{Op: ssa.OpAMD64MOVQstore, Aux: r}, // r = zero
		{Op: ssa.OpAMD64LEAQ, Aux: r},      // &r
		{Op: ssa.OpAMD64MOVQload},          // *(&r)
		{Op: ssa.OpVarDef, Aux: r},         // the VarDef for r
		{Op: ssa.OpAMD64MOVQstore, Aux: r}, // r = *(&r)
	}
}

// TestX12ReviewInitializedResultNotLiveAtEntry is the liveness refutation from
// the review of goppc-xprg.12: the store that initializes the result kills it
// before the address is taken, so the result is not live at entry, and nothing
// in liveness would have rejected the graph. The value order is the one the
// fixed scheduler produces for it.
//
// Kept in out/ppc-mac/x12-evidence/review-counterexample as
// liveness_x12_review_test.go.
func TestX12ReviewInitializedResultNotLiveAtEntry(t *testing.T) {
	r := newResult()
	if entryUevar(t, resultGraph(r), r) {
		t.Fatal("initialized result incorrectly live at entry")
	}
	t.Log("LEAQ after the initializing store is not live at entry: the store kills r")
}

// TestX12ReviewResultAddrBeforeKillIsLiveAtEntry is the control for the test
// above and the reason the scheduler has to order the address at all: with the
// address of the result ahead of the store that kills it, the result is live
// on entry, which is the "bad live variable at entry" compiler error.
//
// The base scheduler produces exactly this order for the graph, so the base
// compiler cannot compile it either, whether or not it schedules the block.
func TestX12ReviewResultAddrBeforeKillIsLiveAtEntry(t *testing.T) {
	r := newResult()
	vs := resultGraph(r)
	// The address first, before the store that kills the result: the order
	// the base scheduler produces for this graph.
	vs[0], vs[1] = vs[1], vs[0]
	if !entryUevar(t, vs, r) {
		t.Fatal("result not live at entry although its address precedes every kill of it")
	}
	t.Log("an address taken before the first kill of the result is live at entry")
}
