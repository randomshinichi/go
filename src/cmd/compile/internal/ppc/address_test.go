// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestLargeOffsetMaterialized checks the signed-16-bit addi boundary at the
// compiler/assembler interface. Before the ppc32 address emitter materialized
// large offsets, this compilation passed 86016 directly to addi and failed
// with an out-of-range immediate diagnostic.
func TestLargeOffsetMaterialized(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "large_offset.go")
	if err := os.WriteFile(src, []byte(`package ppc
import "unsafe"
//go:nosplit
func largeOffset(p *byte) *byte { return (*byte)(unsafe.Add(unsafe.Pointer(p), 86016)) }
`), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-S", "-o", filepath.Join(dir, "large_offset.o"), src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile large offset: %v\n%s", err, out)
	}
	dump := string(out)
	for _, op := range []string{"ADDIS", "ORI", "ADD"} {
		if !strings.Contains(dump, op) {
			t.Fatalf("large offset lowering lacks %s materialization/add; compiler output:\n%s", op, dump)
		}
	}
}
