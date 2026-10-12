// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os

// RootFd returns the directory descriptor a Root holds.
//
// os.Root has no exported accessor for it, but the close-on-exec tests have to
// observe the flag on the Root handle itself: newRoot is one of the three places
// that falls back to syscall.CloseOnExec when supportsCloseOnExec is false.
var RootFd = func(r *Root) int { return r.root.fd }

// MarkInRoot marks f as if it had been opened inside a Root, and InRoot reads
// that mark back.
//
// They exist for one test: statat_darwin_ppc.go refuses to stat a directory
// entry for a File in a Root, and on this platform every Root operation is
// ENOTSUP, so an ordinary program cannot construct such a File and the refusal
// would otherwise stay dormant. Nothing here opens a Root or makes a confined
// operation work -- the File keeps the descriptor os.Open gave it, and the mark
// only reaches code that asks whether the File came from a Root.
var MarkInRoot = func(f *File, inRoot bool) { f.inRoot = inRoot }

// InRoot reports the mark MarkInRoot sets.
var InRoot = func(f *File) bool { return f.inRoot }
