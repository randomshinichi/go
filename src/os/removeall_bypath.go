// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (js && wasm) || plan9 || (darwin && ppc64)

package os

// removeAll is the path-based algorithm for the platforms that have no
// descriptor-based one (removeall_at.go). darwin/ppc32 is deliberately absent:
// there the path-based code is the fallback behind a runtime check, so it is
// named removeAllByPath and removeAll comes from removeall_at.go.
func removeAll(path string) error { return removeAllByPath(path) }
