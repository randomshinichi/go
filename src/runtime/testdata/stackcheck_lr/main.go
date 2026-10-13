// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program is the SIGPROF witness for runtime·stackcheck on ppc32 (goppc-1ye4,
// criterion 4). It is NOT built by any go test: it needs the sigprof hook in
// overlay/ (see README below) and only has meaning on darwin/ppc.
//
// Background. A split-stack function on ppc32 does MFLR R0; STW R0,8(R1); then
// BL runtime·stackcheck BEFORE it has a frame. BL overwrites LR with the address of
// the next instruction, so until the function's later MFLR/LWZ/STWU, LR == PC at
// the instruction right after the BL. A profiling signal that lands there makes
// the traceback see pc == lr && sp == fp and throw "traceback stuck".
// runtime·stackcheck now returns with LR reloaded from the 8(R1) slot, so LR is
// the caller's address there and the equality cannot occur.
//
// The hook (overlay/proc.go, overlay/stackcheck_lr_1ye4c.go) records, for every
// SIGPROF sample, the interrupted PC and LR. This program loops on a function whose
// prologue runs the stack check, then reports how many samples were taken, how many
// had pc == lr, and a histogram of where in the function the samples fell.
//
// Usage: stackcheck_lr MODE SECONDS HZ [skip]
//
//	MODE asm   loop on bigFrame, an assembly function with a 2432-byte frame; the
//	           stack is large enough, so stackcheck takes the "no growth" return.
//	MODE go    loop on a compiled Go function with a large zeroed frame (same
//	           prologue, longer body).
//	MODE recurse  every iteration starts a fresh goroutine that recurses 300 deep in
//	           compiled Go (about 70 KB of frames), so stackcheck's "grow" return and
//	           morestack/newstack run about six times per goroutine. Run it with HZ 0:
//	           see the note on SIGPROF below.
//	skip       do not run the traceback for samples with pc == lr, so an unfixed
//	           build counts them instead of throwing.
//
// Known limits, recorded where the next reader will look:
//   - SIGPROF livelock. On ppc32 the 64-bit atomics share one spinlock
//     (internal/runtime/atomic/atomic_ppc.go lock64), and the profiling handler's
//     profBuf.write takes it. A signal landing while the interrupted code holds it
//     spins forever on the same thread. Long asm/go runs therefore sometimes never
//     exit, and recurse mode (many 64-bit atomics) stalls almost at once with HZ > 0.
//   - Sampling does not reach the window. The hazardous PC is one instruction long
//     and the timer interrupt is attributed to instructions that wait at completion;
//     an instruction after a correctly predicted BLR returns almost never is one. So
//     the unfixed build is demonstrated by injecting the signal at that PC from gdb
//     (see gdb_window.cmds), not by waiting for a timer sample.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"time"
)

// bigFrame is in frame_ppc.s.
func bigFrame()

var sink int

// bigFrame16 amortizes the loop and jitter overhead so that a larger share of
// the samples falls inside bigFrame's prologue.
//
//go:noinline
func bigFrame16() {
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
	bigFrame()
}

//go:noinline
func goBig(i int) int {
	var a [600]int32
	a[i&511] = int32(i)
	return int(a[(i+1)&511])
}

// recurse grows the stack through compiler-emitted split-stack prologues.
//
//go:noinline
func recurse(n int) int {
	var pad [48]int32
	pad[n&31] = int32(n)
	if n == 0 {
		return int(pad[0])
	}
	return recurse(n-1) + int(pad[(n+1)&31])
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("usage: stackcheck_lr asm|go|recurse SECONDS HZ [skip]")
		os.Exit(2)
	}
	mode := os.Args[1]
	secs, _ := strconv.Atoi(os.Args[2])
	hz, _ := strconv.Atoi(os.Args[3])
	runtime.Sigprof1ye4cSkipEq = len(os.Args) > 4 && os.Args[4] == "skip"

	runtime.SetCPUProfileRate(hz)
	deadline := time.Now().Add(time.Duration(secs) * time.Second)
	rng := uint32(2463534242)
	calls := 0
	next := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if time.Now().After(next) {
			n, eq := runtime.Sigprof1ye4cCounts()
			fmt.Printf("progress calls=%d samples=%d eq=%d\n", calls, n, eq)
			next = next.Add(10 * time.Second)
		}
		for k := 0; k < 20000; k++ {
			// xorshift jitter so the loop period does not alias the timer tick.
			rng ^= rng << 13
			rng ^= rng >> 17
			rng ^= rng << 5
			for j := uint32(0); j < rng&15; j++ {
				sink += int(j)
			}
			switch mode {
			case "asm":
				bigFrame16()
				calls += 15
			case "go":
				sink += goBig(k)
			case "recurse":
				done := make(chan struct{})
				go func() { sink += recurse(300); close(done) }()
				<-done
			}
			calls++
		}
	}
	fmt.Println("loop done")
	runtime.SetCPUProfileRate(0)
	fmt.Println("profile off")

	n, eq := runtime.Sigprof1ye4cCounts()
	fmt.Printf("mode=%s secs=%d hz=%d calls=%d skip=%v\n", mode, secs, hz, calls, runtime.Sigprof1ye4cSkipEq)
	fmt.Printf("morestack-newstack-calls=%d\n", runtime.Newstack1ye4cCount())
	fmt.Printf("sigprof-samples=%d  pc==lr=%d\n", n, eq)
	hist := map[string]int{}
	pcs0, lrs0 := runtime.Sigprof1ye4cSamples()
	for i := 0; i < len(pcs0) && i < 3; i++ {
		fmt.Printf("first-sample[%d] pc=%#x lr=%#x\n", i, pcs0[i], lrs0[i])
	}
	// The instruction right after "BL runtime.stackcheck" is bigFrame+0x14 (checked against go tool objdump in the report). That is the
	// only PC where the function has no frame yet and LR would be the stale return
	// address. pc == lr elsewhere is the ordinary instruction after a call in a
	// function that already has a frame (sp != fp), which the traceback accepts.
	var atWindow, atWindowEq, otherEq int
	pcs, lrs := runtime.Sigprof1ye4cSamples()
	for i := range pcs {
		pc, lr := uintptr(pcs[i]), uintptr(lrs[i])
		f := runtime.FuncForPC(pc)
		name, off := "?", uintptr(0)
		if f != nil {
			name, off = f.Name(), pc-f.Entry()
		}
		tag := ""
		if pc == lr {
			tag = "  PC==LR"
		}
		if off == 0x14 && name == "main.bigFrame" {
			atWindow++
			if pc == lr {
				atWindowEq++
			}
		} else if pc == lr {
			otherEq++
		}
		hist[fmt.Sprintf("%s+%#x%s", name, off, tag)]++
	}
	fmt.Printf("window-pc(after BL stackcheck)-samples=%d  of which pc==lr=%d   pc==lr-elsewhere=%d\n", atWindow, atWindowEq, otherEq)
	keys := make([]string, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return hist[keys[i]] > hist[keys[j]] })
	for i, k := range keys {
		if i >= 40 {
			break
		}
		fmt.Printf("%6d  %s\n", hist[k], k)
	}
}
