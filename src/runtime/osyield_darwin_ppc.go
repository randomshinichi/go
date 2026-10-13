// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package runtime

// osyield uses Leopard's swtch_pri Mach trap (trap -59), which does not take
// the pthread cancellation lock that makes usleep unsafe in signal handlers.
//
//go:nosplit
func osyield_no_g() {
	sched_yield()
}

//go:nosplit
func osyield() {
	sched_yield()
}

func sched_yield()
