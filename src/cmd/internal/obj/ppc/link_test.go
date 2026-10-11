// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// Words independently assembled with clang --target=powerpc-unknown-linux-gnu:
// b .+12; lis 3,(target+12)@ha; addi 3,3,(target+12)@l; bl callee+4; blr.
// A local branch crosses the expanded address pair, so a one-word PC estimate
// would produce the wrong displacement even if the pair itself were correct.
func TestLargeNamedAddressMaterialized(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Offset: 32768}}
	addr := &obj.Prog{
		As:   AADDI,
		From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_AUTO, Offset: 33356},
		Reg:  REGSP,
		To:   obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3},
	}
	text.Link = addr
	s := ctxt.Lookup("large")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	if len(s.P) != 8 {
		t.Fatalf("large named address emitted %d bytes, want addis/addi pair: %x", len(s.P), s.P)
	}
	if got := binary.BigEndian.Uint32(s.P[:4]); got != 0x3c610001 {
		t.Errorf("addis = %08x, want 3c610001", got)
	}
	if got := binary.BigEndian.Uint32(s.P[4:]); got != 0x3863024c {
		t.Errorf("addi = %08x, want 3863024c", got)
	}
}

func TestRelaxConditionalBranch(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	text := &obj.Prog{As: obj.ATEXT}
	cmp := &obj.Prog{As: ACMPWI, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}, To: obj.Addr{Type: obj.TYPE_CONST}}
	branch := &obj.Prog{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	text.Link, cmp.Link = cmp, branch
	var prev = branch
	for i := 0; i < 9000; i++ {
		p := &obj.Prog{As: AORI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}}
		prev.Link = p
		prev = p
	}
	falseExit := []*obj.Prog{
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 1}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}},
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 1}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}},
		{As: ASC},
	}
	for _, p := range falseExit {
		prev.Link = p
		prev = p
	}
	target := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	successExit := []*obj.Prog{
		target,
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 1}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}},
		{As: ASC},
	}
	for _, p := range successExit {
		prev.Link = p
		prev = p
	}
	branch.To.SetTarget(target)
	s := ctxt.Lookup("long-conditional")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	if s.Size != 4*(9000+10) {
		t.Fatalf("relaxed text size %d, want %d", s.Size, 4*(9000+10))
	}
	// The branch is inverted and reaches over the long AB and its local target.
	// The AB then reaches the original target, more than the BC range away.
	bc := binary.BigEndian.Uint32(s.P[4:8])
	if bc != 0x40820008 { // bc BO=4, BI=2, +8
		t.Fatalf("relaxed conditional = %08x, want 40820008", bc)
	}
	wantAB, err := Encode(AB, int32(target.Pc-branch.Pc-4))
	if err != nil {
		t.Fatalf("expected relaxed AB displacement to encode: %v", err)
	}
	if got := binary.BigEndian.Uint32(s.P[8:12]); got != wantAB {
		t.Fatalf("long branch = %08x, want %08x", got, wantAB)
	}
	if got := binary.BigEndian.Uint32(s.P[12:16]); got != 0x60000000 {
		t.Fatalf("skip target = %08x, want ori r0,r0,0", got)
	}
	qemu, err := exec.LookPath("qemu-ppc")
	if err != nil {
		t.Skip("qemu-ppc unavailable: relaxed-branch execution inconclusive")
	}
	for _, tc := range []struct{ input, want int }{{0, 0}, {1, 1}} {
		words := []uint32{0x38600000 | uint32(tc.input)} // li r3,input for cmpwi
		for i := 0; i < len(s.P); i += 4 {
			words = append(words, binary.BigEndian.Uint32(s.P[i:i+4]))
		}
		path := filepath.Join(t.TempDir(), "relaxed-branch.elf")
		writePPCELF(t, path, words)
		cmd := exec.Command(qemu, "-cpu", "750", path)
		if out, err := cmd.CombinedOutput(); err != nil {
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != tc.want {
				t.Fatalf("qemu input=%d: err=%v want exit=%d\n%s", tc.input, err, tc.want, out)
			}
		} else if tc.want != 0 {
			t.Fatalf("qemu input=%d exited 0, want %d", tc.input, tc.want)
		}
	}
}

