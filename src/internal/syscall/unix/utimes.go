// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !wasip1 && !(darwin && ppc)

// darwin/ppc is excluded because Leopard's utimensat needs dispatch of its own
// (at_darwin_ppc.go): a real directory descriptor with AT_SYMLINK_NOFOLLOW is
// served by the setattrlist(FSOPT_NOFOLLOW) compound, which cannot be expressed
// as a linkname to syscall.utimensat. That file still reaches the path form
// through the same linkname, so there is one implementation of the path cases.

package unix

import (
	"syscall"
	_ "unsafe" // for //go:linkname
)

//go:linkname Utimensat syscall.utimensat
func Utimensat(dirfd int, path string, times *[2]syscall.Timespec, flag int) error
