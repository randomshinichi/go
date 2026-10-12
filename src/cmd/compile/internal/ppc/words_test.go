// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

// This test closes the loop from Go source to instruction words for the 64-bit
// operations the ppc32 back end lowers, and it does the comparison against an
// assembler that has nothing to do with this toolchain.
//
// It compiles a fixture with the compiler built from this tree, reads the
// instruction words out of the compiler's own -S dump, assembles the same
// instruction sequence with clang's PowerPC 32-bit big-endian assembler, and
// compares. A wrong operand order, a wrong opcode, a dropped carry or a wrong
// displacement all show up as a word mismatch.
//
// Two things are deliberately spelled out rather than derived:
//
//   - the sequences below are the recipes DESIGN.md §5 gives for the 64-bit
//     operations, transcribed by hand from the -S output and read against the
//     architecture: addc/adde for add, subfc/subfe for sub, mullw/mulhwu plus
//     the two cross products for mul. Writing them out is what makes the test
//     able to catch a change in the lowering; deriving them from the same
//     compiler under test would not.
//   - the displacements are derived from the measured 24-byte linkage area
//     (docs/darwin-ppc32-abi-measured.md), plus declaration-order stack slots.
//     Frameless leaf arguments begin at 24; two uint64 args put the result at
//     40. Non-leaves additionally allocate a 16-byte-aligned total frame.

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"internal/testenv"
)

// wordRecipes pairs a Go function with the instruction sequence it must lower to,
// in clang's ppc32 syntax.
var wordRecipes = []struct {
	fn    string
	clang string
}{
	// The minimal global-load repro: a signed immediate compare and a CR0.LT
	// branch. Symbol immediates remain zero until a linker applies the reloc.
	{"main.branch", `	lis 3,main.G@ha
	addi 3,3,main.G@l
	lwz 3,0(3)
	cmpwi 3,0
	bc 4,0,.+20
	li 3,1
	addi 4,1,24
	stw 3,0(4)
	blr
	li 3,2
	addi 4,1,24
	stw 3,0(4)
	blr
`},
	// Compare the same G/H bit patterns with signed and unsigned semantics.
	// Canonicalization swaps the input order; CR0.GT now means G < H.
	{"main.signed", `	lis 3,main.H@ha
	addi 3,3,main.H@l
	lwz 3,0(3)
	lis 4,main.G@ha
	addi 4,4,main.G@l
	lwz 4,0(4)
	cmpw 3,4
	bc 4,1,.+20
	li 3,1
	addi 4,1,24
	stw 3,0(4)
	blr
	li 3,2
	addi 4,1,24
	stw 3,0(4)
	blr
`},
	{"main.unsigned", `	lis 3,main.G@ha
	addi 3,3,main.G@l
	lwz 3,0(3)
	lis 4,main.H@ha
	addi 4,4,main.H@l
	lwz 4,0(4)
	cmplw 4,3
	bc 4,1,.+20
	li 3,1
	addi 4,1,24
	stw 3,0(4)
	blr
	li 3,2
	addi 4,1,24
	stw 3,0(4)
	blr
`},
	// A zero-argument non-leaf still needs linkage space: align(24,16)=32.
	{"main.call", `	mflr 0
	stw 0,8(1)
	li 0,0
	stwu 1,-32(1)
	bl main.callee
	lwz 31,40(1)
	mtlr 31
	addi 1,1,32
	blr
`},
	// 64-bit add: addc on the low words leaves the carry in XER[CA], adde on the
	// high words adds it back. The high word lands at the lower address.
	{"main.add", `	addi 3,1,40
	addi 4,3,4
	lwz 5,28(1)
	lwz 6,36(1)
	addc 5,5,6
	lwz 6,24(1)
	lwz 7,32(1)
	adde 6,6,7
	stw 6,0(3)
	stw 5,0(4)
	blr
`},
	// 64-bit sub: subfc leaves the inverted borrow, subfe consumes it, which is
	// why subf's operands are the other way round from add's.
	{"main.sub", `	addi 3,1,40
	addi 4,3,4
	lwz 5,28(1)
	lwz 6,36(1)
	subfc 5,6,5
	lwz 6,24(1)
	lwz 7,32(1)
	subfe 6,7,6
	stw 6,0(3)
	stw 5,0(4)
	blr
`},
	// 64-bit mul: the low word is mullw, and the high word is
	// hi(alo*blo) + alo*bhi + ahi*blo. The two cross products and the two adds
	// are the terms a hand-written split multiply forgets.
	{"main.mul", `	addi 3,1,40
	addi 4,3,4
	lwz 5,28(1)
	lwz 6,36(1)
	mullw 8,5,6
	mulhwu 7,5,6
	lwz 9,32(1)
	mullw 5,9,5
	lwz 9,24(1)
	mullw 6,6,9
	add 5,5,6
	add 5,7,5
	stw 5,0(3)
	stw 8,0(4)
	blr
`},
	// 32-bit subtract: the plain unsigned form, which is one subf with the
	// operands the other way round from how the source reads.
	{"main.sub32", `	lwz 3,24(1)
	lwz 4,28(1)
	subf 3,4,3
	addi 4,1,32
	stw 3,0(4)
	blr
`},
	// 64-bit or: the halves are independent, so this one is covered to show that
	// the split itself is not doing anything clever.
	{"main.or", `	addi 3,1,40
	addi 4,3,4
	lwz 5,32(1)
	lwz 6,24(1)
	or 5,5,6
	stw 5,0(3)
	lwz 3,36(1)
	lwz 5,28(1)
	or 3,3,5
	stw 3,0(4)
	blr
`},
}

