// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

func init() { register("CloneStackBoundary", cloneStackBoundary) }

func cloneBoundaryChildPC() uintptr
func cloneBoundaryStart(flags int32, stk, mp, gp unsafe.Pointer, fn uintptr) int32

//go:noinline
func cloneBoundaryStack() *[8192 + 64]byte {
	// A raw child SP must not move if the parent's goroutine stack grows.
	return new([8192 + 64]byte)
}

func cloneStackBoundary() {
	// Leave writable sentinels above the exclusive stack bound. A child
	// callee's linkage writes must stay below it, even without a Go g.
	stack := cloneBoundaryStack()
	top := (uintptr(unsafe.Pointer(&stack[8192])) + 15) &^ 15
	guard := (*[8]uint32)(unsafe.Pointer(top))
	for i := range guard {
		guard[i] = 0xa5a5a5a5
	}
	var done uint32
	// Use the runtime's thread flags, also supported by linux-user QEMU.
	// With mp nil, clone leaves gp in R8 without treating it as a Go g.
	pid := cloneBoundaryStart(0x50f00, unsafe.Pointer(top), nil, unsafe.Pointer(&done), cloneBoundaryChildPC())
	if pid < 0 {
		panic("clone failed")
	}
	deadline := time.Now().Add(5 * time.Second)
	for atomic.LoadUint32(&done) == 0 {
		if time.Now().After(deadline) {
			panic("clone child timeout")
		}
		time.Sleep(time.Millisecond)
	}
	for _, v := range guard {
		if v != 0xa5a5a5a5 {
			panic("clone wrote above child stack")
		}
	}
	runtime.KeepAlive(stack)
	println("OK")
}
