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

// The ppc32 prologue carries the caller's return address in R0 for exactly two
// instructions, and R0 is not preserved across an async preemption:
// runtime.asyncPreempt's generated prologue zeroes it (genPPC32 in
// runtime/mkpreempt.go saves R3..R29 only). A preemption landing on the STW
// therefore stores a bogus return address into the caller's linkage area, and
// the function later returns to PC 0 or the unwinder fatals with "traceback did
// not unwind completely".
//
// The fix (bead goppc-lau1, for goppc-lye2) marks the pair as
// PCDATA_UnsafePoint = UnsafePointRestart1, so isAsyncSafePoint backs the
// resumption PC off to the MFLR and the store re-executes with a reloaded R0.
// Two properties are load-bearing, and this file pins both:
//
//  1. The marker adds NO code bytes. A real instruction inserted at function
//     entry would shift the frame and nosplit accounting for every non-leaf
//     function on the target.
//  2. The marked range is at most 20 bytes wide. runtime/preempt.go throws
//     "bad restart PC" when pc-startpc > 20, so an over-wide window is worse
//     than none: it turns a rare corruption into a deterministic fatal. That is
//     the failure mode this change can itself introduce, so it is checked here
//     by name.

// insnSize is ppc32's fixed instruction width, used to turn a count of
// instructions into a byte width without depending on the assembler having run.
const insnSize = 4

// windowVerdict is what the gate needs to know about a prologue's unsafe-point
// marking. It is derived from the Prog list only, so it can be run against both
// a real preprocessed function and a deliberately broken hand-built one.
type windowVerdict struct {
	pairFound    bool  // there is an MFLR Rt ; STW Rt,8(R1) prologue
	linkReg      int16 // the register the pair carries LR through
	markedAtMFLR bool  // Restart1/2 in effect at the MFLR
	startsAtMFLR bool  // that value's run begins at the MFLR itself
	markedAtSTW  bool  // the same run still covers the STW
	closed       bool  // the run ends immediately after the STW
	insnCount    int   // real instructions inside the run
	width        int   // insnCount*4: the byte width the runtime tests against 20
	worstDispl   int   // largest pc-startpc a preemption could observe
}

// scorePrologueWindow reports how the unsafe-point table covers a function's
// MFLR/STW prologue, if it has one.
//
// It models what the runtime actually does, rather than pattern-matching on the
// marker progs: linkpcln collapses the Prog list into a table of (value, pc)
// runs, and pcdatavalue2 hands isAsyncSafePoint the value in effect at the
// interrupted pc together with the pc at which that value's run began. So this
// walks the list once to record, for every REAL instruction, the unsafe-point
// value in effect there and the index of the first real instruction of that
// value's run, then reads the prologue pair out of that model. PCDATA progs are
// skipped in the real-instruction index because they carry no bytes -- which is
// precisely why they cannot shift the window's width.
func scorePrologueWindow(text *obj.Prog) windowVerdict {
	v := windowVerdict{linkReg: -1}

	type insn struct {
		p        *obj.Prog
		val      int64
		runStart int // index of the first real insn in this value's run
	}
	var insns []*insn
	val, runStart := int64(abi.UnsafePointSafe), 0
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			if p.To.Offset != val {
				val, runStart = p.To.Offset, len(insns)
			}
			continue
		}
		insns = append(insns, &insn{p: p, val: val, runStart: runStart})
	}

	// Locate the prologue pair: an MFLR into a GPR whose very next real
	// instruction stores that same register to the caller's linkage area.
	i := -1
	for k := 0; k+1 < len(insns); k++ {
		a, b := insns[k].p, insns[k+1].p
		if a.As == AMFLR && a.To.Type == obj.TYPE_REG && b.As == ASTW &&
			b.From.Type == obj.TYPE_REG && b.From.Reg == a.To.Reg &&
			b.To.Type == obj.TYPE_MEM && b.To.Reg == REGSP && b.To.Offset == 8 {
			i, v.linkReg = k, a.To.Reg
			break
		}
	}
	if i < 0 {
		return v
	}
	v.pairFound = true

	val1 := insns[i].val
	v.markedAtMFLR = isRestartValue(val1)
	v.startsAtMFLR = v.markedAtMFLR && insns[i].runStart == i
	v.markedAtSTW = insns[i+1].val == val1 && insns[i+1].runStart == i

	// End of the run: the first real instruction no longer under val1.
	j := i
	for j < len(insns) && insns[j].val == val1 {
		j++
	}
	v.insnCount = j - i
	v.width = v.insnCount * insnSize
	// The window is closed only if it stopped exactly at the STW: j must be i+2.
	v.closed = j == i+2
	// Worst case the runtime measures: a preemption taken on the last instruction
	// still inside the run. Only instructions up to the STW can be preempted here
	// if the window is closed; if it leaked, every leaked instruction counts.
	last := j - 1
	if v.closed {
		last = i + 1
	}
	if last >= i {
		v.worstDispl = (last - i) * insnSize
	}
	return v
}

