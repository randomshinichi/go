// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/sys"
	"encoding/binary"
	"fmt"
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

	// autosize is the local frame size the compiler put on the TEXT, exactly as
	// ppc64 reads it. It is what turns a name's own offset into an SP-relative
	// displacement (see resolveOffset).
	autosize := p.To.Offset

	// Assign function-relative PCs before encoding so local branches can resolve
	// their block targets. This is an assembler fixup only; external/linker
	// relocations remain a separate ppc32 linker task.
	pc := int64(0)
	for q := p.Link; q != nil; q = q.Link {
		q.Pc = pc
		if q.As != obj.ATEXT && q.As != obj.ANOP && q.As != obj.AFUNCDATA && q.As != obj.APCDATA {
			pc += 4
		}
	}
	words := make([]uint32, 0, 16)
	for q := p.Link; q != nil; q = q.Link {
		w, ok := encodeProg(ctxt, q, autosize)
		if !ok {
			return // diag already issued
		}
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

// resolveOffset turns an operand's displacement into the byte offset the
// instruction carries.
//
// A NAME_AUTO or NAME_PARAM reference is a reference into the current function's
// stack frame; its final address is SP-relative and needs no relocation, because
// the frame is laid out before the instruction is encoded. The compiler's offset
// is relative to the frame base, so the frame size is added for a local and the
// fixed linkage area as well for a parameter -- the same arithmetic ppc64 and
// arm use, and inheriting the same open question about what FixedFrameSize is on
// Darwin ppc32 (DESIGN.md §11 Q-3).
//
// Anything else that carries a symbol (NAME_EXTERN, NAME_STATIC, NAME_GOTREF,
// TLS) needs a relocation record, and where a ppc32 symbol reference has to be
// split (lis/ori, or a pc-relative form) depends on the 32-bit Mach-O relocation
// set, which the Mach-O lane owns (Q-4). Emitting the instruction without its
// relocation would be a silently wrong address, so those are refused by name.
func resolveOffset(ctxt *obj.Link, p *obj.Prog, a *obj.Addr, autosize int64) (int64, error) {
	switch a.Name {
	case obj.NAME_NONE:
		if a.Sym != nil {
			return 0, fmt.Errorf("%v: %v has a symbol but no name to resolve it with", p, Mnemonic(Inst(p.As)))
		}
		return a.Offset, nil
	case obj.NAME_AUTO:
		return autosize + a.Offset, nil
	case obj.NAME_PARAM:
		return autosize + a.Offset + ctxt.Arch.FixedFrameSize, nil
	}
	return 0, fmt.Errorf("%v: %v refers to %s (%v), and the ppc32 relocation path does not exist yet (brief 3.4)",
		p, Mnemonic(Inst(p.As)), a.Sym, a.Name)
}

// encodeProg returns the words for one Prog, or ok=false after issuing a
// diagnostic. A nil, true result is a pseudo-instruction that occupies no space.
func encodeProg(ctxt *obj.Link, p *obj.Prog, autosize int64) (w []uint32, ok bool) {
	switch p.As {
	case obj.ATEXT, obj.ANOP, obj.AFUNCDATA, obj.APCDATA:
		return nil, true
	case obj.ADUFFZERO, obj.ADUFFCOPY:
		ctxt.Diag("ppc32: %v needs the runtime's duffzero/duffcopy, which does not exist for this target yet", p)
		return nil, false
	}

	bad := func(err error) ([]uint32, bool) {
		ctxt.Diag("ppc32: %v", err)
		return nil, false
	}

	// Operand convention, shared with cmd/compile/internal/ppc/ssa.go:
	//
	//	op rd,ra,rb   To=rd, From.Reg=ra, Reg=rb
	//	op rd,ra      To=rd, From.Reg=ra
	//	op rd,ra,si   To=rd, Reg=ra, From={TYPE_CONST|TYPE_ADDR, Offset=si}
	//	op ra,rs,ui   To=ra, Reg=rs, From={TYPE_CONST, Offset=ui}
	//	op rd,d(ra)   To=rd, From={TYPE_MEM, Reg=ra, Offset=d, Name, Sym}
	//	op rs,d(ra)   To={TYPE_MEM, Reg=ra, Offset=d, Name, Sym}, From=rs
	//
	// `rd`/`rs` name one field in asm.go's table; which spelling a form uses is
	// the form's business, not the assembler's.
	rd := func() int32 { return int32(p.To.Reg - REG_R0) }
	ra := func() int32 { return int32(p.From.Reg - REG_R0) }
	rs := func() int32 { return int32(p.Reg - REG_R0) }
	rb := func() int32 { return int32(p.Reg - REG_R0) }
	reg := func() int32 { return int32(p.Reg - REG_R0) }

	emit := func(args ...int32) ([]uint32, bool) {
		enc, err := Encode(Inst(p.As), args...)
		if err != nil {
			return bad(err)
		}
		return []uint32{enc}, true
	}

	switch p.As {
	// Three-register forms: To=rd, From=ra, Reg=rb.
	case AADD, AADDC, AADDE, ASUBF, ASUBFC, ASUBFE,
		AMULLW, AMULHW, AMULHWU,
		AAND, AANDC, AOR, AORC, AXOR, ANAND, ANOR, AEQV,
		ASLW, ASRW, ASRAW, AROTLW:
		return emit(rd(), ra(), rb())

	// Two-register forms: To=rd, From=ra.
	case ANEG, ACNTLZW, AEXTSB, AEXTSH:
		return emit(rd(), ra())

	// Register-immediate forms rd,ra,si: To=rd, Reg=ra, From.Offset=si. A stack
	// name arrives as TYPE_ADDR; an ordinary constant as TYPE_CONST.
	case AADDI, AADDIC, AADDIS:
		if p.From.Type != obj.TYPE_CONST && p.From.Type != obj.TYPE_ADDR {
			return bad(fmt.Errorf("%v: %v needs a constant operand", p, Mnemonic(Inst(p.As))))
		}
		off, err := resolveOffset(ctxt, p, &p.From, autosize)
		if err != nil {
			return bad(err)
		}
		return emit(rd(), reg(), int32(off))

	// Immediate forms that write their first operand: To=ra, Reg=rs,
	// From.Offset=the immediate (ui or sh).
	case AORI, AXORI, AORIS, AXORIS, AANDI_, AANDIS_, ASRAWI:
		return emit(rd(), rs(), int32(p.From.Offset))

	// Loads: To=rd, From=d(ra).
	case ALWZ, ALBZ, ALHZ, ALHA:
		if p.From.Type != obj.TYPE_MEM {
			return bad(fmt.Errorf("%v: %v needs a memory operand", p, Mnemonic(Inst(p.As))))
		}
		d, err := resolveOffset(ctxt, p, &p.From, autosize)
		if err != nil {
			return bad(err)
		}
		return emit(rd(), int32(d), ra())

	// Stores: To=d(ra), From=rs.
	case ASTW, ASTB, ASTH:
		if p.To.Type != obj.TYPE_MEM {
			return bad(fmt.Errorf("%v: %v needs a memory destination", p, Mnemonic(Inst(p.As))))
		}
		d, err := resolveOffset(ctxt, p, &p.To, autosize)
		if err != nil {
			return bad(err)
		}
		return emit(int32(p.From.Reg-REG_R0), int32(d), int32(p.To.Reg-REG_R0))

	case AB, ABC:
		if p.To.Type != obj.TYPE_BRANCH || p.To.Target() == nil {
			return bad(fmt.Errorf("%v: %v has no local branch target", p, Mnemonic(Inst(p.As))))
		}
		disp := p.To.Target().Pc - p.Pc
		if p.As == AB {
			return emit(int32(disp))
		}
		return emit(int32(p.From.Offset), int32(p.Reg), int32(disp))

	case ABLR:
		return emit()
	}

	return bad(fmt.Errorf("no ppc32 encoding for %v (as %v)", p, p.As))
}

// symOperand returns the operand carrying a symbol, if any. One symbol per Prog
// is all the ppc32 baseline forms can have.
func symOperand(p *obj.Prog) *obj.Addr {
	for _, a := range []*obj.Addr{&p.From, &p.To} {
		if a.Sym != nil {
			return a
		}
	}
	return nil
}
