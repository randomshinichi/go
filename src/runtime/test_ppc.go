// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// wbProtocolCheck drives the compiler-private gcWriteBarrier8 protocol from
// assembly, the way generated code does. It is in test_ppc.s; the assertion is
// TestPPC32WriteBarrierProtocol.
func wbProtocolCheck(iters uint32) (bad, flushes uint32)
