// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && !ppc64

package runtime

// darwin/ppc64 yields with sched_yield instead; see sys_darwin_ppc64.go.

//go:nosplit
func osyield_no_g() {
	usleep_no_g(1)
}

//go:nosplit
func osyield() {
	usleep(1)
}
