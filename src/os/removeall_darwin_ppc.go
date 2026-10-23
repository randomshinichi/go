// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os

import "internal/syscall/unix"

// RemoveAll on darwin/ppc32 is the descriptor-based algorithm
// (removeall_at.go), the same one every other Darwin target uses, now that the
// anchored *at operations exist here (internal/syscall/unix/at_darwin_ppc.go).
// That is what makes Root.RemoveAll work: the descriptor walk removes entries
// relative to directories it holds open, so an entry swapped for a symlink while
// the tree is being walked cannot redirect it outside the root. The path-based
// algorithm (removeall_noat.go) that this port used instead has no such
// protection; it is kept only as the fallback below.
//
// The anchored backend exists only on the kernel it was measured against
// (kern.osrelease 9.8.x). On any other Darwin/ppc kernel -- an older 9.x or
// Tiger -- the descriptor walk cannot run, and RemoveAll falls back to the
// path-based algorithm rather than failing: that is the behaviour every
// Darwin/ppc build had before the backend existed, and it is what keeps
// t.TempDir teardown working there. Root.RemoveAll has no such fallback by
// design: it must be confined or refused, so on those kernels it still returns
// ENOTSUP (removeAllFrom is the descriptor walk; there is no path-based
// substitute inside a Root).
//
// SECURITY-RELEVANT DIVERGENCE, fallback only: on a kernel without the backend,
// RemoveAll keeps the pre-Go-1.12 algorithm's exposure to a symlink race (another
// process that can write to a directory being walked may swap an entry for a
// symlink between the check and the removal). Do not run RemoveAll on a tree an
// untrusted party can modify on such a kernel.
func confinedRemoveAll() bool { return unix.AnchoredAvailable() }
