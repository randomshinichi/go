// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import "cmd/internal/obj"

// Register numbers as cmd/internal/obj sees them. The SSA register table in
// cmd/compile/internal/ssa/opGen.go stores one of these in each Register's
// objNum, so ppc/ssa.go writes these numbers into Progs. They are a different
// numbering from the architectural numbers R0..R31 / F0..F31 that Encode
// consumes: architecturally the GPR and FPR spaces overlap (F0 is register 0 of
// its class), but a Prog needs one number per register, so the two classes are
// laid out consecutively here.
const (
	REG_R0 = obj.RBasePPC + iota
	REG_R1
	REG_R2
	REG_R3
	REG_R4
	REG_R5
	REG_R6
	REG_R7
	REG_R8
	REG_R9
	REG_R10
	REG_R11
	REG_R12
	REG_R13
	REG_R14
	REG_R15
	REG_R16
	REG_R17
	REG_R18
	REG_R19
	REG_R20
	REG_R21
	REG_R22
	REG_R23
	REG_R24
	REG_R25
	REG_R26
	REG_R27
	REG_R28
	REG_R29
	REG_R30
	REG_R31
	REG_F0
	REG_F1
	REG_F2
	REG_F3
	REG_F4
	REG_F5
	REG_F6
	REG_F7
	REG_F8
	REG_F9
	REG_F10
	REG_F11
	REG_F12
	REG_F13
	REG_F14
	REG_F15
	REG_F16
	REG_F17
	REG_F18
	REG_F19
	REG_F20
	REG_F21
	REG_F22
	REG_F23
	REG_F24
	REG_F25
	REG_F26
	REG_F27
	REG_F28
	REG_F29
	REG_F30
	REG_F31
)

// The architectural names the rest of the compiler uses for the three
// special-purpose registers of the Go-internal convention (DESIGN.md 4).
const (
	REGZERO = REG_R0  // reads as the literal zero in the RA position
	REGSP   = REG_R1  // the stack pointer
	REGG    = REG_R30 // holds g
)

// ppcDWARFRegisters maps obj register numbers to DWARF register numbers.
//
// The numbering below is the PowerPC DWARF convention this fork already uses for
// ppc64 (cmd/internal/obj/ppc64/a.out.go, PPC64DWARFRegisters): GPRs 0-31, FPRs
// 32-63, LR 65, CTR 66 and XER 76. Whether Apple's toolchain numbers registers
// the same way for a 32-bit Darwin process is not established here — the ppc32
// DWARF and Mach-O surface is the open Mach-O lane (DESIGN.md §11 Q-4) — so this
// map is a debug-info label only, it is not an ABI claim, and it must be
// re-derived from an Apple-built object before debug info is relied on.
var ppcDWARFRegisters = map[int16]int16{}

func init() {
	f := func(from, to, base int16) {
		for r := from; r <= to; r++ {
			ppcDWARFRegisters[r] = r - from + base
		}
	}
	f(REG_R0, REG_R31, 0)
	f(REG_F0, REG_F31, 32)
}
