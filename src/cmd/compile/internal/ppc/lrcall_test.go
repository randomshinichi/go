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

const lrCallFixture = `package main
type caller interface { Call() int }
//go:noinline
func callClosure(f func() int) int { return f() }
//go:noinline
func callInterface(x caller) int { return x.Call() }
`

// checkLRCallWords checks the real instruction words emitted for the supplied
// Go functions. MTLR R12 followed by BLRL is the LR-indirect-call sequence;
// CTR-based forms are unsafe because asyncPreempt does not preserve CTR.
func checkLRCallWords(fn string, words []uint32) error {
	if len(words) == 0 {
		return fmt.Errorf("%s: compiler listing has no instructions", fn)
	}
	pairs := 0
	mtctr, bctrl := -1, -1
	for i, word := range words {
		if word&0xfc1fffff == 0x7c0903a6 && mtctr < 0 {
			mtctr = i
		}
		if word == 0x4e800421 && bctrl < 0 {
			bctrl = i
		}
		if word == 0x7d8803a6 && i+1 < len(words) && words[i+1] == 0x4e800021 {
			pairs++
		}
	}
	if bctrl >= 0 {
		return fmt.Errorf("%s: found BCTRL at instruction %d", fn, bctrl)
	}
	if mtctr >= 0 {
		return fmt.Errorf("%s: found MTCTR at instruction %d", fn, mtctr)
	}
	if pairs == 0 {
		return fmt.Errorf("%s: no adjacent MTLR R12/BLRL sequence", fn)
	}
	return nil
}

func TestLRIndirectCallLowering(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "lrcall.go")
	if err := os.WriteFile(src, []byte(lrCallFixture), 0644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "lrcall.o")
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-p", "main", "-S", "-o", obj, src)
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH=ppc")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compile ppc32 indirect-call fixture: %v\n%s", err, out)
	}
	listing := string(out)
	for _, fn := range []string{"main.callClosure", "main.callInterface"} {
		if err := checkLRCallWords(fn, emittedWords(t, listing, fn)); err != nil {
			t.Error(err)
		}
	}
}

func TestLRIndirectCallEmptyListingRejected(t *testing.T) {
	err := checkLRCallWords("main.callClosure", nil)
	if err == nil || !strings.Contains(err.Error(), "listing has no instructions") {
		t.Fatalf("empty listing error = %v, want explicit no-instructions failure", err)
	}
}
