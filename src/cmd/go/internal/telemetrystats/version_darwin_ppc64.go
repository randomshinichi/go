// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !cmd_go_bootstrap && darwin && ppc64

// DISABLED FEATURE, RECORDED HERE ON PURPOSE (darwin/ppc64 port).
//
// The vendored cmd/vendor/golang.org/x/sys/unix carries type and syscall tables
// (ztypes/zsyscall/zerrors/zsysnum) only for darwin/amd64 and darwin/arm64, so
// unix.Uname cannot be compiled for darwin/ppc64; that single call is the only
// reason the ORDINARY go command could not be built for this port
// (`go list -deps cmd/go` shows cmd/go/internal/telemetrystats as the sole
// importer of golang.org/x/sys/unix in the whole ordinary closure, and
// src/cmd/go/main.go:218 telemetrystats.Increment() as its only caller).
//
// This file applies upstream's own "this platform cannot report the OS version"
// path, exactly as version_other.go spells it for the non-unix platforms.
// Consequence, stated plainly: the host-OS-version telemetry counters are not
// reported on darwin/ppc64. Nothing else is disabled - package loading, the
// build system, module fetching and the TLS/x509 path are the ordinary ones.
// Do not read this as full support; see
// out/stage2-20261003/D18-native-build/REPORT.md (disabled-feature table).
package telemetrystats

import "cmd/internal/telemetry/counter"

func incrementVersionCounters() {
	counter.Inc("go/platform:version-not-supported")
}