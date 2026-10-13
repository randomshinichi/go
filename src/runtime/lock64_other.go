// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc

package runtime

// lock64Held is false where 64-bit atomics have no shared spinlock that a
// signal handler could find held by the thread it interrupted.
func lock64Held() bool { return false }
