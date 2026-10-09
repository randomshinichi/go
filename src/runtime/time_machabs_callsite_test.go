// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"internal/testenv"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestNanotime1CallSiteSyntax is a syntactic tripwire for the Darwin
// nanotime1 call site, runnable on any host that has the runtime source
// (it reads sys_darwin.go from the test's working directory, and skips when
// that file is absent, e.g. in a copied test binary). It requires nanotime1 to
// end in `return machTimeToNanos(...)` and to contain no multiplication or
// division of its own, so the old inline `t *= numer; t /= denom` cannot return
// unnoticed on a machine that cannot run Darwin. It checks the shape of the
// code, not its arithmetic; TestNanotimeCallSite (darwin only) checks the
// arithmetic.
func TestNanotime1CallSiteSyntax(t *testing.T) {
	src, err := os.ReadFile("sys_darwin.go")
	if err != nil {
		t.Skipf("runtime source not available: %v", err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "sys_darwin.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var fn *ast.FuncDecl
	for _, d := range f.Decls {
		if d, ok := d.(*ast.FuncDecl); ok && d.Name.Name == "nanotime1" {
			fn = d
		}
	}
	if fn == nil || fn.Body == nil {
		t.Fatal("sys_darwin.go: func nanotime1 not found")
	}
	stmts := fn.Body.List
	ret, ok := stmts[len(stmts)-1].(*ast.ReturnStmt)
	var call *ast.CallExpr
	if ok && len(ret.Results) == 1 {
		call, _ = ret.Results[0].(*ast.CallExpr)
	}
	if call == nil {
		t.Fatalf("%s: nanotime1 does not end in a single return of a call", fset.Position(fn.Pos()))
	}
	if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "machTimeToNanos" {
		t.Errorf("%s: nanotime1 returns %T, want a call to machTimeToNanos", fset.Position(call.Pos()), call.Fun)
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.BinaryExpr:
			if n.Op == token.MUL || n.Op == token.QUO || n.Op == token.REM {
				t.Errorf("%s: nanotime1 does its own %s; the conversion belongs in machTimeToNanos", fset.Position(n.Pos()), n.Op)
			}
		case *ast.AssignStmt:
			if n.Tok == token.MUL_ASSIGN || n.Tok == token.QUO_ASSIGN || n.Tok == token.REM_ASSIGN {
				t.Errorf("%s: nanotime1 does its own %s; the conversion belongs in machTimeToNanos", fset.Position(n.Pos()), n.Tok)
			}
		}
		return true
	})
}

// TestNanotime1NoSplitCallsPPC compiles the runtime for darwin/ppc and requires
// that nothing nanotime1 calls has a stack check. nanotime1 runs inside
// exitsyscall while stackguard0 is stackPreempt and throwsplit is set, so any
// stack check it reaches is fatal ("stack split at bad time"). On ppc32 a
// 64-bit / or % in machTimeToNanos compiled to calls to runtime.uint64div and
// uint64mod, which have stack checks. runtime.panicdivide, reached only for a
// zero denominator, is the one exception. Functions written in assembly are
// not in the listing and are not followed.
func TestNanotime1NoSplitCallsPPC(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the runtime for darwin/ppc")
	}
	testenv.MustHaveGoBuild(t)
	cmd := exec.Command(testenv.GoToolPath(t), "build", "-gcflags=runtime=-S", "-o", os.DevNull, "runtime")
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc", "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", cmd, err, out)
	}
	header := regexp.MustCompile(`^(\S+) STEXT\b(.*)`)
	call := regexp.MustCompile(`\tCALL\t([^(\s]+)\(SB\)`)
	nosplit := map[string]bool{}
	calls := map[string][]string{}
	var fn string
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		line := sc.Text()
		if m := header.FindStringSubmatch(line); m != nil {
			fn = m[1]
			nosplit[fn] = strings.Contains(m[2], " nosplit")
			continue
		}
		if m := call.FindStringSubmatch(line); m != nil && fn != "" {
			calls[fn] = append(calls[fn], m[1])
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	const root = "runtime.nanotime1"
	if !nosplit[root] {
		t.Fatalf("%s is missing from the listing or not nosplit", root)
	}
	from := map[string]string{root: ""}
	queue := []string{root}
	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		for _, c := range calls[fn] {
			if _, seen := from[c]; seen {
				continue
			}
			from[c] = fn
			ns, listed := nosplit[c]
			if !listed || c == "runtime.panicdivide" {
				continue
			}
			if !ns {
				path := c
				for p := fn; p != ""; p = from[p] {
					path = p + " -> " + path
				}
				t.Errorf("has a stack check: %s", path)
				continue
			}
			queue = append(queue, c)
		}
	}
}
