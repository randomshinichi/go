// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"bytes"
	"cmd/internal/obj"
	"internal/abi"
	"testing"
)

// cmd/compile/internal/ppc/ggen.go zerorange() zeroes an ambiguously-live stack
// range by materializing the value zero in R0 and storing it into every word of
// the range:
//
//	ADDI $0, R0, R0        ; li r0, 0
//	STW  R0, off+0(SP)
//	STW  R0, off+4(SP)
//	…
//
// R0 is NOT preserved across an asynchronous preemption: runtime.asyncPreempt's
// generated ppc32 prologue saves R3..R29, R31, CR, CTR, XER and the FPRs, but not
// R0 (genPPC32 in runtime/mkpreempt.go). A preemption landing on the second or a
// later store therefore resumes with the preemption's own scratch in R0 and
// writes that into a GC-scanned stack slot. Bead goppc-2s7q.
//
// The register cannot be changed. R31 is the assembler's out-of-range
// displacement temporary (link.go's encodeProg materializes the address in R31
// and stores the source register through it), so a store *from* R31 stores the
// address -- a deterministic corruption traded for a racy one, which is why the
// first candidate fix was rejected -- and every other GPR is allocatable. The run
// is marked instead: PCDATA_UnsafePoint = UnsafePointUnsafe from the `li` through
// the last store, so isAsyncSafePoint (runtime/preempt.go) refuses a preemption
// anywhere inside it.
//
// Two properties are load-bearing, and this file pins both:
//
//  1. The markers add NO code bytes. A real instruction would shift frame and
//     nosplit accounting for every function that zeroes a range.
//  2. The marking covers the whole run and ends with it. A run that is marked
//     only from the second store leaves the `li` preemptible with a dead R0
//     before it; a run that leaks past the last store keeps unrelated code
//     unpreemptible.
//
// The byte-identity control reuses stripUnsafeMarkers (stackcheckunsafe_test.go),
// which deletes every Unsafe marker, reproducing the pre-fix emitter.

// zeroRangeFixture builds the Prog stream the compiler emits at function entry:
// the zerorange sequence followed by a real instruction, in a function with a
// frame big enough that the store displacements are outside the D-form range.
// It returns the Li and the stores by pointer, so the checks below identify the
// run by construction rather than by re-running the production predicate.
func zeroRangeFixture(t *testing.T, locals int64, stores int) (*obj.Link, *obj.LSym, *obj.Prog, *obj.Prog, []*obj.Prog) {
	t.Helper()
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("zerorange")
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals, Val: int32(4)}}
	prog := func(p *obj.Prog, as obj.As, ftype obj.AddrType, freg int16, foff int64, ttype obj.AddrType, treg int16, toff int64) *obj.Prog {
		q := &obj.Prog{As: as}
		q.From = obj.Addr{Type: ftype, Reg: freg, Offset: foff}
		q.To = obj.Addr{Type: ttype, Reg: treg, Offset: toff}
		q.Link = p.Link
		p.Link = q
		return q
	}
	tail := &obj.Prog{As: obj.ARET}
	text.Link = tail
	// The compiler's zerorange: li r0,0 then cnt/PtrSize stores from R0, at
	// displacements that, in this fixture, the D-form cannot encode.
	li := prog(text, AADDI, obj.TYPE_CONST, 0, 0, obj.TYPE_REG, REG_R0, 0)
	li.Reg = REG_R0
	var sts []*obj.Prog
	last := li
	for i := 0; i < stores; i++ {
		st := prog(last, ASTW, obj.TYPE_REG, REG_R0, 0, obj.TYPE_MEM, REGSP, locals-4*int64(stores-i))
		sts = append(sts, st)
		last = st
	}
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	return ctxt, s, text, li, sts
}

// zeroRangeVerdict is what the gate needs to know about how the unsafe-point
// table covers one zerorange run. It models the runtime's lookup: linkpcln turns
// the Prog list into (value, pc) runs and pcdatavalue2 hands isAsyncSafePoint the
// value in effect at the interrupted pc, so the value is read per REAL
// instruction and PCDATA progs (which carry no bytes) are skipped.
type zeroRangeVerdict struct {
	liValue     int64 // unsafe-point value in effect at the `li r0,0`
	storeValues []int64
	afterValue  int64 // value in effect at the first real instruction past the run
	found       bool
}

// scoreZeroRangeWindow reads the unsafe-point value in effect at the Li, at each
// store, and at the first real instruction after the last store.
func scoreZeroRangeWindow(text, li *obj.Prog, stores []*obj.Prog) zeroRangeVerdict {
	v := zeroRangeVerdict{liValue: abi.UnsafePointSafe, afterValue: abi.UnsafePointSafe}
	if li == nil {
		return v
	}
	val := int64(abi.UnsafePointSafe)
	seen := 0
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			val = p.To.Offset
			continue
		}
		if p == li {
			v.liValue, v.found = val, true
			continue
		}
		if seen < len(stores) {
			if p != stores[seen] {
				continue
			}
			v.storeValues = append(v.storeValues, val)
			seen++
			continue
		}
		// First real instruction past the run.
		v.afterValue = val
		break
	}
	return v
}

