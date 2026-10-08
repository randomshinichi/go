// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/sys"
	"encoding/binary"
)

// Linkppc is the assembler interface for GOARCH=ppc. It is a sibling of
// obj/ppc64's Linkppc64, not a variant of it: the 32-bit target has its own
// register width, its own frame layout and its own instruction subset
// (DESIGN.md §10).
var Linkppc = obj.LinkArch{
	Arch:           sys.ArchPPC,
	Init:           buildop,
	Preprocess:     preprocess,
	Assemble:       assemble,
	Progedit:       progedit,
	DWARFRegisters: ppcDWARFRegisters,
	UnaryDst:       unaryDst,
}

// unaryDst lists the ppc32 forms whose operand is written rather than read, so
// the assembler's dependency tracking does not treat it as a source.
var unaryDst = map[obj.As]bool{
	ALWZ:   true,
	ALBZ:   true,
	ALHZ:   true,
	ALHA:   true,
	AMFLR:  true,
	AMFCTR: true,
}

// preprocess and progedit are the two per-Prog fix-up hooks obj calls on the way
// to the assembler. ppc32 needs neither yet: it emits no pseudo-instructions that
// have to be expanded (a return is already the architectural blr), and the
// pc-relative and relocation handling they would normally do belongs to the
// Mach-O/linker lane (DESIGN.md §13 brief 3.4). They exist so that the assembler
// path is a no-op rather than a nil call.
func preprocess(*obj.Link, *obj.LSym, obj.ProgAlloc) {}

func progedit(*obj.Link, *obj.Prog, obj.ProgAlloc) {}

// buildop is the per-Link initialisation hook. The ppc32 instruction table is
// package-level immutable data (asm.go), not per-Link state, so there is nothing
// to rebuild; the hook exists because obj.LinkArch requires it and because a
// future relocation table would be installed here.
func buildop(*obj.Link) {}

// assemble turns the Progs of one function into instruction words. It uses the
// same form table as Encode: this function chooses *which* operands a Prog
// carries, Encode does the masking and range checking. Refusing an op by name is
// deliberate — an assembler that quietly emits nothing is worse than one that
// fails, and this slice of the backend only lowers the scalar forms below.
func assemble(ctxt *obj.Link, s *obj.LSym, _ obj.ProgAlloc) {
	p := s.Func().Text
	if p == nil || p.Link == nil {
		return // external function or section symbol
	}

	words := make([]uint32, 0, 16)
	for q := p.Link; q != nil; q = q.Link {
		if q.As == obj.ATEXT {
			continue
		}
		w, ok := encodeProg(ctxt, q)
		if !ok {
			return // diag already issued
		}
		if w == nil {
			continue // zero-width pseudo-instruction
		}
		q.Pc = int64(len(words) * 4)
		words = append(words, w...)
	}

	s.Size = int64(len(words) * 4)
	s.Grow(s.Size)
	bp := s.P
	for _, w := range words {
		binary.BigEndian.PutUint32(bp, w)
		bp = bp[4:]
	}
}

// encodeProg returns the words for one Prog, or ok=false after issuing a
// diagnostic. A nil, true result is a pseudo-instruction that occupies no space.
func encodeProg(ctxt *obj.Link, p *obj.Prog) (w []uint32, ok bool) {
	switch p.As {
	case obj.ATEXT, obj.ANOP, obj.AFUNCDATA, obj.APCDATA:
		return nil, true
	case obj.ADUFFZERO, obj.ADUFFCOPY:
		ctxt.Diag("ppc32: %v needs the runtime's duffzero/duffcopy, which does not exist for this target yet", p)
		return nil, false
	}

	bad := func(format string, args ...any) ([]uint32, bool) {
		ctxt.Diag("ppc32: "+format, args...)
		return nil, false
	}

	// Operand convention, shared with cmd/compile/internal/ppc/ssa.go:
	//
	//	op rd,ra,rb   To=rd, From.Reg=ra, Reg=rb
	//	op rd,ra      To=rd, From.Reg=ra
	//	op rd,ra,si   To=rd, Reg=ra, From={TYPE_CONST, Offset=si}
	//	op ra,rs,ui   To=ra, Reg=rs, From={TYPE_CONST, Offset=ui}
	//
	// `rd`/`rs` name one field in asm.go's table; which spelling a form uses is
	// the form's business, not the assembler's.
	rd := func() int32 { return int32(p.To.Reg - REG_R0) }
	ra := func() int32 { return int32(p.From.Reg - REG_R0) }
	rs := func() int32 { return int32(p.Reg - REG_R0) }
	rb := func() int32 { return int32(p.Reg - REG_R0) }
	reg := func() int32 { return int32(p.Reg - REG_R0) }

	switch p.As {
	// Three-register forms: To=rd, From=ra, Reg=rb.
	case AADD, AADDC, AADDE, ASUBF, ASUBFC, ASUBFE,
		AMULLW, AMULHW, AMULHWU,
		AAND, AANDC, AOR, AORC, AXOR, ANAND, ANOR, AEQV,
		ASLW, ASRW, ASRAW, AROTLW:
		enc, err := Encode(Inst(p.As), rd(), ra(), rb())
		if err != nil {
			return bad("%v: %v", p, err)
		}
		return []uint32{enc}, true

	// Two-register forms: To=rd, From=ra.
	case ANEG, ACNTLZW, AEXTSB, AEXTSH:
		enc, err := Encode(Inst(p.As), rd(), ra())
		if err != nil {
			return bad("%v: %v", p, err)
		}
		return []uint32{enc}, true

	// Register-immediate forms rd,ra,si: To=rd, Reg=ra, From.Offset=si.
	case AADDI, AADDIC, AADDIS:
		if p.From.Type != obj.TYPE_CONST {
			return bad("%v: %v needs a constant operand", p, Mnemonic(Inst(p.As)))
		}
		enc, err := Encode(Inst(p.As), rd(), reg(), int32(p.From.Offset))
		if err != nil {
			return bad("%v: %v", p, err)
		}
		return []uint32{enc}, true

	// Immediate forms that write their first operand: To=ra, Reg=rs,
	// From.Offset=the immediate (ui or sh).
	case AORI, AXORI, AORIS, AXORIS, AANDI_, AANDIS_, ASRAWI:
		enc, err := Encode(Inst(p.As), rd(), rs(), int32(p.From.Offset))
		if err != nil {
			return bad("%v: %v", p, err)
		}
		return []uint32{enc}, true

	case ABLR:
		enc, _ := Encode(ABLR)
		return []uint32{enc}, true
	}

	return bad("no ppc32 encoding for %v (as %v)", p, p.As)
}
