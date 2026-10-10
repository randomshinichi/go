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

func TestGoStackFrames(t *testing.T) {
	for _, tc := range []struct {
		name          string
		locals, frame int64
		call          bool
	}{
		{"frameless-leaf", 0, 0, false},
		{"leaf-local", 4, 32, false},
		{"minimum-nonleaf", 0, 32, true},
		{"rounded-nonleaf", 40, 64, true},
		{"large-local", 4096, 4128, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: tc.locals, Val: int32(4)}}
			load := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_PARAM, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
			ret := &obj.Prog{As: ABLR}
			text.Link, load.Link = load, ret
			if tc.call {
				call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}, Link: ret}
				load.Link = call
			}
			s := ctxt.Lookup("fixture")
			s.Set(obj.AttrNoSplit, true)
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			if s.Func().Args != 4 || s.Func().Locals != int32(tc.locals) || s.Leaf() == tc.call {
				t.Fatalf("wrong frame metadata: args=%d locals=%d leaf=%v", s.Func().Args, s.Func().Locals, s.Leaf())
			}
			if text.To.Offset != tc.frame || tc.frame%16 != 0 {
				t.Fatalf("allocated frame %d, want aligned %d", text.To.Offset, tc.frame)
			}
			if off, err := resolveOffset(ctxt, load, &load.From, tc.frame); err != nil || off != tc.frame+24 {
				t.Fatalf("incoming parameter offset %d, %v: want frame+24", off, err)
			}
			var spadj int32
			for p := text.Link; p != nil; p = p.Link {
				if p.As == ABLR {
					if spadj != 0 {
						t.Fatalf("return with SP delta %d", spadj)
					}
				} else {
					spadj += p.Spadj
				}
			}
			assemble(ctxt, s, nil)
			word := func(i int) uint32 { return binary.BigEndian.Uint32(s.P[4*i:]) }
			if tc.call {
				// Independently assembled shape: mflr0; stw0,8(SP); li0,0;
				// stwu SP,-N(SP). LR store precedes allocation, not follows it.
				want := []uint32{0x7c0802a6, 0x90010008, 0x38000000, 0x94210000 | uint32(uint16(-tc.frame))}
				for i, w := range want {
					if word(i) != w {
						t.Errorf("prologue word %d = %08x, want %08x", i, word(i), w)
					}
				}
				n := len(s.P) / 4
				wantRet := []uint32{0x38210000 | uint32(tc.frame), 0x83e10008, 0x7fe803a6, 0x4e800020}
				for i, w := range wantRet {
					if word(n-4+i) != w {
						t.Errorf("epilogue word %d = %08x, want %08x", i, word(n-4+i), w)
					}
				}
			} else if tc.frame == 0 && (len(s.P) != 8 || word(0) != 0x80610018 || word(1) != 0x4e800020) {
				t.Fatalf("frameless leaf has unexpected prologue: %x", s.P)
			}
		})
	}
}

func TestNoFrameNonLeafPreservesIncomingLR(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("noframe")
	s.Set(obj.AttrNoSplit, true)
	s.Set(obj.AttrNoFrame, true)
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Val: int32(0)}}
	// morestack must observe its incoming LR before calling another function.
	capture := &obj.Prog{As: AMFLR, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R8}}
	call := &obj.Prog{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}}
	ret := &obj.Prog{As: obj.ARET}
	text.Link, capture.Link, call.Link = capture, call, ret
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	if text.To.Offset != 0 || s.Leaf() || text.Link != capture || capture.Link != call || call.Link != ret {
		t.Fatalf("NOFRAME nonleaf gained an automatic prologue or epilogue")
	}
	if ret.As != ABLR {
		t.Fatalf("NOFRAME return = %v, want BLR", ret.As)
	}
	assemble(ctxt, s, nil)
	if len(s.P) != 12 || binary.BigEndian.Uint32(s.P) != 0x7d0802a6 {
		t.Fatalf("NOFRAME code overwrote LR/linkage before explicit MFLR R8: %x", s.P)
	}
}

