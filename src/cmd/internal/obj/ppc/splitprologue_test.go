// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

// TestSplitPrologueRecheck pins the control flow of the prologue the ppc32
// emitter generates for a splittable frame. The guard is an inline comparison
// against g->stackguard0 (it used to be `BL runtime.stackcheck`, which made the
// returning PC fatal for the unwinder -- bead goppc-h890), and the return from
// runtime.morestack must branch back to that guard rather than fall through into
// the frame allocation.
//
// The reason is runtime.newstack: it can return without having grown the stack
// at all (its preemption path hands the goroutine back later, and
// runtime.shrinkstack sizes a stack from a stack pointer this frame has not
// been subtracted from yet, runtime/stack.go). A function that resumes from
// morestack onto such a stack and allocates K immediately has an
// under-provisioned frame. ppc64 has always had the back edge ("BR start",
// obj/ppc64/obj9.go); ppc32 was missing it, which was latent until route A
// gave the reflect call wrappers a declared frame.
//
// The test is written to fail on the pre-fix emitter: it also deletes the back
// edge from the emitted prologue to reproduce the pre-fix shape and requires
// the same checker to reject it, so a checker that cannot fail cannot pass.
func TestSplitPrologueRecheck(t *testing.T) {
	for _, tc := range []struct {
		name   string
		locals int64
	}{
		{"d-form-frame", 4096},
		{"indexed-frame", 65512},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The emitted control flow.
			_, _, text := splitFixture(t, tc.locals)
			frame := text.To.Offset
			if ok, why := checkSplitPrologueRecheck(text, frame, int64(Linkppc.Arch.PtrSize)); !ok {
				t.Fatalf("generated split prologue does not re-check the guard after morestack: %s\n%s", why, dumpProgs(text))
			}

			// Negative control: the pre-fix shape, i.e. this prologue with
			// the back edge deleted, must be rejected for that reason.
			if !removeBackEdge(text) {
				t.Fatal("no back edge to delete: the positive check above cannot have been real")
			}
			ok, why := checkSplitPrologueRecheck(text, frame, int64(Linkppc.Arch.PtrSize))
			if ok {
				t.Fatal("checker accepts the pre-fix prologue, whose morestack return falls into the allocation")
			}
			if !strings.Contains(why, "back to the guard check") {
				t.Fatalf("checker rejects the pre-fix prologue, but for the wrong reason: %s", why)
			}

			// The encoding. A backward branch is the one path a forward-only
			// prologue never exercises, so check the word, not just the Prog.
			ctxt, s, clean := splitFixture(t, tc.locals)
			back := onlyBackEdge(clean)
			if back == nil {
				t.Fatal("no back edge in a fresh fixture")
			}
			assemble(ctxt, s, nil)
			disp := back.To.Target().Pc - back.Pc
			if disp >= 0 {
				t.Fatalf("back edge at pc %d targets pc %d, which is not earlier", back.Pc, back.To.Target().Pc)
			}
			want, err := Encode(AB, int32(disp))
			if err != nil {
				t.Fatalf("back-edge displacement %d does not encode: %v", disp, err)
			}
			if got := binary.BigEndian.Uint32(s.P[back.Pc:]); got != want {
				t.Errorf("back edge word = %08x, want %08x (b %d)", got, want, disp)
			}
		})
	}
}

// TestSplitPrologueRecheckNotForUnsplitFrames is the boundary of the change:
// frames that do not split get no guard check at all, so nothing about them
// (assembly functions that hand-write their own stackcheck block included) may
// acquire a morestack loop.
func TestSplitPrologueRecheckNotForUnsplitFrames(t *testing.T) {
	for _, tc := range []struct {
		name   string
		attr   obj.Attribute
		locals int64
	}{
		{"nosplit", obj.AttrNoSplit, 4096},
		{"noframe", obj.AttrNoFrame, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt := obj.Linknew(&Linkppc)
			ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
			s := ctxt.Lookup("unsplit")
			s.Set(tc.attr, true)
			text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: tc.locals, Val: int32(4)}}
			text.Link = &obj.Prog{As: obj.ARET}
			s.NewFuncInfo().Text = text
			preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
			for _, p := range progsOf(text) {
				if p.As == ABL && p.To.Sym != nil {
					t.Errorf("%s function gained a call to %s", tc.name, p.To.Sym.Name)
				}
			}
			if back := onlyBackEdge(text); back != nil {
				t.Errorf("%s function gained a back edge: %v", tc.name, back)
			}
		})
	}
}

