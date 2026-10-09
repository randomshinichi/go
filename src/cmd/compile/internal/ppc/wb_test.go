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
	"strconv"
	"strings"
	"testing"
)

// This exercises the real WB helpers and real compiler barrier stores, with a
// flush recorder that deliberately destroys every promised preserved register.
// go_asm.h uses a synthetic g->m->p layout; this tests symbolic offset use and
// the private call protocol, not actual runtime layout or GC marking correctness.
func TestWriteBarrierQEMU(t *testing.T) {
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
	if d := os.Getenv("PPC32_WB_EVIDENCE"); d != "" {
		dir = d
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	header := "#define g_m 0\n#define m_p 0\n#define p_wbBuf 0\n#define wbBuf_next 0\n#define wbBuf_end 4\n"
	if err := os.WriteFile(filepath.Join(dir, "go_asm.h"), []byte(header), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "asm", "-p", "runtime", "-I", dir, "-I", filepath.Join(testenv.GOROOT(t), "pkg/include"), "-S", "-o", filepath.Join(dir, "wb.o"), filepath.Join(testenv.GOROOT(t), "src/runtime/wb_ppc.s"))
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "wb.dump"), out, 0644)
	if err != nil {
		t.Fatalf("WB asm: %v\n%s", err, out)
	}
	dump := string(out)
	source := filepath.Join(dir, "wb.go")
	writePPCFixture(t, source, []byte("package main\n//go:nosplit\nfunc pointerStore(d **byte,s *byte) { *d = s }\n"), 0644)
	cmd = exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-d=ssa/check/on", "-S", "-o", filepath.Join(dir, "wb_go.o"), source)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err = cmd.CombinedOutput()
	writePPCFixture(t, filepath.Join(dir, "wb_go.dump"), out, 0644)
	if err != nil {
		t.Fatalf("WB compile: %v\n%s", err, out)
	}
	dump += string(out)
	preserved := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 20, 21, 22, 23, 24, 25, 26, 27, 28}
	for _, tc := range []struct {
		name                    string
		entries                 int
		flush, compiled, mutant bool
	}{
		{"reserve1", 1, false, false, false}, {"reserve2", 2, false, false, false}, {"reserve3", 3, false, false, false}, {"reserve4", 4, false, false, false}, {"reserve5", 5, false, false, false}, {"reserve6", 6, false, false, false}, {"reserve7", 7, false, false, false}, {"reserve8", 8, false, false, false},
		{"flush1", 1, true, false, false}, {"flush2", 2, true, false, false}, {"flush3", 3, true, false, false}, {"flush4", 4, true, false, false}, {"flush5", 5, true, false, false}, {"flush6", 6, true, false, false}, {"flush7", 7, true, false, false}, {"flush8", 8, true, false, false},
		{"compiled-reserve", 2, false, true, false}, {"compiled-flush", 2, true, true, false}, {"negative-eight-byte-entries", 2, false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var a strings.Builder
			a.WriteString("fragment_start:\nclrrwi 1,1,4\naddi 1,1,-4096\nmr 2,1\nli 0,0\nlis 3,0x1357\nori 3,3,0x9bdf\nstw 3,-1024(1)\nstw 3,2044(1)\naddi 30,1,1600\naddi 3,1,1604\nstw 3,1600(1)\naddi 3,1,1608\nstw 3,1604(1)\naddi 3,1,544\nstw 3,1612(1)\n")
			next := 512
			if tc.flush {
				next = 544
			}
			fmt.Fprintf(&a, "addi 3,1,%d\nstw 3,1608(1)\nli 3,0\nstw 3,1900(1)\n", next)
			for off := 508; off <= 548; off += 4 {
				a.WriteString(storeWord(off, 0xa5a5a5a5))
			}
			if tc.compiled {
				a.WriteString("addi 3,1,1400\nstw 3,24(1)\n")
				a.WriteString(storeWord(28, 0x55667788))
				a.WriteString(storeWord(1400, 0x11223344))
			} else {
				for _, r := range preserved {
					fmt.Fprintf(&a, "li %d,%d\n", r, 0x100+r)
				}
			}
			entry := fmt.Sprintf("runtime.gcWriteBarrier%d", tc.entries)
			if tc.compiled {
				entry = "main.pointerStore"
			}
			a.WriteString("bl entry\ncmpw 1,2\nbne harness_bad\n")
			if tc.compiled {
				for _, v := range []struct {
					off  int
					want uint32
				}{{512, 0x55667788}, {516, 0x11223344}, {520, 0xa5a5a5a5}, {1400, 0x55667788}} {
					fmt.Fprintf(&a, "lwz 3,%d(1)\n", v.off)
					a.WriteString(checkReg(3, v.want))
				}
			} else {
				for _, r := range preserved {
					fmt.Fprintf(&a, "cmpwi %d,%d\nbne harness_result_bad\n", r, 0x100+r)
				}
				a.WriteString("addi 31,1,512\ncmpw 29,31\nbne harness_result_bad\n")
			}
			fmt.Fprintf(&a, "lwz 3,1608(1)\naddi 4,1,%d\ncmpw 3,4\nbne harness_result_bad\nlwz 3,1900(1)\ncmpwi 3,%d\nbne harness_bad\n", 512+4*tc.entries, boolInt(tc.flush))
			a.WriteString("addi 4,1,1600\ncmpw 30,4\nbne harness_bad\n")
			for _, off := range []int{508, 544, 548} {
				fmt.Fprintf(&a, "lwz 3,%d(1)\n", off)
				a.WriteString(checkReg(3, 0xa5a5a5a5))
			}
			for _, off := range []int{-1024, 2044} {
				fmt.Fprintf(&a, "lwz 3,%d(1)\n", off)
				a.WriteString(checkReg(3, 0x13579bdf))
			}
			a.WriteString("li 3,0\nli 0,1\nsc\nharness_result_bad:\nli 3,1\nli 0,1\nsc\nharness_bad:\nli 3,2\nli 0,1\nsc\n")
			labels := map[string]string{entry: "entry", "gcWriteBarrier": "wb_helper", "runtime.wbBufFlush": "record_flush", "runtime.writeBarrier": "barrier_flag"}
			if tc.compiled {
				labels["runtime.gcWriteBarrier2"] = "wb_entry"
			}
			functions := []string{entry, "gcWriteBarrier"}
			if tc.compiled {
				functions = append(functions, "runtime.gcWriteBarrier2")
			}
			changed := 0
			for _, fn := range functions {
				fmt.Fprintf(&a, "%s:\n", labels[fn])
				words := emittedWords(t, dump, fn)
				relocs := fixtureFunctionRelocs(t, dump, fn)
				for i := 0; i < len(words); i++ {
					w := words[i]
					if rel, ok := relocs[i*4]; ok {
						target, ok := labels[rel.target]
						if !ok {
							t.Fatalf("unknown WB relocation %v", rel)
						}
						if rel.kind == "R_ADDRPOWER" {
							r := (w >> 21) & 31
							fmt.Fprintf(&a, "lis %d,(0x11000+%s-fragment_start)@ha\naddi %d,%d,(0x11000+%s-fragment_start)@l\n", r, target, r, r, target)
							i++
							continue
						}
						if w == 0x48000001 {
							fmt.Fprintf(&a, "bl %s\n", target)
						} else if w == 0x48000000 {
							fmt.Fprintf(&a, "b %s\n", target)
						} else {
							t.Fatalf("not branch %08x", w)
						}
					} else {
						if tc.mutant && fn == "runtime.gcWriteBarrier2" && w == 0x3ba00008 {
							w = 0x3ba00010
							changed++
						}
						fmt.Fprintf(&a, ".long 0x%08x\n", w)
					}
				}
			}
			if tc.mutant && changed != 1 {
				t.Fatalf("mutated %d size setters", changed)
			}
			a.WriteString("record_flush:\nlwz 31,1900(2)\naddi 31,31,1\nstw 31,1900(2)\naddi 31,2,512\nstw 31,1608(2)\n")
			// wbBufFlush is an ordinary Go call; these GPRs are caller-saved.
			for _, r := range append(append([]int{}, preserved...), 18, 19, 29, 30) {
				fmt.Fprintf(&a, "li %d,-1\n", r)
			}
			a.WriteString("blr\nbarrier_flag:\n.long 1\n")
			want := 0
			if tc.mutant {
				want = 1
			}
			executeIntELF(t, qemu, clang, dir, tc.name, a.String(), want, changed)
		})
	}
	censusPPCObjects(t, dir, "wb.o", "wb_go.o")
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// Census every instruction in the Go objects, not just QEMU's executed paths.
// GNU spelling matters: the shared ppc64 decoder prints MOVD for G3-legal addi.
func censusPPCObjects(t *testing.T, dir string, objects ...string) {
	t.Helper()
	allowed := strings.Fields("mflr mtlr li lis stw stwu stb sth lwz lbz lha lhz extsh extsb add addi subf neg mullw mulhwu mulhw divw divwu and or ori xor slw srw sraw srawi rotlw rlwnm cmpwi cmplwi cmpw cmplw beq bne blt bge bgt ble b blr nop")
	set := map[string]bool{}
	for _, s := range allowed {
		set[s] = true
	}
	counts := map[string]int{}
	for _, o := range objects {
		cmd := exec.Command(testenv.GoToolPath(t), "tool", "objdump", "-gnu", filepath.Join(dir, o))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("objdump %s: %v\n%s", o, err, out)
		}
		writePPCFixture(t, filepath.Join(dir, o+".objdump"), out, 0644)
		n := 0
		for _, line := range strings.Split(string(out), "\n") {
			if !strings.HasPrefix(line, "  ") {
				continue
			}
			if strings.Contains(line, "R_CALLPOWER:runtime.") {
				if !strings.Contains(line, "48000001") {
					t.Fatalf("unexpected panicdivide call instruction: %s", line)
				}
				n++
				continue
			}
			parts := strings.Split(line, " // ")
			if len(parts) != 2 {
				t.Fatalf("undecoded instruction: %s", line)
			}
			fields := strings.Fields(parts[1])
			if len(fields) == 0 || !set[fields[0]] {
				t.Fatalf("outside scalar 750 floor: %s", line)
			}
			counts[fields[0]]++
			n++
		}
		if n == 0 {
			t.Fatalf("empty census for %s", o)
		}
		t.Logf("objdump census %s: %d instructions, all scalar 750-floor forms", o, n)
	}
	var lines []string
	for op, n := range counts {
		lines = append(lines, op+"\t"+strconv.Itoa(n))
	}
	writePPCFixture(t, filepath.Join(dir, "census.tsv"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
