// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"testing"
)

func TestSchedule(t *testing.T) {
	c := testConfig(t)
	cases := []fun{
		c.Fun("entry",
			Bloc("entry",
				Valu("mem0", OpInitMem, types.TypeMem, 0, nil),
				Valu("ptr", OpConst64, c.config.Types.Int64, 0xABCD, nil),
				Valu("v", OpConst64, c.config.Types.Int64, 12, nil),
				Valu("mem1", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem0"),
				Valu("mem2", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem1"),
				Valu("mem3", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "sum", "mem2"),
				Valu("l1", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem1"),
				Valu("l2", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem2"),
				Valu("sum", OpAdd64, c.config.Types.Int64, 0, nil, "l1", "l2"),
				Goto("exit")),
			Bloc("exit",
				Exit("mem3"))),
	}
	for _, c := range cases {
		schedule(c.f)
		if !isSingleLiveMem(c.f) {
			t.Error("single-live-mem restriction not enforced by schedule for func:")
			printFunc(c.f)
		}
	}
}

func isSingleLiveMem(f *Func) bool {
	for _, b := range f.Blocks {
		var liveMem *Value
		for _, v := range b.Values {
			for _, w := range v.Args {
				if w.Type.IsMemory() {
					if liveMem == nil {
						liveMem = w
						continue
					}
					if w != liveMem {
						return false
					}
				}
			}
			if v.Type.IsMemory() {
				liveMem = v
			}
		}
	}
	return true
}

func TestStoreOrder(t *testing.T) {
	// In the function below, v2 depends on v3 and v4, v4 depends on v3, and v3 depends on store v5.
	// storeOrder did not handle this case correctly.
	c := testConfig(t)
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("mem0", OpInitMem, types.TypeMem, 0, nil),
			Valu("a", OpAdd64, c.config.Types.Int64, 0, nil, "b", "c"),                        // v2
			Valu("b", OpLoad, c.config.Types.Int64, 0, nil, "ptr", "mem1"),                    // v3
			Valu("c", OpNeg64, c.config.Types.Int64, 0, nil, "b"),                             // v4
			Valu("mem1", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "v", "mem0"), // v5
			Valu("mem2", OpStore, types.TypeMem, 0, c.config.Types.Int64, "ptr", "a", "mem1"),
			Valu("ptr", OpConst64, c.config.Types.Int64, 0xABCD, nil),
			Valu("v", OpConst64, c.config.Types.Int64, 12, nil),
			Goto("exit")),
		Bloc("exit",
			Exit("mem2")))

	CheckFunc(fun.f)
	order := storeOrder(fun.f.Blocks[0].Values, fun.f.newSparseSet(fun.f.NumValues()), make([]int32, fun.f.NumValues()))

	// check that v2, v3, v4 is sorted after v5
	var ai, bi, ci, si int
	for i, v := range order {
		switch v.ID {
		case 2:
			ai = i
		case 3:
			bi = i
		case 4:
			ci = i
		case 5:
			si = i
		}
	}
	if ai < si || bi < si || ci < si {
		t.Logf("Func: %s", fun.f)
		t.Errorf("store order is wrong: got %v, want v2 v3 v4 after v5", order)
	}
}

func TestCarryChainOrder(t *testing.T) {
	// In the function below, there are two carry chains that have no dependencies on each other,
	// one is A1 -> A1carry -> A1Carryvalue, the other is A2 -> A2carry -> A2Carryvalue. If they
	// are not scheduled properly, the carry will be clobbered, causing the carry to be regenerated.
	c := testConfigARM64(t)
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("mem0", OpInitMem, types.TypeMem, 0, nil),
			Valu("x", OpARM64MOVDconst, c.config.Types.UInt64, 5, nil),
			Valu("y", OpARM64MOVDconst, c.config.Types.UInt64, 6, nil),
			Valu("z", OpARM64MOVDconst, c.config.Types.UInt64, 7, nil),
			Valu("A1", OpARM64ADDSflags, types.NewTuple(c.config.Types.UInt64, types.TypeFlags), 0, nil, "x", "z"), // x+z, set flags
			Valu("A1carry", OpSelect1, types.TypeFlags, 0, nil, "A1"),
			Valu("A2", OpARM64ADDSflags, types.NewTuple(c.config.Types.UInt64, types.TypeFlags), 0, nil, "y", "z"), // y+z, set flags
			Valu("A2carry", OpSelect1, types.TypeFlags, 0, nil, "A2"),
			Valu("A1value", OpSelect0, c.config.Types.UInt64, 0, nil, "A1"),
			Valu("A1Carryvalue", OpARM64ADCzerocarry, c.config.Types.UInt64, 0, nil, "A1carry"), // 0+0+A1carry
			Valu("A2value", OpSelect0, c.config.Types.UInt64, 0, nil, "A2"),
			Valu("A2Carryvalue", OpARM64ADCzerocarry, c.config.Types.UInt64, 0, nil, "A2carry"), // 0+0+A2carry
			Valu("ValueSum", OpARM64ADD, c.config.Types.UInt64, 0, nil, "A1value", "A2value"),
			Valu("CarrySum", OpARM64ADD, c.config.Types.UInt64, 0, nil, "A1Carryvalue", "A2Carryvalue"),
			Valu("Sum", OpARM64AND, c.config.Types.UInt64, 0, nil, "ValueSum", "CarrySum"),
			Goto("exit")),
		Bloc("exit",
			Exit("mem0")),
	)

	CheckFunc(fun.f)
	schedule(fun.f)

	// The expected order is A1 < A1carry < A1Carryvalue < A2 < A2carry < A2Carryvalue.
	// There is no dependency between the two carry chains, so it doesn't matter which
	// comes first and which comes after, but the unsorted position of A1 is before A2,
	// so A1Carryvalue < A2.
	var ai, bi, ci, di, ei, fi int
	for i, v := range fun.f.Blocks[0].Values {
		switch {
		case fun.values["A1"] == v:
			ai = i
		case fun.values["A1carry"] == v:
			bi = i
		case fun.values["A1Carryvalue"] == v:
			ci = i
		case fun.values["A2"] == v:
			di = i
		case fun.values["A2carry"] == v:
			ei = i
		case fun.values["A2Carryvalue"] == v:
			fi = i
		}
	}
	if !(ai < bi && bi < ci && ci < di && di < ei && ei < fi) {
		t.Logf("Func: %s", fun.f)
		t.Errorf("carry chain order is wrong: got %v, want V%d after V%d after V%d after V%d after V%d after V%d,",
			fun.f.Blocks[0], fun.values["A1"].ID, fun.values["A1carry"].ID, fun.values["A1Carryvalue"].ID,
			fun.values["A2"].ID, fun.values["A2carry"].ID, fun.values["A2Carryvalue"].ID)
	}
}

