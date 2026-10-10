// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && (ppc || ppc64)

package net

import "syscall"

// Leopard has no per-socket keepalive interval or count options. Ignore only
// ENOPROTOOPT so those unsupported KeepAliveConfig fields leave the kernel's
// system defaults in effect; other socket errors must still prevent the call.
func darwinKeepAliveOptionError(err error) error {
	if err == syscall.ENOPROTOOPT {
		return nil
	}
	return err
}
