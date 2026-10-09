// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
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

// preprocess implements the Go stack ABI. Allocatable GPRs/FPRs are
// caller-saved (DESIGN D-2); SSA spills values live across calls into locals.
// There is no C callee-save block and no register argument shadow area.
// Stack growth is a separate runtime milestone: these frames do not split.
func preprocess(ctxt *obj.Link, s *obj.LSym, newprog obj.ProgAlloc) {
	text := s.Func().Text
	if text == nil {
		return
	}
	locals := text.To.Offset
	if locals < 0 || locals > 32752 {
		ctxt.Diag("ppc32: TEXT frame size %d outside supported D-form frame range", locals)
		return
	}
	leaf := true
	for p := text.Link; p != nil; p = p.Link {
		switch p.As {
		case obj.ACALL, ABL, ABCTRL, obj.ADUFFCOPY, obj.ADUFFZERO:
			leaf = false
		}
	}
	frame := int64(0)
	if locals != 0 || !leaf {
		frame = (locals + ctxt.Arch.FixedFrameSize + 15) &^ 15
	}
	// Stack operands currently use signed D-form displacements; refuse a
	// larger frame rather than truncate its allocation or unwind metadata.
	if frame > 32752 {
		ctxt.Diag("ppc32: frame %d exceeds supported D-form frame range", frame)
		return
	}
	if s.NoFrame() && frame != 0 {
		ctxt.Diag("ppc32: NOFRAME function requires an allocated frame")
		return
	}
	s.Func().Locals = int32(locals)
	if args, ok := text.To.Val.(int32); ok {
		s.Func().Args = args
	}
	text.To.Offset = frame
	s.Set(obj.AttrLeaf, leaf)

	reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
	mem := func(off int64) obj.Addr { return obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: off} }
	appendOp := func(p *obj.Prog, as obj.As, from, to obj.Addr) *obj.Prog {
		q := obj.Appendp(p, newprog)
		q.As, q.From, q.To = as, from, to
		return q
	}
	p := text
	if !leaf {
		p = appendOp(p, AMFLR, obj.Addr{}, reg(REG_R0))
		// Save in the caller's linkage area BEFORE moving SP, as measured.
		p = appendOp(p, ASTW, reg(REG_R0), mem(8))
		// Go treats R0 as zero for stores (not merely the RA=0 address rule).
		p = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST}, reg(REG_R0))
		p.Reg = REG_R0
	}
	if frame != 0 {
		p = appendOp(p, ASTWU, reg(REGSP), mem(-frame))
		p.Spadj = int32(frame)
	}
	for p := text.Link; p != nil; p = p.Link {
		if p.As == obj.AGETCALLERPC {
			if leaf {
				p.As = AMFLR
			} else {
				p.As = ALWZ
				p.From = mem(frame + 8) // saved LR in caller's linkage area
			}
		}
		if p.As != ABLR && p.As != obj.ARET {
			continue
		}
		if p.From.Type != obj.TYPE_NONE {
			ctxt.Diag("ppc32: unsupported BECOME: %v", p)
			continue
		}
		// As on ppc64, RET symbol is a tail call, not a call followed by
		// a return. Keep the target while replacing RET with the epilogue.
		// SSA's tail-call expansion reuses incoming arguments/results, so
		// restoring SP exposes the original caller's result area unchanged.
		target := p.To
		retAs := ABLR
		if target.Sym != nil {
			retAs = AB
		} else {
			target = obj.Addr{}
		}
		if frame == 0 {
			p.As, p.To = retAs, target
			continue
		}
		// Rewrite in place so branches to a return also execute its epilogue.
		p.As, p.From, p.To = AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: frame}, reg(REGSP)
		p.Reg = REGSP
		p.Spadj = -int32(frame)
		if !leaf {
			p = appendOp(p, ALWZ, mem(8), reg(REG_R31))
			p = appendOp(p, AMTLR, reg(REG_R31), obj.Addr{})
		}
		p = appendOp(p, retAs, obj.Addr{}, target)
		// Balance the linear PC/SP table after the return (ppc64 does likewise).
		p.Spadj = int32(frame)
	}
}

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
	// Assembly input PCs span the file; pcln needs a function-relative TEXT
	// origin just like the instruction PCs below (including a second TEXT).
	p.Pc = 0

	// Assign function-relative PCs before encoding so local branches can resolve
	// their block targets. This is an assembler fixup only; external/linker
	// relocations remain a separate ppc32 linker task.
	pc := int64(0)
	for q := p.Link; q != nil; q = q.Link {
		q.Pc = pc
		if q.As != obj.ATEXT && q.As != obj.ANOP && q.As != obj.AFUNCDATA && q.As != obj.APCDATA {
			pc += 4
			if isSymbolAddress(q) {
				pc += 4 // addis + addi, not one addi
			}
		}
	}
	words := make([]uint32, 0, 16)
	for q := p.Link; q != nil; q = q.Link {
		w, ok := encodeProg(ctxt, q, autosize)
		if !ok {
			return // diag already issued
		}
		if isSymbolAddress(q) {
			s.AddRel(ctxt, obj.Reloc{Off: int32(q.Pc), Siz: 8, Type: objabi.R_ADDRPOWER, Sym: q.From.Sym, Add: q.From.Offset})
		} else if (q.As == obj.ACALL || q.As == ABL || q.As == AB || q.As == obj.AJMP) && symOperand(q) != nil {
			target := symOperand(q)
			s.AddRel(ctxt, obj.Reloc{Off: int32(q.Pc), Siz: 4, Type: objabi.R_CALLPOWER, Sym: target.Sym, Add: target.Offset})
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

// isSymbolAddress recognizes the only external address form implemented here:
// an absolute address relative to the compiler's pseudo-SB, not PIC/GOT/TLS.
func isSymbolAddress(p *obj.Prog) bool {
	return p.As == AADDI && p.From.Type == obj.TYPE_ADDR &&
		(p.From.Name == obj.NAME_EXTERN || p.From.Name == obj.NAME_STATIC) &&
		p.From.Sym != nil && p.Reg == 0 // pseudo-SB has no hardware register
}

// resolveOffset turns an operand's displacement into the byte offset the
// instruction carries.
//
// A NAME_AUTO or NAME_PARAM reference is a reference into the current function's
// stack frame; its final address is SP-relative and needs no relocation, because
// the frame is laid out before the instruction is encoded. The compiler's offset
// is relative to the frame base, so the frame size is added for a local and the
// fixed linkage area as well for a parameter -- the same arithmetic ppc64 and
// arm use. FixedFrameSize is the measured 24-byte linkage area; autosize is
// the total allocated, 16-byte-aligned frame, or zero for a frameless leaf.
//
// Symbol operands are not displacements. Only the absolute address pair in
// encodeProg is supported; other external, GOT and TLS accesses are refused
// rather than silently emitting an address without its relocation.
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
	case ALWARX:
		return emit(rd(), ra(), rb())

	case ASTWCX_:
		return emit(int32(p.From.Reg-REG_R0), int32(p.To.Reg-REG_R0), rb())

	case AADD, AADDC, AADDE, ASUBF, ASUBFC, ASUBFE,
		AMULLW, AMULHW, AMULHWU, ADIVW, ADIVWU,
		AAND, AANDC, AOR, AORC, AXOR, ANAND, ANOR, AEQV,
		ASLW, ASRW, ASRAW, AROTLW:
		return emit(rd(), ra(), rb())

	// Two-register forms: To=rd, From=ra.
	case ANEG, ACNTLZW, AEXTSB, AEXTSH:
		return emit(rd(), ra())

	// Register-immediate forms rd,ra,si: To=rd, Reg=ra, From.Offset=si. A stack
	// name arrives as TYPE_ADDR; an ordinary constant as TYPE_CONST.
	case AADDI, AADDIC, AADDIS:
		if isSymbolAddress(p) {
			// R_ADDRPOWER supplies high-adjusted/low signed halves to this
			// D-form pair. Leave both immediates zero in the object.
			hi, err := Encode(AADDIS, rd(), 0, 0)
			if err != nil {
				return bad(err)
			}
			lo, err := Encode(AADDI, rd(), rd(), 0)
			if err != nil {
				return bad(err)
			}
			return []uint32{hi, lo}, true
		}
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
	case ALWZ, ALBZ, ALHZ, ALHA, ALFS, ALFD:
		if p.From.Type != obj.TYPE_MEM && !(p.From.Type == obj.TYPE_ADDR && (p.From.Name == obj.NAME_AUTO || p.From.Name == obj.NAME_PARAM)) {
			return bad(fmt.Errorf("%v: %v needs a memory operand", p, Mnemonic(Inst(p.As))))
		}
		d, err := resolveOffset(ctxt, p, &p.From, autosize)
		if err != nil {
			return bad(err)
		}
		base := ra()
		if p.From.Name == obj.NAME_AUTO || p.From.Name == obj.NAME_PARAM {
			base = int32(REGSP - REG_R0)
		}
		dest := rd()
		if p.As == ALFS || p.As == ALFD {
			dest = int32(p.To.Reg - REG_F0)
		}
		return emit(dest, int32(d), base)

	// Stores: To=d(ra), From=rs.
	case ASTW, ASTWU, ASTB, ASTH, ASTFS, ASTFD:
		if p.To.Type != obj.TYPE_MEM && !(p.To.Type == obj.TYPE_ADDR && (p.To.Name == obj.NAME_AUTO || p.To.Name == obj.NAME_PARAM)) {
			return bad(fmt.Errorf("%v: %v needs a memory destination", p, Mnemonic(Inst(p.As))))
		}
		d, err := resolveOffset(ctxt, p, &p.To, autosize)
		if err != nil {
			return bad(err)
		}
		base := p.To.Reg - REG_R0
		if p.To.Name == obj.NAME_AUTO || p.To.Name == obj.NAME_PARAM {
			base = REGSP - REG_R0
		}
		src := int32(p.From.Reg - REG_R0)
		if p.As == ASTFS || p.As == ASTFD {
			src = int32(p.From.Reg - REG_F0)
		}
		return emit(src, int32(d), int32(base))

	// All integer comparisons target CR0. Unlike a three-GPR operation,
	// From and To are the two inputs, not a source and a destination.
	case ACMPW, ACMPLW:
		if p.From.Type != obj.TYPE_REG || p.To.Type != obj.TYPE_REG {
			return bad(fmt.Errorf("%v: word compare needs two registers", p))
		}
		return emit(0, ra(), rd())
	case ACMPWI, ACMPLWI:
		if p.From.Type != obj.TYPE_REG || p.To.Type != obj.TYPE_CONST {
			return bad(fmt.Errorf("%v: immediate compare needs a register and constant", p))
		}
		return emit(0, ra(), int32(p.To.Offset))

	case AFMR, AFRSP, AFNEG, AFCTIWZ:
		return emit(int32(p.To.Reg-REG_F0), int32(p.From.Reg-REG_F0))
	case AFADD, AFADDS, AFSUB, AFSUBS, AFMUL, AFMULS, AFDIV, AFDIVS:
		return emit(int32(p.To.Reg-REG_F0), int32(p.From.Reg-REG_F0), int32(p.Reg-REG_F0))
	case AFCMPU:
		return emit(0, int32(p.From.Reg-REG_F0), int32(p.To.Reg-REG_F0))

	case obj.ACALL, ABL:
		if p.To.Sym != nil && p.To.Name == obj.NAME_EXTERN && p.To.Type == obj.TYPE_MEM {
			if p.To.Offset&3 != 0 {
				return bad(fmt.Errorf("%v: unaligned direct call addend", p))
			}
			enc, err := Encode(ABL, 0)
			if err != nil {
				return bad(err)
			}
			return []uint32{enc}, true
		}
		if p.As == ABL && p.To.Type == obj.TYPE_BRANCH && p.To.Target() != nil {
			return emit(int32(p.To.Target().Pc - p.Pc))
		}
		return bad(fmt.Errorf("%v: direct call needs an external symbol or local BL target", p))

	case obj.AJMP, AB, ABC:
		target := &p.To
		if p.As == obj.AJMP && p.To.Type == obj.TYPE_NONE {
			target = &p.From // generic asm represents JMP's symbol in From
		}
		if p.As == obj.AJMP && (target.Type == obj.TYPE_REG || target.Type == obj.TYPE_MEM) && target.Name == obj.NAME_NONE {
			var as obj.As
			switch target.Reg {
			case REG_CTR:
				as = ABCTR
			case REG_LR:
				as = ABLR
			default:
				return bad(fmt.Errorf("%v: indirect JMP target must be CTR or LR", p))
			}
			enc, err := Encode(Inst(as))
			if err != nil {
				return bad(err)
			}
			return []uint32{enc}, true
		}
		if (p.As == obj.AJMP || p.As == AB) && target.Sym != nil &&
			(target.Type == obj.TYPE_BRANCH || (target.Type == obj.TYPE_MEM && (target.Name == obj.NAME_EXTERN || target.Name == obj.NAME_STATIC))) {
			if target.Offset&3 != 0 {
				return bad(fmt.Errorf("%v: unaligned tail branch addend", p))
			}
			// JMP's generic assembler Prog lowers to the same non-linking
			// relative branch as AB. R_CALLPOWER resolves the symbolic target.
			enc, err := Encode(AB, 0)
			if err != nil {
				return bad(err)
			}
			return []uint32{enc}, true
		}
		if target.Type != obj.TYPE_BRANCH || target.Target() == nil {
			return bad(fmt.Errorf("%v: %v has no local branch target", p, Mnemonic(Inst(p.As))))
		}
		disp := target.Target().Pc - p.Pc
		if p.As == obj.AJMP || p.As == AB {
			enc, err := Encode(AB, int32(disp))
			if err != nil {
				return bad(err)
			}
			return []uint32{enc}, true
		}
		return emit(int32(p.From.Offset), int32(p.Reg), int32(disp))

	case ASYNC, AISYNC, ASC:
		return emit()

	case AMFCR, AMFCTR:
		if p.From.Type == obj.TYPE_REG {
			return emit(ra())
		}
		return emit(rd())
	case AMFLR:
		if p.To.Type == obj.TYPE_REG {
			return emit(rd()) // linker-generated prologue
		}
		return emit(ra())
	case AMTCRF:
		return emit(int32(p.From.Offset), int32(p.To.Reg-REG_R0))
	case AMTLR, AMTCTR:
		return emit(ra())
	case ABLR, ABLRL, ABCTR, ABCTRL:
		return emit()
	case ABCLR, ABCLRL, ABCCTR, ABCCTRL:
		return emit(int32(p.From.Offset), int32(p.Reg-REG_R0), int32(p.To.Offset))
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
