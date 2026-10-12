// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package startlinetest

// PanicSlot is defined in panicslots_ppc.s for runtime_test.TestPPC32PanicBoundsSlots.
// It loads 1000+k into R(3+k) for every bounds save slot k = 0..6 and calls
// runtime.panicBounds or runtime.panicExtend with hand-written
// PCDATA_PanicBounds metadata naming the slots for this site. It lives here,
// not in runtime, because panicCheck1 turns a bounds failure whose caller is
// named "runtime.*" into a throw, which cannot be recovered.
func PanicSlot(site uint32)
