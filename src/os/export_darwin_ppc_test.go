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
