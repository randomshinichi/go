// run

// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

// Exercise PPC32's saved LR, closure context, and stack-copy protocol without
// fmt/testing, which require legacy Darwin imports unrelated to the runtime.
package main

import "runtime"

//go:noinline
func grow(n int, f func()) int {
	var pad [256]byte
	pad[n&255] = byte(n)
	if n == 0 {
		f()
		return 0
	}
	return grow(n-1, f) + int(pad[n&255])
}

//go:noinline
func callers() {
	var pcs [16]uintptr
	n := runtime.Callers(1, pcs[:])
	if n < 3 {
		panic("short caller chain")
	}
	if runtime.FuncForPC(pcs[0]-1).Name() != "main.callers" {
		for _, pc := range pcs[:n] {
			println("caller", runtime.FuncForPC(pc-1).Name())
		}
		panic("wrong innermost caller")
	}
	found := false
	for _, pc := range pcs[:n] {
		if runtime.FuncForPC(pc-1).Name() == "main.grow" {
			found = true
		}
	}
	if !found {
		panic("lost grow frame")
	}
}

func main() {
	const capture = 137
	done := make(chan int)
	go func(v int) {
		// This wrapper must begin with gogo's restored closure context.
		var walk func(int) int
		walk = func(n int) int {
			// Retain both the context and stack locals across morestack.
			var pad [256]byte
			pad[n&255] = byte(n)
			if n == 0 {
				runtime.Gosched()
				return v
			}
			return walk(n-1) + int(pad[n&255])
		}
		if got := walk(96); got != capture+96*97/2 {
			panic("closure context or stack locals lost")
		}
		if got := grow(96, callers); got != 96*97/2 {
			panic("copied stack locals lost")
		}
		done <- v
	}(capture)
	if <-done != capture {
		panic("goroutine capture lost")
	}
	println("ppc32 runtime stack/context OK")
}
