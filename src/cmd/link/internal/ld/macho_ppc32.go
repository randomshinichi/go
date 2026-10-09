// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	imacho "cmd/internal/macho"
	"cmd/internal/sys"
)

// newPPC32ThreadLoad emits the measured Darwin PPC_THREAD_STATE layout.
// srr0 is state word 0 (load data word 2), not otool's printed final field.
func newPPC32ThreadLoad(entry uint32) *MachoLoad {
	ml := newMachoLoad(sys.ArchPPC, imacho.LC_UNIXTHREAD, 40+2)
	ml.data[0] = 1  // PPC_THREAD_STATE flavor
	ml.data[1] = 40 // PPC_THREAD_STATE_COUNT (160 state bytes / 4)
	ml.data[2] = entry
	return ml
}
