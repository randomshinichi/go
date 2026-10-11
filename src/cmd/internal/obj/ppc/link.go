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
	"internal/abi"
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
	if locals < 0 || locals > int64(^uint32(0)>>1)-ctxt.Arch.FixedFrameSize-15 {
		ctxt.Diag("ppc32: TEXT frame size %d exceeds the supported 32-bit stack range", locals)
		return
	}
	// ABLRL is the compiler's indirect call: a branch through LR with link, so
	// it carries no symbol operand and cannot be recognized as a call by name.
	// Missing it here would build a making-calls function as a leaf -- no frame,
	// no saved LR -- which is exactly the class of defect this port has already
	// paid for once.
	leaf := true
	for p := text.Link; p != nil; p = p.Link {
		switch p.As {
		case obj.ACALL, ABL, ABCTRL, ABLRL, obj.ADUFFCOPY, obj.ADUFFZERO:
			leaf = false
		}
	}
	// A split-capable function with locals calls the stack guard helper before
	// allocating its frame, so its incoming link register must be saved too.
	if locals != 0 && !s.NoSplit() {
		leaf = false
	}
	frame := int64(0)
	if locals != 0 || !leaf {
		frame = (locals + ctxt.Arch.FixedFrameSize + 15) &^ 15
	}
	if frame > int64(^uint32(0)>>1) {
		ctxt.Diag("ppc32: frame %d exceeds the supported 32-bit stack range", frame)
		return
	}
	if s.NoFrame() {
		if locals != 0 {
			ctxt.Diag("ppc32: NOFRAME function requires zero locals")
			return
		}
		frame = 0 // assembly manages SP and LR itself, even if it calls.
	}
	s.Func().Locals = int32(locals)
	if args, ok := text.To.Val.(int32); ok {
		s.Func().Args = args
	}
	text.To.Offset = frame
	s.Set(obj.AttrLeaf, leaf)

	// Describe this function's static stack-pointer movement, so that the SPWRITE
	// detector in assemble() can tell an accounted-for SP write from an
	// unannounced stack switch, and so the PCSP table that the unwinder (and the
	// linker's nosplit check, which reads the same table) sees describes this
	// target. This runs on the source Progs only: everything preprocess appends
	// below sets its own Spadj because it knows the frame it is building.
	//
	// The decision is per function, not per Prog, because a function that writes
	// SP in a form no static table can describe is marked FuncFlagSPWrite and the
	// unwinder then refuses to unwind through it: describing its other SP writes
	// would buy the unwinder nothing, while the linker's nosplit accounting reads
	// the same PCSP table and would start counting stack this function pushes by
	// hand. Leave such a function entirely undescribed; the detector marks it, as
	// it does on ppc64, and the nosplit budget is unchanged from before this
	// bookkeeping existed.
	describeSP := true
	for p := text.Link; p != nil; p = p.Link {
		if _, ok := staticSPAdj(p); !ok && writesSP(p) {
			describeSP = false
			break
		}
	}
	if describeSP {
		for p := text.Link; p != nil; p = p.Link {
			if adj, ok := staticSPAdj(p); ok {
				p.Spadj = adj
			}
		}
	}

	reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
	mem := func(off int64) obj.Addr { return obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: off} }
	appendOp := func(p *obj.Prog, as obj.As, from, to obj.Addr) *obj.Prog {
		q := obj.Appendp(p, newprog)
		q.As, q.From, q.To = as, from, to
		return q
	}
	appendFrameConst := func(p *obj.Prog, value int64, scratch int16) *obj.Prog {
		p = appendOp(p, AADDIS, obj.Addr{Type: obj.TYPE_CONST, Offset: (value + 0x8000) >> 16}, reg(scratch))
		p.Reg = REG_R0
		p = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: int64(int16(value))}, reg(scratch))
		p.Reg = scratch
		return p
	}
	p := text
	if !leaf && !s.NoFrame() {
		p = appendOp(p, AMFLR, obj.Addr{}, reg(REG_R0))
		// Save in the caller's linkage area BEFORE moving SP, as measured.
		p = appendOp(p, ASTW, reg(REG_R0), mem(8))
		// Go treats R0 as zero for stores (not merely the RA=0 address rule).
		p = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST}, reg(REG_R0))
		p.Reg = REG_R0
	}
	if frame != 0 && !s.NoSplit() {
		// The guard check is a LOOP, not a straight line. runtime.newstack can
		// return WITHOUT growing the stack: its preemption path may put the
		// goroutine back on a stack the shrinker measured while this frame was
		// still unallocated (runtime/stack.go: newstack, shrinkstack). If the
		// prologue then proceeded straight to the allocation, that allocation
		// would not have been accounted for by the shrink sizing, and the frame
		// would be under-provisioned -- silent stack corruption. So resume at
		// morestack's return goes back to re-run the guard, exactly as ppc64 does
		// (obj/ppc64/obj9.go emits BR start). The frame size is rematerialized
		// inside the loop (nothing in it survives the call), and the loop body
		// carries no SP delta, so re-entering it never disturbs the PCSP table.
		recheck := appendOp(p, obj.ANOP, obj.Addr{}, obj.Addr{})
		p = recheck
		if frame <= 32752 {
			p = appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: frame}, reg(REG_R3))
			p.Reg = REG_R0
		} else {
			p = appendFrameConst(p, frame, REG_R3)
		}
		p = appendOp(p, ABL, obj.Addr{}, obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("runtime.stackcheck")})
		p = appendOp(p, ACMPWI, reg(REG_R3), obj.Addr{Type: obj.TYPE_CONST})
		branch := appendOp(p, ABC, obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, obj.Addr{Type: obj.TYPE_BRANCH})
		branch.Reg = 2
		p = appendOp(branch, ALWZ, mem(8), reg(REG_R5))
		morestack := "runtime.morestack_noctxt"
		if s.NeedCtxt() {
			morestack = "runtime.morestack"
		}
		p = appendOp(p, ABL, obj.Addr{}, obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup(morestack)})
		// Unconditional back edge to the guard, mirroring ppc64's BR start. It MUST
		// precede `done`, else the "enough stack" branch skips past it into the
		// allocation and the back edge becomes dead code.
		p = appendOp(p, AB, obj.Addr{}, obj.Addr{Type: obj.TYPE_BRANCH})
		p.To.SetTarget(recheck)
		done := appendOp(p, obj.ANOP, obj.Addr{}, obj.Addr{})
		branch.To.SetTarget(done)
		p = done
	}
	if frame != 0 {
		if frame <= 32752 {
			p = appendOp(p, ASTWU, reg(REGSP), mem(-frame))
		} else {
			p = appendFrameConst(p, -frame, REG_R31)
			p = appendOp(p, ASTWUX, reg(REGSP), obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP})
			p.Reg = REG_R31
		}
		p.Spadj = int32(frame)
	}
	for p := text.Link; p != nil; p = p.Link {
		if p.As == obj.AGETCALLERPC {
			if leaf {
				p.As = AMFLR
			} else if frame <= 32752 {
				p.As = ALWZ
				p.From = mem(frame + 8) // saved LR in caller's linkage area
			} else {
				// Materialize SP+frame+8 because the linkage-area slot is
				// outside D-form reach for a large local frame.
				ret := p.To
				p.As = AADDIS
				p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: (frame + 8 + 0x8000) >> 16}
				p.Reg = REG_R0
				p.To = reg(REG_R31)
				q := appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: int64(int16(frame + 8))}, reg(REG_R31))
				q.Reg = REG_R31
				q = appendOp(q, AADD, reg(REGSP), reg(REG_R31))
				q.Reg = REG_R31
				q = appendOp(q, ALWZ, obj.Addr{Type: obj.TYPE_MEM, Reg: REG_R31, Offset: 8}, ret)
				p = q
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
		var stackAdj *obj.Prog
		if frame <= 32752 {
			p.As, p.From, p.To = AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: frame}, reg(REGSP)
			p.Reg = REGSP
			stackAdj = p
		} else {
			// Replace RET with the first instruction of the epilogue. Appending
			// after RET would leave the stack restoration unreachable.
			p.As = AADDIS
			p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: (frame + 0x8000) >> 16}
			p.Reg = REG_R0
			p.To = reg(REG_R31)
			q := appendOp(p, AADDI, obj.Addr{Type: obj.TYPE_CONST, Offset: int64(int16(frame))}, reg(REG_R31))
			q.Reg = REG_R31
			q = appendOp(q, AADD, reg(REGSP), reg(REGSP))
			q.Reg = REG_R31
			stackAdj = q
			p = q
		}
		stackAdj.Spadj = -int32(frame)
		if !leaf {
			p = appendOp(p, ALWZ, mem(8), reg(REG_R31))
			p = appendOp(p, AMTLR, reg(REG_R31), obj.Addr{})
		}
		p = appendOp(p, retAs, obj.Addr{}, target)
		// Balance the linear PC/SP table after the return (ppc64 does likewise).
		p.Spadj = int32(frame)
	}
}

