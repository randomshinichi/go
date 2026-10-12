// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !(darwin && ppc64) && !(darwin && ppc)

package os

// supportsCloseOnExec reports whether the platform supports the
// O_CLOEXEC flag.
// On Darwin, the O_CLOEXEC flag was introduced in OS X 10.7 (Darwin 11.0.0).
// See https://support.apple.com/kb/HT1633.
// The two Darwin PowerPC ports run on Leopard (10.5), which predates it, so
// both take the false constant in sys_darwin_ppc64.go and sys_darwin_ppc.go
// even though every other Darwin port takes the true one here.
// On FreeBSD, the O_CLOEXEC flag was introduced in version 8.3.
const supportsCloseOnExec = true
