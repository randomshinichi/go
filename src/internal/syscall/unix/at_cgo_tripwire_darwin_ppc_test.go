// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build cgo && darwin && ppc

package unix_test

import "testing"

// TestCgoEnabledTripwireDarwinPPC fails the moment cgo is enabled for
// darwin/ppc, which the port disables (cmd/dist/build.go). The private-cwd
// compound rests on two properties that a cgo build removes:
//
//   - no foreign callbacks run on Go threads, so nothing enters application code
//     while a thread carries the private cwd installed by the native scope;
//   - no foreign handler performs relative-path I/O on the thread that holds it.
//
// If this fires, the compound design has to be revisited before enablement
// (docs/leopard-at-compat-design.md, "Signals/cgo"), not merely re-tested.
func TestCgoEnabledTripwireDarwinPPC(t *testing.T) {
	t.Fatal("cgo is enabled for darwin/ppc: the private-cwd compound's justification excludes foreign callbacks and handlers; revisit docs/leopard-at-compat-design.md before enabling cgo on this platform")
}