// splitFixture builds the smallest split-capable function: locals plus a
// return, preprocessed but not assembled.
func splitFixture(t *testing.T, locals int64) (*obj.Link, *obj.LSym, *obj.Prog) {
	t.Helper()
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("split")
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals, Val: int32(4)}}
	text.Link = &obj.Prog{As: obj.ARET}
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	if frame := text.To.Offset; frame <= 0 {
		t.Fatalf("fixture frame %d is not split-capable", frame)
	}
	return ctxt, s, text
}

func progsOf(text *obj.Prog) []*obj.Prog {
	var progs []*obj.Prog
	for p := text.Link; p != nil; p = p.Link {
		progs = append(progs, p)
	}
	return progs
}

func progIndex(progs []*obj.Prog) map[*obj.Prog]int {
	pos := make(map[*obj.Prog]int, len(progs))
	for i, p := range progs {
		pos[p] = i
	}
	return pos
}

// onlyBackEdge returns the single unconditional local branch in the function
// whose target is an earlier Prog, or nil.
func onlyBackEdge(text *obj.Prog) *obj.Prog {
	progs := progsOf(text)
	pos := progIndex(progs)
	var found *obj.Prog
	for _, p := range progs {
		if p.As != AB || p.To.Type != obj.TYPE_BRANCH || p.To.Target() == nil {
			continue
		}
		if j, ok := pos[p.To.Target()]; ok && j < pos[p] {
			found = p
		}
	}
	return found
}

// removeBackEdge deletes the emitted back edge, reproducing the pre-fix
// emitter's control flow. It reports whether it found one.
func removeBackEdge(text *obj.Prog) bool {
	pos := progIndex(progsOf(text))
	for prev := text; prev != nil && prev.Link != nil; prev = prev.Link {
		p := prev.Link
		if p.As != AB || p.To.Type != obj.TYPE_BRANCH || p.To.Target() == nil {
			continue
		}
		if j, ok := pos[p.To.Target()]; ok && j < pos[p] {
			prev.Link = p.Link
			return true
		}
	}
	return false
}

