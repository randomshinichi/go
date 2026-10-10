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

const indirectFixture = `package runtime
//go:noinline
//go:nosplit
func invoke(f func() uint32) uint32 { return f() }
//go:noinline
//go:nosplit
func nested(f func() uint32) uint32 { a:=invoke(f); b:=invoke(f); return a+b }
//go:noinline
//go:nosplit
func captured(x uint32) uint32 {
 f:=func() uint32 { return x }
 return invoke(f)
}
//go:noinline
//go:nosplit
func nestedCaptured(x uint32) uint32 {
 inner:=func() uint32 { return x }
 outer:=func() uint32 { return nested(inner) }
 return invoke(outer)
}
type reader interface { Read() uint32 }
//go:noinline
//go:nosplit
func inter(x reader) uint32 { return x.Read() }
//go:noinline
//go:nosplit
func method(p *uint32) uint32 { return *p }
type g struct{}
func getg() *g
//go:noinline
//go:nosplit
func readG() *g { return getg() }
//go:noinline
//go:nosplit
func readGAfterCall(p *uint32) *g { method(p); return getg() }
type tailEnd struct { value uint32 }
//go:noinline
//go:nosplit
func (p *tailEnd) Read() uint32 { return p.value }
type tailFirst struct { *tailEnd }
// tailLink.Read is an assembly declaration. The fragment adapter binds it to
// the entire compiled tailFirst.Read wrapper, whose receiver has the same
// one-pointer layout. This makes a two-hop chain without editing either body.
type tailLink struct { *tailEnd }
func (p *tailLink) Read() uint32
type tailOuter struct { *tailLink }
`

// GetCallerSP is intrinsified in internal/runtime/sys, not runtime. This
// non-leaf verifies that its address accounts for the allocated frame, the
// same address tail-call expansion uses for incoming arguments/results.
const callerSPFixture = `package sys
func GetCallerSP() uintptr
func method(p *uint32) uint32
//go:noinline
//go:nosplit
func callerSP(p *uint32, expected uintptr) uint32 {
 method(p)
 if GetCallerSP() == expected { return 1 }
 return 0
}
`