func TestStackGuardPrologue(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("split")
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: 4}}
	text.Link = &obj.Prog{As: obj.ARET}
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	var stackcheck, morestack, skipBranch bool
	for p := text.Link; p != nil; p = p.Link {
		if p.As == ABL && p.To.Sym != nil {
			stackcheck = stackcheck || p.To.Sym.Name == "runtime.stackcheck"
			morestack = morestack || p.To.Sym.Name == "runtime.morestack_noctxt"
		}
		if p.As == ABC && p.From.Offset == 12 && p.Reg == 2 && p.To.Target() != nil {
			skipBranch = true
		}
	}
	if !stackcheck || !morestack || !skipBranch {
		t.Fatalf("stack-split prologue missing check/call/skip: stackcheck=%v morestack=%v branch=%v", stackcheck, morestack, skipBranch)
	}
}

func TestStackGuardContextSelection(t *testing.T) {
	for _, needCtxt := range []bool{false, true} {
		ctxt := obj.Linknew(&Linkppc)
		ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
		s := ctxt.Lookup("split")
		s.Set(obj.AttrNeedCtxt, needCtxt)
		text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: 4}}
		text.Link = &obj.Prog{As: obj.ARET}
		s.NewFuncInfo().Text = text
		preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
		want := "runtime.morestack_noctxt"
		if needCtxt {
			want = "runtime.morestack"
		}
		found := false
		for p := text.Link; p != nil; p = p.Link {
			if p.As == ABL && p.To.Sym != nil && p.To.Sym.Name == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("NeedCtxt=%v: missing call to %s", needCtxt, want)
		}
	}
}

func TestLargeFramePrologue(t *testing.T) {
	for _, tc := range []struct {
		name  string
		frame int64
	}{
		{"largest-d-form", 32752},
		{"first-indexed", 32768},
		{"above-32k", 65536},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			s := ctxt.Lookup("large")
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: tc.frame - ctxt.Arch.FixedFrameSize}}
			text.Link = &obj.Prog{As: obj.ARET}
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			if text.To.Offset != tc.frame || tc.frame%16 != 0 {
				t.Fatalf("allocated frame %d, want aligned frame %d", text.To.Offset, tc.frame)
			}
			var stackcheck, morestack, directAllocation, indexedAllocation, sourceReturnRemains bool
			var allocationSPAdj, restoreSPAdj int32
			var tail []obj.As
			for p := text.Link; p != nil; p = p.Link {
				if p.As == ABL && p.To.Sym != nil {
					stackcheck = stackcheck || p.To.Sym.Name == "runtime.stackcheck"
					morestack = morestack || p.To.Sym.Name == "runtime.morestack_noctxt"
				}
				directAllocation = directAllocation || p.As == ASTWU
				indexedAllocation = indexedAllocation || p.As == ASTWUX
				if p.As == ASTWU || p.As == ASTWUX {
					allocationSPAdj = p.Spadj
				}
				if p.Spadj < 0 {
					restoreSPAdj = p.Spadj
				}
				sourceReturnRemains = sourceReturnRemains || p.As == obj.ARET
				tail = append(tail, p.As)
			}
			if !stackcheck || !morestack {
				t.Fatalf("large growable frame lacks stack-growth path: stackcheck=%v morestack=%v", stackcheck, morestack)
			}
			if tc.frame <= 32752 && (!directAllocation || indexedAllocation) {
				t.Fatalf("frame %d should use D-form stwu only", tc.frame)
			}
			if tc.frame > 32752 && (directAllocation || !indexedAllocation) {
				t.Fatalf("frame %d should use indexed stwux only", tc.frame)
			}
			if sourceReturnRemains || len(tail) < 6 {
				t.Fatalf("frame %d retains source RET instead of replacing it with an epilogue: %v", tc.frame, tail)
			}
			if allocationSPAdj != int32(tc.frame) || restoreSPAdj != -int32(tc.frame) {
				t.Fatalf("frame %d has stack-map deltas allocation=%d restore=%d", tc.frame, allocationSPAdj, restoreSPAdj)
			}
			wantTail := []obj.As{AADDI, ALWZ, AMTLR, ABLR}
			if tc.frame > 32752 {
				wantTail = []obj.As{AADDI, AADD, ALWZ, AMTLR, ABLR}
			}
			for i := range wantTail {
				if tail[len(tail)-len(wantTail)+i] != wantTail[i] {
					t.Fatalf("frame %d epilogue tail %v, want suffix %v", tc.frame, tail, wantTail)
				}
			}
		})
	}
}

