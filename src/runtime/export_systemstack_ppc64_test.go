// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/runtime/sys"
	"unsafe"
)

// SystemstackTailReturn exercises both the real switch and nested noswitch
// returns, with captured context read after the nested closure returns.
func SystemstackTailReturn() int {
	n := 1
	systemstack(func() {
		n += 2
		systemstack(func() { n += 4 })
		n += 8
	})
	return n
}

type SystemstackTailSnapshotResult struct {
	RetPC, TailPC       uintptr
	RetDelta, TailDelta int32
	Ret, Tail           [64]uintptr
	RetN, TailN         int
}

//go:noinline
func SystemstackTailSnapshot() (out SystemstackTailSnapshotResult) {
	systemstack(func() { systemstackTailSnapshot(&out) })
	return
}

//go:noinline
func systemstackTailSnapshot(out *SystemstackTailSnapshotResult) {
	entry := abi.FuncPCABI0(systemstack)
	// Locate the actual instructions, rather than assuming source offsets.
	// systemstack has two ordinary returns and one non-linking CTR branch.
	for pc := entry; findfunc(pc).entry() == entry; pc += 4 {
		switch *(*uint32)(unsafe.Pointer(pc)) {
		case 0x4e800020: // BLR
			if out.RetPC == 0 {
				out.RetPC = pc
			}
		case 0x4e800420: // BCCTR, LK clear
			if out.TailPC != 0 {
				throw("multiple systemstack tail branches")
			}
			out.TailPC = pc
		}
	}
	if out.RetPC == 0 || out.TailPC == 0 {
		return
	}
	f := findfunc(entry)
	out.RetDelta = funcspdelta(f, out.RetPC)
	out.TailDelta = funcspdelta(f, out.TailPC)

	// Model a SIGPROF interruption after LR and SP have been restored. The
	// ordinary RET and the tail branch must expose identical callers to the
	// unwinder, including its jump back from g0 to the goroutine stack. This
	// is deterministic interrupted-state emulation, not signal delivery.
	sp, lr := sys.GetCallerSP(), sys.GetCallerPC()
	var u unwinder
	u.initAt(out.RetPC, sp, lr, getg(), unwindTrap|unwindJumpStack|unwindSilentErrors)
	out.RetN = tracebackPCs(&u, 0, out.Ret[:])
	u.initAt(out.TailPC, sp, lr, getg(), unwindTrap|unwindJumpStack|unwindSilentErrors)
	out.TailN = tracebackPCs(&u, 0, out.Tail[:])
}
