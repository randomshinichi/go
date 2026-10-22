// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix

// Exports for the external tests: the platform gate decides whether the
// anchored backend may run at all, so its decision and its parser are asserted
// directly rather than being inferred from the behaviour of ten call sites.

func AtCompatRelease(release string) (ok, parsed bool) { return atCompatRelease(release) }

func AtCompatEnabledForTest() bool { return atCompatEnabled() }
