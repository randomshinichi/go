// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"internal/testenv"
)

const runtimeIntegerFixture = `package main
import "unsafe"
// Frontend constants must agree with the four-byte backend.
var _ [4]byte = [unsafe.Sizeof(uintptr(0))]byte{}
var _ [0]byte = [^uintptr(0)>>63]byte{}
//go:nosplit
func shl(x, n uint32) uint32 { return x << n }
//go:nosplit
func shr(x, n uint32) uint32 { return x >> n }
//go:nosplit
func sar(x int32, n uint32) int32 { return x >> n }
//go:nosplit
func compare(x, y uint32) bool { return x <= y }
//go:nosplit
func small(x, y uint32) uint32 { return uint32(uint8(x)+uint8(y)) }
//go:nosplit
func signedbyte(p *int8) int32 { return int32(*p) }
//go:nosplit
func constant() uint32 { return 0x89abcdef }
//go:nosplit
func explicitnil(p *[8192]byte) byte { return p[4096] }
//go:nosplit
func complement(x uint32) uint32 { return ^x }
//go:nosplit
func smallcompare(x, y uint32) bool { return int8(x) < int8(y) }
//go:noinline
func zeroAggregate(p *[17]uint32) { *p = [17]uint32{} }
`

func compileRuntimeIntegers(t *testing.T, dir string) string {
	t.Helper()
	testenv.MustHaveGoBuild(t)
	path := filepath.Join(dir, "runtime_integers.go")
	if err := os.WriteFile(path, []byte(runtimeIntegerFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-S", "-o", filepath.Join(dir, "integers.o"), path)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ppc32 integer compile: %v\n%s", err, out)
	}
	return string(out)
}

func TestRuntimeIntegerLowering(t *testing.T) {
	out := compileRuntimeIntegers(t, t.TempDir())
	for _, name := range []string{"shl", "shr", "sar", "compare", "small", "signedbyte", "constant", "explicitnil", "complement", "smallcompare", "zeroAggregate"} {
		if len(emittedWords(t, out, "main."+name)) == 0 {
			t.Errorf("no machine instructions for %s", name)
		}
	}
	// The offset load cannot stand in for the nil check (4096 is beyond the
	// protected zero page); it must first touch the original pointer.
	if !strings.Contains(out, "LBZ\t(R3), R31") && !strings.Contains(out, "LBZ\t(R4), R31") {
		t.Fatalf("explicit nil check did not emit a faulting byte load:\n%s", out)
	}
	start := strings.Index(out, "main.zeroAggregate STEXT")
	if start < 0 {
		t.Fatal("no machine instructions for zeroAggregate")
	}
	end := strings.Index(out[start+len("main.zeroAggregate STEXT"):], "\nmain.")
	if end >= 0 {
		end += start + len("main.zeroAggregate STEXT")
	} else {
		end = len(out)
	}
	zero := out[start:end]
	if !strings.Contains(zero, "ADDI\t$0, R0, R31") || !strings.Contains(zero, "STW\tR31,") {
		t.Fatalf("LoweredZero did not materialize zero in R31 before stores:\n%s", zero)
	}
	if strings.Contains(zero, "STW\tR0, (R3)") {
		t.Fatalf("LoweredZero still stores from assumed-zero R0:\n%s", zero)
	}
}

// Independent Linux ELF wrapper supplies Go stack arguments. This executes
// integer functions under a 750 emulator, not the Darwin runtime or its ABI.
func TestRuntimeIntegerLoweringQEMU(t *testing.T) {
	qemu, err := exec.LookPath("qemu-ppc")
	if err != nil {
		t.Skip("qemu-ppc unavailable")
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang unavailable")
	}
	dir := t.TempDir()
	out := compileRuntimeIntegers(t, dir)
	for _, tc := range []struct {
		name    string
		x, n    uint32
		want    uint32
		boolean bool
	}{
		{"shl", 3, 31, 0x80000000, false},
		{"shl", 3, 64, 0, false},
		{"shr", 0x80000000, 65, 0, false},
		{"sar", 0x80000000, 64, 0xffffffff, false},
		{"sar", 0x7fffffff, 255, 0, false},
		{"small", 255, 2, 1, false},
		{"compare", 1, 2, 1, true},
		{"compare", 0xffffffff, 2, 0, true},
		{"smallcompare", 255, 1, 1, true},
		{"constant", 0, 0, 0x89abcdef, false},
		{"complement", 0x12345678, 0, 0xedcba987, false},
	} {
		t.Run(fmt.Sprintf("%s/%x/%d", tc.name, tc.x, tc.n), func(t *testing.T) {
			var asm strings.Builder
			fmt.Fprintf(&asm, "addi 1,1,-128\nlis 3,%d\nori 3,3,%d\nstw 3,24(1)\nlis 4,%d\nori 4,4,%d\nstw 4,28(1)\nbl fragment\n", int16(tc.x>>16), tc.x&65535, int16(tc.n>>16), tc.n&65535)
			load, off := "lwz", 32
			if tc.boolean {
				load = "lbz"
			}
			if tc.name == "constant" {
				off = 24
			}
			if tc.name == "complement" {
				off = 28
			}
			fmt.Fprintf(&asm, "%s 3,%d(1)\nlis 4,%d\nori 4,4,%d\ncmpw 3,4\nli 3,1\nbne exit\nli 3,0\nexit:\nli 0,1\nsc\nfragment:\n", load, off, int16(tc.want>>16), tc.want&65535)
			for _, word := range emittedWords(t, out, "main."+tc.name) {
				fmt.Fprintf(&asm, ".long 0x%08x\n", word)
			}
			path := filepath.Join(dir, "integer.elf")
			writePPCTextELF(t, path, clangTextWords(t, clang, dir, asm.String()))
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", path)
			if output, err := cmd.CombinedOutput(); err != nil {
				if ctx.Err() != nil {
					t.Fatalf("INCONCLUSIVE: emulator timeout: %v", ctx.Err())
				}
				t.Fatalf("integer result mismatch: %v\n%s", err, output)
			}
		})
	}
}
