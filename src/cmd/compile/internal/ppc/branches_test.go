// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"internal/testenv"
)

// This runs only the emitted comparison/branch tail, not a Go function or ABI.
// An independent assembly wrapper supplies input registers, a scratch result
// slot and LR. No Go prologue, Darwin loader or symbol relocation is exercised.
func TestComparisonBranchesQEMU(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	qemu, err := exec.LookPath("qemu-ppc")
	if err != nil {
		t.Skip("qemu-ppc unavailable: comparison execution inconclusive")
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang unavailable: independent wrapper assembly inconclusive")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "fixture.go")
	if err := os.WriteFile(src, []byte(wordFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-S", src)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	// -1 and 1 are the same input bits in both tests. Signed must take
	// the true edge (1); unsigned must take the false edge (2).
	for _, tc := range []struct {
		fn           string
		r3, r4, want int
	}{
		{"main.signed", 1, -1, 1},
		{"main.unsigned", -1, 1, 2},
		{"main.signed", -1, 1, 2},
		{"main.unsigned", 1, -1, 1},
	} {
		t.Run(fmt.Sprintf("%s/%d/%d", tc.fn, tc.r3, tc.r4), func(t *testing.T) {
			words := emittedWords(t, string(out), tc.fn)
			if len(words) != 16 {
				t.Fatalf("comparison fixture changed: %d words, want 16", len(words))
			}
			// Drop only the two relocated global-address/load prefixes.
			var asm strings.Builder
			fmt.Fprintf(&asm, "addi 1,1,-32\nli 3,%d\nli 4,%d\nbl fragment\nlwz 3,24(1)\nli 0,1\nsc\nfragment:\n", tc.r3, tc.r4)
			for _, w := range words[6:] {
				fmt.Fprintf(&asm, ".long 0x%08x\n", w)
			}
			text := clangTextWords(t, clang, dir, asm.String())
			path := filepath.Join(dir, "compare.elf")
			writePPCTextELF(t, path, text)
			log := filepath.Join(dir, "decode.log")
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", "-d", "in_asm", "-D", log, path)
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("INCONCLUSIVE: QEMU timeout: %v", ctx.Err())
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != tc.want {
				t.Fatalf("comparison result: %v, want exit %d\n%s", err, tc.want, out)
			}
			decoded, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			compare := "cmpw"
			if tc.fn == "main.unsigned" {
				compare = "cmplw"
			}
			if !strings.Contains(string(decoded), compare) || !strings.Contains(string(decoded), "ble") {
				t.Fatalf("QEMU did not decode the expected compare/branch:\n%s", decoded)
			}
			t.Logf("750 instruction fragment returned %d; QEMU decoded %s and ble", tc.want, compare)
		})
	}
}

// A single RX ELF32 segment needs no linker: there are no unresolved symbols
// in this wrapper. The Linux syscall exit is only an emulator harness boundary.
func writePPCTextELF(t *testing.T, path string, words []uint32) {
	t.Helper()
	const offset = 0x1000
	const base = 0x10000
	b := make([]byte, offset+4*len(words))
	copy(b, []byte{0x7f, 'E', 'L', 'F', 1, 2, 1})
	be := binary.BigEndian
	be.PutUint16(b[16:], 2)  // ET_EXEC
	be.PutUint16(b[18:], 20) // EM_PPC
	be.PutUint32(b[20:], 1)
	be.PutUint32(b[24:], base+offset)
	be.PutUint32(b[28:], 52)
	be.PutUint16(b[40:], 52)
	be.PutUint16(b[42:], 32)
	be.PutUint16(b[44:], 1)
	p := b[52:]
	be.PutUint32(p, 1) // PT_LOAD
	be.PutUint32(p[8:], base)
	be.PutUint32(p[12:], base)
	be.PutUint32(p[16:], uint32(len(b)))
	be.PutUint32(p[20:], uint32(len(b)))
	be.PutUint32(p[24:], 5) // PF_R | PF_X
	be.PutUint32(p[28:], 0x1000)
	for i, w := range words {
		be.PutUint32(b[offset+4*i:], w)
	}
	if err := os.WriteFile(path, b, 0o755); err != nil {
		t.Fatal(err)
	}
}
