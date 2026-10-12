// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"bytes"
	"cmd/internal/obj"
	"fmt"
	"internal/abi"
	"sort"
	"testing"
)

// The ppc32 split prologue saves LR into the caller's linkage area and then
// calls runtime.stackcheck (and runtime.morestack on the slow path) BEFORE the
// frame exists. Those calls overwrite LR with an address inside the function
// itself, while the PCSP delta is still zero. At such a PC the unwinder, which
// reads LR from the register for a frameless innermost frame, would name this
// function as its own caller; right after the BL, PC == LR and SP == FP, which
// is the "traceback stuck" throw in runtime/traceback.go. (Bead goppc-1ye4.)
//
// Marking the run PCDATA_UnsafePoint = Unsafe keeps those PCs out of
// isAsyncSafePoint. The scorer below counts every real instruction that is
//
//	frameless (no Spadj has been applied before it), AND
//	reachable with LR clobbered by one of the prologue's own calls, AND
//	not marked UnsafePointUnsafe.
//
// "Reachable" follows the control flow, not the listing order: the guard check is
// a loop (morestack returns to the recheck), so the instructions at the top of
// the loop run a second time with a stale LR. A listing-order scorer would miss
// them.

type staleLRReport struct {
	stale  []int // pc offsets from TEXT of every frameless, stale-LR instruction not marked Unsafe
	insns  int   // frameless, stale-LR instructions in total, marked or not
	allocs int   // frame-allocating instructions found (must be 1)
}

// scoreStackCheckStaleLR models the unsafe-point table the way linkpcln builds
// it (the value in effect at each real instruction, PCDATA and NOP carrying no
// bytes) and walks the control flow from every prologue call. text must have been
// preprocessed AND assembled so that Pc is meaningful.
func scoreStackCheckStaleLR(text *obj.Prog) staleLRReport {
	var r staleLRReport
	type insn struct {
		p   *obj.Prog
		val int64
	}
	var insns []*insn
	index := map[*obj.Prog]int{}
	val := int64(abi.UnsafePointSafe)
	pending := []*obj.Prog{} // no-width progs waiting for the next real instruction
	flush := func(i int) {
		for _, q := range pending {
			index[q] = i
		}
		pending = pending[:0]
	}
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			val = p.To.Offset
			continue
		}
		if p.As == obj.APCDATA || p.As == obj.AFUNCDATA || p.As == obj.ANOP {
			pending = append(pending, p)
			continue
		}
		flush(len(insns))
		index[p] = len(insns)
		insns = append(insns, &insn{p: p, val: val})
	}

	// Frameless: no Spadj on any earlier instruction. The allocating instruction
	// itself is still frameless at its own PC (the delta applies after it).
	alloc := -1
	for i, in := range insns {
		if in.p.Spadj != 0 && alloc < 0 {
			alloc = i
		}
	}
	if alloc < 0 {
		return r
	}
	r.allocs = 1

	// LR is clobbered after any call that comes before the allocation; walk
	// forward from each such call's successor.
	succ := func(i int) []int {
		p := insns[i].p
		var out []int
		target := func() {
			if p.To.Type == obj.TYPE_BRANCH && p.To.Target() != nil {
				if j, ok := index[p.To.Target()]; ok {
					out = append(out, j)
				}
			}
		}
		switch p.As {
		case AB:
			target()
		case ABC:
			target()
			out = append(out, i+1)
		default:
			out = append(out, i+1)
		}
		return out
	}
	seen := map[int]bool{}
	var work []int
	for i := 0; i < alloc; i++ {
		if insns[i].p.As == ABL {
			work = append(work, i+1)
		}
	}
	for len(work) > 0 {
		i := work[len(work)-1]
		work = work[:len(work)-1]
		if i > alloc || i >= len(insns) || seen[i] {
			continue
		}
		seen[i] = true
		if i < alloc {
			work = append(work, succ(i)...)
		}
	}
	for i := range seen {
		r.insns++
		if insns[i].val != abi.UnsafePointUnsafe {
			r.stale = append(r.stale, int(insns[i].p.Pc-text.Pc))
		}
	}
	sort.Ints(r.stale)
	return r
}