const wordFixture = `package main

var G int
var H int
func branch() int { if G < 0 { return 1 }; return 2 }
func signed() int { if G < H { return 1 }; return 2 }
func unsigned() int { if uint32(G) < uint32(H) { return 1 }; return 2 }
func callee()
//go:nosplit
func call() { callee() }

func add(a, b uint64) uint64 { return a + b }
func sub(a, b uint64) uint64 { return a - b }
func mul(a, b uint64) uint64 { return a * b }
func or(a, b uint64) uint64  { return a | b }
func sub32(a, b uint32) uint32 { return a - b }
`

var dumpLine = regexp.MustCompile(`^\t0x[0-9a-f]+ ((?:[0-9a-f]{2} )+)\s`)
var progLine = regexp.MustCompile(`^\t0x[0-9a-f]+ [0-9]+ \(`)

// TestEmittedWordsMatchClang assembles each recipe with clang and compares the
// words with what the compiler emitted for the matching Go function.
func TestEmittedWordsMatchClang(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang not found: the independent assembler for this comparison is unavailable")
	}
	goTool := testenv.GoToolPath(t)

	dir := t.TempDir()
	src := filepath.Join(dir, "fixture.go")
	if err := os.WriteFile(src, []byte(wordFixture), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goTool, "tool", "compile", "-p", "main", "-S", src)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile -S of the ppc fixture failed: %v\n%s", err, out)
	}

	for _, r := range wordRecipes {
		t.Run(r.fn, func(t *testing.T) {
			got := emittedWords(t, string(out), r.fn)
			ofile := clangAssemble(t, clang, dir, r.clang)
			want := elf32beTextWords(t, ofile)
			gotRelocs := emittedRelocs(string(out), r.fn)
			wantRelocs := clangRelocs(t, ofile)
			if !reflect.DeepEqual(gotRelocs, wantRelocs) {
				t.Errorf("%s relocations: emitted %v, clang %v", r.fn, gotRelocs, wantRelocs)
			}
			if len(got) != len(want) {
				t.Fatalf("%s emitted %d words, clang assembled %d", r.fn, len(got), len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("%s word %d: emitted %08x, clang %08x", r.fn, i, got[i], want[i])
				}
			}
		})
	}
}

