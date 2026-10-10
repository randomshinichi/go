// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// expected holds one row of testdata/expected_words.tsv: an instruction, its
// operands in the package's operand order, the expected instruction word, and
// where that word came from.
type expected struct {
	mnem       string
	args       []int32
	word       uint32
	provenance string
}

// readExpected parses testdata/expected_words.tsv. The words in that file are
// NOT this package's output: they are produced by
// out/ppc-mac/ppc32-emitter/evidence/derive_expected_words.py from clang's
// PowerPC assembler (target powerpc-unknown-linux-gnu) and from the sealed
// isa-v1 corpus, and the derivation refuses to run when the sealed corpus's
// sha256 does not match. Regenerating the file must not consult this package.
func readExpected(t *testing.T) []expected {
	t.Helper()
	data, err := os.ReadFile("testdata/expected_words.tsv")
	if err != nil {
		t.Fatal(err)
	}
	var rows []expected
	for i, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 4 {
			t.Fatalf("testdata/expected_words.tsv:%d: %d fields, want 4", i+1, len(f))
		}
		w, err := strconv.ParseUint(strings.TrimPrefix(f[2], "0x"), 16, 32)
		if err != nil {
			t.Fatalf("testdata/expected_words.tsv:%d: %v", i+1, err)
		}
		var args []int32
		if f[1] != "" {
			for _, a := range strings.Split(f[1], ",") {
				v, err := strconv.ParseInt(a, 10, 32)
				if err != nil {
					t.Fatalf("testdata/expected_words.tsv:%d: %v", i+1, err)
				}
				args = append(args, int32(v))
			}
		}
		rows = append(rows, expected{f[0], args, uint32(w), f[3]})
	}
	if len(rows) == 0 {
		t.Fatal("testdata/expected_words.tsv has no cases")
	}
	return rows
}

// TestExpectedWords is the 3.2-B word-equality gate: every baseline form must
// encode to the word an independent assembler produces for the same instruction.
func TestExpectedWords(t *testing.T) {
	rows := readExpected(t)
	provenance := map[string]int{}
	var emitted []uint32
	for _, r := range rows {
		op, err := Lookup(r.mnem)
		if err != nil {
			t.Errorf("Lookup(%q): %v", r.mnem, err)
			continue
		}
		got, err := Encode(op, r.args...)
		if err != nil {
			t.Errorf("Encode(%s %v): %v", r.mnem, r.args, err)
			continue
		}
		emitted = append(emitted, got)
		if got != r.word {
			t.Errorf("%s %v: got 0x%08X, independent expectation 0x%08X (%s)",
				r.mnem, r.args, got, r.word, r.provenance)
		}
		provenance[r.provenance]++
	}
	t.Logf("%d cases: %v", len(rows), provenance)
	if path := os.Getenv("PPC_EMIT_WORDS"); path != "" {
		// Evidence hook for the frozen isa_check.py leg of DESIGN §13 brief
		// 3.2-B: dump what this encoder produced, in fixture order.
		var b strings.Builder
		for _, w := range emitted {
			fmt.Fprintf(&b, "%08x\n", w)
		}
		if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d emitted words to %s", len(emitted), path)
	}
}

// TestCorpusWordsReachTheEncoder pins the two forms the sealed isa-v1 corpus
// contains that the baseline also encodes, so that the corpus cross-check in the
// expected-words file is exercised through the whole table rather than by name.
func TestCorpusWordsReachTheEncoder(t *testing.T) {
	// tiger-ppc32-750-addi = addi r3,r0,1; tiger-ppc32-750-blr = blr.
	if got, err := Encode(AADDI, 3, 0, 1); err != nil || got != 0x38600001 {
		t.Errorf("addi 3,0,1 = 0x%08X, %v; the sealed corpus word is 0x38600001", got, err)
	}
	if got, err := Encode(ABLR); err != nil || got != 0x4e800020 {
		t.Errorf("blr = 0x%08X, %v; the sealed corpus word is 0x4e800020", got, err)
	}
}

// TestForbiddenFormsRefused is the other half of the 3.2-B gate: the forms the
// G3 cannot execute, and the forms design decision D-1 forbids outright, must be
// refused by name — not silently absent, and never emitted.
func TestForbiddenFormsRefused(t *testing.T) {
	forbidden := []string{
		// G3-illegal, named by DESIGN §13 brief 3.2-B and by isa-v1.
		"isel", "fctidz", "vaddubm", "vxor", "vor", "xxlor", "friz", "eciwx",
		// 64-bit-only (D-1).
		"ld", "std", "ldarx", "stdcx.", "mulld", "divd", "sld", "rldicl", "cntlzd",
		"fctid", "fcfid",
		// Not G3-legal.
		"popcntb", "lq", "stq",
		// Legality unestablished on 750 (Q-8/Q-9).
		"lwsync", "stfiwx", "fsqrt", "fres",
		// Privileged: the baseline is user-mode code.
		"mfmsr", "mtmsr",
	}
	for _, m := range forbidden {
		if _, err := Lookup(m); err == nil {
			t.Errorf("Lookup(%q) succeeded; the G3 scalar baseline must refuse it", m)
			continue
		}
		if _, ok := Refused(m); !ok {
			t.Errorf("Refused(%q) = false; a refusal must carry a recorded reason", m)
		}
	}
	// Control: the same check must not refuse a form that IS in the baseline, or
	// the list above would pass for the wrong reason.
	for _, m := range []string{"addi", "blr", "lwarx", "stwcx.", "fctiwz", "sync"} {
		if _, err := Lookup(m); err != nil {
			t.Errorf("Lookup(%q) = %v; this form is in the baseline", m, err)
		}
		if _, ok := Refused(m); ok {
			t.Errorf("Refused(%q) = true; this form is in the baseline", m)
		}
	}
}

