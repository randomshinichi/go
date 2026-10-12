//go:build darwin && ppc && arc4randomtest

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"os"
	"strconv"
	"testing"
	"unsafe"

	rt "runtime"
)

// TestArc4randomBufTrampolineShim exercises the Darwin/ppc assembly adapter
// directly. Run with DYLD_INSERT_LIBRARIES naming the deterministic shim and
// ARC4RANDOM_SHIM_COUNT naming its call-count file.
func TestArc4randomBufTrampolineShim(t *testing.T) {
	countPath := os.Getenv("ARC4RANDOM_SHIM_COUNT")
	if countPath == "" || os.Getenv("DYLD_INSERT_LIBRARIES") == "" {
		t.Skip("requires the deterministic arc4random DYLD shim")
	}

	type argsBlock struct {
		p     unsafe.Pointer
		n     int32
		guard [32]byte
	}
	const sentinel = byte(0xa7)
	var storage [48]byte
	lengths := []int{0, 1, 4, 5, 8, 32}
	for _, n := range lengths {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			for i := range storage {
				storage[i] = sentinel
			}
			var args argsBlock
			args.p = unsafe.Pointer(&storage[1]) // intentionally unaligned
			args.n = int32(n)
			for i := range args.guard {
				args.guard[i] = sentinel
			}

			before := shimArc4randomCalls(t, countPath)
			rt.Arc4randomBufTrampolineTestCall(unsafe.Pointer(&args))
			after := shimArc4randomCalls(t, countPath)
			wantCalls := (n + 3) / 4
			if after-before != wantCalls {
				t.Errorf("arc4random calls = %d, want %d", after-before, wantCalls)
			}
			if args.p != unsafe.Pointer(&storage[1]) || args.n != int32(n) {
				t.Errorf("argument block overwritten: p=%p n=%d", args.p, args.n)
			}
			for i, b := range args.guard {
				if b != sentinel {
					t.Errorf("argument-block guard[%d] = %#x, want %#x", i, b, sentinel)
				}
			}

			for i := range storage {
				want := sentinel
				if i >= 1 && i < n+1 {
					wordIndex := before + (i-1)/4
					shift := uint((i - 1) % 4 * 8)
					word := uint32(0x10203040 + wordIndex*0x01010101)
					want = byte(word >> shift)
				}
				if storage[i] != want {
					t.Errorf("storage[%d] = %#x, want %#x", i, storage[i], want)
				}
			}
		})
	}
}

func shimArc4randomCalls(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("invalid shim call count %q: %v", data, err)
	}
	return n
}
