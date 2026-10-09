// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"context"
	"fmt"
	"internal/abi"
	"internal/testenv"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const intFixture = `package main
//go:nosplit
func div(a,b int32) int32 { return a/b }
//go:nosplit
func mod(a,b int32) int32 { return a%b }
//go:nosplit
func udiv(a,b uint32) uint32 { return a/b }
//go:nosplit
func umod(a,b uint32) uint32 { return a%b }
//go:nosplit
func div16(a,b int16) int16 { return a/b }
//go:nosplit
func mod16(a,b int16) int16 { return a%b }
//go:nosplit
func udiv16(a,b uint16) uint16 { return a/b }
//go:nosplit
func umod16(a,b uint16) uint16 { return a%b }
//go:nosplit
func div8(a,b int8) int8 { return a/b }
//go:nosplit
func mod8(a,b int8) int8 { return a%b }
//go:nosplit
func udiv8(a,b uint8) uint8 { return a/b }
//go:nosplit
func umod8(a,b uint8) uint8 { return a%b }
//go:nosplit
func rotate(x uint32, k int) uint32 { return x<<uint(k&31) | x>>uint((-k)&31) }
//go:nosplit
func index(p *[4]byte, i int) byte { return p[i] }
//go:nosplit
func indexU(p *[4]byte, i uint) byte { return p[i] }
//go:nosplit
func extend(p *[4]byte, i int64) byte { return p[i] }
//go:nosplit
func extendU(p *[4]byte, i uint64) byte { return p[i] }
//go:nosplit
func pointerCopy(d,s *struct { p [8]*byte }) { *d = *s }
//go:nosplit
func pointerStore(d **byte,s *byte) { *d = s }
`