// TestXERForms pins the two named SPR forms for XER, word by word.
//
// They are not decoration: XER[CA] is where this target keeps the carry of the
// multi-word add/sub chains the compiler threads through no register (PPCOps.go:
// gp2flags1carry has an empty output mask), so runtime.asyncPreempt must save
// and restore it (genPPC32 in src/runtime/mkpreempt.go). The expected words were
// produced by two independent assemblers, neither of them this package:
//
//   - clang 22.1.8, -target powerpc-unknown-linux-gnu, on `mfxer 31` /
//     `mtxer 31` / `mfxer 3` / `mtxer 3` (readelf -x .text),
//   - the ppc64 encoder in this tree, on `MOVD XER, R31` = 0x7FE102A6 and
//     `MOVD R31, XER` = 0x7FE103A6 (go tool asm -S for linux/ppc64), which is
//     the same mfspr/mtspr-with-SPR-1 pair.
//
// The register field is bits 6-10 and the SPR sits in the base: 0x7C0102A6 is
// mfspr with SPR 1, exactly as 0x7C0802A6 is mflr (SPR 8) and 0x7C0902A6 is
// mfctr (SPR 9) in the table above.
func TestXERForms(t *testing.T) {
	cases := []struct {
		mnem string
		arg  int32
		word uint32
	}{
		{"mfxer", 0, 0x7C0102A6},
		{"mfxer", 3, 0x7C6102A6},
		{"mfxer", 31, 0x7FE102A6},
		{"mtxer", 0, 0x7C0103A6},
		{"mtxer", 3, 0x7C6103A6},
		{"mtxer", 31, 0x7FE103A6},
	}
	for _, c := range cases {
		op, err := Lookup(c.mnem)
		if err != nil {
			t.Errorf("Lookup(%q): %v", c.mnem, err)
			continue
		}
		got, err := Encode(op, c.arg)
		if err != nil {
			t.Errorf("Encode(%s %d): %v", c.mnem, c.arg, err)
			continue
		}
		if got != c.word {
			t.Errorf("%s %d = 0x%08X, want 0x%08X", c.mnem, c.arg, got, c.word)
		}
	}
	// Negative control: the named forms are additive. Generic SPR access, and
	// the floating-point control forms whose 750 legality is the open question
	// Q-8, must still be refused — this change must not become an escape hatch.
	for _, m := range []string{"mfspr", "mtspr", "mffs", "mtfsf", "mtfsfi", "mtfsb0"} {
		if _, err := Lookup(m); err == nil {
			t.Errorf("Lookup(%q) succeeded; generic SPR access and the Q-8 FP control forms stay refused", m)
		}
		if _, ok := Refused(m); !ok {
			t.Errorf("Refused(%q) = false; the refusal must keep its recorded reason", m)
		}
	}
	// And the register model already names XER, so the refusal above is about
	// encodings, not about the register being unknown.
	if got := Rconv(REG_XER); got != "XER" {
		t.Errorf("Rconv(REG_XER) = %q, want %q", got, "XER")
	}
}

// TestReservedBranchFormsRefused checks the encoding-level rule the corpus's
// "branch-bad" row exercises: bclr with BO=23 is a reserved encoding, so the
// encoder must reject the operand combination rather than emit the word.
func TestReservedBranchFormsRefused(t *testing.T) {
	if _, err := Encode(ABCLR, 23, 0, 0); err == nil {
		t.Error("bclr 23,0,0 encoded; BO=23 is the reserved form of the corpus's branch-bad row")
	}
	// BO values the frozen checker accepts must still encode, so the refusal
	// above is not a blanket rejection of bclr.
	for _, bo := range []int32{16, 20} {
		if _, err := Encode(ABCLR, bo, 0, 0); err != nil {
			t.Errorf("bclr %d,0,0: %v; BO=%d passes the isa_check rule", bo, err, bo)
		}
	}
}