// TestStackPointerAccounted verifies the two halves of the SP bookkeeping that
// assemble's detector depends on: an assembly stack movement written as a
// constant is described by Spadj (so the function need not be marked
// FuncFlagSPWrite), and a stack movement no static table can describe is not
// described (so the function is marked). Getting either half wrong is how a
// function whose declared frame already accounts for its movement ends up
// marked SPWRITE, which the unwinder then treats as fatal for any frame that is
// not the innermost one.
func TestStackPointerAccounted(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []*obj.Prog
		// want SP adj for each body Prog, and whether the function must end up
		// with FuncFlagSPWrite.
		wantAdj    []int32
		wantSPWrite bool
	}{
		{
			// The shape runtime/preempt_ppc.s ends with: a constant restore of
			// the frame the assembler generated a prologue for.
			name:        "constant-restore-is-described",
			body:        []*obj.Prog{{As: AADDI, Reg: REGSP, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 448}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}}},
			wantAdj:     []int32{-448},
			wantSPWrite: false,
		},
		{
			// The shape runtime/asm_ppc.s uses to switch stacks: a constant
			// push whose restore later reloads SP from memory. The push is
			// describable, but nothing describes the function as a whole, so the
			// function must stay marked and no Prog may claim a delta.
			name: "register-restore-leaves-function-marked",
			body: []*obj.Prog{
				{As: AADDI, Reg: REGSP, From: obj.Addr{Type: obj.TYPE_CONST, Offset: -32}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}},
				{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: 24}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}},
			},
			wantAdj:     []int32{0, 0},
			wantSPWrite: true,
		},
		{
			// The shape runtime/asm_reflect_ppc.s uses: a runtime-variable
			// update of SP. No static value exists, so the function is marked
			// and nothing is described.
			name: "indexed-update-is-not-describable",
			body: []*obj.Prog{
				{As: ASTWUX, From: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}, Reg: REG_R4},
			},
			wantAdj:     []int32{0},
			wantSPWrite: true,
		},
		{
			// A hand-written frame allocation: store with update. Like the
			// generated prologue, this is a constant adjustment.
			name:        "store-with-update-is-described",
			body:        []*obj.Prog{{As: ASTWU, From: obj.Addr{Type: obj.TYPE_REG, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: -256}}},
			wantAdj:     []int32{256},
			wantSPWrite: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			// The detector refuses to mark compiler output: a proc that
			// writes SP is a diagnostics-worthy event unless the input is
			// assembly, which is what it is here.
			ctxt.IsAsm = true
			s := ctxt.Lookup("fixture")
			s.Set(obj.AttrNoSplit, true)
			s.Set(obj.AttrNoFrame, true)
			text := &obj.Prog{As: obj.ATEXT}
			last := text
			for _, p := range tc.body {
				last.Link = p
				last = p
			}
			last.Link = &obj.Prog{As: obj.ARET}
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			// The body Progs stay in order in front of the return, so index
			// them from the TEXT.
			i := 0
			for p := text.Link; p != nil && i < len(tc.body); p, i = p.Link, i+1 {
				if p.Spadj != tc.wantAdj[i] {
					t.Errorf("%v: Spadj = %d, want %d", p, p.Spadj, tc.wantAdj[i])
				}
			}
			assemble(ctxt, s, nil)
			if got := s.Func().FuncFlag&abi.FuncFlagSPWrite != 0; got != tc.wantSPWrite {
				t.Errorf("FuncFlagSPWrite = %v, want %v", got, tc.wantSPWrite)
			}
		})
	}
}
