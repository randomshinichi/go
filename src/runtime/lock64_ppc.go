// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "internal/runtime/atomic"

// lock64Held reports whether the 64-bit atomic spinlock is held.
// See atomic.Lock64Held.
//
//go:nosplit
func lock64Held() bool { return atomic.Lock64Held() }