// TestScheduleResultAddrAfterVarDef checks that in the entry block the value
// that first kills a result parameter is scheduled before every address
// computed for that result, even when the VarDef is not ready until a chain of
// other memory operations (the open-coded defer bits here) has been scheduled
// behind some address computations of its own. Liveness counts an address-of
// as a use (liveness.affectedVar), and liveness.prologue rejects a result that
// is live on entry to the function (liveness/plive.go:855-863), so the four
// field addresses here must follow the VarDef.
//
// The result is a pointer so that the VarDef is valid SSA: ssa/check.go:325
// requires a VarDef to be a merge candidate or to have a pointer type.
func TestScheduleResultAddrAfterVarDef(t *testing.T) {
	c := testConfig(t)
	typ := c.config.Types
	result := c.Temp(typ.BytePtr)
	result.Class = ir.PPARAMOUT
	deferBits := c.Temp(typ.UInt8)
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, types.TypeMem, 0, nil),
			Valu("sp", OpSP, typ.Uintptr, 0, nil),
			Valu("raddr", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 0, result, "sp"),
			Valu("rfield1", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 8, nil, "raddr"),
			Valu("rfield2", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 16, nil, "raddr"),
			Valu("rfield3", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 24, nil, "raddr"),
			Valu("baddr", OpAMD64LEAQ, typ.UInt8.PtrTo(), 0, deferBits, "sp"),
			Valu("zero", OpAMD64MOVLconst, typ.UInt8, 0, nil),
			Valu("bstore", OpAMD64MOVBstore, types.TypeMem, 0, nil, "baddr", "zero", "start"),
			Valu("blive", OpVarLive, types.TypeMem, 0, deferBits, "bstore"),
			Valu("rdef", OpVarDef, types.TypeMem, 0, result, "blive"),
			Valu("v", OpAMD64MOVQconst, typ.Int64, 7, nil),
			Valu("s0", OpAMD64MOVQstore, types.TypeMem, 0, nil, "raddr", "v", "rdef"),
			Valu("s1", OpAMD64MOVQstore, types.TypeMem, 0, nil, "rfield1", "v", "s0"),
			Valu("s2", OpAMD64MOVQstore, types.TypeMem, 0, nil, "rfield2", "v", "s1"),
			Valu("s3", OpAMD64MOVQstore, types.TypeMem, 0, nil, "rfield3", "v", "s2"),
			Goto("exit")),
		Bloc("exit",
			Exit("s3")))
	CheckFunc(fun.f)
	schedule(fun.f)
	CheckFunc(fun.f)
	pos := map[string]int{}
	for i, v := range fun.f.Entry.Values {
		pos[v.String()] = i
	}
	rdef := pos[fun.values["rdef"].String()]
	for _, name := range []string{"raddr", "rfield1", "rfield2", "rfield3"} {
		if p := pos[fun.values[name].String()]; p < rdef {
			t.Errorf("%s scheduled at %d, before VarDef of the result at %d", name, p, rdef)
		}
	}
}

