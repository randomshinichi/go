// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// signalMNoSplit is signalM for the held-stripe hooks. runtime.signalM has a
// stack-splitting prologue, and a splitting prologue is an entry into newstack,
// which is how a pending cooperative preemption request runs the scheduler. The
// scheduler makes 64-bit atomics of its own, so entering it while a stripe is
// held can wait for the thread that holds it (goppc-weook). Here the whole
// chain is nosplit: this function, getpid and tgkill (assembly, NOSPLIT|NOFRAME
// in sys_linux_ppc.s, which issue the system call directly). The emitted code
// is audited in the commit message / evidence, not assumed.
//
//go:nosplit
func signalMNoSplit(mp *m, sig int) {
	tgkill(getpid(), int(mp.procid), sig)
}
