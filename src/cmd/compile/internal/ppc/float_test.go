// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"context"
	"fmt"
	"internal/testenv"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const floatTransportFixture = `package main
//go:noinline
//go:nosplit
func echo64(x float64) float64 { return x }
//go:noinline
//go:nosplit
func echo32(x float32) float32 { return x }
//go:noinline
//go:nosplit
func nested64(pad uint32, x float64) float64 { return echo64(x) }
//go:noinline
//go:nosplit
func nested32(x float32) float32 { return echo32(x) }
//go:noinline
//go:nosplit
func round(x float64) float64 { return float64(float32(x)) }
//go:noinline
//go:nosplit
func copy64(p,q *float64) { *q = *p }
//go:noinline
//go:nosplit
func copy32(p,q *float32) { *q = *p }
//go:noinline
//go:nosplit
func constant64() float64 { return -123.5 }
//go:noinline
//go:nosplit
func constant32() float32 { return 123.5 }
func smash()
//go:noinline
//go:nosplit
func survive(p,q *float64) { x:=*p; smash(); *q=x }
//go:noinline
//go:nosplit
func survive32(p,q *float32) { x:=*p; smash(); *q=x }
`

// Compile whole functions with the tree's freshly installed compiler. The ELF
// adapter only resolves fixture-local calls; no function body is reconstructed.
func compileFloatFixture(t *testing.T, source string, functions []string, spillFunc string) (string, string) {
	t.Helper()
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	if root := os.Getenv("PPC32_FLOAT_EVIDENCE"); root != "" {
		dir = filepath.Join(root, t.Name())
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	src := filepath.Join(dir, "fixture.go")
	if err := os.WriteFile(src, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	args := []string{"tool", "compile", "-p", "main", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "fixture.o")}
	if spillFunc != "" {
		args = append(args, "-d=ssa/regalloc/dump="+spillFunc)
	}
	args = append(args, src)
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if e := os.WriteFile(filepath.Join(dir, "fixture.dump"), out, 0644); e != nil {
		t.Fatal(e)
	}
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	if spillFunc != "" {
		paths, _ := filepath.Glob(filepath.Join(dir, spillFunc+"*regalloc.dump"))
		if len(paths) != 1 {
			t.Fatalf("missing regalloc spill evidence: %v", paths)
		}
		s, e := os.ReadFile(paths[0])
		if e != nil {
			t.Fatal(e)
		}
		stores, loads := strings.Count(string(s), "StoreReg"), strings.Count(string(s), "LoadReg")
		if stores == 0 || loads == 0 {
			t.Fatalf("no demonstrated spills: stores=%d reloads=%d", stores, loads)
		}
		t.Logf("DEMONSTRATED %s: StoreReg=%d LoadReg=%d", spillFunc, stores, loads)
	}
	return dir, string(out)
}

func setWord(off int, bits uint32) string {
	return fmt.Sprintf("lis 3,%d\nori 3,3,%d\nstw 3,%d(1)\n", int16(bits>>16), bits&65535, off)
}
func checkWord(off int, bits uint32) string {
	return fmt.Sprintf("lwz 3,%d(1)\nlis 4,%d\nori 4,4,%d\ncmpw 3,4\nbne result_bad\n", off, int16(bits>>16), bits&65535)
}
func setDouble(off int, bits uint64) string {
	return setWord(off, uint32(bits>>32)) + setWord(off+4, uint32(bits))
}
func checkDouble(off int, bits uint64) string {
	return checkWord(off, uint32(bits>>32)) + checkWord(off+4, uint32(bits))
}

// The caller and canaries enclose every entire emitted function. R2 is not
// allocatable; smash really overwrites all 32 FPRs before returning.
func runFloatELF(t *testing.T, dir, dump string, functions []string, body string, mutate func(string, uint32) uint32) {
	t.Helper()
	qemu, e := exec.LookPath("qemu-ppc")
	if e != nil {
		t.Skip("INCONCLUSIVE: qemu-ppc unavailable")
	}
	clang, e := exec.LookPath("clang")
	if e != nil {
		t.Skip("INCONCLUSIVE: clang unavailable")
	}
	var asm strings.Builder
	asm.WriteString("fragment_start:\nclrrwi 1,1,4\naddi 1,1,-8192\nmr 2,1\nli 0,0\n")
	asm.WriteString(setWord(4092, 0x13579bdf) + setWord(-4096, 0x13579bdf))
	asm.WriteString(body)
	asm.WriteString("cmpw 1,2\nbne stack_bad\n")
	asm.WriteString(checkWord(4092, 0x13579bdf) + checkWord(-4096, 0x13579bdf))
	asm.WriteString("li 3,0\nli 0,1\nsc\nresult_bad:\nli 3,1\nli 0,1\nsc\nstack_bad:\nli 3,2\nli 0,1\nsc\nfn_smash:\n")
	for i := 0; i < 32; i++ {
		fmt.Fprintf(&asm, "fsub %d,%d,%d\n", i, i, i)
	}
	asm.WriteString("blr\n")
	changes := 0
	for _, fn := range functions {
		fmt.Fprintf(&asm, "fn_%s:\n", fn)
		words := emittedWords(t, dump, "main."+fn)
		for _, w := range words {
			if w>>26 == 50 || w>>26 == 54 { // lfd/stfd: only aligned private scratch
				if (w>>16)&31 != 1 || w&7 != 0 {
					t.Fatalf("%s has unproven double alignment: %08x", fn, w)
				}
			}
		}
		calls := fixtureLocalCalls(t, dump, "main."+fn)
		for i, w := range words {
			if target, ok := calls[i*4]; ok {
				if w != 0x48000001 {
					t.Fatalf("non-call relocation: %08x", w)
				}
				fmt.Fprintf(&asm, "bl fn_%s\n", strings.TrimPrefix(target, "main."))
			} else {
				if mutate != nil {
					v := mutate(fn, w)
					if v != w {
						changes++
					}
					w = v
				}
				fmt.Fprintf(&asm, ".long 0x%08x\n", w)
			}
		}
	}
	if mutate != nil && changes == 0 {
		t.Fatal("negative control changed no instruction")
	}
	name := strings.ReplaceAll(t.Name(), "/", "_")
	if e := os.WriteFile(filepath.Join(dir, name+".s"), []byte(asm.String()), 0644); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(dir, name+".elf")
	writePPCTextELF(t, path, clangTextWords(t, clang, dir, asm.String()))
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	log := filepath.Join(dir, name+".decode")
	cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", "-d", "in_asm", "-D", log, path)
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
	status := fmt.Sprintf("command=%s -cpu 750 -d in_asm -D %s %s\nexit=%d\nmutations=%d\n%s", qemu, log, path, rc, changes, out)
	if e := os.WriteFile(filepath.Join(dir, name+".status"), []byte(status), 0644); e != nil {
		t.Fatal(e)
	}
	want := 0
	if mutate != nil {
		want = 1
	}
	if rc != want {
		t.Fatalf("QEMU exit=%d want=%d\n%s", rc, want, out)
	}
	if mutate != nil {
		t.Logf("negative control FAIL as required: full-result mismatch exit=1 mutations=%d", changes)
	} else {
		t.Log("PASS: exact bits, SP restored, canaries intact; qemu-ppc -cpu 750 exit=0")
	}
}

func TestFloatTransportQEMU(t *testing.T) {
	fns := []string{"echo64", "echo32", "round", "copy64", "copy32", "constant64", "constant32", "survive", "survive32", "nested32", "nested64"}
	dir, dump := compileFloatFixture(t, floatTransportFixture, fns, "")
	var body strings.Builder
	// Both stack alignments; pointer accesses and call arguments/results.
	for _, b := range []uint64{0, 1 << 63, 0x3ff23456789abcde, 0x7ff0000000000000, 0xfff0000000000000, 1, 0x000fffffffffffff} {
		body.WriteString(setDouble(24, b) + "bl fn_echo64\n" + checkDouble(32, b))
		body.WriteString(setWord(24, 17) + setDouble(28, b) + "bl fn_nested64\n" + checkDouble(36, b))
		for _, off := range []int{256, 260} {
			body.WriteString(setDouble(off, b) + fmt.Sprintf("addi 3,1,%d\nstw 3,24(1)\naddi 3,1,280\nstw 3,28(1)\nbl fn_copy64\n", off) + checkDouble(280, b))
			body.WriteString(setDouble(off, b) + fmt.Sprintf("addi 3,1,%d\nstw 3,24(1)\naddi 3,1,284\nstw 3,28(1)\nbl fn_survive\n", off) + checkDouble(284, b))
		}
	}
	for _, b := range []uint32{0, 1 << 31, 0x3f923456, 0x7f800000, 0xff800000, 1, 0x007fffff} {
		body.WriteString(setWord(24, b) + "bl fn_echo32\n" + checkWord(28, b))
		body.WriteString(setWord(24, b) + "bl fn_nested32\n" + checkWord(28, b))
		body.WriteString(setWord(256, b) + "addi 3,1,256\nstw 3,24(1)\naddi 3,1,260\nstw 3,28(1)\nbl fn_survive32\n" + checkWord(260, b))
		body.WriteString(setWord(256, b) + "addi 3,1,256\nstw 3,24(1)\naddi 3,1,260\nstw 3,28(1)\nbl fn_copy32\n" + checkWord(260, b))
	}
	body.WriteString("bl fn_constant64\n" + checkDouble(24, math.Float64bits(-123.5)) + "bl fn_constant32\n" + checkWord(24, math.Float32bits(123.5)))
	x := 1.0 + math.Ldexp(1, -24) + math.Ldexp(1, -40)
	body.WriteString(setDouble(24, math.Float64bits(x)) + "bl fn_round\n" + checkDouble(32, math.Float64bits(float64(float32(x)))))
	t.Run("bits-calls-alignment-clobber-rounding", func(t *testing.T) { runFloatELF(t, dir, dump, fns, body.String(), nil) })
	t.Run("negative-store32", func(t *testing.T) {
		runFloatELF(t, dir, dump, fns, body.String(), func(fn string, w uint32) uint32 {
			if fn == "echo32" && w>>26 == 52 {
				return 0x90000000 | (w & 0x001fffff)
			} // stfs -> stw R0
			return w
		})
	})
}

func TestFloatPressureQEMU(t *testing.T) {
	for _, width := range []int{32, 64} {
		for _, conversion := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/conversion-%t", width, conversion), func(t *testing.T) {
				var src strings.Builder
				src.WriteString("package main\n//go:noinline\n//go:nosplit\nfunc pressure(")
				for i := 0; i < 40; i++ {
					fmt.Fprintf(&src, "p%d,q%d *float%d,", i, i, width)
				}
				src.WriteString(") {\n")
				for i := 0; i < 40; i++ {
					fmt.Fprintf(&src, "x%d:=*p%d\n", i, i)
				}
				if conversion {
					src.WriteString("z:=uint32(x39)\n")
				}
				for i := 0; i < 40; i++ {
					if conversion && i == 39 {
						fmt.Fprintf(&src, "*q39=float%d(z)\n", width)
					} else {
						fmt.Fprintf(&src, "*q%d=x%d\n", i, i)
					}
				}
				src.WriteString("}\n")
				dir, dump := compileFloatFixture(t, src.String(), []string{"pressure"}, "pressure")
				var body strings.Builder
				for i := 0; i < 40; i++ {
					x := float64(i) + 0.25
					if width == 64 {
						body.WriteString(setDouble(1028+i*8, math.Float64bits(x)))
					} else {
						body.WriteString(setWord(1028+i*8, math.Float32bits(float32(x))))
					}
					fmt.Fprintf(&body, "addi 3,1,%d\nstw 3,%d(1)\naddi 3,1,%d\nstw 3,%d(1)\n", 1028+i*8, 24+i*8, 2052+i*8, 28+i*8)
				}
				body.WriteString("bl fn_pressure\n")
				for i := 0; i < 40; i++ {
					x := float64(i) + 0.25
					if conversion && i == 39 {
						x = 39
					}
					if width == 64 {
						body.WriteString(checkDouble(2052+i*8, math.Float64bits(x)))
					} else {
						body.WriteString(checkWord(2052+i*8, math.Float32bits(float32(x))))
					}
				}
				runFloatELF(t, dir, dump, []string{"pressure"}, body.String(), nil)
			})
		}
	}
}