// TestNeverEmitsZeroWord is a property check against the corpus's "zero" row
// (0x00000000, no instruction): no baseline form may encode to it, even with
// all-zero operands.
func TestNeverEmitsZeroWord(t *testing.T) {
	n := 0
	for i := range forms {
		f := &forms[i]
		args := make([]int32, len(f.fields))
		word, err := Encode(f.op, args...)
		if err != nil {
			t.Errorf("%s with zero operands: %v", f.mnem, err)
			continue
		}
		if word == 0 {
			t.Errorf("%s with zero operands encodes to 0x00000000, which is not an instruction", f.mnem)
		}
		n++
	}
	if n < 100 {
		t.Errorf("only %d forms checked; the baseline should be far larger", n)
	}
}

// TestOperandRangeChecked makes sure the encoder is not a masker: out-of-range
// registers and immediates must be errors, and the two 16-bit immediate spaces
// (signed and unsigned) must be distinguished.
func TestOperandRangeChecked(t *testing.T) {
	bad := []struct {
		mnem string
		args []int32
	}{
		{"addi", []int32{32, 4, 0}},          // register 32 does not exist
		{"addi", []int32{-1, 4, 0}},          // negative register
		{"addi", []int32{3, 4, 32768}},       // si is signed 16-bit
		{"ori", []int32{3, 4, 65536}},        // ui is unsigned 16-bit
		{"ori", []int32{3, 4, -1}},           // ui is not signed
		{"lwz", []int32{3, 32768, 4}},        // d is signed 16-bit
		{"add", []int32{3, 4}},               // too few operands
		{"add", []int32{3, 4, 5, 6}},         // too many operands
		{"srawi", []int32{3, 4, 32}},         // shift amount is 5 bits
		{"rlwinm", []int32{3, 4, 32, 0, 31}}, // shift amount is 5 bits
		{"cmpwi", []int32{8, 4, 0}},          // CR field is 3 bits
		{"mtcrf", []int32{0x100, 4}},         // CR mask is 8 bits
		{"b", []int32{6}},                    // branch displacement must be word aligned
		{"b", []int32{1 << 26}},              // beyond the 26-bit byte range
		{"bc", []int32{12, 0, 2}},            // displacement must be word aligned
	}
	for _, c := range bad {
		op, err := Lookup(c.mnem)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", c.mnem, err)
		}
		if _, err := Encode(op, c.args...); err == nil {
			t.Errorf("Encode(%s %v) succeeded; this operand set must be rejected", c.mnem, c.args)
		}
	}
}

// TestEmitBigEndian checks the byte stream, which is what the linker will
// consume: a ppc32 word goes to memory most-significant byte first.
func TestEmitBigEndian(t *testing.T) {
	buf, err := Emit(nil, ABLR)
	if err != nil {
		t.Fatal(err)
	}
	buf, err = Emit(buf, AADDI, 3, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x4e, 0x80, 0x00, 0x20, 0x38, 0x60, 0x00, 0x01}
	if string(buf) != string(want) {
		t.Errorf("Emit = % x, want % x", buf, want)
	}
}

// TestMnemonicTableIntegrity checks the invariants the rest of the package
// relies on: the opcode constants are in the same order as the table, Anames is
// indexed by op-AXXX, mnemonics are unique, and no form's fixed bits contain
// anything outside the opcode and fixed-field positions.
func TestMnemonicTableIntegrity(t *testing.T) {
	names := Anames()
	if names[0] != "XXX" {
		t.Errorf("Anames()[0] = %q, want XXX", names[0])
	}
	seen := map[string]bool{}
	for i := range forms {
		f := &forms[i]
		if f.op != AXXX+Inst(1+i) {
			t.Errorf("forms[%d] (%s) = 0x%x, want %d", i, f.mnem, uint32(f.op), int(AXXX)+1+i)
		}
		if seen[f.mnem] {
			t.Errorf("duplicate mnemonic %q", f.mnem)
		}
		seen[f.mnem] = true
		if names[int(f.op)-int(AXXX)] != strings.ToUpper(f.mnem) {
			t.Errorf("Anames()[%d] = %q, want %q", int(f.op)-int(AXXX), names[int(f.op)-int(AXXX)], strings.ToUpper(f.mnem))
		}
		if Mnemonic(f.op) != f.mnem {
			t.Errorf("Mnemonic(0x%x) = %q, want %q", uint32(f.op), Mnemonic(f.op), f.mnem)
		}
		if got, err := Lookup(strings.ToUpper(f.mnem)); err != nil || got != f.op {
			t.Errorf("Lookup(%q) = 0x%x, %v; want 0x%x", strings.ToUpper(f.mnem), uint32(got), err, uint32(f.op))
		}
		// The fixed bits may carry the primary opcode (bits 0-5), the XO field
		// (bits 21-30), an Rc bit and the operands of an alias that takes none
		// (blr is bclr 20,0; mflr is mfspr rd,8), but they must not overlap a
		// field the form actually encodes, which would silently merge two
		// operands.
		for _, fl := range f.fields {
			_, mask := fl.position()
			shift, _ := fl.position()
			if f.base&(mask<<shift) != 0 {
				t.Errorf("%s: fixed bits 0x%08X overlap the %s field (0x%08X)",
					f.mnem, f.base, fl, mask<<shift)
			}
		}
	}
}
