// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

import "unsafe"

//go:noescape
func asmcgocall_no_g(fn, arg unsafe.Pointer)

// getfp is only used by profiling and traceback code that can tolerate an
// unavailable frame pointer, as on the existing ppc64 runtime.
func getfp() uintptr { return 0 }