// initializedResultCycle schedules the graph of
// TestX12ReviewInitializedResultCycle with a result of type res: a store
// initializes the result before its VarDef, the address of the result is
// taken, and the result is read back through that address.
func initializedResultCycle(c *Conf, res *types.Type) {
	typ := c.config.Types
	result := c.Temp(res)
	result.Class = ir.PPARAMOUT
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, types.TypeMem, 0, nil),
			Valu("sp", OpSP, typ.Uintptr, 0, nil),
			Valu("zero", OpAMD64MOVQconst, typ.BytePtr, 0, nil),
			Valu("init", OpAMD64MOVQstore, types.TypeMem, 0, result, "sp", "zero", "start"),
			Valu("raddr", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 0, result, "sp"),
			Valu("read", OpAMD64MOVQload, typ.BytePtr, 0, nil, "raddr", "init"),
			Valu("rdef", OpVarDef, types.TypeMem, 0, result, "init"),
			Valu("write", OpAMD64MOVQstore, types.TypeMem, 0, result, "sp", "read", "rdef"),
			Goto("exit")),
		Bloc("exit",
			Exit("write")))
	CheckFunc(fun.f)
	schedule(fun.f)
	CheckFunc(fun.f)
}

// TestX12ReviewInitializedResultCycle schedules a result whose address is
// taken and whose memory is initialized by a store before its VarDef. The
// store is the first kill of the result, and it is on the VarDef's memory
// chain ahead of it, so an edge from the VarDef to the address would have to
// be satisfied at the same time as the load-before-store edge that puts the
// VarDef after the load through that address: the block would have no
// schedule at all.
//
// This is the counterexample from the review of goppc-xprg.12, kept in
// out/ppc-mac/x12-evidence/review-counterexample as ssa_x12_review_test.go.
func TestX12ReviewInitializedResultCycle(t *testing.T) {
	c := testConfig(t)
	initializedResultCycle(c, c.config.Types.BytePtr)
}