func writePPCELF(t *testing.T, path string, words []uint32) {
	t.Helper()
	const offset, base = 0x1000, 0x10000
	b := make([]byte, offset+4*len(words))
	copy(b, []byte{0x7f, 'E', 'L', 'F', 1, 2, 1})
	be := binary.BigEndian
	be.PutUint16(b[16:], 2)  // executable
	be.PutUint16(b[18:], 20) // EM_PPC
	be.PutUint32(b[20:], 1)
	be.PutUint32(b[24:], base+offset)
	be.PutUint32(b[28:], 52)
	be.PutUint16(b[40:], 52)
	be.PutUint16(b[42:], 32)
	be.PutUint16(b[44:], 1)
	ph := b[52:]
	be.PutUint32(ph, 1) // PT_LOAD
	be.PutUint32(ph[8:], base)
	be.PutUint32(ph[12:], base)
	be.PutUint32(ph[16:], uint32(len(b)))
	be.PutUint32(ph[20:], uint32(len(b)))
	be.PutUint32(ph[24:], 5) // PF_R | PF_X
	be.PutUint32(ph[28:], 0x1000)
	for i, w := range words {
		be.PutUint32(b[offset+4*i:], w)
	}
	if err := os.WriteFile(path, b, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestAssembleSymbolAddressAndCall(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	target := ctxt.Lookup("target")
	callee := ctxt.Lookup("callee")
	text := &obj.Prog{As: obj.ATEXT}
	branch := &obj.Prog{As: AB, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	addr := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_EXTERN, Sym: target, Offset: 12}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: callee, Offset: 4}}
	ret := &obj.Prog{As: ABLR}
	text.Link, branch.Link, addr.Link, call.Link = branch, addr, call, ret
	branch.To.SetTarget(call)
	s := ctxt.Lookup("fixture")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	want := []uint32{0x4800000c, 0x3c600000, 0x38630000, 0x48000001, 0x4e800020}
	if s.Size != int64(4*len(want)) {
		t.Fatalf("text size %d, want %d", s.Size, 4*len(want))
	}
	for i, w := range want {
		if got := binary.BigEndian.Uint32(s.P[4*i:]); got != w {
			t.Errorf("word %d: got %08x, clang %08x", i, got, w)
		}
	}
	wantRelocs := []obj.Reloc{
		{Off: 4, Siz: 8, Type: objabi.R_ADDRPOWER, Sym: target, Add: 12},
		{Off: 12, Siz: 4, Type: objabi.R_CALLPOWER, Sym: callee, Add: 4},
	}
	if !reflect.DeepEqual(s.R, wantRelocs) {
		t.Errorf("relocations = %+v, want %+v", s.R, wantRelocs)
	}
}

// clang: cmpw 3,4; cmplw 3,4; cmpwi 3,-1; cmplwi 3,65535.
func TestAssembleSymbolJMP(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	target := ctxt.Lookup("target")
	text := &obj.Prog{As: obj.ATEXT}
	jump := &obj.Prog{As: obj.AJMP, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: target}}
	text.Link = jump
	s := ctxt.Lookup("fixture")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	if s.Size != 4 || binary.BigEndian.Uint32(s.P) != 0x48000000 {
		t.Fatalf("JMP encoding = % x, want 48 00 00 00 (b target)", s.P)
	}
	want := []obj.Reloc{{Off: 0, Siz: 4, Type: objabi.R_CALLPOWER, Sym: target}}
	if !reflect.DeepEqual(s.R, want) {
		t.Fatalf("JMP relocations = %+v, want %+v", s.R, want)
	}
}

func TestAssembleRegisterTargetJMP(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		target int16
		want   uint32
	}{
		{REG_CTR, 0x4e800420},
		{REG_LR, 0x4e800020},
	} {
		p := &obj.Prog{As: obj.AJMP, From: obj.Addr{Type: obj.TYPE_REG, Reg: tc.target}}
		words, ok := encodeProg(ctxt, p, 0)
		if !ok || len(words) != 1 || words[0] != tc.want {
			t.Errorf("%v = %08x, want %08x", p, words, tc.want)
		}
	}
}

func TestEncodePreemptSpecialRegisters(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		name string
		prog *obj.Prog
		want uint32
	}{
		{"mfcr", &obj.Prog{As: AMFCR, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}}, 0x7fe00026},
		{"mfctr", &obj.Prog{As: AMFCTR, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}}, 0x7fe902a6},
		{"mtcrf", &obj.Prog{As: AMTCRF, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 0xff}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R31}}, 0x7feff120},
	} {
		words, ok := encodeProg(ctxt, tc.prog, 0)
		if !ok || len(words) != 1 || words[0] != tc.want {
			t.Errorf("%s = %08x, want %08x", tc.name, words, tc.want)
		}
	}
}