// writesSP reports whether p assigns a value to the stack pointer at all.
// ppc32's operand convention puts the destination in To, so a register write to
// SP is visible there; the indexed store-with-update forms name the updated
// register as their base (To.Reg with a memory destination).
func writesSP(p *obj.Prog) bool {
	if p.To.Type == obj.TYPE_REG && p.To.Reg == REGSP {
		// A compare reads its operands and writes only CR, even though the
		// destination field carries the second register.
		switch p.As {
		case ACMPW, ACMPLW, ACMPWI, ACMPLWI, AFCMPU, AFCMPO:
			return false
		}
		return true
	}
	if (p.As == ASTWU || p.As == ASTWUX) && p.To.Type == obj.TYPE_MEM &&
		(p.To.Reg == REGSP || p.To.Name == obj.NAME_AUTO || p.To.Name == obj.NAME_PARAM) {
		return true
	}
	return false
}

// staticSPAdj returns the stack-pointer adjustment p performs when that
// adjustment is a compile-time constant, and whether it has one. Spadj is
// signed the way every other target signs it: positive means SP was pushed, so
// it is the negation of the value the instruction adds to SP.
//
//	ADDI $si, SP, SP      SP = SP + si     Spadj = -si
//	STWU RS, off(SP)      SP = SP + off    Spadj = -off
//
// This is the ppc32 spelling of the two cases ppc64 handles in
// obj/ppc64/obj9.go (AADD with a constant and AMOVDU): ppc32 spells the
// constant add ADDI, and the frame allocation STWU. An immediate outside
// D-form range is encoded by assemble's encoder as an ADDIS/ADDI pair, so the
// value in the Prog is not the whole adjustment; that case is deliberately not
// described.
func staticSPAdj(p *obj.Prog) (int32, bool) {
	switch p.As {
	case AADDI:
		if p.To.Type != obj.TYPE_REG || p.To.Reg != REGSP || p.Reg != REGSP || p.From.Type != obj.TYPE_CONST {
			return 0, false
		}
		if p.From.Offset < -32768 || p.From.Offset > 32767 {
			return 0, false
		}
		return -int32(p.From.Offset), true
	case ASTWU:
		if p.To.Type != obj.TYPE_MEM || p.To.Reg != REGSP || p.To.Name != obj.NAME_NONE {
			return 0, false
		}
		return -int32(p.To.Offset), true
	}
	return 0, false
}

