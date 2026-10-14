// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestRuntimeGoexitIsFrameless is the regression guard for bead goppc-339n.
//
// newproc1 records gp.stktopsp as the SP at the fabricated call to
// runtime.goexit and sets the goroutine's return address to
// goexit+PCQuantum. When the goroutine's function returns, execution continues
// at goexit+PCQuantum, which must be the call to goexit1, with SP unchanged.
// traceback.go's finishInternal then throws "traceback did not unwind
// completely" for any goroutine that has returned, unless the unwound outermost
// frame's SP equals stktopsp exactly. So goexit may not move SP.
//
// The only thing that keeps it from moving SP is the NOFRAME keyword on its TEXT
// line: without it, preprocess (link.go) gives a non-leaf function a 32-byte
// frame. Nothing else in the tree notices that keyword being lost; the throwing
// unwind is reached by GC scanning and stack growth, as well as heap dumps and
// panic handling. Profiler walks use silent errors and cannot cause this throw.
//
// The test takes the goexit TEXT block verbatim from the real runtime source of
// each architecture, assembles it with the assembler built from this tree (so
// preprocess is the code under test), and reads the assembler's -S listing,
// which is the post-preprocess program with final offsets. It requires that
//
//   - the function header reports locals=0x0, i.e. no frame was inserted;
//   - the second instruction, at offset 4 (sys.PCQuantum on ppc and ppc64), is
//     the BL to runtime.goexit1: that is what the fabricated return address
//     reaches, and an inserted prologue or a missing leading dead instruction
//     both move it;
//   - the one instruction before it does not write R1.
//
// The last point is a bounded form of "does not move SP without the pcsp
// table saying so": with the call pinned at offset 4 there is exactly one
// instruction that can run before it, and it is checked directly. A general
// SP-versus-pcsp consistency check for arbitrary functions is out of scope here.
func TestRuntimeGoexitIsFrameless(t *testing.T) {
	if gcflags := os.Getenv("GO_GCFLAGS"); gcflags != "" {
		t.Fatalf("required guard %s cannot execute: GO_GCFLAGS=%q disables child builds", t.Name(), gcflags)
	}
	if !testenv.HasGoBuild() {
		t.Fatalf("required guard %s cannot execute: go build is unavailable", t.Name())
	}
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("required guard %s cannot locate its source file", t.Name())
	}
	testFile, err := filepath.Abs(testFile)
	if err != nil {
		t.Fatalf("required guard %s cannot make its source path absolute: %v", t.Name(), err)
	}
	goroot := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "..", "..", ".."))
	goroot, err = filepath.EvalSymlinks(goroot)
	if err != nil {
		t.Fatalf("required guard %s cannot resolve its source GOROOT: %v", t.Name(), err)
	}
	if inherited := os.Getenv("GOROOT"); inherited != "" {
		inherited, err = filepath.Abs(inherited)
		if err == nil {
			inherited, err = filepath.EvalSymlinks(inherited)
		}
		if err != nil {
			t.Fatalf("required guard %s cannot resolve inherited GOROOT: %v", t.Name(), err)
		}
		if inherited != goroot {
			t.Fatalf("required guard %s refuses mismatched GOROOT: source tree is %s, environment names %s", t.Name(), goroot, inherited)
		}
	}
	goTool := filepath.Join(goroot, "bin", "go")
	if info, err := os.Stat(goTool); err != nil || info.Mode()&0o111 == 0 {
		t.Fatalf("required guard %s cannot execute worktree Go tool %s: %v", t.Name(), goTool, err)
	}

	asm := filepath.Join(t.TempDir(), "asm")
	build := exec.Command(goTool, "build", "-o", asm, "cmd/asm")
	build.Env = append(os.Environ(), "GOROOT="+goroot, "GOOS=", "GOARCH=", "GOFLAGS=")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building cmd/asm from the tree under test: %v\n%s", err, out)
	}

	// A -run filter can omit child subtests while leaving the parent green.
	executed := make(map[string]bool, 2)
	for _, tc := range []struct{ goarch, file string }{
		{"ppc", "asm_ppc.s"},
		{"ppc64", "asm_ppc64x.s"},
	} {
		t.Run(tc.goarch, func(t *testing.T) {
			executed[tc.goarch] = true
			src, err := os.ReadFile(filepath.Join(goroot, "src", "runtime", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			body, ok := textBlock(string(src), "runtime·goexit(SB)")
			if !ok {
				t.Fatalf("%s has no TEXT runtime·goexit(SB)", tc.file)
			}
			dir := t.TempDir()
			in := filepath.Join(dir, "goexit.s")
			if err := os.WriteFile(in, []byte("#include \"textflag.h\"\n"+body), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(asm, "-S", "-p", "runtime", "-I", filepath.Join(goroot, "src", "runtime"),
				"-o", filepath.Join(dir, "goexit.o"), in)
			cmd.Env = append(os.Environ(), "GOROOT="+goroot, "GOOS=darwin", "GOARCH="+tc.goarch)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("assembling runtime.goexit from %s: %v\n%s\n%s", tc.file, err, out, body)
			}
			if msg := checkGoexitListing(string(out)); msg != "" {
				t.Errorf("%s runtime·goexit: %s\nsource:\n%s\nlisting:\n%s", tc.file, msg, body, out)
			}
		})
	}
	for _, goarch := range []string{"ppc", "ppc64"} {
		if !executed[goarch] {
			t.Errorf("required guard architecture check %s did not execute", goarch)
		}
	}
}

// TestRuntimeGoexitGuardMissingBuildControl verifies that the required guard
// fails, rather than skips, when its child-build capability is unavailable.
func TestRuntimeGoexitGuardMissingBuildControl(t *testing.T) {
	goTool := testenv.GoToolPath(t)
	goroot := testenv.GOROOT(t)
	cmd := exec.Command(goTool, "test", "-run", "^TestRuntimeGoexitIsFrameless$", "-count=1", ".")
	cmd.Dir = filepath.Join(goroot, "src", "cmd", "internal", "obj", "ppc")
	cmd.Env = append(os.Environ(), "GO_GCFLAGS=review21")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("goexit guard passed with GO_GCFLAGS=review21; expected missing-build failure:\n%s", out)
	}
	if !strings.Contains(string(out), "--- FAIL: TestRuntimeGoexitIsFrameless") ||
		!strings.Contains(string(out), "required guard TestRuntimeGoexitIsFrameless cannot execute") {
		t.Fatalf("missing-build control did not name the failed required guard:\n%s", out)
	}
	if strings.Contains(string(out), "--- SKIP: TestRuntimeGoexitIsFrameless") {
		t.Fatalf("missing-build control skipped the required guard:\n%s", out)
	}
}

