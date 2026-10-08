// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj/ppc"
)

// Init registers the ppc32 back end with the compiler driver. It is modelled on
// cmd/compile/internal/arm: a 32-bit RISC target whose 64-bit operations are
// decomposed into 32-bit ones, with no register calling convention and no
// hardware g register.
func Init(arch *ssagen.ArchInfo) {
	arch.LinkArch = &ppc.Linkppc
	arch.REGSP = ppc.REGSP
	arch.MAXWIDTH = (1 << 32) - 1

	arch.ZeroRange = zerorange
	arch.Ginsnop = ginsnop
	// Pad locals/outgoing arguments before stack zeroing computes its offsets.
	// A zero-sized leaf remains frameless; obj/ppc supplies a minimum frame
	// for non-leaves even when they have no outgoing arguments.
	arch.PadFrame = func(frame int64) int64 {
		if frame == 0 {
			return 0
		}
		fixed := arch.LinkArch.Arch.FixedFrameSize
		return (frame+fixed+15)&^15 - fixed
	}

	arch.SSAMarkMoves = func(s *ssagen.State, b *ssa.Block) {}
	arch.SSAGenValue = ssaGenValue
	arch.SSAGenBlock = ssaGenBlock
}