// checkSplitPrologueRecheck verifies the generated split prologue's control
// flow end to end: the frame size is materialized for the inline guard, the
// guard loads g->stackguard0 through the g register, a failed check calls
// runtime.morestack, and the morestack return branches back to the
// rematerialization of the frame size, which falls through to the same guard
// load. It returns a reason when the shape is wrong.
func checkSplitPrologueRecheck(text *obj.Prog, frame, ptrSize int64) (bool, string) {
	progs := progsOf(text)
	pos := progIndex(progs)

	guard, morestack, alloc := -1, -1, -1
	for i, p := range progs {
		if stackGuardLoad(p, ptrSize) {
			if guard < 0 {
				guard = i
			}
		}
		if p.As == ABL && p.To.Sym != nil {
			switch p.To.Sym.Name {
			case "runtime.stackcheck":
				// The old shape. It is the defect, not the fix: the branch-and-link
				// leaves the next PC with pc == lr and no frame established.
				return false, "the guard is a call to runtime.stackcheck again, which makes the PC it returns to fatal for the unwinder"
			case "runtime.morestack", "runtime.morestack_noctxt":
				morestack = i
			}
		}
		if alloc < 0 && p.Spadj != 0 {
			alloc = i
		}
	}
	switch {
	case guard < 0:
		return false, "no inline stack-guard load of g->stackguard0"
	case morestack < 0:
		return false, "no call to runtime.morestack"
	case morestack < guard:
		return false, "the runtime.morestack call precedes the guard check"
	case alloc < 0:
		return false, "no frame allocation"
	case alloc < morestack:
		return false, "the frame allocation precedes the runtime.morestack call"
	}

	back := -1
	for i := morestack + 1; i < alloc; i++ {
		p := progs[i]
		if p.As != AB || p.To.Type != obj.TYPE_BRANCH || p.To.Target() == nil {
			continue
		}
		if j, ok := pos[p.To.Target()]; ok && j < i {
			back = i
		}
	}
	if back < 0 {
		return false, "no unconditional branch back to the guard check after runtime.morestack"
	}

	// It must land on the frame size rematerialization, and that must be the
	// guard the morestack call was reached from -- not merely somewhere earlier in
	// the function. Zero-width progs (the loop anchor ANOP and the unsafe-point
	// marker that opens the region at it) do not start a PC of their own, so the
	// branch lands on the first real instruction after them.
	at := pos[progs[back].To.Target()]
	for at < len(progs) && isZeroWidthProg(progs[at]) {
		at++
	}
	if at >= len(progs) {
		return false, "the back edge lands past the end of the prologue"
	}
	val, next, ok := frameSizeMaterialized(progs, at)
	if !ok {
		return false, "the back edge does not land on a materialization of the frame size"
	}
	if val != frame {
		return false, fmt.Sprintf("the back edge rematerializes %d, want the frame size %d", val, frame)
	}
	if next > guard {
		return false, "the rematerialized frame size does not fall through to the stack-guard load"
	}
	// Falling through is the whole contract, because the guard consumes the frame
	// size in R3. Nothing in between may branch-and-link: the PC such a call
	// returns to would read pc == lr with the frame not yet established, which is
	// the fatal combination bead goppc-h890 is about.
	for _, p := range progs[next:guard] {
		if setsLinkRegister(p) {
			return false, "a branch-and-link sits between the frame size and the guard load"
		}
	}

	// The check runs on the caller's SP: no instruction on the loop may move
	// it. Positive Spadj means SP was pushed.
	for i := at; i <= morestack; i++ {
		if progs[i].Spadj != 0 {
			return false, fmt.Sprintf("stack delta %d on the guard-check loop at %v", progs[i].Spadj, progs[i].As)
		}
	}
	if progs[alloc].Spadj != int32(frame) {
		return false, fmt.Sprintf("frame allocation delta %d, want %d", progs[alloc].Spadj, frame)
	}
	return true, ""
}

// frameSizeMaterialized decodes the frame-size materialization that feeds the
// inline guard's subtraction: either a single `addi $K,r0,r3`, or the
// `addis`/`addi` pair a frame too large for D-form needs. next is the index of
// the first Prog after it.
func frameSizeMaterialized(progs []*obj.Prog, at int) (val int64, next int, ok bool) {
	p := progs[at]
	switch {
	case p.As == AADDI && p.To.Type == obj.TYPE_REG && p.To.Reg == REG_R3 && p.Reg == REG_R0:
		return p.From.Offset, at + 1, true
	case p.As == AADDIS && p.To.Type == obj.TYPE_REG && p.To.Reg == REG_R3 && p.Reg == REG_R0:
		if at+1 >= len(progs) {
			return 0, 0, false
		}
		q := progs[at+1]
		if q.As != AADDI || q.To.Type != obj.TYPE_REG || q.To.Reg != REG_R3 || q.Reg != REG_R3 {
			return 0, 0, false
		}
		return (p.From.Offset << 16) + q.From.Offset, at + 2, true
	}
	return 0, 0, false
}

func dumpProgs(text *obj.Prog) string {
	var b strings.Builder
	for p := text; p != nil; p = p.Link {
		fmt.Fprintf(&b, "\t%v\tSpadj=%d\n", p.As, p.Spadj)
	}
	return b.String()
}
