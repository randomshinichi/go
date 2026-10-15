package ppc

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPPC32AssemblyDoesNotAssumeR0IsZero(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../"))
	asmPath := filepath.Join(src, "runtime/asm_ppc.s")
	asm, err := os.ReadFile(asmPath)
	if err != nil {
		t.Fatal(err)
	}
	checkGogo := func(text string) bool {
		start := strings.Index(text, "TEXT runtime·gogo(SB)")
		if start < 0 {
			return false
		}
		end := strings.Index(text[start+1:], "\nTEXT ")
		if end < 0 {
			end = len(text) - start - 1
		}
		body := text[start : start+1+end]
		last := -1
		for _, instruction := range []string{
			"ADDI R6, R0, $0",
			"STW R6, gobuf_sp(R5)",
			"STW R6, gobuf_lr(R5)",
			"STW R6, gobuf_ctxt(R5)",
		} {
			position := strings.Index(body, instruction)
			if position <= last {
				return false
			}
			last = position
		}
		return !strings.Contains(body, "STW R0, gobuf_")
	}
	if !checkGogo(string(asm)) {
		t.Fatal("gogo must materialize zero in R6 before clearing all gobuf fields")
	}

	// Negative control: restore the original bad source operand. The exact
	// gogo contract must reject it, proving this check can detect regression.
	broken := strings.Replace(string(asm), "STW R6, gobuf_ctxt(R5)", "STW R0, gobuf_ctxt(R5)", 1)
	if broken == string(asm) {
		t.Fatal("negative-control injection point missing")
	}
	if checkGogo(broken) {
		t.Fatal("negative control accepted an assumed-zero R0 store")
	}

	for _, item := range []struct {
		path  string
		funcs []string
	}{
		{asmPath, []string{"runtime·mcall", "runtime·gosave_systemstack_switch", "runtime·systemstack"}},
		{filepath.Join(src, "reflect/asm_ppc.s"), []string{"·makeFuncStub", "·methodValueCall"}},
		{filepath.Join(src, "internal/runtime/atomic/atomic_ppc.s"), []string{"·casRaw"}},
	} {
		data, err := os.ReadFile(item.path)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range item.funcs {
			start := strings.Index(string(data), "TEXT "+name+"(SB)")
			if start < 0 {
				t.Fatalf("%s: function %s not found", item.path, name)
			}
			end := strings.Index(string(data)[start+1:], "\nTEXT ")
			if end < 0 {
				end = len(data) - start - 1
			}
			body := string(data)[start : start+1+end]
			r0IsZero := false
			for _, line := range strings.Split(body, "\n") {
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				opcode := fields[0]
				if opcode == "ADDI" && len(fields) >= 4 && fields[1] == "R0," && fields[2] == "R0," && fields[3] == "$0" {
					r0IsZero = true
					continue
				}
				storeUsesR0AsZero := len(fields) >= 3 && fields[1] == "R0," &&
					((opcode == "STW" && (strings.Contains(fields[2], "g_sched+") || strings.Contains(fields[2], "ARG_REGARGS"))) ||
						(opcode == "STB" && (strings.Contains(fields[2], "LOCAL_RETVALID") || strings.Contains(fields[2], "ret+12(FP)"))))
				if storeUsesR0AsZero && !r0IsZero {
					t.Errorf("%s: %s uses R0 as zero without materializing it before %s", item.path, name, line)
				}
				if opcode == "BL" || opcode == "BCTRL" || (len(fields) > 1 && fields[1] == "R0," && opcode != "STW" && opcode != "STB") {
					r0IsZero = false
				}
			}
		}
	}
}
