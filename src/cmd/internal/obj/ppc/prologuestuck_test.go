// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"encoding/binary"
	"fmt"
	"internal/abi"
	"strings"
	"testing"
)

// This file is the gate for bead goppc-h890: the inserted ppc32 prologue must
// never offer the unwinder a PC that is simultaneously
//
//	(a) preemptible -- the runtime is willing to stop the goroutine there, and
//	    the SIGPROF sampler will walk it regardless of that willingness;
//	(b) frame-less -- the function's own SP adjustment has not happened yet, so
//	    the PCSP delta is 0 and the unwinder's frame.fp equals frame.sp; and
//	(c) self-referential -- the link register holds that very PC, so
//	    frame.pc == frame.lr.
//
// All three at once is the fatal combination: src/runtime/traceback.go:491-497
// then prints "runtime: traceback stuck" and throws, unconditionally -- not
// gated by unwindPrintErrors/unwindSilentErrors -- so the walk cannot make
// progress and cannot report softly either. (a) is what the toolchain controls;
// (b) and (c) are what a prologue shape causes. Condition (c) is produced by a
// branch-and-link instruction, because `bl` sets LR to the following PC: with
// `BL runtime.stackcheck` five instructions ahead of the frame allocation, a
// sample or preemption taken at the CMPW right after it was all three, and it
// killed `archive/zip` on real hardware at pc=0x11d30bc =
// bufio.(*Writer).Write+0x14 (see the bead).
//
// The shape rule that follows, and that the checker below enforces: no
// branch-and-link may sit between function entry and the instruction that
// establishes the frame, on any control-flow path a sample can land on, unless
// that path is marked PCDATA_UnsafePoint=Unsafe.
//
// The checker models the runtime's decision, not a mnemonic pattern:
//   - unsafe-point value in effect at a PC, the way linkpcln's funcpctab builds
//     it and pcdatavalue2 hands it to isAsyncSafePoint (a PCDATA prog is
//     zero-width and changes the value from the next real instruction on),
//   - preemptible means "isAsyncSafePoint would take the sample at all", which
//     for a Go function means the value is not UnsafePointUnsafe
//     (src/runtime/preempt.go:443-449). Restart1 and Restart2 ARE preemptible:
//     they return true with a backed-off resume PC (src/runtime/preempt.go:485),
//     so a restart window does not by itself close this hole -- it only moves
//     where the goroutine stops,
//   - the SP delta in effect *before* an instruction executes, which is the
//     sum of Spadj of the preceding instructions, exactly what funcspdelta
//     reports for that PC,
//   - the link register, tracked through the instruction stream. It enters as
//     the caller's return address -- an address in the caller, which is never a
//     PC of this prologue -- and every branch-and-link rewrites it with the
//     address of the next real instruction, which is precisely the state (c).

// prologueInsn is one real (byte-bearing) instruction of a function, with the
// three facts the runtime would observe if it sampled there.
type prologueInsn struct {
	p         *obj.Prog
	pc        int64 // byte offset from the function's entry, 4 apart
	unsafeVal int64 // PCDATA_UnsafePoint in effect at pc
	spBefore  int64 // SP delta accumulated by earlier instructions
	lr        int64 // LR's value; externalLinkReg means the caller's link register
}

// externalLinkReg stands for "LR still holds the address the caller branched to
// this function with". It is an address in the caller, so it cannot equal any
// PC of this function; a negative sentinel keeps that fact out of the
// arithmetic without pretending to know the caller's layout.
const externalLinkReg = int64(-1)

// modelFunction walks a (preprocessed, optionally assembled) Prog list and
// reports every real instruction with the state an asynchronous sample would
// observe there. It mirrors the assembler's own PC assignment -- 4 bytes per
// real instruction, zero bytes for PCDATA/FUNCDATA/ANOP -- which
// TestPrologueStuckPCsNeverPreemptible checks against a real assembly run so
// that the model cannot quietly drift from the emitted layout.
func modelFunction(text *obj.Prog) []prologueInsn {
	var insns []prologueInsn
	pc, acc, lr := int64(0), int64(0), externalLinkReg
	val := int64(abi.UnsafePointSafe) // linkpcln seeds the table with Safe
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			val = p.To.Offset
			continue
		}
		if p.As == obj.APCDATA || p.As == obj.AFUNCDATA || p.As == obj.ANOP {
			continue // zero-width: no PC of its own, cannot change any of the above
		}
		insns = append(insns, prologueInsn{p: p, pc: pc, unsafeVal: val, spBefore: acc, lr: lr})
		// A branch-and-link leaves LR pointing at the next real instruction.
		if setsLinkRegister(p) {
			lr = pc + insnSize
		}
		pc += insnSize
		// Spadj is the delta this instruction causes, so it belongs to the state
		// of the instructions that follow it, not to its own.
		acc += int64(p.Spadj)
	}
	return insns
}