func TestEncodeIndirectBranchForms(t *testing.T) {
	for _, tc := range []struct {
		as   obj.As
		bo   int32
		bi   int32
		want uint32
	}{
		{ABCLR, 16, 0, 0x4e000020},
		{ABCLRL, 16, 0, 0x4e000021},
		// bclr/bclrl take their target from LR, so a CTR-decrement BO is legal;
		// bcctr/bcctrl branch to the CTR and must NOT decrement it (valid bo=12
		// here; the old test pinned the invalid decrement bo=16 — R2's finding).
		{ABCCTR, 12, 0, 0x4d800420},
		{ABCCTRL, 12, 0, 0x4d800421},
	} {
		word, err := Encode(tc.as, tc.bo, tc.bi, 0)
		if err != nil || word != tc.want {
			t.Errorf("Encode(%v,%d,%d,0) = %#08x,%v want %#08x", tc.as, tc.bo, tc.bi, word, err, tc.want)
		}
	}
	// The fixed (no-operand) forms still encode their baked-in canonical word.
	for _, tc := range []struct {
		as   obj.As
		want uint32
	}{{ABLR, 0x4e800020}, {ABLRL, 0x4e800021}, {ABCTR, 0x4e800420}, {ABCTRL, 0x4e800421}} {
		word, err := Encode(tc.as)
		if err != nil || word != tc.want {
			t.Errorf("Encode(%v) = %#08x,%v want %#08x", tc.as, word, err, tc.want)
		}
	}
	// R2-D1: bcctr/bcctrl with a CTR-decrement BO (bo&4==0) is the invalid form
	// and must be refused, unlike the same BO on bclr.
	for _, bo := range []int32{16, 0, 1, 2, 3} {
		if _, err := Encode(ABCCTR, bo, 0, 0); err == nil {
			t.Errorf("bcctr bo=%d accepted; BO[2]=0 decrements the target CTR (invalid)", bo)
		}
		if _, err := Encode(ABCCTRL, bo, 0, 0); err == nil {
			t.Errorf("bcctrl bo=%d accepted; BO[2]=0 decrements the target CTR (invalid)", bo)
		}
	}
	// R2-D1: a nonzero branch hint is not valid on this floor (bits 16-20
	// reserved) and must be refused on the indirect forms.
	if _, err := Encode(ABCLRL, 20, 0, 1); err == nil {
		t.Errorf("bclrl with nonzero hint (bh=1) accepted; must refuse on the floor")
	}
	if _, err := Encode(ABCCTR, 12, 0, 1); err == nil {
		t.Errorf("bcctr with nonzero hint (bh=1) accepted; must refuse on the floor")
	}
	// Positive: the reserved-BO check that already existed still refuses bclr 23.
	if _, err := Encode(ABCLR, 23, 0, 0); err == nil {
		t.Errorf("bclr 23 accepted; BO=23 reserved")
	}
}

func TestEncodeCompareProgs(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		as   obj.As
		to   obj.Addr
		word uint32
	}{
		{ACMPW, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R4}, 0x7c032000},
		{ACMPLW, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R4}, 0x7c032040},
		{ACMPWI, obj.Addr{Type: obj.TYPE_CONST, Offset: -1}, 0x2c03ffff},
		{ACMPLWI, obj.Addr{Type: obj.TYPE_CONST, Offset: 65535}, 0x2803ffff},
	} {
		p := &obj.Prog{As: tc.as, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}, To: tc.to}
		words, ok := encodeProg(ctxt, p, 0)
		if !ok || len(words) != 1 || words[0] != tc.word {
			t.Errorf("%v = %08x, want clang %08x", p, words, tc.word)
		}
	}
}

// TestRelaxConditionalBranchRejectsUninvertibleBO proves goppc-xprg.16: the
// long-branch relaxation must only sense-invert BO=4/011zy-adjacent forms and
// must reject a hand-written out-of-range BC whose BO it cannot correctly
// invert (BO=20 "always" -> 28 skips the inserted long branch, never reaching
// the target; a CTR-decrement form must not be re-sensed). The compiler/runtime
// emit only BO=4/12, so this path is reachable only from assembly.
func TestRelaxConditionalBranchRejectsUninvertibleBO(t *testing.T) {
	// build a long-branch scenario: cmp; BC $bo -> target; 9000 nops; target.
	runAssemble := func(bo int32) (diagnosed, panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		ctxt := obj.Linknew(&Linkppc)
		ctxt.DiagFunc = func(string, ...interface{}) { diagnosed = true }
		text := &obj.Prog{As: obj.ATEXT}
		cmp := &obj.Prog{As: ACMPWI, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}, To: obj.Addr{Type: obj.TYPE_CONST}}
		branch := &obj.Prog{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: int64(bo)}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}}
		text.Link, cmp.Link = cmp, branch
		prev := branch
		for i := 0; i < 9000; i++ {
			p := &obj.Prog{As: AORI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R0}}
			prev.Link = p
			prev = p
		}
		target := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
		prev.Link = target
		branch.To.SetTarget(target)
		s := ctxt.Lookup("relax-bo")
		s.NewFuncInfo().Text = text
		assemble(ctxt, s, nil)
		return
	}

	// Must be rejected: BO=20 (1z1zz always) and BO=16 (1z00y CTR!=0).
	for _, bo := range []int32{20, 16} {
		diagnosed, panicked := runAssemble(bo)
		if !diagnosed && !panicked {
			t.Errorf("BC $%d out-of-range was silently mis-relaxed; expected a diagnostic", bo)
		}
	}
	// Must succeed (no diagnosis): the CR-test forms 4 and 12.
	for _, bo := range []int32{4, 12} {
		if diagnosed, _ := runAssemble(bo); diagnosed {
			t.Errorf("BC $%d is an invertible CR-test form but was diagnosed", bo)
		}
	}
}

func TestRejectUnsupportedSymbolAddress(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	diagnosed := false
	ctxt.DiagFunc = func(string, ...interface{}) { diagnosed = true }
	p := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_GOTREF, Sym: ctxt.Lookup("target")}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	if _, ok := encodeProg(ctxt, p, 0); ok || !diagnosed {
		t.Fatal("GOT address silently accepted without a supported relocation path")
	}
}
