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
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"internal/testenv"

	"cmd/compile/internal/ssagen"
)

func TestPadGoStackFrame(t *testing.T) {
	var arch ssagen.ArchInfo
	Init(&arch)
	for _, tc := range []struct{ raw, padded int64 }{{0, 0}, {4, 8}, {8, 8}, {12, 24}, {40, 40}, {4096, 4104}} {
		got := arch.PadFrame(tc.raw)
		if got != tc.padded || (got != 0 && (got+arch.LinkArch.Arch.FixedFrameSize)%16 != 0) {
			t.Errorf("PadFrame(%d) = %d, want %d and aligned total", tc.raw, got, tc.padded)
		}
	}
}

const frameFixture = `package main
//go:noinline
//go:nosplit
func leaf(x int) int { return x*3+1 }
//go:noinline
//go:nosplit
func middle(x int) int { return leaf(x)+2 }
//go:noinline
//go:nosplit
func outer(x int) int { z := middle(x+1); return middle(z)+x }
//go:noinline
//go:nosplit
func sum12(a,b,c,d,e,f,g,h,i,j,k,l int) int { return a+b+c+d+e+f+g+h+i+j+k+l }
//go:noinline
//go:nosplit
func many() int { return sum12(1,2,3,4,5,6,7,8,9,10,11,12) }
//go:noinline
//go:nosplit
func wideleaf(x uint64) uint64 { return x+0x100000001 }
//go:noinline
//go:nosplit
func wide(x uint64) uint64 { return wideleaf(x)+x }
//go:noinline
//go:nosplit
func spilled(x int) int { y:=middle(x); z:=middle(x+1); return y+z }
//go:noinline
//go:nosplit
func branchCall(x int) int { if x<0 { return middle(x) }; return middle(x)+1 }
`

