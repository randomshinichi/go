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
// Third leg, on top of user and sys: fwd / fwd-sys call runtime.sigfwd, the
// forwarder sigfwdgo uses to hand a signal to whatever handler was installed
// before Go's, with a real -O0 C handler that homes its register arguments into
// the caller's parameter area. Thus the case that matters is a handler whose
// homing stores run past sigfwd's frame. See helpers.c (home3, home8fwd) and
// runFwd; it reports sigfwd's frame size as measured from inside the call, and
// sigfwd-home8 additionally proves the saved LR survived a handler that homes
// all eight register words.
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

// sigfwd is runtime.sigfwd, the darwin/ppc32 signal forwarder, reached the way
// sigfwdgo reaches it: by name, through the same ABI. The runtime declaration is
// func sigfwd(fn uintptr, sig uint32, info *siginfo, ctx unsafe.Pointer) -- the
// three argument words are the C ABI's r3-r5, which is what the handler gets.
//
//go:linkname sigfwd runtime.sigfwd
func sigfwd(fn uintptr, sig uint32, info unsafe.Pointer, ctx unsafe.Pointer)

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

// home_report, as filled in by the C handlers (helpers.c). Index names here are
// the contract between the two sides.
const (
	fwdSig = 30 // SIGUSR1 on Darwin; sigfwd only passes it on

	// ctxSentinel is the ctx word the handlers home, and a corruption sentinel:
	// it has to be a value no real backchain could be, so that the handler's own
	// shape check (bit 4: the word at caller_SP+frame is above caller_SP) fires
	// when the homing has already written this value there. Leopard ppc32 stacks
	// are at 0x0183xxxx (25 MB) or near the top of memory, so a 57 KB word
	// qualifies; 0x0BADF00D, the first choice, does NOT -- it is above the stack
	// and reads as a plausible backchain. The exact check for whether the state
	// survived is the independent before/after comparison in runFwd.
	ctxSentinel = 0xDEAD

	repFrame     = 0  // sigfwd's frame size, measured from its own backchain
	repCallerSP  = 1  // sigfwd's SP at the call (the handler's backchain)
	repNWords    = 2  // register argument words the handler homed
	repBad       = 3  // 0 when the caller's frame honoured the ABI contract
	repLinkage0  = 4  // the word at caller_SP+frame: that frame's own backchain
	repArgWord   = 5  // 8 words: what the handler homed at caller_SP+24+4*i
	repSavedLR   = 13 // the word sigfwd saved as its return address
	repChecksum  = 14 // the handler's parameters: did the arguments arrive
	repReportLen = 16
)

func readU32(p uintptr) uint32 { return *(*uint32)(unsafe.Pointer(p)) }

func fwdReport(addr uintptr) [repReportLen]uint32 {
	var r [repReportLen]uint32
	for i := range r {
		r[i] = readU32(addr + uintptr(4*i))
	}
	return r
}

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

type fwdCheck struct {
	name   string
	fn     uintptr // the C handler
	ctx    uintptr // the ctx word sigfwd passes on, also a corruption sentinel
	nwords int     // register argument words that handler homes
}

// runFwd drives one sigfwd case and reports what the call itself measured:
//
//   - home_report, filled in by the C handler from inside the call (helpers.c),
//   - the linkage word zero of the frame that called sigfwd, read directly
//     before and after the call, which must not change.
//
// The caller is a closure, so its SP is stable across the call and gsp() -- an
// assembly NOFRAME function -- reports it; the word at that SP is the backchain
// its own prologue planted, and it is exactly the word a homing handler reaches
// when sigfwd's frame is only 32 bytes (sigfwd_SP+32 is that caller's SP+0).
func runFwd(leg string, c fwdCheck, onSystem bool, report uintptr) {
	println("ENTER", leg, c.name, "fn", c.fn, "sig", fwdSig)
	info := make([]byte, 128)
	infop := uintptr(unsafe.Pointer(&info[0]))
	var before, after uintptr
	call := func() {
		_, sp := gsp()
		before = uintptr(readU32(sp))
		sigfwd(c.fn, fwdSig, unsafe.Pointer(infop), unsafe.Pointer(c.ctx))
		_, sp = gsp()
		after = uintptr(readU32(sp))
	}
	if onSystem {
		systemstack(call)
	} else {
		call()
	}
	r := fwdReport(report)
	ok := r[repBad] == 0 &&
		int(r[repNWords]) == c.nwords &&
		r[repFrame] >= 56 &&
		r[repArgWord] == fwdSig &&
		uintptr(r[repArgWord+1]) == infop &&
		uintptr(r[repArgWord+2]) == c.ctx &&
		before == after
	verdict := "PASS"
	if !ok {
		verdict = "FAIL"
		failed = true
	}
	println(verdict, leg, c.name, "frame", int32(r[repFrame]), "bad", int32(r[repBad]),
		"nwords", int32(r[repNWords]), "callerSP", int32(r[repCallerSP]),
		"linkage0", int32(r[repLinkage0]), "savedLR", int32(r[repSavedLR]),
		"linkage0-pre", int32(before), "post", int32(after))
	println("    ", leg, c.name, "argwords homed at caller_SP+24..+52",
		int32(r[repArgWord]), int32(r[repArgWord+1]), int32(r[repArgWord+2]),
		int32(r[repArgWord+3]), int32(r[repArgWord+4]), int32(r[repArgWord+5]),
		int32(r[repArgWord+6]), int32(r[repArgWord+7]), "checksum", int32(r[repChecksum]))
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
	// The sigfwd witness: two -O0 C handlers that home their register arguments
	// into the caller's parameter area, plus the global they report through.
	home3 := sym(lib, "home3")
	home8fwd := sym(lib, "home8fwd")
	homeReport := sym(lib, "home_report")
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
	// sigfwd cases. home3 is the three-argument handler shape (its third homed
	// word is the one that leaves a 32-byte sigfwd frame); home8fwd homes all
	// eight register words, so it also reaches the saved LR at sigfwd_SP+40.
	fwdChecks := []fwdCheck{
		{name: "sigfwd-home3", fn: home3, ctx: ctxSentinel, nwords: 3},
		{name: "sigfwd-home8", fn: home8fwd, ctx: ctxSentinel, nwords: 8},
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
	// The forwarder, driven through runtime.sigfwd itself, from a goroutine stack
	// and from g0 (sigtrampgo forwards from the signal stack, which is a g0
	// context, so both shapes are worth running).
	for _, leg := range []string{"fwd", "fwd-sys"} {
		if onlyLeg != "" && onlyLeg != leg {
			continue
		}
		onSystem := leg == "fwd-sys"
		leg := leg
		go func() {
			for _, c := range fwdChecks {
				if strings.HasPrefix(c.name, onlyCase) {
					runFwd(leg, c, onSystem, homeReport)
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
