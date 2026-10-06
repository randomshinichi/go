// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
)

// TestGetGCMaskOnDemand exercises runtime.getGCMaskOnDemand: every
// iteration creates a fresh array type whose pointer mask is too large to
// be emitted statically (reflect.ArrayOf sets TFlagGCMaskOnDemand), so the
// mask is built on first use by whoever needs it first -- GC mark workers
// scanning the array objects and write barriers in typedmemmove -- while
// the others wait for it. A wrong or lost mask makes the GC free the
// objects the arrays point to, which the churn below then overwrites.
func TestGetGCMaskOnDemand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))

	const (
		firstLen   = 16385 // smallest array of pointers with an on-demand mask (abi.MaxPtrmaskBytes)
		iterations = 24
		objects    = 6
		slots      = 3
	)
	ptrType := reflect.TypeFor[*int]()
	for it := 0; it < iterations; it++ {
		n := firstLen + it
		at := reflect.ArrayOf(n, ptrType)
		idx := [slots]int{0, n / 2, n - 1}

		// srcs hold live pointers; dsts start empty and receive a copy of
		// the matching src.
		srcs := make([]reflect.Value, objects)
		dsts := make([]reflect.Value, objects)
		for i := range srcs {
			srcs[i] = reflect.New(at)
			dsts[i] = reflect.New(at)
			for s, j := range idx {
				v := new(int)
				*v = 1000*i + s + 1
				srcs[i].Elem().Index(j).Set(reflect.ValueOf(v))
			}
		}

		// Concurrent first use: collections scan the arrays while other
		// goroutines copy whole arrays (write barriers) and allocate.
		var wg sync.WaitGroup
		for g := 0; g < 3; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runtime.GC()
			}()
		}
		for i := range dsts {
			wg.Add(1)
			go func() {
				defer wg.Done()
				dsts[i].Elem().Set(srcs[i].Elem())
			}()
		}
		wg.Wait()

		// Everything referenced must have survived; churn so that freed
		// memory would be reused and show up as a changed value. The srcs
		// are dropped first so that only the dsts keep the integers alive.
		srcs = nil
		var sink [][]int
		for k := 0; k < 2000; k++ {
			sink = append(sink, []int{-k})
		}
		runtime.GC()
		for i := range dsts {
			for s, j := range idx {
				got := *dsts[i].Elem().Index(j).Interface().(*int)
				if want := 1000*i + s + 1; got != want {
					t.Fatalf("iteration %d: copy %d slot %d = %d, want %d", it, i, s, got, want)
				}
			}
		}
		runtime.KeepAlive(sink)
	}
}