// stripUnsafeMarkers deletes every unsafe-point marker whose value is Unsafe,
// reproducing the pre-fix emitter. It reports how many it removed.
func stripUnsafeMarkers(text *obj.Prog) int {
	n := 0
	for prev := text; prev.Link != nil; {
		if p := prev.Link; isUnsafePointPCDATA(p) && p.To.Offset == abi.UnsafePointUnsafe {
			prev.Link = p.Link
			n++
			continue
		}
		prev = prev.Link
	}
	return n
}

func assembledSplitFixture(t *testing.T, locals int64) (*obj.LSym, *obj.Prog) {
	t.Helper()
	ctxt, s, text := splitFixture(t, locals)
	assemble(ctxt, s, nil)
	return s, text
}

func TestStackCheckWindowIsUnsafe(t *testing.T) {
	for _, tc := range []struct {
		name   string
		locals int64
		// every frameless, stale-LR instruction: ADDI frame, [BL stackcheck, CMPWI,
		// BC, LWZ R5, BL morestack, B] ; STWU. The large-frame form needs ADDIS+ADDI
		// for the guard constant and again for the allocation constant.
		total int
		// pc offsets (from TEXT) the review measured on the enough-stack branch of
		// the 40-byte frame: right after the BL, the BC, and the STWU.
		measured []int
	}{
		{"small-frame", 24, 8, []int{0x14, 0x18, 0x28}},
		{"indexed-frame", 65512, 11, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Fixed arm: no instruction on a stale-LR, frameless path is
			// advertised as anything but Unsafe.
			_, text := assembledSplitFixture(t, tc.locals)
			got := scoreStackCheckStaleLR(text)
			if got.allocs != 1 {
				t.Fatalf("scorer found %d frame allocations, want 1", got.allocs)
			}
			if got.insns != tc.total {
				t.Fatalf("scorer sees %d frameless stale-LR instructions, want %d (it is not measuring the prologue it thinks it is)\n%s",
					got.insns, tc.total, dumpProgs(text))
			}
			if len(got.stale) != 0 {
				t.Fatalf("frameless stale-LR PCs advertised as safe: %#x\n%s", got.stale, dumpProgs(text))
			}

			// Control arm, same fixture shape: deleting the marker must make the
			// scorer find every one of them. If it reads 0 here the zero above
			// proves nothing.
			ctxt, s, ctext := splitFixture(t, tc.locals)
			if stripUnsafeMarkers(ctext) == 0 {
				t.Fatal("no Unsafe marker to delete: the fixed arm did not carry one")
			}
			assemble(ctxt, s, nil)
			broken := scoreStackCheckStaleLR(ctext)
			if len(broken.stale) != tc.total {
				t.Fatalf("with the marker removed the scorer reports %d PCs %#x, want %d",
					len(broken.stale), broken.stale, tc.total)
			}
			have := map[int]bool{}
			for _, pc := range broken.stale {
				have[pc] = true
			}
			for _, pc := range tc.measured {
				if !have[pc] {
					t.Errorf("review-measured stale PC +%#x missing from the unmarked arm %#x", pc, broken.stale)
				}
			}
		})
	}
}

// TestStackCheckWindowAddsNoCodeBytes: the marker is PCDATA only. The assembled
// bytes with and without it must be identical, so frame and nosplit accounting
// are untouched.
func TestStackCheckWindowAddsNoCodeBytes(t *testing.T) {
	for _, locals := range []int64{24, 65512} {
		s, _ := assembledSplitFixture(t, locals)
		ctxt, s2, text2 := splitFixture(t, locals)
		stripUnsafeMarkers(text2)
		assemble(ctxt, s2, nil)
		if !bytes.Equal(s.P, s2.P) {
			t.Errorf("locals=%d: marker changed the emitted code (%d vs %d bytes)", locals, len(s.P), len(s2.P))
		}
	}
}