// TestGoStackCallsQEMU executes entire compiled functions, not selected body
// instructions. The only substitutions are R_CALLPOWER words: clang resolves
// these fixture-local BL labels. This is a fragment adapter, not a port linker.
// Linux exit and initial stack setup are harness code, not Darwin evidence.
func TestGoStackCallsQEMU(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	qemu, err := exec.LookPath("qemu-ppc")
	if err != nil {
		t.Skip("qemu-ppc unavailable: frame execution inconclusive")
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang unavailable: fragment assembly inconclusive")
	}
	dir := t.TempDir()
	// Optional machine-output retention for the port's lane runner.
	if evidence := os.Getenv("PPC32_FRAME_EVIDENCE"); evidence != "" {
		dir = evidence
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(dir, "frames.go")
	if err := os.WriteFile(src, []byte(frameFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-S", "-o", filepath.Join(dir, "frames.o"), src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err := os.WriteFile(filepath.Join(dir, "frames.dump"), out, 0o644); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}

	// The first call result cannot be recomputed or left in the outgoing
	// result slot (the second call overwrites it). Pin a real local spill and
	// reload at SP+44: raw outgoing8+local4 pads to24, total48, local y=-4.
	var store, reload bool
	for _, w := range emittedWords(t, string(out), "main.spilled") {
		store = store || w&0xfc1fffff == 0x9001002c
		reload = reload || w&0xfc1fffff == 0x8001002c
	}
	if !store || !reload {
		t.Fatal("live-across-call fixture no longer exercises a local spill/reload")
	}

	for _, tc := range []struct {
		name, entry, setup, check string
		want                      int
		mutant                    bool
	}{
		{"nested", "outer", "li 3,5\nstw 3,24(1)\n", "lwz 3,28(1)\n", 71, false},
		{"twelve-stack-args", "many", "", "lwz 3,24(1)\n", 78, false},
		{"live-gpr-spill", "spilled", "li 3,5\nstw 3,24(1)\n", "lwz 3,28(1)\n", 39, false},
		{"branch-negative", "branchCall", "li 3,-5\nstw 3,24(1)\n", "lwz 3,28(1)\n", 244, false},
		{"branch-positive", "branchCall", "li 3,5\nstw 3,24(1)\n", "lwz 3,28(1)\n", 19, false},
		// x=0x00000001_ffffffff, result=2*x+0x100000001=0x4_ffffffff.
		{"split-uint64", "wide", "li 3,1\nstw 3,24(1)\nli 3,-1\nstw 3,28(1)\n", "lwz 3,32(1)\ncmpwi 3,4\nbne harness_bad\nlwz 3,36(1)\n", 255, false},
		{"wrong-lr-slot", "outer", "li 3,5\nstw 3,24(1)\n", "lwz 3,28(1)\n", 71, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asm strings.Builder
			// Initial Linux SP is not a Go frame. Reserve a private, aligned
			// caller area and keep its address in g (reserved from allocation).
			fmt.Fprintf(&asm, "clrrwi 1,1,4\naddi 1,1,-256\nmr 30,1\nli 0,0\nstw 0,4(1)\n%sbl fn_%s\ncmpw 1,30\nbne harness_bad\n%sli 0,1\nsc\nharness_bad:\nli 3,250\nli 0,1\nsc\n", tc.setup, tc.entry, tc.check)
			mutated := 0
			for _, fn := range []string{"leaf", "middle", "outer", "sum12", "many", "wideleaf", "wide", "spilled", "branchCall"} {
				words := emittedWords(t, string(out), "main."+fn)
				fmt.Fprintf(&asm, "fn_%s:\n", fn)
				// Check SP alignment at every actual Go call boundary, before
				// the callee's entire (unmodified) emitted function.
				asm.WriteString("andi. 31,1,15\nbne harness_bad\n")
				calls := fixtureLocalCalls(t, string(out), "main."+fn)
				for i, w := range words {
					if target, ok := calls[i*4]; ok {
						if w != 0x48000001 {
							t.Fatalf("call word %08x not BL placeholder", w)
						}
						fmt.Fprintf(&asm, "bl fn_%s\n", strings.TrimPrefix(target, "main."))
					} else {
						if tc.mutant && w == 0x83e10008 {
							w = 0x83e10004 // restore LR from reserved CR slot, not LR
							mutated++
						}
						fmt.Fprintf(&asm, ".long 0x%08x\n", w)
					}
				}
			}
			if tc.mutant && mutated == 0 {
				t.Fatal("negative control changed no LR restore")
			}
			if err := os.WriteFile(filepath.Join(dir, tc.name+".s"), []byte(asm.String()), 0o644); err != nil {
				t.Fatal(err)
			}
			text := clangTextWords(t, clang, dir, asm.String())
			path := filepath.Join(dir, tc.name+".elf")
			writePPCTextELF(t, path, text)
			log := filepath.Join(dir, tc.name+".decode")
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", "-d", "in_asm", "-D", log, path)
			cmd.Dir = dir // keep any emulator core file in lane evidence
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("INCONCLUSIVE: QEMU timeout: %v", ctx.Err())
			}
			rc := 0
			if exit, ok := err.(*exec.ExitError); ok {
				rc = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			status := fmt.Sprintf("command=%s -cpu 750 -d in_asm -D %s %s\nexit=%d\nmutated_lr_restores=%d\n%s", qemu, log, path, rc, mutated, output)
			if err := os.WriteFile(filepath.Join(dir, tc.name+".status"), []byte(status), 0o644); err != nil {
				t.Fatal(err)
			}
			if tc.mutant {
				if rc != -1 || !strings.Contains(string(output), "target signal 11") {
					t.Fatalf("wrong LR-slot control did not cause expected target SIGSEGV: exit=%d\n%s", rc, output)
				}
				t.Logf("wrong LR slot rejected: exit=%d, modified %d restores", rc, mutated)
			} else {
				if rc != tc.want {
					t.Fatalf("QEMU result %d, want %d\n%s", rc, tc.want, output)
				}
				t.Logf("complete Go functions under qemu-ppc -cpu 750 returned %d; call-boundary alignment and caller SP checked", rc)
			}
		})
	}
}

var fixtureCallReloc = regexp.MustCompile(`^(\d+)\+4 R_CALLPOWER (main\.\w+)\+0$`)

func fixtureLocalCalls(t *testing.T, out, fn string) map[int]string {
	t.Helper()
	calls := make(map[int]string)
	for _, rel := range emittedRelocs(out, fn) {
		m := fixtureCallReloc.FindStringSubmatch(rel)
		if m == nil {
			t.Fatalf("fragment has a non-local-call relocation: %s", rel)
		}
		off, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatal(err)
		}
		calls[off] = m[2]
	}
	return calls
}
