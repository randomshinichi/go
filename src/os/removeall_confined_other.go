// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(darwin && ppc)

package os

// removeAllPathFallback reports "not handled" everywhere except darwin/ppc32,
// where the descriptor-based algorithm depends on a kernel that may not have the
// anchored *at operations; see removeall_darwin_ppc.go.
func removeAllPathFallback(path string) (handled bool, err error) { return false, nil }