// setsLinkRegister reports whether executing p leaves LR holding the following
// PC. These are the forms that branch with the link bit set: exactly what
// preprocess counts as "making calls" when it decides whether a function is a
// leaf, plus the DUFF intrinsics, which are calls as well. Plain AB/BLR do not
// touch LR.
func setsLinkRegister(p *obj.Prog) bool {
	switch p.As {
	case ABL, obj.ACALL, ABCTRL, ABLRL, obj.ADUFFCOPY, obj.ADUFFZERO:
		return true
	}
	return false
}

// stuckProloguePCs returns every instruction that satisfies all three
// conditions above. A non-empty result means a stack walk starting there throws.
func stuckProloguePCs(insns []prologueInsn) []prologueInsn {
	var bad []prologueInsn
	for _, in := range insns {
		frameless := in.spBefore == 0
		preemptible := in.unsafeVal != abi.UnsafePointUnsafe
		selfRef := in.lr == in.pc && in.lr != externalLinkReg
		if frameless && preemptible && selfRef {
			bad = append(bad, in)
		}
	}
	return bad
}

// describePCs renders findings with the mnemonic and the modelled state, so a
// failure names the shape instead of just counting.
func describePCs(bad []prologueInsn) string {
	var b strings.Builder
	for _, in := range bad {
		fmt.Fprintf(&b, "\tpc=%d %s spBefore=%d lr=%d unsafePoint=%d\n",
			in.pc, Mnemonic(Inst(in.p.As)), in.spBefore, in.lr, in.unsafeVal)
	}
	return b.String()
}

func TestPrologueStuckPCsNeverPreemptible(t *testing.T) {
	for _, tc := range []struct {
		name     string
		locals   int64
		needCtxt bool
	}{
		{"d-form-frame", 4096, false},
		{"indexed-frame", 65512, false},
		{"largest-d-form", 32752 - 16, false},
		{"small-frame", 4, false},
		{"needs-context", 64, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			s := ctxt.Lookup("growable")
			s.Set(obj.AttrNeedCtxt, tc.needCtxt)
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: tc.locals, Val: int32(4)}}
			text.Link = &obj.Prog{As: obj.ARET}
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })

			insns := modelFunction(text)
			if bad := stuckProloguePCs(insns); len(bad) > 0 {
				t.Fatalf("%d preemptible frame-less PC(s) with pc == lr; a traceback or SIGPROF sample that starts here hits the unconditional throw(\"traceback stuck\") at src/runtime/traceback.go:491:\n%s",
					len(bad), describePCs(bad))
			}

			// Non-vacuity. The prologue is frame-less and preemptible over
			// several instructions by design (the MFLR/STW pair, the zeroing of
			// R0, the frame allocation itself). If a refactor ever makes the
			// checker see no such instruction at all, it has stopped measuring
			// something and the pass above means nothing.
			seen := 0
			for _, in := range insns {
				if in.spBefore == 0 && in.unsafeVal != abi.UnsafePointUnsafe {
					seen++
				}
			}
			if seen == 0 {
				t.Fatalf("checker is vacuous: no frame-less preemptible PC in the prologue of a non-leaf function\n%s", dumpProgs(text))
			}

			// The instrument's modelled PCs must be the emitted ones, or the
			// reasoning above is about a different program than the one that
			// ships. assemble() assigns the real PCs; the prologue's local
			// branches are short, so nothing is relaxed and the two must agree
			// instruction by instruction.
			assemble(ctxt, s, nil)
			m := 0
			for _, in := range insns {
				if in.p.Pc != in.pc {
					t.Fatalf("instruction %d modelled at pc %d but assembled at pc %d -- the model no longer matches the emitted layout",
						m, in.pc, in.p.Pc)
				}
				m++
			}
			if want := int(s.Size / insnSize); m != want {
				t.Fatalf("model covered %d instructions, assembled function is %d bytes", m, want)
			}

			// Byte-level check of the load the model says is there. lwz r4,8(r30) is
			// opcode 32<<26 | RT=4<<21 | RA=30<<16 | D=8 = 0x809e0008, and it must occur
			// exactly once: this is the g->stackguard0 read that replaced the call, and a
			// second one would mean a duplicated guard (which would also mean the guard
			// region markers do not line up with the code they describe).
			const guardWord = uint32(0x809e0008)
			hits := 0
			for i := 0; i+4 <= len(s.P); i += 4 {
				if binary.BigEndian.Uint32(s.P[i:]) == guardWord {
					hits++
				}
			}
			if hits != 1 {
				t.Fatalf("assembled function contains %d copies of the inline guard load 0x809e0008, want exactly 1", hits)
			}
			// And no branch-and-link to the old helper may survive anywhere.
			for p := text.Link; p != nil; p = p.Link {
				if p.As == ABL && p.To.Sym != nil && p.To.Sym.Name == "runtime.stackcheck" {
					t.Fatalf("assembled prologue still links to runtime.stackcheck: %v", p)
				}
			}
		})
	}
}