// TestX12ReviewInitializedResultCycleAggregateResult is the review of
// goppc-10mz's counterexample: the same graph with a result type that
// liveness counts a single store to as a kill although the type is not a
// scalar. liveness.isfat (liveness/plive.go:1524) recurses through a
// one-element array or a one-field struct, so for [1]*T and struct{p *T} a
// store to the result IS the first kill of it, and the scheduler's
// fullWrite has to say so or the edge it adds comes from the VarDef -- the
// edge this graph cannot satisfy.
//
// [2]*T is deliberately absent: isfat([2]*T) is true, so liveness agrees with
// fullWrite that one store does not cover the result, and this graph is then
// genuinely unschedulable for both compilers.
func TestX12ReviewInitializedResultCycleAggregateResult(t *testing.T) {
	for _, tc := range []struct {
		name string
		res  func(*Types) *types.Type
	}{
		{"pointer", func(t *Types) *types.Type { return t.BytePtr }},
		{"array1", func(t *Types) *types.Type { return types.NewArray(t.BytePtr, 1) }},
		{"struct1pointer", func(t *Types) *types.Type {
			return types.NewStruct([]*types.Field{{Type: t.BytePtr}})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testConfig(t)
			initializedResultCycle(c, tc.res(&c.config.Types))
		})
	}
}

// TestX12ReviewSecondVarDef schedules a second VarDef of the same result. Only
// the first kill of the result constrains the addresses computed for it: a
// second VarDef redefines a result that is already live, so ordering the
// address after it would cycle with the store that the first VarDef is
// chained behind.
//
// Also from out/ppc-mac/x12-evidence/review-counterexample.
func TestX12ReviewSecondVarDef(t *testing.T) {
	c := testConfig(t)
	typ := c.config.Types
	result := c.Temp(typ.BytePtr)
	result.Class = ir.PPARAMOUT
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, types.TypeMem, 0, nil),
			Valu("sp", OpSP, typ.Uintptr, 0, nil),
			Valu("zero", OpAMD64MOVQconst, typ.BytePtr, 0, nil),
			Valu("raddr", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 0, result, "sp"),
			Valu("rdef0", OpVarDef, types.TypeMem, 0, result, "start"),
			Valu("write0", OpAMD64MOVQstore, types.TypeMem, 0, nil, "raddr", "zero", "rdef0"),
			Valu("rdef1", OpVarDef, types.TypeMem, 0, result, "write0"),
			Valu("write1", OpAMD64MOVQstore, types.TypeMem, 0, nil, "raddr", "zero", "rdef1"),
			Goto("exit")),
		Bloc("exit",
			Exit("write1")))
	CheckFunc(fun.f)
	schedule(fun.f)
	CheckFunc(fun.f)
}

// TestScheduleResultAddrAfterFirstKill checks the invariant the entry-block
// constraint is there to establish: no address of a result is scheduled above
// the value that kills the result, so liveness never sees the result as used
// before it is defined on entry to the function. The store that initializes
// this result comes before its VarDef, and the base scheduler does not order
// the address after the store either -- it schedules the address first, which
// liveness then rejects with "bad live variable at entry".
func TestScheduleResultAddrAfterFirstKill(t *testing.T) {
	c := testConfig(t)
	typ := c.config.Types
	result := c.Temp(typ.BytePtr)
	result.Class = ir.PPARAMOUT
	fun := c.Fun("entry",
		Bloc("entry",
			Valu("start", OpInitMem, types.TypeMem, 0, nil),
			Valu("sp", OpSP, typ.Uintptr, 0, nil),
			Valu("zero", OpAMD64MOVQconst, typ.BytePtr, 0, nil),
			Valu("init", OpAMD64MOVQstore, types.TypeMem, 0, result, "sp", "zero", "start"),
			Valu("raddr", OpAMD64LEAQ, typ.BytePtr.PtrTo(), 0, result, "sp"),
			Valu("read", OpAMD64MOVQload, typ.BytePtr, 0, nil, "raddr", "init"),
			Valu("rdef", OpVarDef, types.TypeMem, 0, result, "init"),
			Valu("write", OpAMD64MOVQstore, types.TypeMem, 0, result, "sp", "read", "rdef"),
			Goto("exit")),
		Bloc("exit",
			Exit("write")))
	CheckFunc(fun.f)
	schedule(fun.f)
	CheckFunc(fun.f)
	pos := map[string]int{}
	for i, v := range fun.f.Entry.Values {
		pos[v.String()] = i
	}
	kill := pos[fun.values["init"].String()]
	if p := pos[fun.values["raddr"].String()]; p < kill {
		t.Errorf("raddr scheduled at %d, before the kill of the result at %d", p, kill)
	}
}