// emittedWords pulls the instruction words of one function out of the compiler's
// -S output. A function's section is a header line, the Progs, and then a hex
// dump of the assembled text symbol.
func emittedWords(t *testing.T, out, fn string) []uint32 {
	t.Helper()
	lines := strings.Split(out, "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, fn+" STEXT") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatalf("%s has no STEXT section in the compiler output", fn)
	}
	var b []byte
	for _, l := range lines[start:] {
		if !strings.HasPrefix(l, "\t") {
			break // next symbol
		}
		if progLine.MatchString(l) {
			continue
		}
		m := dumpLine.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		for _, f := range strings.Fields(m[1]) {
			v, err := strconv.ParseUint(f, 16, 8)
			if err != nil {
				t.Fatalf("bad byte %q in the -S dump: %v", f, err)
			}
			b = append(b, byte(v))
		}
	}
	if len(b) == 0 || len(b)%4 != 0 {
		t.Fatalf("%s: extracted %d bytes from the -S dump, want a non-empty multiple of 4", fn, len(b))
	}
	words := make([]uint32, len(b)/4)
	for i := range words {
		words[i] = binary.BigEndian.Uint32(b[4*i:])
	}
	return words
}

// clangTextWords assembles src with clang's ppc32 assembler and returns the words
// of the .text section.
func clangTextWords(t *testing.T, clang, dir, src string) []uint32 {
	t.Helper()
	return elf32beTextWords(t, clangAssemble(t, clang, dir, src))
}

func clangAssemble(t *testing.T, clang, dir, src string) string {
	t.Helper()
	sfile := filepath.Join(dir, "recipe.s")
	if err := os.WriteFile(sfile, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	ofile := filepath.Join(dir, "recipe.o")
	out, err := exec.Command(clang, "--target=powerpc-unknown-linux-gnu", "-c", "-o", ofile, sfile).CombinedOutput()
	if err != nil {
		t.Fatalf("clang failed to assemble the recipe: %v\n%s", err, out)
	}
	return ofile
}

// elf32beTextWords reads the .text section of a 32-bit big-endian ELF object.
func elf32beTextWords(t *testing.T, path string) []uint32 {
	t.Helper()
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(blob) < 0x34 || blob[0] != 0x7f || blob[1] != 'E' || blob[2] != 'L' || blob[3] != 'F' || blob[4] != 1 || blob[5] != 2 {
		t.Fatalf("%s is not a 32-bit big-endian ELF object", path)
	}
	be := binary.BigEndian
	shoff := int(be.Uint32(blob[0x20:]))
	shentsize := int(be.Uint16(blob[0x2e:]))
	shnum := int(be.Uint16(blob[0x30:]))
	shstrndx := int(be.Uint16(blob[0x32:]))
	sh := func(i int) (name, off, size uint32) {
		base := shoff + i*shentsize
		return be.Uint32(blob[base:]), be.Uint32(blob[base+16:]), be.Uint32(blob[base+20:])
	}
	_, strOff, _ := sh(shstrndx)
	for i := 0; i < shnum; i++ {
		nameOff, off, size := sh(i)
		b := blob[int(strOff)+int(nameOff):]
		end := 0
		for end < len(b) && b[end] != 0 {
			end++
		}
		if string(b[:end]) != ".text" {
			continue
		}
		if size == 0 || size%4 != 0 {
			t.Fatalf("%s: .text size %d is not a non-empty multiple of 4", path, size)
		}
		data := blob[off : off+size]
		words := make([]uint32, size/4)
		for j := range words {
			words[j] = be.Uint32(data[4*j:])
		}
		return words
	}
	t.Fatalf("%s has no .text section", path)
	return nil
}

// TestWordFixtureIsWellFormed keeps a silent typo in a recipe from turning into a
// confusing failure: every recipe must assemble, and its word count must match
// the number of instructions written.
func TestWordFixtureIsWellFormed(t *testing.T) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang not found")
	}
	dir := t.TempDir()
	for _, r := range wordRecipes {
		n := len(strings.Split(strings.TrimSpace(r.clang), "\n"))
		words := clangTextWords(t, clang, dir, r.clang)
		if len(words) != n {
			t.Errorf("%s: recipe has %d instructions but clang emitted %d words", r.fn, n, len(words))
		}
	}
}