// textBlock returns the lines of the TEXT directive naming sym through the line
// before the next blank line, which ends a function in the runtime assembly.
func textBlock(src, sym string) (string, bool) {
	lines := strings.Split(src, "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "TEXT") || !strings.Contains(l, sym) {
			continue
		}
		j := i
		for j < len(lines) && strings.TrimSpace(lines[j]) != "" {
			j++
		}
		return strings.Join(lines[i:j], "\n") + "\n", true
	}
	return "", false
}

var (
	listingHeader = regexp.MustCompile(`^runtime\.goexit STEXT .*\blocals=(0x[0-9a-f]+)`)
	listingInsn   = regexp.MustCompile(`^\s+0x[0-9a-f]{4} (\d+) \(\S+:\d+\)\t(\S+)\t(.*)$`)
)

// checkGoexitListing applies the rules above to an assembler -S listing and
// returns a description of the first violated rule, or "" if there is none.
func checkGoexitListing(listing string) string {
	type insn struct {
		off      int
		op, args string
	}
	var insns []insn
	locals := ""
	for _, l := range strings.Split(listing, "\n") {
		if m := listingHeader.FindStringSubmatch(l); m != nil {
			locals = m[1]
		} else if m := listingInsn.FindStringSubmatch(l); m != nil {
			switch m[2] {
			case "TEXT", "PCDATA", "FUNCDATA", "NOP":
				continue
			}
			off, _ := strconv.Atoi(m[1])
			insns = append(insns, insn{off, m[2], m[3]})
		}
	}
	if locals == "" {
		return "no 'runtime.goexit STEXT ... locals=' header in the assembler listing"
	}
	if locals != "0x0" {
		return "goexit was given a frame (locals=" + locals + "); it must be NOFRAME so SP at the goexit1 call equals gp.stktopsp"
	}
	if len(insns) < 2 {
		return "goexit has fewer than two instructions"
	}
	if c := insns[1]; c.off != 4 || (c.op != "BL" && c.op != "CALL") || !strings.HasPrefix(c.args, "runtime.goexit1(SB)") {
		return "the instruction at goexit+PCQuantum (offset 4) is not a call to runtime.goexit1: it is " +
			c.op + " " + c.args + " at offset " + strconv.Itoa(c.off) +
			"; newproc1 returns the goroutine there"
	}
	f := insns[0]
	if ops := strings.Split(f.args, ", "); ops[len(ops)-1] == "R1" {
		return "the instruction before the goexit1 call writes SP: " + f.op + " " + f.args
	}
	return ""
}
