// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"internal/abi"
	"os"
	"testing"
	"unsafe"
)

func TestProcRegionFilename(t *testing.T) {
	t.Run("text", func(t *testing.T) {
		address := uint64(abi.FuncPCABIInternal(TestProcRegionFilename))
		filename := regionFilename(address)
		t.Logf("address=%#x filename=%q", address, filename)
		got, err := os.Stat(filename)
		if err != nil {
			t.Fatalf("regionFilename(%#x) = %q: %v", address, filename, err)
		}
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		want, err := os.Stat(executable)
		if err != nil {
			t.Fatal(err)
		}
		if !os.SameFile(got, want) {
			t.Fatalf("regionFilename(%#x) = %q, want executable %q", address, filename, executable)
		}
	})
	t.Run("shortBuffer", func(t *testing.T) {
		address := uint64(abi.FuncPCABIInternal(TestProcRegionFilename))
		buf := make([]byte, _MAXPATHLEN)
		// libc requires MAXPATHLEN bytes even when the path would fit.
		r := proc_regionfilename(os.Getpid(), address, unsafe.SliceData(buf), int64(len(buf)-1))
		if r != 0 {
			t.Fatalf("proc_regionfilename with short buffer = %d, want 0", r)
		}
	})
}
