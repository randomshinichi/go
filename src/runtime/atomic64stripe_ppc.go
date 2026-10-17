// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "internal/runtime/atomic"

// atomic64AnyStripeHeld reports whether any 64-bit atomic stripe is held.
// See atomic.Atomic64AnyStripeHeld for why the answer is "any" and why it is sound.
//
//go:nosplit
func atomic64AnyStripeHeld() bool { return atomic.Atomic64AnyStripeHeld() }