func isRestartValue(val int64) bool {
	return val == abi.UnsafePointRestart1 || val == abi.UnsafePointRestart2
}

func nonLeafFixture(ctxt *obj.Link) *obj.LSym {
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 0, Val: int32(4)}}
	load := &obj.Prog{As: ALWZ, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_PARAM, Reg: REGSP}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	ret := &obj.Prog{As: ABLR}
	call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("callee")}, Link: ret}
	text.Link, load.Link = load, call
	s := ctxt.Lookup("fixture")
	s.NewFuncInfo().Text = text
	return s
}

func TestPrologueRestartWindow(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := nonLeafFixture(ctxt)
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	text := s.Func().Text

	// 1. Exact shape: Restart1 opens before the MFLR, Safe is restored directly
	// after the STW, and no real instruction was inserted between them.
	first := text.Link
	if !isUnsafePointPCDATA(first) || first.To.Offset != abi.UnsafePointRestart1 {
		t.Fatalf("first prog after TEXT = %v (unsafe point %d), want PCDATA_UnsafePoint=%d (Restart1)",
			first, first.To.Offset, abi.UnsafePointRestart1)
	}
	mflr, stw, after := first.Link, first.Link.Link, first.Link.Link.Link
	if mflr.As != AMFLR || mflr.To.Type != obj.TYPE_REG || mflr.To.Reg != REG_R0 {
		t.Fatalf("second prog = %v, want MFLR R0", mflr)
	}
	if stw.As != ASTW || stw.From.Type != obj.TYPE_REG || stw.From.Reg != REG_R0 ||
		stw.To.Type != obj.TYPE_MEM || stw.To.Reg != REGSP || stw.To.Offset != 8 {
		t.Fatalf("third prog = %v, want STW R0,8(R1)", stw)
	}
	if !isUnsafePointPCDATA(after) || after.To.Offset != abi.UnsafePointSafe {
		t.Fatalf("prog after the STW = %v (unsafe point %d), want PCDATA_UnsafePoint restored to %d (Safe)",
			after, after.To.Offset, abi.UnsafePointSafe)
	}
	// The store is already committed by here, so R0 no longer matters and the
	// window must not reach over this instruction.
	if after.Link == nil || after.Link.As != AADDI || after.Link.To.Reg != REG_R0 {
		t.Fatalf("instruction after the window = %v, want ADDI zeroing R0", after.Link)
	}

	// 2. No code bytes at entry, proven on the assembled words: mflr r0 /
	// stw r0,8(r1) / li r0,0, the pre-fix shape, byte for byte.
	assemble(ctxt, s, nil)
	word := func(i int) uint32 { return binary.BigEndian.Uint32(s.P[4*i:]) }
	for i, w := range []uint32{0x7c0802a6, 0x90010008, 0x38000000} {
		if word(i) != w {
			t.Errorf("prologue word %d = %08x, want %08x -- the marker must not emit code", i, word(i), w)
		}
	}

	// 3. Width, from the PCs the assembler actually assigned. The run spans the
	// pair and nothing else.
	if stw.Pc-mflr.Pc != insnSize {
		t.Fatalf("STW is %d bytes from the MFLR, want %d", stw.Pc-mflr.Pc, insnSize)
	}
	width := after.Link.Pc - mflr.Pc
	if width > 20 {
		t.Fatalf("Restart1 window is %d bytes wide; runtime/preempt.go throws \"bad restart PC\" above 20", width)
	}
	if width != 2*insnSize {
		t.Errorf("Restart1 window is %d bytes wide, want %d (two instructions)", width, 2*insnSize)
	}
	// A preemption is only taken where the value reads Restart1 -- the MFLR or
	// the STW -- so the worst case the runtime sees is pc-startpc = 4.
	if worst := stw.Pc - mflr.Pc; worst > 20 {
		t.Fatalf("worst-case pc-startpc is %d, over the runtime's 20-byte limit", worst)
	}

	// 4. The scorer agrees with the hand-checked shape above.
	if got := scorePrologueWindow(text); !got.pairFound || !got.markedAtMFLR ||
		!got.startsAtMFLR || !got.markedAtSTW || !got.closed || got.width != 2*insnSize ||
		got.worstDispl != insnSize || got.linkReg != REG_R0 {
		t.Fatalf("scorer rejects the correct window: %+v", got)
	}

	// 5. A leaf function emits no prologue and so must carry no marker: an
	// unreachable Restart1 would advertise a sequence that does not exist.
	ctxt2 := obj.Linknew(&Linkppc)
	ctxt2.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	ltext := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 0, Val: int32(4)}}
	ltext.Link = &obj.Prog{As: ABLR}
	ls := ctxt2.Lookup("leafty")
	ls.NewFuncInfo().Text = ltext
	preprocess(ctxt2, ls, func() *obj.Prog { return new(obj.Prog) })
	for p := ltext.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			t.Errorf("leaf function carries an unsafe-point marker: %v", p)
		}
		if p.As == AMFLR {
			t.Errorf("leaf function carries a prologue MFLR: %v", p)
		}
	}
	if got := scorePrologueWindow(ltext); got.pairFound {
		t.Errorf("scorer finds a prologue pair in a leaf function: %+v", got)
	}
}