// TestStuckPCCheckerRejectsPreFixPrologue is the negative control the gate needs
// to be evidence: it rebuilds the prologue this package emitted before the
// guard was inlined -- frame size, BL runtime.stackcheck, CMPW, BEQ, morestack,
// back edge, allocation -- and requires the same checker to reject it, naming
// the CMPW right after the guard call. A checker that cannot fail is not
// evidence, and this is the exact shape hardware proved fatal (bead goppc-h890).
func TestStuckPCCheckerRejectsPreFixPrologue(t *testing.T) {
	text := preFixSplitPrologue()
	insns := modelFunction(text)
	bad := stuckProloguePCs(insns)
	if len(bad) == 0 {
		t.Fatalf("checker accepts the pre-fix prologue, whose post-stackcheck CMPW is pc == lr with no frame:\n%s", dumpProgs(text))
	}
	// The instruction right after `BL runtime.stackcheck` is the first one the
	// runtime could be stopped on, and it is the one observed on pmg5.
	if bad[0].p.As != ACMPWI {
		t.Fatalf("checker's first finding is %v, want the CMPW immediately after the guard call", bad[0].p.As)
	}
	// And the morestack return is the second hole the inline guard also closes.
	sawMorestackReturn := false
	for _, in := range bad {
		if in.p.As == AB && in.p.To.Type == obj.TYPE_BRANCH && in.p.To.Target() != nil {
			sawMorestackReturn = true
		}
	}
	if !sawMorestackReturn {
		t.Fatalf("checker did not report the PC after the runtime.morestack call as a stuck risk:\n%s", dumpProgs(text))
	}
}

// preFixSplitPrologue hand-builds the inserted prologue as it was emitted before
// the guard was inlined, markers included: a Restart1 window over MFLR/STW and
// nothing else, so everything from the zeroing of R0 onward reads as Safe.
func preFixSplitPrologue() *obj.Prog {
	mkr := func(val int64) *obj.Prog {
		return &obj.Prog{As: obj.APCDATA,
			From: obj.Addr{Type: obj.TYPE_CONST, Offset: abi.PCDATA_UnsafePoint},
			To:   obj.Addr{Type: obj.TYPE_CONST, Offset: val}}
	}
	reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
	at := func(off int64) obj.Addr {
		return obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: off}
	}
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 128, Val: int32(4)}}
	seq := []*obj.Prog{
		mkr(abi.UnsafePointRestart1),
		{As: AMFLR, To: reg(REG_R0)},
		{As: ASTW, From: reg(REG_R0), To: at(8)},
		mkr(abi.UnsafePointSafe),
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: reg(REG_R0)},
		{As: obj.ANOP},
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 128}, Reg: REG_R0, To: reg(REG_R3)},
		{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN}},
		{As: ACMPWI, From: reg(REG_R3), To: obj.Addr{Type: obj.TYPE_CONST}},
		{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}},
		{As: ALWZ, From: at(8), To: reg(REG_R5)},
		{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN}},
		{As: AB, From: obj.Addr{Type: obj.TYPE_CONST}, To: obj.Addr{Type: obj.TYPE_BRANCH}},
		{As: obj.ANOP},
		{As: ASTWU, From: reg(REGSP), To: at(-128)},
	}
	for i := 0; i+1 < len(seq); i++ {
		seq[i].Link = seq[i+1]
	}
	text.Link = seq[0]
	// The back edge returns to the frame-size materialization, the guard's "done"
	// ANOP follows the morestack loop, and the BEQ skips to that ANOP.
	seq[12].To.SetTarget(seq[5])
	seq[9].To.SetTarget(seq[13])
	seq[13].Link = seq[14]
	return text
}

// stackGuardLoad reports whether p is the inline prologue guard's load of
// g.stackguard0 -- `lwz r4, 2*ptrSize(r30)`. The tests in this package that pin
// the prologue shape anchor on it rather than on a call, because the call is the
// defect (bead goppc-h890): `BL runtime.stackcheck` left the very next PC with
// pc == lr and no frame, which is the unconditional throw("traceback stuck").
func stackGuardLoad(p *obj.Prog, ptrSize int64) bool {
	return p.As == ALWZ && p.From.Type == obj.TYPE_MEM && p.From.Reg == REGG &&
		p.From.Offset == 2*ptrSize && p.To.Type == obj.TYPE_REG && p.To.Reg == REG_R4
}

// enoughStackBranch reports the guard's "enough stack, skip the grow path"
// branch: a branch conditional on the unsigned-compare LT bit being clear
// (BO=4 tests the condition negated, BI=0 is bit 0 of CR0).
func enoughStackBranch(p *obj.Prog) bool {
	return p.As == ABC && p.From.Type == obj.TYPE_CONST && p.From.Offset == 4 &&
		p.Reg == 0 && p.To.Type == obj.TYPE_BRANCH && p.To.Target() != nil
}

// isZeroWidthProg reports whether p occupies no bytes, so that a branch whose
// target prog is one of these lands on the next real instruction. PCDATA and
// FUNCDATA pseudo-ops are emitted with no width by both the assembler and
// linkpcln (any pseudo-prog sharing its successor's PC is dropped), and the
// inserted prologue uses ANOP as a branch anchor for exactly that reason.
func isZeroWidthProg(p *obj.Prog) bool {
	return p.As == obj.ANOP || p.As == obj.APCDATA || p.As == obj.AFUNCDATA
}