// This is the existing whole-function ELF adapter, extended to arithmetic,
// bulk memory and bounds helpers. Compiler/assembler bodies are never replaced
// with hand-authored recipes. Only symbol relocations are resolved. Bounds
// recorders stand in for panicBounds32/X's decoder, not for unwinding/recover.
func TestIntegerAndMemoryQEMU(t *testing.T) {
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
	if d := os.Getenv("PPC32_INT_EVIDENCE"); d != "" {
		dir = d
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	src := intFixture
	for _, n := range []int{0, 1, 3, 4, 7, 8, 31, 256} {
		src += fmt.Sprintf("\n//go:nosplit\nfunc zero%d(d *[%d]byte) { *d = [%d]byte{} }\n//go:nosplit\nfunc move%d(d,s *struct { x [%d]byte }) { *d = *s }\n", n, n, n, n, n)
	}
	for _, n := range []int{1, 2, 64} {
		src += fmt.Sprintf("\n//go:nosplit\nfunc wordzero%d(d *[%d]uint32) { *d = [%d]uint32{} }\n//go:nosplit\nfunc wordmove%d(d,s *struct { x [%d]uint32 }) { *d = *s }\n", n, n, n, n, n)
	}
	source := filepath.Join(dir, "ints.go")
	writePPCFixture(t, source, []byte(src), 0644)
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "ints.o"), source)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	output, err := cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "ints.dump"), output, 0644)
	if err != nil {
		t.Fatalf("compile: %v\n%s", err, output)
	}
	dump := string(output)
	// Intrinsification is source-symbol based. Compile a real math/bits caller
	// so this exercises the generic RotateLeft32 op, not a hand-emitted opcode.
	rotateSrc := filepath.Join(dir, "rotate.go")
	writePPCFixture(t, rotateSrc, []byte("package bits\nfunc RotateLeft32(x uint32, k int) uint32\n//go:nosplit\nfunc rotate(x uint32,k int) uint32 { return RotateLeft32(x,k) }\n"), 0644)
	cmd = exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "math/bits", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "rotate.o"), rotateSrc)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	output, err = cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "rotate.dump"), output, 0644)
	if err != nil {
		t.Fatalf("rotate compile: %v\n%s", err, output)
	}
	dump += string(output)
	if !strings.Contains(functionDump(dump, "math/bits.rotate"), "ROTLW") {
		t.Fatal("rotation fixture does not exercise RotateLeft32 lowering")
	}
	// Pointer aggregate copies must retain the compiler-inserted bulk barrier.
	copyDump := functionDump(dump, "main.pointerCopy")
	if !strings.Contains(copyDump, "runtime.wbMove") && !strings.Contains(copyDump, "runtime.typedmemmove") {
		t.Fatal("pointer copy lost barrier-aware runtime copy")
	}
	cmd = exec.Command(testenv.GoToolPath(t), "tool", "asm", "-p", "runtime", "-I", filepath.Join(testenv.GOROOT(t), "pkg/include"), "-S", "-o", filepath.Join(dir, "panic.o"), filepath.Join(testenv.GOROOT(t), "src/runtime/panic_ppc.s"))
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	output, err = cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "panic.dump"), output, 0644)
	if err != nil {
		t.Fatalf("panic asm: %v\n%s", err, output)
	}
	dump += string(output)

	run := func(name, entry, setup, check string, mutant string, want int) {
		t.Run(name, func(t *testing.T) {
			var a strings.Builder
			a.WriteString("clrrwi 1,1,4\naddi 1,1,-4096\nmr 2,1\nli 0,0\nlis 3,0x1357\nori 3,3,0x9bdf\nstw 3,2044(1)\nstw 3,-1024(1)\n")
			a.WriteString(setup)
			a.WriteString("bl fn_entry\nharness_after:\ncmpw 1,2\nbne harness_bad\n")
			a.WriteString(check)
			a.WriteString("lis 4,0x1357\nori 4,4,0x9bdf\nlwz 3,2044(1)\ncmpw 3,4\nbne harness_bad\nlwz 3,-1024(1)\ncmpw 3,4\nbne harness_bad\nli 3,0\nli 0,1\nsc\nharness_result_bad:\nli 3,1\nli 0,1\nsc\nharness_bad:\nli 3,2\nli 0,1\nsc\n")
			changed := 0
			labels := map[string]string{entry: "fn_entry", "runtime.panicBounds": "fn_bounds", "runtime.panicExtend": "fn_extend", "runtime.panicBounds32": "record_bounds", "runtime.panicBounds32X": "record_extend", "runtime.panicdivide": "record_divide"}
			for _, fn := range []string{entry, "runtime.panicBounds", "runtime.panicExtend"} {
				fmt.Fprintf(&a, "%s:\n", labels[fn])
				words := emittedWords(t, dump, fn)
				relocs := fixtureFunctionRelocs(t, dump, fn)
				for i, w := range words {
					if rel, ok := relocs[i*4]; ok {
						label, ok := labels[rel.target]
						if !ok {
							t.Fatalf("unsupported target %s", rel.target)
						}
						if rel.kind != "R_CALLPOWER" || w != 0x48000001 {
							t.Fatalf("unexpected relocation %v %08x", rel, w)
						}
						fmt.Fprintf(&a, "bl %s\n", label)
						if fn == entry && strings.Contains(rel.target, "panic") {
							fmt.Fprintf(&a, "panic_pc_%s:\n", strings.TrimPrefix(rel.target, "runtime."))
						}
					} else {
						if mutant == "quotient" && fn == entry && w&0xfc0007ff == 0x7c0003d6 {
							w = w&^uint32(0x7ff) | 0x396
							changed++
						}
						if mutant == "move-source" && fn == entry && w&0xfc1f0000 == 0x88040000 {
							w++ // load one byte past the declared source offset
							changed++
						}
						if mutant == "clear" && fn == entry && w&0xfc000000 == 0x98000000 {
							w = 0x60000000
							changed++
						}
						if mutant == "panic-save" && fn == "runtime.panicBounds" && w == 0x90610020 {
							w = 0x90610024
							changed++
						}
						fmt.Fprintf(&a, ".long 0x%08x\n", w)
					}
				}
			}
			if mutant != "" && changed == 0 {
				t.Fatal("negative control mutated no instruction")
			}
			// Recorders not used by a case are fatal, including an unexpected zero divisor.
			a.WriteString("record_divide:\n")
			if strings.HasPrefix(name, "divide-zero") {
				fmt.Fprintf(&a, "addi 1,1,%d\nb harness_after\n", functionFrame(t, dump, entry))
			} else {
				a.WriteString("b harness_bad\n")
			}
			for _, ext := range []bool{false, true} {
				label, helper := "record_bounds", "panicBounds"
				if ext {
					label, helper = "record_extend", "panicExtend"
				}
				fmt.Fprintf(&a, "%s:\n", label)
				if strings.HasPrefix(name, "bounds") && strings.Contains(entry, "extend") == ext {
					meta := panicMetadata(t, dump, entry, helper)
					code, signed, xreg, yreg, x, y := abi.BoundsDecode(meta)
					expectedSigned := !strings.HasSuffix(entry, "U")
					if code != abi.BoundsIndex || signed != expectedSigned || !xreg || !yreg {
						t.Fatalf("unexpected metadata %d: code=%d signed=%t", meta, code, signed)
					}
					// Check actual helper argument PC and exact save-area address.
					fmt.Fprintf(&a, "lwz 3,24(1)\nbl recorder_pc_%s\nrecorder_pc_%s:\nmflr 4\naddi 4,4,panic_pc_%s-recorder_pc_%s\ncmpw 3,4\nbne harness_result_bad\nlwz 5,28(1)\naddi 4,1,32\ncmpw 5,4\nbne harness_bad\n", helper, helper, helper, helper)
					if ext {
						fmt.Fprintf(&a, "lwz 3,%d(5)\n", 4*(x>>2))
						a.WriteString(checkReg(3, 0x12345678))
						fmt.Fprintf(&a, "lwz 3,%d(5)\n", 4*(x&3))
						a.WriteString(checkReg(3, 0x89abcdef))
					} else {
						fmt.Fprintf(&a, "lwz 3,%d(5)\n", 4*x)
						a.WriteString(checkReg(3, 0xffffffff))
					}
					fmt.Fprintf(&a, "lwz 3,%d(5)\n", 4*y)
					a.WriteString(checkReg(3, 4))
					// A panic never returns. Check back chains then explicitly return the
					// recorder to the harness, rather than claiming runtime recover works.
					a.WriteString("lwz 3,0(1)\naddi 4,1,96\ncmpw 3,4\nbne harness_bad\n")
					fmt.Fprintf(&a, "lwz 3,0(3)\ncmpw 3,2\nbne harness_bad\naddi 1,1,%d\nb harness_after\n", 96+functionFrame(t, dump, entry))
				} else {
					a.WriteString("b harness_bad\n")
				}
			}
			executeIntELF(t, qemu, clang, dir, name, a.String(), want, changed)
		})
	}
	for _, v := range [][2]int32{{-7, 3}, {-2147483648, -1}, {2147483647, -1}, {-7, -3}, {7, -3}} {
		for _, op := range []string{"div", "mod"} {
			val := int64(v[0]) / int64(v[1])
			if op == "mod" {
				val = int64(v[0]) % int64(v[1])
			}
			setup := stackWords(uint32(v[0]), uint32(v[1]))
			check := "lwz 3,32(1)\n" + checkReg(3, uint32(val))
			run(fmt.Sprintf("%s-%d-%d", op, v[0], v[1]), "main."+op, setup, check, "", 0)
		}
	}
	run("negative-division", "main.div", stackWords(0xfffffff9, 3), "lwz 3,32(1)\n"+checkReg(3, 0xfffffffe), "quotient", 1)
	for _, v := range [][2]uint32{{0xffffffff, 1}, {0xffffffff, 0x80000000}, {0x80000000, 3}, {0x80000000, 0xffffffff}, {0xffffffff, 0xffffffff}, {0x12345678, 0x10001}, {0, 7}} {
		// Host 64-bit division is independent of the ppc word divider under test.
		for _, op := range []string{"udiv", "umod"} {
			val := uint64(v[0]) / uint64(v[1])
			if op == "umod" {
				val = uint64(v[0]) % uint64(v[1])
			}
			run(fmt.Sprintf("%s-%08x-%08x", op, v[0], v[1]), "main."+op, stackWords(v[0], v[1]), "lwz 3,32(1)\n"+checkReg(3, uint32(val)), "", 0)
		}
	}
	for _, op := range []string{"div", "mod", "udiv", "umod"} {
		run("divide-zero-"+op, "main."+op, stackWords(7, 0), "", "", 0)
	}
	for _, width := range []int{8, 16} {
		for _, signed := range []bool{true, false} {
			bits := width
			mask := uint32(1<<bits) - 1
			min := int32(-1 << (bits - 1))
			max := int32(1<<(bits-1) - 1)
			fn := "udiv"
			if signed {
				fn = "div"
			}
			if width == 8 {
				fn += "8"
			} else {
				fn += "16"
			}
			if !signed {
				min, max = 0, int32(mask)
			}
			vectors := [][2]int32{{min, -1}, {max, 1}, {min, 1}, {max, max}}
			for _, op := range []string{fn, strings.Replace(fn, "div", "mod", 1)} {
				for _, v := range vectors {
					var result int64
					if signed {
						if op == fn {
							result = int64(v[0]) / int64(v[1])
						} else {
							result = int64(v[0]) % int64(v[1])
						}
					} else if op == fn {
						result = int64(uint32(v[0])&mask) / int64(uint32(v[1])&mask)
					} else {
						result = int64(uint32(v[0])&mask) % int64(uint32(v[1])&mask)
					}
					want := uint32(result)
					if signed && want&(1<<(bits-1)) != 0 {
						want |= ^mask
					}
					setup := smallIntWords(width, uint32(v[0]), uint32(v[1]))
					run(fmt.Sprintf("%s-%d-%d", op, v[0], v[1]), "main."+op, setup, smallIntResultCheck(width, signed, want), "", 0)
				}
				run("divide-zero-"+op, "main."+op, smallIntWords(width, 7, 0), "", "", 0)
			}
		}
	}
	for _, k := range []int{-1, 0, 31, 32, 65} {
		run(fmt.Sprintf("rotate-%d", k), "math/bits.rotate", stackWords(0x89abcdef, uint32(k)), "lwz 3,32(1)\n"+checkReg(3, bits.RotateLeft32(0x89abcdef, k)), "", 0)
	}
	for _, fn := range []string{"index", "indexU", "extend", "extendU"} {
		setup := "addi 3,1,512\nstw 3,24(1)\n" + storeWord(28, 0xffffffff)
		if strings.Contains(fn, "extend") {
			setup = "addi 3,1,512\nstw 3,24(1)\n" + storeWord(28, 0x12345678) + storeWord(32, 0x89abcdef)
		}
		run("bounds-"+fn, "main."+fn, setup, "", "", 0)
	}
	run("bounds-negative-save", "main.index", "addi 3,1,512\nstw 3,24(1)\n"+storeWord(28, 0xffffffff), "", "panic-save", 1)
	for _, n := range []int{0, 1, 3, 4, 7, 8, 31, 256} {
		for _, op := range []string{"zero", "move"} {
			setup, check := memoryCase(n, op == "move", false, false)
			run(fmt.Sprintf("%s-%d", op, n), fmt.Sprintf("main.%s%d", op, n), setup, check, "", 0)
			if op == "move" {
				setup, check = memoryCase(n, true, true, false)
				run(fmt.Sprintf("identical-%d", n), fmt.Sprintf("main.move%d", n), setup, check, "", 0)
			}
		}
	}
	setup, check := memoryCase(31, false, false, false)
	run("negative-zero", "main.zero31", setup, check, "clear", 1)
	setup, check = memoryCase(31, true, false, false)
	run("negative-move", "main.move31", setup, check, "move-source", 1)
	for _, n := range []int{1, 2, 64} {
		for _, op := range []string{"zero", "move"} {
			setup, check := memoryCase(n*4, op == "move", false, true)
			run(fmt.Sprintf("word-%s-%d", op, n), fmt.Sprintf("main.word%s%d", op, n), setup, check, "", 0)
		}
	}
	censusPPCObjects(t, dir, "ints.o", "rotate.o", "panic.o")
}

