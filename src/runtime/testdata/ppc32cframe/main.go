// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

// ppc32cframe exercises the darwin/ppc C-call frame contract through the
// production path, syscall.rawsyscalln -> runtime.libcCall -> asmcgocall ->
// syscallN_trampoline -> a REAL C function, from both stack modes:
//
//	user: a goroutine stack. asmcgocall switches to g0 and back
//	      (its R20 flag is 1).
//	sys:  runtime.systemstack, i.e. already on g0. asmcgocall calls directly
//	      (its R20 flag is 0). This is the leg that proves goppc-j5yb.
//
// The C functions come from libframes.dylib (helpers.c, compiled by the
// Leopard system gcc). Usage: ppc32cframe /path/to/libframes.dylib
//
// Build (cross, on the host; see build.sh in this directory):
//
//	GOOS=darwin GOARCH=ppc go build -ldflags=-checklinkname=0 -o ppc32cframe .
//
// -checklinkname=0 is needed because the probe pulls syscall.rawsyscalln and
// runtime.systemstack by name; a probe must use the production entry points.
//
// Optional arguments narrow the run, for isolating one failure:
//
//	ppc32cframe lib.dylib [user|sys] [case-name-prefix]
//
// Every check prints one line; the exit status is 0 only when every line
// is PASS. A crash is also a failure: the ENTER line before each call
// shows where the process was.
package main

import (
	"os"
	"strings"
	"unsafe"
)

//go:cgo_import_dynamic libc_dlopen dlopen "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_dlsym dlsym "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic _ _ "/usr/lib/libSystem.B.dylib"

//go:linkname rawsyscalln syscall.rawsyscalln
func rawsyscalln(fn uintptr, args ...uintptr) (r1, r2, err uintptr)

//go:linkname systemstack runtime.systemstack
func systemstack(fn func())

var dlopenAddr, dlsymAddr uintptr

func gsp() (gptr, stk uintptr)

const rtldDefault = ^uintptr(1) // (void *)-2, RTLD_DEFAULT on Darwin

// record is filled in by call while running on whichever stack the leg uses.
// It is allocated before the leg starts because g0 must not allocate.
type record struct {
	g0, sp0, g1, sp1 uintptr
	r1               uintptr
	sentinels        bool
}

const sentinelBase = 0xC0DE0000

//go:noinline
func call(fn uintptr, args []uintptr, rec *record) {
	var sent [16]uintptr // stack sentinels in this frame, above the C call's area
	for i := range sent {
		sent[i] = sentinelBase + uintptr(i)
	}
	rec.g0, rec.sp0 = gsp()
	rec.r1, _, _ = rawsyscalln(fn, args...)
	rec.g1, rec.sp1 = gsp()
	rec.sentinels = true
	for i := range sent {
		if sent[i] != sentinelBase+uintptr(i) {
			rec.sentinels = false
		}
	}
}

func cstr(s string) uintptr {
	b := append([]byte(s), 0)
	keep = append(keep, b)
	return uintptr(unsafe.Pointer(&b[0]))
}

var keep [][]byte

func sym(lib uintptr, name string) uintptr {
	r, _, _ := rawsyscalln(dlsymAddr, lib, cstr(name))
	if r == 0 {
		println("FAIL dlsym", name)
		os.Exit(2)
	}
	return r
}

type check struct {
	name string
	fn   uintptr
	args []uintptr
	want uintptr
	// verify runs after the call on the user stack and may veto the result.
	verify func() bool
}

var failed bool

func run(leg string, c check, onSystem bool) {
	rec := new(record)
	println("ENTER", leg, c.name, "nargs", len(c.args))
	if onSystem {
		systemstack(func() { call(c.fn, c.args, rec) })
	} else {
		call(c.fn, c.args, rec)
	}
	ok := rec.r1 == c.want && rec.g0 == rec.g1 && rec.sp0 == rec.sp1 && rec.sentinels
	if c.verify != nil && !c.verify() {
		ok = false
	}
	verdict := "PASS"
	if !ok {
		verdict = "FAIL"
		failed = true
	}
	println(verdict, leg, c.name, "got", int32(rec.r1), "want", int32(c.want),
		"g", rec.g0, rec.g1, "sp", rec.sp0, rec.sp1, "sentinels", rec.sentinels)
}

func main() {
	if len(os.Args) < 2 {
		println("usage: ppc32cframe /path/to/libframes.dylib")
		os.Exit(2)
	}
	lib, _, _ := rawsyscalln(dlopenAddr, cstr(os.Args[1]), 2 /* RTLD_NOW */)
	if lib == 0 {
		println("FAIL dlopen", os.Args[1])
		os.Exit(2)
	}
	n10 := sym(lib, "n10")
	n12 := sym(lib, "n12")
	home8 := sym(lib, "home8")
	snprintf := sym(rtldDefault, "snprintf")
	printf := sym(rtldDefault, "printf")
	fflush := sym(rtldDefault, "fflush")

	buf := make([]byte, 128)
	bufp := uintptr(unsafe.Pointer(&buf[0]))
	const want8 = "1 2 3 4 5 6 7 8"
	checks := []check{
		{name: "n10(1..8,9,10)=910", fn: n10,
			args: []uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, want: 910},
		{name: "n12(1..8,9..12)=9101112", fn: n12,
			args: []uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, want: 9101112},
		{name: "home8(r3-r10 homed)=0", fn: home8,
			args: []uintptr{0x1111, 0x2222, 0x3333, 0x4444, 0x5555, 0x6666, 0x7777, 0x8888}, want: 0},
		{name: "snprintf(11 words)", fn: snprintf,
			args: []uintptr{bufp, 128, cstr("%d %d %d %d %d %d %d %d"), 1, 2, 3, 4, 5, 6, 7, 8},
			want: uintptr(len(want8)),
			verify: func() bool {
				return string(buf[:len(want8)]) == want8 && buf[len(want8)] == 0
			}},
		{name: "printf(9 words)", fn: printf,
			args: []uintptr{cstr("printf-control: %d %d %d %d %d %d %d %d\n"), 1, 2, 3, 4, 5, 6, 7, 8},
			want: uintptr(len("printf-control: 1 2 3 4 5 6 7 8\n"))},
		{name: "fflush(0)", fn: fflush, args: []uintptr{0}, want: 0},
	}
	onlyLeg, onlyCase := "", ""
	if len(os.Args) > 2 {
		onlyLeg = os.Args[2]
	}
	if len(os.Args) > 3 {
		onlyCase = os.Args[3]
	}
	done := make(chan bool)
	for _, onSystem := range []bool{false, true} {
		leg := "user"
		if onSystem {
			leg = "sys"
		}
		if onlyLeg != "" && onlyLeg != leg {
			continue
		}
		onSystem := onSystem
		go func() {
			for _, c := range checks {
				if strings.HasPrefix(c.name, onlyCase) {
					run(leg, c, onSystem)
				}
			}
			done <- true
		}()
		<-done
	}
	if failed {
		println("RESULT FAIL")
		os.Exit(1)
	}
	println("RESULT PASS")
}
