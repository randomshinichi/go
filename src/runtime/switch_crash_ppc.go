// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

//go:noescape
func switchToCrashStack0(fn func()) // implemented in PPC32 assembly
