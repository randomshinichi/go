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
// The call pair needs no window, but the automatically inserted MFLR R0 / STW
// prologue does: a restart must reload LR before retrying its store. The window
// check below is load-bearing alongside leaf classification and the encoded
// call pair. BLRL carries no symbol, but still requires a frame and saved LR.
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
	}
	if err := indirectCallWindowError(text); err != nil {
		t.Error(err)
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

// indirectCallWindowError checks the preprocessed NOSPLIT fixture, not the
// compiler's pre-prologue stream. The only allowed unsafe-point markers bracket
// the LR-save pair; the restart destination must be the MFLR itself. The shared
// scorer models value/run-start at each instruction, not just marker presence.
func indirectCallWindowError(text *obj.Prog) error {
	foundMarker := false
	for p := text.Link; p != nil; p = p.Link {
		foundMarker = foundMarker || isUnsafePointPCDATA(p)
	}
	if !foundMarker {
		return fmt.Errorf("missing LR-save unsafe-point markers")
	}
	v := scorePrologueWindow(text)
	if !v.pairFound || v.linkReg != REG_R0 {
		return fmt.Errorf("missing MFLR R0 / STW R0,8(SP) LR-save pair: %+v", v)
	}
	if !v.markedAtMFLR || !v.startsAtMFLR || !v.markedAtSTW {
		return fmt.Errorf("LR-save restart must begin at MFLR and cover STW: %+v", v)
	}
	if !v.closed || v.width != 2*insnSize || v.worstDispl != insnSize {
		return fmt.Errorf("LR-save restart must end immediately after STW: %+v", v)
	}
	markers := 0
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			markers++
			value := p.To.Offset
			want := int64(abi.UnsafePointRestart1)
			if markers == 2 {
				want = abi.UnsafePointSafe
			}
			if markers > 2 || value != want {
				return fmt.Errorf("unexpected unsafe-point marker %d: value %d, want %d", markers, value, want)
			}
			continue
		}
		if p.As == AMTCTR || p.As == ABCTRL {
			return fmt.Errorf("CTR-dependent indirect call emitted")
		}
	}
	// Exactly this Restart1/Safe bracket, closing after the prologue STW,
	// leaves no unsafe window over the later MTLR/BLRL call pair.
	if markers != 2 {
		return fmt.Errorf("got %d unsafe-point markers, want 2", markers)
	}
	return nil
}

// These controls mutate fresh output from preprocess, as the epilogue controls
// do, rather than restating an expected prologue by hand. Each must reach the
// same checker as the load-bearing test and fail for its own defect.
func TestIndirectCallWindowRejectsBrokenShapes(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"no-markers", "missing LR-save unsafe-point markers"},
		{"restart-after-MFLR", "LR-save restart must begin at MFLR and cover STW"},
		{"window-covers-call", "LR-save restart must end immediately after STW"},
		{"wrong-link-slot", "missing MFLR R0 / STW R0,8(SP) LR-save pair"},
		{"Restart2", "unexpected unsafe-point marker 1"},
		{"extra-call-marker", "unexpected unsafe-point marker 3"},
		{"CTR-call", "CTR-dependent indirect call emitted"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, text := indirectCallFixture(t, true)
			if err := indirectCallWindowError(text); err != nil {
				t.Fatalf("unmutated fixture: %v", err)
			}
			open := text.Link
			mflr, stw := open.Link, open.Link.Link
			close := stw.Link
			switch tc.name {
			case "no-markers":
				text.Link, stw.Link = mflr, close.Link
			case "restart-after-MFLR":
				text.Link, mflr.Link, open.Link = mflr, open, stw
			case "window-covers-call":
				stw.Link = close.Link
				for p := stw; p != nil; p = p.Link {
					if p.As == ABLRL {
						close.Link, p.Link = p.Link, close
						break
					}
				}
			case "wrong-link-slot":
				stw.To.Offset = 0
			case "Restart2":
				open.To.Offset = abi.UnsafePointRestart2
			case "extra-call-marker":
				for p := close; p != nil; p = p.Link {
					if p.Link != nil && p.Link.As == AMTLR && p.Link.From.Reg == REG_R12 {
						marker := *open
						marker.Link, p.Link = p.Link, &marker
						break
					}
				}
			case "CTR-call":
				for p := text.Link; p != nil; p = p.Link {
					if p.As == AMTLR && p.From.Reg == REG_R12 {
						p.As = AMTCTR
					}
					if p.As == ABLRL {
						p.As = ABCTRL
					}
				}
			}
			err := indirectCallWindowError(text)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("broken emitted shape: got %v, want rejection %q", err, tc.want)
			}
			t.Logf("%s rejected: %v", tc.name, err)
		})
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
