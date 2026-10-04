// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// supportsCloseOnExec reports whether the platform supports the
// O_CLOEXEC flag.
// Mac OS X 10.5 (Leopard) does not: O_CLOEXEC arrived in OS X 10.7, and
// syscall.O_CLOEXEC is 0 on darwin/ppc64, so open is never asked for
// close-on-exec. With false, the opens in file_unix.go (openFileNolog,
// openDirNolog) and root_unix.go (newRoot) follow up with
// syscall.CloseOnExec, an explicit fcntl(F_SETFD, FD_CLOEXEC).
const supportsCloseOnExec = false