func functionDump(dump, fn string) string {
	start := strings.Index(dump, fn+" STEXT")
	if start < 0 {
		return ""
	}
	rows := strings.Split(dump[start:], "\n")
	for i := 1; i < len(rows); i++ {
		if !strings.HasPrefix(rows[i], "\t") {
			return strings.Join(rows[:i], "\n")
		}
	}
	return strings.Join(rows, "\n")
}
func functionFrame(t *testing.T, dump, fn string) int {
	t.Helper()
	m := regexp.MustCompile(`TEXT\s+` + regexp.QuoteMeta(fn) + `\(SB\).*\$(\d+)-`).FindStringSubmatch(functionDump(dump, fn))
	if m == nil {
		t.Fatalf("frame not found for %s", fn)
	}
	n, _ := strconv.Atoi(m[1])
	return n
}
func panicMetadata(t *testing.T, dump, fn, helper string) int {
	t.Helper()
	metadata := -1
	re := regexp.MustCompile(`PCDATA\s+\$4, \$(\d+)`)
	for _, line := range strings.Split(functionDump(dump, fn), "\n") {
		if m := re.FindStringSubmatch(line); m != nil {
			metadata, _ = strconv.Atoi(m[1])
		}
		if strings.Contains(line, "\tCALL\truntime."+helper+"(SB)") && metadata >= 0 {
			return metadata
		}
	}
	t.Fatalf("no bounds metadata for %s in %s", helper, fn)
	return 0
}
func checkReg(r int, w uint32) string {
	return fmt.Sprintf("lis 4,%d\nori 4,4,%d\ncmpw %d,4\nbne harness_result_bad\n", int16(w>>16), w&65535, r)
}
func storeWord(off int, w uint32) string {
	return fmt.Sprintf("lis 3,%d\nori 3,3,%d\nstw 3,%d(1)\n", int16(w>>16), w&65535, off)
}
func stackWords(a, b uint32) string { return storeWord(24, a) + storeWord(28, b) }
func smallIntResultCheck(width int, signed bool, want uint32) string {
	load := "lbz"
	if width == 16 {
		load = "lhz"
	}
	check := fmt.Sprintf("%s 3,28(1)\n", load)
	if signed {
		if width == 8 {
			check += "extsb 3,3\n"
		} else {
			check += "extsh 3,3\n"
		}
	}
	return check + checkReg(3, want)
}
func smallIntWords(width int, a, b uint32) string {
	if width == 8 {
		return fmt.Sprintf("li 3,%d\nstb 3,24(1)\nli 3,%d\nstb 3,25(1)\n", uint8(a), uint8(b))
	}
	return fmt.Sprintf("lis 3,%d\nori 3,3,%d\nsth 3,24(1)\nlis 3,%d\nori 3,3,%d\nsth 3,26(1)\n", int16(a>>16), uint16(a), int16(b>>16), uint16(b))
}
func memoryCase(n int, moving, identical, aligned bool) (string, string) {
	base, source := 513, 1025
	if aligned {
		base, source = 512, 1024
	}
	setup := fmt.Sprintf("addi 3,1,%d\nstw 3,24(1)\naddi 3,1,%d\nstw 3,28(1)\n", base, source)
	if identical {
		setup += fmt.Sprintf("addi 3,1,%d\nstw 3,28(1)\n", base)
	}
	for i := -1; i <= n; i++ {
		setup += fmt.Sprintf("li 3,165\nstb 3,%d(1)\nli 3,%d\nstb 3,%d(1)\n", base+i, (i+257)&255, source+i)
	}
	check := ""
	for i := -1; i <= n; i++ {
		want := 0
		if i < 0 || i == n || identical {
			want = 165
		} else if moving {
			want = (i + 257) & 255
		}
		check += fmt.Sprintf("lbz 3,%d(1)\ncmpwi 3,%d\nbne harness_result_bad\n", base+i, want)
		check += fmt.Sprintf("lbz 3,%d(1)\ncmpwi 3,%d\nbne harness_bad\n", source+i, (i+257)&255)
	}
	return setup, check
}
func writePPCFixture(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
}

func executeIntELF(t *testing.T, qemu, clang, dir, name, asm string, want, changed int) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".s"), []byte(asm), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name+".elf")
	writePPCTextELF(t, path, clangTextWords(t, clang, dir, asm))
	log := filepath.Join(dir, name+".decode")
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, qemu, "-cpu", "750", "-d", "in_asm", "-D", log, path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("INCONCLUSIVE: emulator timeout: %v", ctx.Err())
	}
	rc := 0
	if e, ok := err.(*exec.ExitError); ok {
		rc = e.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	status := fmt.Sprintf("command=%s -cpu 750 -d in_asm -D %s %s\nexit=%d\nmutations=%d\n%s", qemu, log, path, rc, changed, out)
	writePPCFixture(t, filepath.Join(dir, name+".status"), []byte(status), 0644)
	if rc != want {
		t.Fatalf("exit=%d want=%d\n%s", rc, want, out)
	}
	t.Logf("exit=%d expected=%d mutations=%d; full bits, caller SP and canaries checked", rc, want, changed)
}
