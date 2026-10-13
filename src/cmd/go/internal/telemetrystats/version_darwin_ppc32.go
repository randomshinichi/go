// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cmd_go_bootstrap && darwin && ppc

// DISABLED FEATURE, RECORDED HERE ON PURPOSE (darwin/ppc32 port).
//
// This file is the ppc32 counterpart of version_darwin_ppc64.go, and exists for
// the identical reason: the vendored cmd/vendor/golang.org/x/sys/unix carries
// its type and syscall tables (ztypes/zsyscall/zerrors/zsysnum) only for
// darwin/amd64 and darwin/arm64, so unix.Uname cannot be compiled for any PPC
// Darwin port. version_unix.go is therefore excluded on both PPC GOARCHes and
// this file serves upstream's "this platform cannot report the OS version" path,
// exactly as version_other.go spells it for the non-unix platforms.
//
// Until this file existed, darwin/ppc32 fell through to version_unix.go (the
// exclusion read `!(darwin && ppc64)`, which catches ppc64 only) and the failure
// was not a runtime surprise but a compile error:
//
//	# cmd/vendor/golang.org/x/sys/unix
//	src/cmd/vendor/golang.org/x/sys/unix/syscall_unix.go:220:38: undefined: _Socklen
//	src/cmd/vendor/golang.org/x/sys/unix/fcntl_darwin.go:15:42: too many errors
//
// and with it the whole ordinary `go` command, which imports this package from
// src/cmd/go/main.go:218 (telemetrystats.Increment is its only caller; this
// package is the sole importer of golang.org/x/sys/unix in the ordinary cmd/go
// closure). Patching cmd/vendor was rejected on purpose: the vendored copy is
// upstream-owned, and a darwin/ppc table set appended to it would be a large,
// permanently-diverging fork of third-party code to work around one routing
// mistake.
//
// Consequence, stated plainly: the host-OS-version telemetry counters are not
// reported on darwin/ppc32 - the `go/platform/host/darwin/major-version:*` and
// `go/platform/host/darwin/version:*` counters that version_unix.go would emit
// from uname(3) are replaced by the single counter
// `go/platform:version-not-supported`. Nothing else is disabled: package
// loading, the build system, module fetching and the TLS/x509 path are the
// ordinary ones. Do not read this as full support for the telemetry feature.
package telemetrystats

import "cmd/internal/telemetry/counter"

func incrementVersionCounters() {
	counter.Inc("go/platform:version-not-supported")
}
