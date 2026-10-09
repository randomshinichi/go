// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ppc encodes 32-bit PowerPC (GOARCH=ppc) instructions for the Go
// toolchain.
//
// Scope: the scalar instruction subset the ppc32 G3 baseline needs, as agreed in
// out/ppc-mac/ppc32-design/DESIGN.md §5–§7 — integer arithmetic including the
// carry forms the 64-bit decomposition uses, logical, shift/rotate and bit-count
// forms, compares, integer load/store (word, half, byte, byte-reversed), the
// lwarx/stwcx. atomic loop pair, sync/isync, branches to relative/absolute,
// LR/CTR/CR, and hardware single/double floating point.
//
// Two rules from that contract are enforced here rather than left to review:
//
//   - Design decision D-1: the baseline never uses a 64-bit-only form, even on a
//     970, because whether the kernel preserves upper GPR halves for a 32-bit
//     Darwin process across signals is unmeasured (Q-6). 64-bit mnemonics are
//     refused by name, not merely absent.
//   - The G3 (750) must be able to execute every word this package emits. Forms
//     illegal or unestablished on 750/74xx — isel, the POWER5 round-to-integer
//     family, VMX/AltiVec, VSX, lwsync, the stfiwx/fsqrt family, eciwx — are
//     refused with a recorded reason. An encoder that emits a word the G3 cannot
//     execute is the defect this port exists to avoid.
//
// What this package is not: it is not a compiler backend, a linker, a runtime or
// a relocation engine, and it does not yet define an obj.LinkArch for ppc. Branch
// operands are byte displacements relative to the branch instruction, already
// resolved; symbolic targets, fixups and relocation records belong to the linker
// lane (DESIGN §13 brief 3.4).
package ppc

import "cmd/internal/obj"

const (
	NREG  = 32 // number of general-purpose registers
	NFREG = 32 // number of floating-point registers
)

// General-purpose registers, by architectural number. R0 reads as the literal
// zero in the RA position of addi/lwz-style forms, R1 is the stack pointer, R30
// holds g and R31 is the assembler temporary under the Go-internal convention
// (DESIGN §4). On Darwin ppc32, R2 and R13 are conservatively reserved but
// neither is a thread pointer in the measured ABI probe. The Go g register is
// R30, so core pure-Go runtime access to g needs no separate TLS mechanism.
const (
	R0 uint8 = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
	R16
	R17
	R18
	R19
	R20
	R21
	R22
	R23
	R24
	R25
	R26
	R27
	R28
	R29
	R30
	R31
)

// Floating-point registers, by architectural number. This is a separate
// numbering space from the GPRs: F1 and R1 are different registers.
const (
	F0 uint8 = iota
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
	F13
	F14
	F15
	F16
	F17
	F18
	F19
	F20
	F21
	F22
	F23
	F24
	F25
	F26
	F27
	F28
	F29
	F30
	F31
)

const (
	SP     = R1
	g      = R30
	REGTMP = R31
)

// Inst is a ppc32 opcode. The values live in obj's PPC opcode subspace so the
// compiler and assembler can carry them in obj.Prog.As.
type Inst = obj.As

// The G3 scalar baseline instruction set. This block is in the same order as the
// forms table in asm.go, and init panics if the two get out of step.
const (
	AXXX Inst = obj.ABasePPC + obj.A_ARCHSPECIFIC + iota

	AADD    // add
	AADDC   // addc
	AADDE   // adde
	AADDI   // addi
	AADDIC  // addic
	AADDIC_ // addic.
	AADDIS  // addis
	ASUBF   // subf
	ASUBFC  // subfc
	ASUBFE  // subfe
	ASUBFIC // subfic
	AMULLW  // mullw
	AMULHW  // mulhw
	AMULHWU // mulhwu
	ADIVW   // divw
	ADIVWU  // divwu
	ANEG    // neg
	AAND    // and
	AANDC   // andc
	AANDI_  // andi.
	AANDIS_ // andis.
	AOR     // or
	AORC    // orc
	AORI    // ori
	AORIS   // oris
	AXOR    // xor
	AXORI   // xori
	AXORIS  // xoris
	ANAND   // nand
	ANOR    // nor
	AEQV    // eqv
	ASLW    // slw
	ASRW    // srw
	ASRAW   // sraw
	ASRAWI  // srawi
	ARLWINM // rlwinm
	ARLWIMI // rlwimi
	ARLWNM  // rlwnm
	AROTLW  // rotlw
	ACNTLZW // cntlzw
	AEXTSB  // extsb
	AEXTSH  // extsh
	ACMPW   // cmpw
	ACMPLW  // cmplw
	ACMPWI  // cmpwi
	ACMPLWI // cmplwi
	ALWZ    // lwz
	ALWZU   // lwzu
	ALWZX   // lwzx
	ALWZUX  // lwzux
	ALBZ    // lbz
	ALBZU   // lbzu
	ALBZX   // lbzx
	ALBZUX  // lbzux
	ALHZ    // lhz
	ALHZU   // lhzu
	ALHZX   // lhzx
	ALHZUX  // lhzux
	ALHA    // lha
	ALHAU   // lhau
	ALHAX   // lhax
	ALHAUX  // lhaux
	ALMW    // lmw
	ALWARX  // lwarx
	ALWBRX  // lwbrx
	ALHBRX  // lhbrx
	ASTW    // stw
	ASTWU   // stwu
	ASTWX   // stwx
	ASTWUX  // stwux
	ASTB    // stb
	ASTBU   // stbu
	ASTBX   // stbx
	ASTBUX  // stbux
	ASTH    // sth
	ASTHU   // sthu
	ASTHX   // sthx
	ASTHUX  // sthux
	ASTMW   // stmw
	ASTWCX_ // stwcx.
	ASTWBRX // stwbrx
	ASTHBRX // sthbrx
	AB      // b
	ABL     // bl
	ABC     // bc
	ABLR    // blr
	ABLRL   // blrl
	ABCTR   // bctr
	ABCTRL  // bctrl
	ABCLR   // bclr
	ABCLRL  // bclrl
	ABCCTR  // bcctr
	ABCCTRL // bcctrl
	AMFCR   // mfcr
	AMFLR   // mflr
	AMFCTR  // mfctr
	AMTCRF  // mtcrf
	AMTLR   // mtlr
	AMTCTR  // mtctr
	ALFS    // lfs
	ALFSU   // lfsu
	ALFSX   // lfsx
	ALFSUX  // lfsux
	ALFD    // lfd
	ALFDU   // lfdu
	ALFDX   // lfdx
	ALFDUX  // lfdux
	ASTFS   // stfs
	ASTFSU  // stfsu
	ASTFSX  // stfsx
	ASTFSUX // stfsux
	ASTFD   // stfd
	ASTFDU  // stfdu
	ASTFDX  // stfdx
	ASTFDUX // stfdux
	AFADD   // fadd
	AFADDS  // fadds
	AFSUB   // fsub
	AFSUBS  // fsubs
	AFMUL   // fmul
	AFMULS  // fmuls
	AFDIV   // fdiv
	AFDIVS  // fdivs
	AFNEG   // fneg
	AFABS   // fabs
	AFMR    // fmr
	AFRSP   // frsp
	AFCTIWZ // fctiwz
	AFCMPU  // fcmpu
	AFCMPO  // fcmpo
	ASYNC   // sync
	AISYNC  // isync
	ASC     // sc

	ALAST
)