// TestStackCheckWindowShape pins where the window opens and closes. The scorer
// above catches an uncovered PC; this catches the two leaks in the other
// direction: a window that stays open past the allocation (hiding the body) and a
// close that restores the wrong value.
func TestStackCheckWindowShape(t *testing.T) {
	for _, locals := range []int64{24, 65512} {
		_, text := assembledSplitFixture(t, locals)
		progs := progsOf(text)
		start, end := -1, -1
		for i, p := range progs {
			if isUnsafePointPCDATA(p) && p.To.Offset == abi.UnsafePointUnsafe {
				start = i
			}
			if p.Spadj > 0 && end < 0 {
				end = i
			}
		}
		if start < 0 || end < 0 {
			t.Fatalf("locals=%d: no Unsafe marker (%d) or no allocation (%d)\n%s", locals, start, end, dumpProgs(text))
		}
		// The first real instruction after the opening marker is the guard's
		// frame-size materialization (or its NOP), and the allocation is followed by
		// the marker restoring Safe.
		if !(progs[end+1].As == obj.APCDATA && progs[end+1].To.Offset == abi.UnsafePointSafe) {
			t.Errorf("locals=%d: prog after the allocation is %v, want PCDATA restoring Safe", locals, progs[end+1].As)
		}
		for i := start + 1; i < end; i++ {
			if isUnsafePointPCDATA(progs[i]) {
				t.Errorf("locals=%d: unsafe-point marker inside the window at %d", locals, i)
			}
		}
	}
}

// TestStackCheckWindowScorerFailsOnBrokenShapes hand-breaks the marker in the
// ways the fix could plausibly be wrong; the scorer must flag each.
func TestStackCheckWindowScorerFailsOnBrokenShapes(t *testing.T) {
	mark := func(p *obj.Prog, val int64) *obj.Prog {
		return &obj.Prog{As: obj.APCDATA, Link: p,
			From: obj.Addr{Type: obj.TYPE_CONST, Offset: abi.PCDATA_UnsafePoint},
			To:   obj.Addr{Type: obj.TYPE_CONST, Offset: val}}
	}
	for name, mutate := range map[string]func(text *obj.Prog) bool{
		// Opens one instruction late: the ADDI that rematerializes the frame size
		// runs again after morestack with a stale LR.
		"opens-after-the-first-instruction": func(text *obj.Prog) bool {
			stripUnsafeMarkers(text)
			for p := text; p != nil; p = p.Link {
				if p.As == AADDI && p.To.Reg == REG_R3 {
					p.Link = mark(p.Link, abi.UnsafePointUnsafe)
					return true
				}
			}
			return false
		},
		// Closes before the allocation: the STWU runs frameless with stale LR.
		"closes-before-the-allocation": func(text *obj.Prog) bool {
			for p := text; p.Link != nil; p = p.Link {
				if p.Link.Spadj > 0 {
					p.Link = mark(p.Link, abi.UnsafePointSafe)
					return true
				}
			}
			return false
		},
		// Covers only the call to stackcheck.
		"covers-only-the-stackcheck-call": func(text *obj.Prog) bool {
			stripUnsafeMarkers(text)
			for prev := text; prev.Link != nil; prev = prev.Link {
				if b := prev.Link; b.As == ABL && b.To.Sym != nil && b.To.Sym.Name == "runtime.stackcheck" {
					b.Link = mark(b.Link, abi.UnsafePointSafe)
					prev.Link = mark(b, abi.UnsafePointUnsafe)
					return true
				}
			}
			return false
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctxt, s, text := splitFixture(t, 24)
			if !mutate(text) {
				t.Fatal("mutation found nothing to change")
			}
			assemble(ctxt, s, nil)
			got := scoreStackCheckStaleLR(text)
			if len(got.stale) == 0 {
				t.Fatalf("scorer accepts the broken marker:\n%s", dumpProgs(text))
			}
			t.Logf("%s: flagged %s", name, fmt.Sprintf("%#x", got.stale))
		})
	}
}
