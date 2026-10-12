// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// PPC32WBProtocolCheck exposes test_ppc.s's driver to runtime_test, which is
// where testing can be imported (see stackguardoffset_ppc_test.go's note).
var PPC32WBProtocolCheck = wbProtocolCheck

// PPC32WBBufEntries is the reservation buffer's capacity in entries.
const PPC32WBBufEntries = wbBufEntries
