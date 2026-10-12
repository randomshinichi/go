// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"fmt"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCallerPCQEMU(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	qemu, err := exec.LookPath("qemu-ppc")
	if err != nil {
		t.Skip("INCONCLUSIVE: qemu-ppc unavailable")
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("INCONCLUSIVE: clang unavailable")
	}
	dir := t.TempDir()
	if d := os.Getenv("PPC32_PC_EVIDENCE"); d != "" {
		dir = d
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(dir, "pc.go")
	fixture := `package sys
func GetCallerPC() uintptr
//go:noinline
//go:nosplit
func other() {}
//go:nosplit
func leafPC() uintptr { return GetCallerPC() }
//go:nosplit
func framePC() uintptr { other(); return GetCallerPC() }
`
	writePPCFixture(t, src, []byte(fixture), 0644)
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "internal/runtime/sys", "-S", "-o", filepath.Join(dir, "pc.o"), src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "pc.dump"), out, 0644)
	if err != nil {
		t.Fatalf("PC compile: %v\n%s", err, out)
	}
	dump := string(out)
	for _, tc := range []struct {
		fn     string
		mutant bool
	}{{"leafPC", false}, {"framePC", false}, {"framePC", true}} {
		name := tc.fn
		if tc.mutant {
			name += "-wrong-saved-slot"
		}
		t.Run(name, func(t *testing.T) {
			var a strings.Builder
			a.WriteString("clrrwi 1,1,4\naddi 1,1,-4096\nmr 2,1\nli 0,0\nstw 0,4(1)\nlis 3,0x1357\nori 3,3,0x9bdf\nstw 3,-1024(1)\nstw 3,2044(1)\nbl entry\nexpected_pc:\ncmpw 1,2\nbne harness_bad\nlwz 3,24(1)\nbl harness_pc\nharness_pc:\nmflr 4\naddi 4,4,expected_pc-harness_pc\ncmpw 3,4\nbne harness_result_bad\n")
			for _, off := range []int{-1024, 2044} {
				fmt.Fprintf(&a, "lwz 3,%d(1)\n", off)
				a.WriteString(checkReg(3, 0x13579bdf))
			}
			a.WriteString("li 3,0\nli 0,1\nsc\nharness_result_bad:\nli 3,1\nli 0,1\nsc\nharness_bad:\nli 3,2\nli 0,1\nsc\nentry:\n")
			fn := "internal/runtime/sys." + tc.fn
			words := emittedWords(t, dump, fn)
			relocs := fixtureFunctionRelocs(t, dump, fn)
			changed := 0
			for i, w := range words {
				if rel, ok := relocs[i*4]; ok {
					if rel.target != "internal/runtime/sys.other" || w != 0x48000001 {
						t.Fatalf("unexpected PC relocation %v", rel)
					}
					a.WriteString("bl other\n")
				} else {
					// Mutate GetCallerPC's result load, not the epilogue's
					// now-identical saved-LR offset into reserved R31.
					if tc.mutant && w&0xfc1fffff == 0x80010028 && (w>>21)&31 != 31 {
						w -= 4
						changed++
					} // allocated32 + caller LR8 -> reserved CR4
					fmt.Fprintf(&a, ".long 0x%08x\n", w)
				}
			}
			a.WriteString("other:\n")
			for _, w := range emittedWords(t, dump, "internal/runtime/sys.other") {
				fmt.Fprintf(&a, ".long 0x%08x\n", w)
			}
			if tc.mutant && changed != 1 {
				t.Fatalf("PC mutant changed %d loads", changed)
			}
			want := 0
			if tc.mutant {
				want = 1
			}
			executeIntELF(t, qemu, clang, dir, name, a.String(), want, changed)
		})
	}
	censusPPCObjects(t, dir, "pc.o")
}
