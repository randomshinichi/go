// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(darwin && ppc)

package os

// confinedRemoveAll reports whether RemoveAll may use the descriptor-based
// algorithm. Everywhere except darwin/ppc32 the answer is fixed at build time;
// see removeall_darwin_ppc.go for the one platform where it depends on the
// kernel.
func confinedRemoveAll() bool { return true }
