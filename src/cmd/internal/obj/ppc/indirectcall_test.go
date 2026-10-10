// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"encoding/binary"
	"internal/abi"
	"testing"
)

// indirectCallFixture builds the Prog list cmd/compile/internal/ppc emits for
// an indirect call -- MTLR R12 ; BLRL, with the call's stack-map PCDATA in
// between -- and runs it through preprocess, which is where the frame and the
// leaf classification are decided.
func indirectCallFixture(t *testing.T, indirect bool) (*obj.Link, *obj.LSym, *obj.Prog) {
	t.Helper()
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE}}
	after := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 4}, Reg: REG_R3, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	ret := &obj.Prog{As: obj.ARET}
	text.Link = after
	after.Link = ret
	if indirect {
		mtlr := &obj.Prog{As: AMTLR, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R12}}
		stackmap := &obj.Prog{As: obj.APCDATA, From: obj.Addr{Type: obj.TYPE_CONST, Offset: abi.PCDATA_StackMapIndex}, To: obj.Addr{Type: obj.TYPE_CONST, Offset: 1}}
		blrl := &obj.Prog{As: ABLRL}
		text.Link, mtlr.Link, stackmap.Link, blrl.Link = mtlr, stackmap, blrl, after
	}
	s := ctxt.Lookup("fixture")
	s.Set(obj.AttrNoSplit, true)
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	return ctxt, s, text
}

// TestIndirectCallUsesLinkRegister pins the fix for the async-preemption hang.
//
// An indirect call branches through LR, not CTR. runtime.asyncPreempt resumes a
// preempted goroutine with "MTCTR <resume PC> ; JMP (CTR)" -- it has to, because
// LR must still hold the interrupted function's link register on resumption --
// so a goroutine stopped *on* a CTR-based indirect call would resume with CTR
// pointing at that call and branch to itself forever. With MTLR+BLRL the
// interrupted LR on the BLRL is already the call target, and the BLRL reloads
// the return address as it branches, so there is no window to guard: the
// instruction stream contains no CTR-dependent call at all.
//
// The load-bearing part of the change in this package is the leaf
// classification: BLRL is register-indirect and carries no symbol, so a
// function that calls through LR must be recognized as a call-making function
// and given a frame and a saved LR.
func TestIndirectCallUsesLinkRegister(t *testing.T) {
	ctxt, s, text := indirectCallFixture(t, true)

	if s.Leaf() {
		t.Errorf("function with an indirect call is classified as a leaf: no frame, no saved LR")
	}
	if text.To.Offset == 0 {
		t.Errorf("indirect-call function got frame size 0")
	}
	for p := text.Link; p != nil; p = p.Link {
		if p.As == ABCTRL {
			t.Errorf("CTR-based indirect call still emitted: %v", p)
		}
		if p.As == obj.APCDATA && p.From.Offset == abi.PCDATA_UnsafePoint {
			t.Errorf("unsafe-point marker emitted; the LR form needs no unsafe window: %v", p)
		}
	}

	assemble(ctxt, s, nil)
	n := len(s.P) / 4
	word := func(i int) uint32 { return binary.BigEndian.Uint32(s.P[4*i:]) }
	pair := false
	for i := 0; i+1 < n; i++ {
		// mtlr r12 = 0x7d8803a6, blrl = 0x4e800021 (ISA encodings; the
		// assembler's own table in asm.go carries the same bases).
		if word(i) == 0x7d8803a6 && word(i+1) == 0x4e800021 {
			pair = true
		}
	}
	if !pair {
		t.Errorf("no mtlr r12 ; blrl pair in the %d encoded words", n)
	}
}

// TestReturnOnlyFunctionStaysLeaf is the control for the leaf classification:
// widening the call set must not make an ordinary returning function non-leaf.
func TestReturnOnlyFunctionStaysLeaf(t *testing.T) {
	_, s, text := indirectCallFixture(t, false)
	if !s.Leaf() {
		t.Errorf("function with no call is not classified as a leaf")
	}
	if text.To.Offset != 0 {
		t.Errorf("leaf function got frame size %d, want 0", text.To.Offset)
	}
}