// marked reports whether the whole run is under UnsafePointUnsafe.
func (v zeroRangeVerdict) marked() bool {
	if !v.found || v.liValue != abi.UnsafePointUnsafe || len(v.storeValues) == 0 {
		return false
	}
	for _, s := range v.storeValues {
		if s != abi.UnsafePointUnsafe {
			return false
		}
	}
	return true
}

// TestZeroRangeRunIsMarkedUnsafe: the run the compiler emits is marked Unsafe for
// async preemption from the `li r0,0` through the last store, and the value in
// effect after it is restored (the run does not leak into unrelated code).
func TestZeroRangeRunIsMarkedUnsafe(t *testing.T) {
	const locals, stores = 96064, 2 // a frame whose store displacements exceed 16 bits
	_, _, text, li, sts := zeroRangeFixture(t, locals, stores)
	v := scoreZeroRangeWindow(text, li, sts)
	if !v.found {
		t.Fatalf("fixture did not survive preprocess: no `li r0,0` found")
	}
	if !v.marked() {
		t.Errorf("zerorange run is not marked Unsafe: value at li=%d, at the %d stores=%v",
			v.liValue, len(sts), v.storeValues)
	}
	if v.afterValue == abi.UnsafePointUnsafe {
		t.Errorf("the Unsafe marking leaks past the run: the instruction after the last store reads %d", v.afterValue)
	}
	if v.afterValue != abi.UnsafePointSafe {
		t.Errorf("value after the run = %d, want the compiler's entry value %d", v.afterValue, abi.UnsafePointSafe)
	}
}

// TestZeroRangeMarkingIsConfinedToTheRun: a `li r0,0` with no store has no live
// R0 value and is not marked, and neither is a store whose source is another
// register. The marking is for the R0-live shape, not for the instruction pair.
func TestZeroRangeMarkingIsConfinedToTheRun(t *testing.T) {
	// No store at all.
	_, _, text, li, _ := zeroRangeFixture(t, 96064, 0)
	if got := inEffectAt(text, li); got == abi.UnsafePointUnsafe {
		t.Errorf("a `li r0,0` with nothing storing R0 is marked Unsafe")
	}
	// A store from R31 -- the rejected candidate fix's shape, where R31 is the
	// assembler's expansion temporary. Nothing reads R0 here, so nothing is marked.
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("zerorange31")
	text2 := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 96064, Val: int32(4)}}
	tail := &obj.Prog{As: obj.ARET}
	text2.Link = tail
	li2 := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}}
	li2.Reg = REG_R0
	st2 := &obj.Prog{As: ASTW, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}, To: obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: 96056}}
	text2.Link, li2.Link, st2.Link = li2, st2, tail
	s.NewFuncInfo().Text = text2
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	if got := inEffectAt(text2, st2); got == abi.UnsafePointUnsafe {
		t.Errorf("a store from R31 after `li r0,0` is marked Unsafe, but R0 is not read there")
	}
}

// inEffectAt returns the unsafe-point value in effect at prog target.
func inEffectAt(text, target *obj.Prog) int64 {
	val := int64(abi.UnsafePointSafe)
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			val = p.To.Offset
			continue
		}
		if p == target {
			return val
		}
	}
	return val
}

// TestZeroRangeMarkersAddNoCodeBytes: the markers are PCDATA only. The assembled
// bytes with and without them must be identical, so frame and nosplit accounting
// are untouched.
func TestZeroRangeMarkersAddNoCodeBytes(t *testing.T) {
	for _, locals := range []int64{128, 96064} {
		for _, stores := range []int{0, 1, 2, 8} {
			ctxt, s, _, _, _ := zeroRangeFixture(t, locals, stores)
			assemble(ctxt, s, nil)
			ctxt2, s2, text2, li2, sts2 := zeroRangeFixture(t, locals, stores)
			if stores > 0 {
				// The control is only worth running if the marking is really there
				// to remove: an empty strip would compare two identical pre-fix
				// listings and pass for the wrong reason.
				if !scoreZeroRangeWindow(text2, li2, sts2).marked() {
					t.Fatalf("locals=%d stores=%d: the run is not marked, so this control is vacuous", locals, stores)
				}
				stripUnsafeMarkers(text2)
				if scoreZeroRangeWindow(text2, li2, sts2).marked() {
					t.Fatalf("locals=%d stores=%d: the Unsafe marker survived stripUnsafeMarkers", locals, stores)
				}
			}
			assemble(ctxt2, s2, nil)
			if !bytes.Equal(s.P, s2.P) {
				t.Errorf("locals=%d stores=%d: markers changed the emitted code (%d vs %d bytes)",
					locals, stores, len(s.P), len(s2.P))
			}
		}
	}
}
