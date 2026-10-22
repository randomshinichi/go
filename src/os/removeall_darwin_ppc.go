// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import "syscall"

// On darwin/ppc (Mac OS X 10.5 Leopard) RemoveAll is still the path-based
// removeall_noat.go, not the descriptor-based removeall_at.go. Leopard has no
// unlinkat/openat, and although most *at operations now have an anchored form
// (internal/syscall/unix/at_darwin_ppc.go), the descriptor-based walk needs the
// directory-reading path to be retired and re-validated first, so it is not
// selected here yet. Root.RemoveAll therefore refuses: the path algorithm's
// symlink race is exactly what Root must not have.
//
// SECURITY-RELEVANT DIVERGENCE from non-PowerPC Darwin targets: removeall_noat.go
// is the pre-Go-1.12 algorithm. It does not have the descriptor-based version's
// protection against symlink races (time-of-check/time-of-use): another process
// that can write to a directory being walked may swap an entry for a symlink
// between the check and the removal. Do not run RemoveAll on a tree an
// untrusted party can modify.
//
// Root.RemoveAll is the one caller that needs real descriptor semantics, so it
// fails with ENOTSUP instead of being approximated.

func removeAllFrom(parentFd sysfdType, base string) error {
	return &PathError{Op: "unlinkat", Path: base, Err: syscall.ENOTSUP}
}
