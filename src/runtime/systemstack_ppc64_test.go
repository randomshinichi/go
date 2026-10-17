// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestSystemstackTailReturn(t *testing.T) {
	if got := runtime.SystemstackTailReturn(); got != 15 {
		t.Fatalf("nested systemstack context/return: got %d, want 15", got)
	}
}

func TestSystemstackTailSnapshot(t *testing.T) {
	if !runtime.SystemstackTailcallSupported {
		t.Skip("systemstack tail call not implemented for this build mode")
	}
	out := runtime.SystemstackTailSnapshot()
	if out.RetPC == 0 || out.TailPC == 0 {
		t.Fatal("systemstack must contain both BLR and non-linking BCCTR")
	}
	t.Logf("restored-SP PCs: RET=%#x delta=%d, BCCTR=%#x delta=%d", out.RetPC, out.RetDelta, out.TailPC, out.TailDelta)
	if out.RetDelta != 0 || out.TailDelta != 0 {
		t.Error("restored SP requires zero PC/SP delta at both branch PCs")
	}
	if out.RetN < 2 || out.TailN < 2 || out.RetN == len(out.Ret) || out.TailN == len(out.Tail) {
		t.Fatalf("empty or truncated interrupted traceback: RET=%d tail=%d", out.RetN, out.TailN)
	}
	// The first PC is the synthetic interruption site, so compare every
	// remaining caller PC, not just the number of frames or function names.
	ret, tail := out.Ret[1:out.RetN], out.Tail[1:out.TailN]
	if !reflect.DeepEqual(ret, tail) {
		t.Errorf("interrupted caller suffixes differ:\nRET: %x\ntail: %x", ret, tail)
	}
	// Also require the caller closure that stale tail metadata used to omit.
	for _, pcs := range [][]uintptr{ret, tail} {
		found := false
		frames := runtime.CallersFrames(pcs)
		for {
			f, more := frames.Next()
			if strings.HasSuffix(f.Function, "runtime.SystemstackTailSnapshot.func1") {
				found = true
			}
			if !more {
				break
			}
		}
		if !found {
			t.Errorf("interrupted traceback omits systemstack caller closure: %x", pcs)
		}
	}
}
