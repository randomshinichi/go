// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompound_mutant

package syscall

import "unsafe"

// Test-only exports for the two broken copies of the native compound in
// at_compound_mutant_darwin_ppc.s. The file that declares them, and this one,
// are compiled only with the goppc_atcompound_mutant build tag, so a production
// build of this package does not contain the mutants at all.

func atCompoundSkipClearMutant(req *atCompoundReq)
func atCompoundClearFailMutant(req *atCompoundReq)

// atCompoundMutantRequest builds the same lstat64 request the production direct
// export builds, for one of the mutant routines.
func atCompoundMutantRequest(anchor int, path string) *atCompoundReq {
	req, err := atNewRequest(SYS_LSTAT64, anchor, path)
	if err != nil {
		panic(err)
	}
	req.out = new(Stat_t)
	req.w1 = uintptr(unsafe.Pointer(req.out))
	return req
}

// AtCompoundSkipClearMutantForTest runs the skip-clear mutant on the calling
// goroutine and reports the three status words. The caller's thread is left with
// a thread-private cwd installed, which is the defect being demonstrated.
func AtCompoundSkipClearMutantForTest(anchor int, path string) (enterErr, identErr, opErr int32) {
	req := atCompoundMutantRequest(anchor, path)
	atCompoundSkipClearMutant(req)
	return req.enterErr, req.identErr, req.opErr
}

// AtCompoundClearFailMutantForTest runs the clear-failure mutant on the calling
// goroutine. It does not return: the routine's own clear fails and the fail-stop
// branch calls raw exit(127).
func AtCompoundClearFailMutantForTest(anchor int, path string) {
	req := atCompoundMutantRequest(anchor, path)
	atCompoundClearFailMutant(req)
}