// TestIndirectAndTailCallsQEMU extends the whole-function frame harness. All
// bodies are compiler output. Only local symbol relocations and the declared
// assembly-method alias are resolved by the ELF adapter, not a Darwin linker.
func TestIndirectAndTailCallsQEMU(t *testing.T) {
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
	if evidence := os.Getenv("PPC32_CALL_EVIDENCE"); evidence != "" {
		dir = evidence
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(dir, "calls.go")
	if err := os.WriteFile(src, []byte(indirectFixture), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "runtime", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "calls.o"), src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	output, err := cmd.CombinedOutput()
	if e := os.WriteFile(filepath.Join(dir, "calls.dump"), output, 0644); e != nil {
		t.Fatal(e)
	}
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	dump := string(output)
	spSrc := filepath.Join(dir, "caller_sp.go")
	if err := os.WriteFile(spSrc, []byte(callerSPFixture), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "internal/runtime/sys", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "caller_sp.o"), spSrc)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	output, err = cmd.CombinedOutput()
	if e := os.WriteFile(filepath.Join(dir, "caller_sp.dump"), output, 0644); e != nil {
		t.Fatal(e)
	}
	if err != nil {
		t.Fatalf("caller SP compile: %v\n%s", err, output)
	}
	dump += string(output)
	functions := []string{"runtime.invoke", "runtime.nested", "runtime.captured", "runtime.captured.func1", "runtime.nestedCaptured", "runtime.nestedCaptured.func1", "runtime.nestedCaptured.func2", "runtime.inter", "runtime.method", "runtime.readG", "runtime.readGAfterCall", "internal/runtime/sys.callerSP", "runtime.(*tailEnd).Read", "runtime.(*tailFirst).Read", "runtime.(*tailOuter).Read"}
	labels := make(map[string]string)
	for i, fn := range functions {
		labels[fn] = fmt.Sprintf("fn_%d", i)
	}
	labels["runtime.(*tailLink).Read"] = labels["runtime.(*tailFirst).Read"]
	labels["internal/runtime/sys.method"] = labels["runtime.method"]
	labels["runtime.stackcheck"] = "fn_stackcheck"
	labels["runtime.morestack"] = "fn_morestack"

	for _, tc := range []struct {
		name, entry, setup string
		off                int
		want               uint32
		mutant             bool
	}{
		{"captured-closure", "runtime.captured", "lis 3,0x4567\nori 3,3,0x89ab\nstw 3,24(1)\n", 28, 0x456789ab, false},
		{"nested-indirect", "runtime.nestedCaptured", "lis 3,0x4567\nori 3,3,0x89ab\nstw 3,24(1)\n", 28, 0x8acf1356, false},
		{"interface-receiver", "runtime.inter", "addi 3,1,128\nstw 3,24(1)\naddi 3,1,160\nstw 3,28(1)\nlis 4,0x89ab\nori 4,4,0xcdef\nstw 4,160(1)\n", 32, 0x89abcdef, false},
		{"getg-sentinel", "runtime.readG", "", 24, 0x76543210, false},
		{"getg-after-call", "runtime.readGAfterCall", "addi 3,1,160\nstw 3,24(1)\nli 3,1\nstw 3,160(1)\n", 28, 0x76543210, false},
		{"allocated-caller-sp", "internal/runtime/sys.callerSP", "addi 3,1,160\nstw 3,24(1)\nstw 2,28(1)\nli 3,1\nstw 3,160(1)\n", 32, 1, false},
		{"tail-chain", "runtime.(*tailOuter).Read", "addi 3,1,128\nstw 3,24(1)\naddi 3,1,144\nstw 3,128(1)\naddi 3,1,160\nstw 3,144(1)\nlis 3,0xfedc\nori 3,3,0xba98\nstw 3,160(1)\n", 28, 0xfedcba98, false},
		{"wrong-closure-context", "runtime.captured", "lis 3,0x4567\nori 3,3,0x89ab\nstw 3,24(1)\n", 28, 0x456789ab, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var asm strings.Builder
			// R2 is deliberately unavailable to SSA. R30 holds an independent g
			// sentinel, not the stack pointer. Canaries enclose all fixture frames and
			// the caller's outgoing argument/result area.
			asm.WriteString("fragment_start:\nclrrwi 1,1,4\naddi 1,1,-2048\nmr 2,1\nlis 30,0x7654\nori 30,30,0x3210\nli 0,0\nlis 3,0x1357\nori 3,3,0x9bdf\nstw 3,508(1)\nstw 3,-1024(1)\n")
			asm.WriteString(tc.setup)
			// Address of the fabricated interface method, computed from a local PC
			// label. No unresolved ELF symbol/address relocation is discarded.
			fmt.Fprintf(&asm, "bl harness_pc\nharness_pc:\nmflr 4\naddi 4,4,%s-harness_pc\nstw 4,140(1)\nbl %s\nharness_return:\nmflr 3\nbl harness_lr\nharness_lr:\nmflr 4\naddi 4,4,-8\ncmpw 3,4\nbne harness_bad\ncmpw 1,2\nbne harness_bad\nlwz 3,%d(1)\nlis 4,%d\nori 4,4,%d\ncmpw 3,4\nbne harness_result_bad\n", labels["runtime.method"], labels[tc.entry], tc.off, int16(tc.want>>16), tc.want&65535)
			asm.WriteString("lis 4,0x1357\nori 4,4,0x9bdf\nlwz 3,508(1)\ncmpw 3,4\nbne harness_bad\nlwz 3,-1024(1)\ncmpw 3,4\nbne harness_bad\nlis 4,0x7654\nori 4,4,0x3210\ncmpw 30,4\nbne harness_bad\nli 3,0\nli 0,1\nsc\nharness_result_bad:\nli 3,1\nli 0,1\nsc\nharness_bad:\nli 3,2\nli 0,1\nsc\n")
			mutated := 0
			tailBranches := 0
			for _, fn := range functions {
				fmt.Fprintf(&asm, "%s:\n", labels[fn])
				// Each tail hop must see the original caller's SP, not an accumulating
				// wrapper frame. This runs before the whole emitted wrapper/leaf body.
				if tc.name == "tail-chain" && strings.Contains(fn, ".Read") {
					asm.WriteString("cmpw 1,2\nbne harness_bad\n")
				}
				words := emittedWords(t, dump, fn)
				relocs := fixtureFunctionRelocs(t, dump, fn)
				for i := 0; i < len(words); i++ {
					w := words[i]
					if rel, ok := relocs[i*4]; ok {
						label, ok := labels[rel.target]
						if !ok {
							t.Fatalf("non-fixture target %s", rel.target)
						}
						if rel.kind == "R_ADDRPOWER" {
							r := (w >> 21) & 31
							if w != 0x3c000000|r<<21 || i+1 == len(words) || words[i+1] != 0x38000000|r<<21|r<<16 {
								t.Fatalf("invalid address pair at %s+%d", fn, i*4)
							}
							fmt.Fprintf(&asm, "lis %d,(0x11000+%s-fragment_start)@ha\naddi %d,%d,(0x11000+%s-fragment_start)@l\n", r, label, r, r, label)
							i++
							continue
						}
						switch w {
						case 0x48000001:
							fmt.Fprintf(&asm, "bl %s\n", label)
						case 0x48000000:
							fmt.Fprintf(&asm, "b %s\n", label)
							tailBranches++
						default:
							t.Fatalf("unexpected branch placeholder %08x", w)
						}
					} else {
						if tc.mutant && fn == "runtime.invoke" && (w == 0x7d8903a6 || w == 0x7d8803a6) { // mtctr R12, or mtlr R12 since goppc-8pgs
							// Context becomes the code address, not a funcval pointer. The target
							// stays valid, so this deterministically produces wrong full bits.
							asm.WriteString("mr 11,12\n")
							mutated++
						}
						fmt.Fprintf(&asm, ".long 0x%08x\n", w)
					}
				}
			}
			asm.WriteString("fn_stackcheck:\nli 3,0\nblr\nfn_morestack:\nli 3,99\nli 0,1\nsc\n")
			if tailBranches < 2 {
				t.Fatalf("fixture lost the two tail branches: %d", tailBranches)
			}
			if tc.mutant && mutated == 0 {
				t.Fatal("negative control changed no context")
			}
			if err := os.WriteFile(filepath.Join(dir, tc.name+".s"), []byte(asm.String()), 0644); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, tc.name+".elf")
			writePPCTextELF(t, path, clangTextWords(t, clang, dir, asm.String()))
			log := filepath.Join(dir, tc.name+".decode")
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", "-d", "in_asm", "-D", log, path)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("INCONCLUSIVE: QEMU timeout: %v", ctx.Err())
			}
			rc := 0
			if exit, ok := err.(*exec.ExitError); ok {
				rc = exit.ExitCode()
			} else if err != nil {
				t.Fatal(err)
			}
			status := fmt.Sprintf("command=%s -cpu 750 -d in_asm -D %s %s\nexit=%d\nmutated_contexts=%d\n%s", qemu, log, path, rc, mutated, out)
			if err := os.WriteFile(filepath.Join(dir, tc.name+".status"), []byte(status), 0644); err != nil {
				t.Fatal(err)
			}
			wantExit := 0
			if tc.mutant {
				wantExit = 1
			}
			if rc != wantExit {
				t.Fatalf("QEMU exit=%d want=%d\n%s", rc, wantExit, out)
			}
			if tc.mutant {
				t.Logf("negative control FAIL as required: full-result mismatch exit=1, wrong R11 context at %d indirect calls", mutated)
			} else {
				t.Logf("PASS: full result=%08x, SP/LR restored, g and both canaries intact; qemu-ppc -cpu 750 exit=0", tc.want)
			}
		})
	}
}
