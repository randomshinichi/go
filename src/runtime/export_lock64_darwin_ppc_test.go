// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// signalMNoSplit is signalM for the held-stripe hooks; see the linux file for
// why it must have no splitting callee. On darwin the chain is this function,
// pthread_kill (//go:nosplit), libcCall (//go:nosplit) and asmcgocall
// (NOSPLIT assembly), which calls pthread_kill_trampoline (NOSPLIT assembly)
// on the g0 stack. This was inspected in built darwin/ppc code only; it has not
// been executed.
//
//go:nosplit
func signalMNoSplit(mp *m, sig int) {
	pthread_kill(pthread(mp.procid), uint32(sig))
}