// progedit is the per-Prog preprocess hook. The stack-pointer bookkeeping for
// this target is not per-Prog: whether an SP write is worth describing depends
// on the whole function, so it is done in preprocess, which is also where
// ppc64 does it (obj/ppc64/obj9.go).
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
func assemble(ctxt *obj.Link, s *obj.LSym, newprog obj.ProgAlloc) {
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
	// their block targets. Conditional branches have only a 16-bit displacement;
	// relax an out-of-range one to an inverted local branch around an unconditional
	// branch. The latter has a 26-bit displacement. Encode keeps checking both
	// instruction ranges, so a target beyond even the unconditional range fails.
	// External/linker relocations remain a separate ppc32 linker task.
	if newprog == nil {
		newprog = func() *obj.Prog { return new(obj.Prog) }
	}
	for {
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
		var relaxed bool
		for q := p.Link; q != nil; q = q.Link {
			if q.As != ABC || q.To.Type != obj.TYPE_BRANCH || q.To.Target() == nil {
				continue
			}
			disp := q.To.Target().Pc - q.Pc
			if disp >= -32768 && disp <= 32764 {
				continue
			}
			target := q.To.Target()
			long := obj.Appendp(q, newprog)
			long.As = AB
			long.To.Type = obj.TYPE_BRANCH
			long.To.SetTarget(target)
			skip := obj.Appendp(long, newprog)
			// ori r0,r0,0 is a real four-byte no-op: unlike ANOP it gives
			// the short conditional branch a concrete instruction to target.
			skip.As = AORI
			skip.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
			skip.Reg = REG_R0
			skip.To = obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}
			q.From.Offset ^= 8 // invert BO; taken means skip the long branch
			q.To.SetTarget(skip)
			relaxed = true
		}
		if !relaxed {
			break
		}
	}
	// PCs above include relaxation instructions and are now final for encoding.
	words := make([]uint32, 0, 16)
	for q := p.Link; q != nil; q = q.Link {
		// An assembly function that writes SP without a statically accounted
		// stack adjustment cannot be unwound from that instruction. Keep this
		// in sync with obj/ppc64: declared frames and Spadj-accounted updates
		// do not need SPWRITE.
		if writesSP(q) && q.Spadj == 0 {
			if f := s.Func(); f.FuncFlag&abi.FuncFlagSPWrite == 0 {
				f.FuncFlag |= abi.FuncFlagSPWrite
				if ctxt.Debugvlog || !ctxt.IsAsm {
					ctxt.Logf("auto-SPWRITE: %s %v\n", s.Name, q)
					if !ctxt.IsAsm {
						ctxt.Diag("invalid auto-SPWRITE in non-assembly")
						ctxt.DiagFlush()
						panic("bad SPWRITE")
					}
				}
			}
		}
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

	case ASTWX, ASTWUX:
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
		if off >= -32768 && off <= 32767 {
			return emit(rd(), reg(), int32(off))
		}
		hi, err := Encode(AADDIS, rd(), reg(), int32((off+0x8000)>>16))
		if err != nil {
			return bad(err)
		}
		lo, err := Encode(AADDI, rd(), rd(), int32(int16(off)))
		if err != nil {
			return bad(err)
		}
		return []uint32{hi, lo}, true

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
		if d >= -32768 && d <= 32767 {
			return emit(dest, int32(d), base)
		}
		// Materialize the displacement from zero, then add the base once below.
		hi, err := Encode(AADDIS, int32(REG_R31-REG_R0), 0, int32((d+0x8000)>>16))
		if err != nil {
			return bad(err)
		}
		lo, err := Encode(AADDI, int32(REG_R31-REG_R0), int32(REG_R31-REG_R0), int32(int16(d)))
		if err != nil {
			return bad(err)
		}
		add, err := Encode(AADD, int32(REG_R31-REG_R0), base, int32(REG_R31-REG_R0))
		if err != nil {
			return bad(err)
		}
		mem, err := Encode(Inst(p.As), dest, 0, int32(REG_R31-REG_R0))
		if err != nil {
			return bad(err)
		}
		return []uint32{hi, lo, add, mem}, true

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
		if d >= -32768 && d <= 32767 {
			return emit(src, int32(d), int32(base))
		}
		// Materialize the displacement from zero, then add the base once below.
		hi, err := Encode(AADDIS, int32(REG_R31-REG_R0), 0, int32((d+0x8000)>>16))
		if err != nil {
			return bad(err)
		}
		lo, err := Encode(AADDI, int32(REG_R31-REG_R0), int32(REG_R31-REG_R0), int32(int16(d)))
		if err != nil {
			return bad(err)
		}
		add, err := Encode(AADD, int32(REG_R31-REG_R0), int32(base), int32(REG_R31-REG_R0))
		if err != nil {
			return bad(err)
		}
		mem, err := Encode(Inst(p.As), src, 0, int32(REG_R31-REG_R0))
		if err != nil {
			return bad(err)
		}
		return []uint32{hi, lo, add, mem}, true

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

	// mfxer/mtxer are single-register SPR forms like mfcr/mtctr above, with the
	// SPR number in the encoding's base rather than an operand. The generic
	// assembler puts a lone register operand in From unless the op is in
	// unaryDst, so both positions are accepted, exactly as for mfcr.
	case AMFXER, AMTXER:
		if p.From.Type == obj.TYPE_REG {
			return emit(ra())
		}
		return emit(rd())
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