// TestPrologueRestartWindowScorerFailsOnBrokenShapes keeps the gate honest. The
// shapes below are the ways this fix could plausibly be wrong, each built by
// hand; the scorer must reject every one. If a future edit makes the scorer
// accept any of them, that is the moment the gate stopped being evidence.
func TestPrologueRestartWindowScorerFailsOnBrokenShapes(t *testing.T) {
	// The correct window, as a baseline for the four rejections below.
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(string, ...interface{}) {}
	s := nonLeafFixture(ctxt)
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	good := scorePrologueWindow(s.Func().Text)
	if !good.markedAtMFLR || !good.markedAtSTW || !good.closed || good.width != 2*insnSize {
		t.Fatalf("baseline good window scored bad: %+v", good)
	}

	mkr := func(val int64) *obj.Prog {
		return &obj.Prog{As: obj.APCDATA,
			From: obj.Addr{Type: obj.TYPE_CONST, Offset: abi.PCDATA_UnsafePoint},
			To:   obj.Addr{Type: obj.TYPE_CONST, Offset: val}}
	}
	pair := func() (text, mflr, stw, zero, ret *obj.Prog) {
		text = &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 0, Val: int32(4)}}
		mflr = &obj.Prog{As: AMFLR, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}}
		stw = &obj.Prog{As: ASTW, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0},
			To: obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: 8}}
		zero = &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0,
			To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}}
		ret = &obj.Prog{As: ABLR}
		text.Link, mflr.Link, stw.Link, zero.Link = mflr, stw, zero, ret
		return
	}

	t.Run("no-marker-at-all", func(t *testing.T) {
		text, _, _, _, _ := pair() // exactly the pre-fix toolchain's output
		got := scorePrologueWindow(text)
		if !got.pairFound || got.markedAtMFLR || got.startsAtMFLR {
			t.Fatalf("unmarked prologue accepted as covered: %+v", got)
		}
	})

	t.Run("window-opens-after-the-MFLR", func(t *testing.T) {
		text, mflr, stw, _, _ := pair()
		mflr.Link = mkr(abi.UnsafePointRestart1) // marker sits between MFLR and STW
		mflr.Link.Link = stw
		got := scorePrologueWindow(text)
		if got.startsAtMFLR || !got.pairFound {
			t.Fatalf("window starting past the MFLR accepted: %+v", got)
		}
	})

	t.Run("window-never-closes", func(t *testing.T) {
		text, mflr, stw, zero, ret := pair()
		text.Link = mkr(abi.UnsafePointRestart1)
		text.Link.Link = mflr
		_ = stw
		_ = zero
		_ = ret // no restoring marker anywhere: the run leaks to the end
		got := scorePrologueWindow(text)
		if !got.markedAtMFLR {
			t.Fatalf("expected the MFLR to read as marked here, got %+v", got)
		}
		if got.closed {
			t.Fatalf("leaking window reported as closed: %+v", got)
		}
		if got.width <= 2*insnSize {
			t.Fatalf("leaking window did not measure wider than the pair: %+v", got)
		}
	})

	t.Run("window-wider-than-20-bytes", func(t *testing.T) {
		text, mflr, stw, zero, ret := pair()
		// Eight instructions under one Restart1 run: 32 bytes, over the limit.
		extra := make([]*obj.Prog, 4)
		for i := range extra {
			extra[i] = &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: int64(i)},
				Reg: REG_R3, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
		}
		text.Link = mkr(abi.UnsafePointRestart1)
		text.Link.Link = mflr
		mflr.Link = stw
		stw.Link = extra[0]
		zero.Link = ret
		for i := 0; i < 3; i++ {
			extra[i].Link = extra[i+1]
		}
		extra[3].Link = zero
		got := scorePrologueWindow(text)
		if got.width <= 20 || !got.markedAtMFLR || !got.markedAtSTW || got.closed {
			t.Fatalf("over-wide window not detected as over-wide: %+v", got)
		}
		if got.width != 8*insnSize || got.worstDispl != 7*insnSize {
			t.Fatalf("over-wide window measured width=%d worstDispl=%d, want %d and %d",
				got.width, got.worstDispl, 8*insnSize, 7*insnSize)
		}
	})

	t.Run("link-register-is-not-R0", func(t *testing.T) {
		// The rejected goppc-hqfk shape: safe by construction, so it must be
		// reported as such rather than as a violation of this gate.
		text, mflr, stw, zero, ret := pair()
		mflr.To.Reg = REG_R31
		stw.From.Reg = REG_R31
		text.Link = mkr(abi.UnsafePointRestart1)
		text.Link.Link = mflr
		_ = zero
		_ = ret
		got := scorePrologueWindow(text)
		if got.linkReg != REG_R31 || !got.pairFound {
			t.Fatalf("scorer did not report the link register it saw: %+v", got)
		}
	})
}
