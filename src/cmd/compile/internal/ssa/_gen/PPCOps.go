// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "strings"

// Notes:
//  - The ppc32 baseline has 32-bit general registers. A 64-bit Go value is split
//    integer into a (hi, lo) pair of 32-bit registers by the generic dec64 decomposition
//    before lowering runs (DESIGN.md §5 strategy D-3), so this table registers
//    the 32-bit residue that decomposition leaves behind, not 64-bit forms.
//  - XER[CA] is the carry bit. It is not a general-purpose register and it cannot
//    be read or written by an ordinary instruction, so it is modelled the way ARM
//    models CPSR.C: as a value of SSA type Flags, produced by an op that returns
//    (result, Flags) and consumed as an explicit argument by the op that adds it
//    in. Every op that writes CA is marked clobberFlags, so the flag allocator
//    treats a register-allocated carry as destroyed by an op that does not
//    produce it — that is what stops a carry-clobbering instruction from being
//    scheduled between the addc and the adde of one 64-bit addition.
//    The CA-clobbering forms of this table are addc, adde, subfc, subfe, sraw and
//    srawi; only addc/subfc (through ADDC/SUBC/ADDEC/SUBEC) also *produce* a
//    Flags value.

// Suffixes encode the bit width of an instruction.
// W (word)      = 32 bit
// H (half word) = 16 bit
// HU            = 16 bit unsigned
// B (byte)      = 8 bit
// BU            = 8 bit unsigned

// regNamesPPC lists the registers the register allocator may use. r2 and r13 are
// deliberately absent: their roles on Darwin ppc32 are open (DESIGN.md §11 Q-2),
// and a backend must not allocate a register whose meaning it has not measured.
// R0 is present as an architectural zero, not as a destination.
var regNamesPPC = []string{
	"R0", // architectural zero
	"SP", // R1
	"SB",
	"R3",
	"R4",
	"R5",
	"R6",
	"R7",
	"R8",
	"R9",
	"R10",
	"R11",
	"R12",
	"R14",
	"R15",
	"R16",
	"R17",
	"R18",
	"R19",
	"R20",
	"R21",
	"R22",
	"R23",
	"R24",
	"R25",
	"R26",
	"R27",
	"R28",
	"R29",
	"g",   // R30
	"R31", // assembler/linker temporary
	"F0",
	"F1",
	"F2",
	"F3",
	"F4",
	"F5",
	"F6",
	"F7",
	"F8",
	"F9",
	"F10",
	"F11",
	"F12",
	"F13",
	"F14",
	"F15",
	"F16",
	"F17",
	"F18",
	"F19",
	"F20",
	"F21",
	"F22",
	"F23",
	"F24",
	"F25",
	"F26",
	"F27",
	"F28",
	"F29",
	"F30",
	"F31",
}

func init() {
	// Make map from reg names to reg integers.
	if len(regNamesPPC) > 64 {
		panic("too many registers")
	}
	num := map[string]int{}
	for i, name := range regNamesPPC {
		num[name] = i
	}
	buildReg := func(s string) regMask {
		m := regMask(0)
		for _, r := range strings.Split(s, " ") {
			if n, ok := num[r]; ok {
				m |= regMask(1) << uint(n)
				continue
			}
			panic("register " + r + " not found")
		}
		return m
	}

	// Common individual register masks. R0, R31 and SB are not allocatable, and
	// g is kept out of the general-purpose set so that it is only clobbered
	// deliberately.
	var (
		gp     = buildReg("R3 R4 R5 R6 R7 R8 R9 R10 R11 R12 R14 R15 R16 R17 R18 R19 R20 R21 R22 R23 R24 R25 R26 R27 R28 R29")
		gpg    = gp | buildReg("g")
		sp     = buildReg("SP")
		sb     = buildReg("SB")
		gpspsb = gp | sp | sb
		fp     = buildReg("F0 F1 F2 F3 F4 F5 F6 F7 F8 F9 F10 F11 F12 F13 F14 F15 F16 F17 F18 F19 F20 F21 F22 F23 F24 F25 F26 F27 F28 F29 F30 F31")
	)
	// Common regInfo. The carry / flags operand of a 3-argument op is listed
	// only when it occupies a register, exactly as ARM does for CPSR: an entry
	// for a Flags argument would be meaningless because there is no register
	// mask for XER.
	var (
		gp01           = regInfo{inputs: nil, outputs: []regMask{gp}}
		gp11           = regInfo{inputs: []regMask{gpg}, outputs: []regMask{gp}}
		gp21           = regInfo{inputs: []regMask{gpg, gpg}, outputs: []regMask{gp}}
		gp21carry      = regInfo{inputs: []regMask{gpg, gpg}, outputs: []regMask{gp, 0}}
		gp2flags1      = regInfo{inputs: []regMask{gp, gp}, outputs: []regMask{gp}}
		gp2flags1carry = regInfo{inputs: []regMask{gp, gp}, outputs: []regMask{gp, 0}}
		gp1cr          = regInfo{inputs: []regMask{gpg}}
		gp2cr          = regInfo{inputs: []regMask{gpg, gpg}}
		callerSave     = regMask(gp | fp | gpg)
		gp22           = regInfo{inputs: []regMask{gpg, gpg}, outputs: []regMask{gp, gp}}
		gpload         = regInfo{inputs: []regMask{gpspsb, 0}, outputs: []regMask{gp}}
		gpstore        = regInfo{inputs: []regMask{gpspsb, gp, 0}}
		fp01           = regInfo{outputs: []regMask{fp}}
		fp11           = regInfo{inputs: []regMask{fp}, outputs: []regMask{fp}}
		fpload         = regInfo{inputs: []regMask{gpspsb, 0}, outputs: []regMask{fp}}
		fpstore        = regInfo{inputs: []regMask{gpspsb, fp, 0}}
	)

	ops := []opData{
		// Doubles stay in one FPR. All double memory transfers go through
		// aligned private scratch and two words: Go slots may be 4 mod 8.
		{name: "FMOVSconst", reg: fp01, aux: "Float64", typ: "Float32"},
		{name: "FMOVDconst", reg: fp01, aux: "Float64", typ: "Float64"},
		{name: "FMR", argLength: 1, reg: fp11, asm: "FMR"},
		{name: "FRSP", argLength: 1, reg: fp11, asm: "FRSP"},
		{name: "FMOVSload", argLength: 2, reg: fpload, aux: "SymOff", typ: "Float32", faultOnNilArg0: true, symEffect: "Read"},
		{name: "FMOVDload", argLength: 2, reg: fpload, aux: "SymOff", typ: "Float64", faultOnNilArg0: true, symEffect: "Read"},
		{name: "FMOVSstore", argLength: 3, reg: fpstore, aux: "SymOff", typ: "Mem", faultOnNilArg0: true, symEffect: "Write"},
		{name: "FMOVDstore", argLength: 3, reg: fpstore, aux: "SymOff", typ: "Mem", faultOnNilArg0: true, symEffect: "Write"},
		// 32-bit arithmetic. ADD/SUB are the plain forms the residue needs;
		// ADDC/SUBC/adde/subfe are the carry pair.
		{name: "ADD", argLength: 2, reg: gp21, asm: "ADD", commutative: true},                            // arg0 + arg1
		{name: "ADDC", argLength: 2, reg: gp21carry, asm: "ADDC", commutative: true, clobberFlags: true}, // arg0 + arg1, XER[CA] out (addc)
		{name: "ADDE", argLength: 3, reg: gp2flags1, asm: "ADDE", clobberFlags: true},                    // arg0 + arg1 + XER[CA] (adde); arg2=flags
		{name: "ADDEC", argLength: 3, reg: gp2flags1carry, asm: "ADDE", clobberFlags: true},              // as ADDE, with XER[CA] out
		{name: "SUB", argLength: 2, reg: gp21, asm: "SUBF"},                                              // arg0 - arg1 (subf with operands swapped)
		{name: "SUBC", argLength: 2, reg: gp21carry, asm: "SUBFC", clobberFlags: true},                   // arg0 - arg1, XER[CA] out (subfc)
		{name: "SUBE", argLength: 3, reg: gp2flags1, asm: "SUBFE", clobberFlags: true},                   // arg0 - arg1 - borrow; arg2=flags
		{name: "SUBEC", argLength: 3, reg: gp2flags1carry, asm: "SUBFE", clobberFlags: true},             // as SUBE, with XER[CA] out

		{name: "MULLW", argLength: 2, reg: gp21, asm: "MULLW", commutative: true},   // arg0 * arg1, low 32
		{name: "MULHW", argLength: 2, reg: gp21, asm: "MULHW", commutative: true},   // (arg0 * arg1) >> 32, signed
		{name: "MULHWU", argLength: 2, reg: gp21, asm: "MULHWU", commutative: true}, // (arg0 * arg1) >> 32, unsigned
		{name: "MULLU", argLength: 2, reg: gp22, asm: "MULLW", commutative: true},   // arg0 * arg1, high 32 in out0, low 32 in out1

		{name: "DIVW", argLength: 2, reg: gp21, asm: "DIVW", clobberFlags: true},
		{name: "DIVWU", argLength: 2, reg: gp21, asm: "DIVWU"},
		{name: "ROTLW", argLength: 2, reg: gp21, asm: "ROTLW"}, // low five count bits, including negative counts
		{name: "NEG", argLength: 1, reg: gp11, asm: "NEG"},     // -arg0

		// Logical.
		{name: "AND", argLength: 2, reg: gp21, asm: "AND", commutative: true}, // arg0 & arg1
		{name: "OR", argLength: 2, reg: gp21, asm: "OR", commutative: true},   // arg0 | arg1
		{name: "XOR", argLength: 2, reg: gp21, asm: "XOR", commutative: true}, // arg0 ^ arg1

		// Shifts. slw/srw use only the low 6 count bits, with bit 5 forcing
		// zero. PPC.rules guards the full Go count against wrapping at 64.
		// sraw/srawi and addic write XER[CA].
		{name: "SLW", argLength: 2, reg: gp21, asm: "SLW"},                       // arg0 << (arg1 & 63), zero-filled
		{name: "SRW", argLength: 2, reg: gp21, asm: "SRW"},                       // arg0 >> (arg1 & 63), zero-filled
		{name: "SRAW", argLength: 2, reg: gp21, asm: "SRAW", clobberFlags: true}, // arg0 >> (arg1 & 63), sign-filled
		{name: "SRAWI", argLength: 1, reg: gp11, asm: "SRAWI", aux: "Int32", clobberFlags: true},

		// Bit and byte manipulation.
		{name: "CNTLZW", argLength: 1, reg: gp11, asm: "CNTLZW"}, // count leading zeros
		{name: "EXTSB", argLength: 1, reg: gp11, asm: "EXTSB"},   // sign-extend the low byte
		{name: "EXTSH", argLength: 1, reg: gp11, asm: "EXTSH"},   // sign-extend the low half word

		// Moves. MOVWconst may need more than one instruction for constants the
		// 16-bit immediate field cannot hold; that expansion is the assembler's
		// business (obj/ppc), not the register allocator's.
		{name: "MOVWconst", argLength: 0, reg: gp01, asm: "ADDI", aux: "Int32", typ: "Int32"},
		{name: "MOVWreg", argLength: 1, reg: gp11, asm: "OR"}, // register-to-register move

		// 32-bit compare. CMP/CMPU in the ppc64 table select 64-bit comparison
		// semantics and are not valid for this 32-bit-register backend.
		{name: "CMPW", argLength: 2, reg: gp2cr, asm: "CMPW", typ: "Flags"},
		{name: "CMPWU", argLength: 2, reg: gp2cr, asm: "CMPLW", typ: "Flags"},
		{name: "CMPWconst", argLength: 1, reg: gp1cr, asm: "CMPWI", aux: "Int32", typ: "Flags"},
		{name: "CMPWUconst", argLength: 1, reg: gp1cr, asm: "CMPLWI", aux: "Int32", typ: "Flags"},
		{name: "InvertFlags", argLength: 1},
		{name: "FlagEQ"},
		{name: "FlagLT"},
		{name: "FlagGT"},
		{name: "Equal", argLength: 1, reg: gp01, typ: "Bool"},
		{name: "NotEqual", argLength: 1, reg: gp01, typ: "Bool"},
		{name: "LessThan", argLength: 1, reg: gp01, typ: "Bool"},
		{name: "LessEqual", argLength: 1, reg: gp01, typ: "Bool"},
		{name: "GreaterThan", argLength: 1, reg: gp01, typ: "Bool"},
		{name: "GreaterEqual", argLength: 1, reg: gp01, typ: "Bool"},

		// gcWriteBarrierN is a compiler-private call: four-byte entries,
		// result in R29, R18/R19 scratch, all other allocatable GPRs and g
		// preserved on both reservation and flush paths. FPRs/flags clobbered.
		{name: "LoweredWB", argLength: 1, aux: "Int64", reg: regInfo{outputs: []regMask{buildReg("R29")}, clobbers: fp | buildReg("R18 R19 R31")}, clobberFlags: true},
		// Bounds metadata uses R3..R9 as save-array slots 0..6. Extend packs
		// two two-bit indices, so its high/low words must be in R3..R6.
		{name: "LoweredPanicBounds", argLength: 3, aux: "Int64", reg: regInfo{inputs: []regMask{buildReg("R3 R4 R5 R6 R7 R8 R9"), buildReg("R3 R4 R5 R6 R7 R8 R9")}}, typ: "Mem", call: true, clobberFlags: true},
		{name: "LoweredPanicExtend", argLength: 4, aux: "Int64", reg: regInfo{inputs: []regMask{buildReg("R3 R4 R5 R6"), buildReg("R3 R4 R5 R6"), buildReg("R3 R4 R5 R6 R7 R8 R9")}}, typ: "Mem", call: true, clobberFlags: true},
		// Scalar copy/clear. Alignment comes from TypSize; R31 is reserved
		// for loaded data, R5 is the loop counter. Only loops change R3/R4.
		{name: "LoweredZero", argLength: 2, aux: "TypSize", reg: regInfo{inputs: []regMask{buildReg("R3")}, clobbers: buildReg("R3 R5")}, typ: "Mem", clobberFlags: true},
		{name: "LoweredMove", argLength: 3, aux: "TypSize", reg: regInfo{inputs: []regMask{buildReg("R3"), buildReg("R4")}, clobbers: buildReg("R3 R4 R5 R31")}, typ: "Mem", clobberFlags: true},
		{name: "LoweredNilCheck", argLength: 2, reg: regInfo{inputs: []regMask{gpspsb | gpg}}, nilCheck: true, faultOnNilArg0: true},
		{name: "LoweredGetClosurePtr", reg: regInfo{outputs: []regMask{buildReg("R11")}}, zeroWidth: true},
		{name: "LoweredGetCallerSP", argLength: 1, reg: gp01, rematerializeable: true},
		{name: "LoweredGetCallerPC", reg: gp01, rematerializeable: true},
		{name: "CALLstatic", argLength: -1, reg: regInfo{clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true},
		{name: "CALLclosure", argLength: -1, reg: regInfo{inputs: []regMask{buildReg("R12"), buildReg("R11"), 0}, clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true},
		{name: "CALLinter", argLength: -1, reg: regInfo{inputs: []regMask{buildReg("R12")}, clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true},
		{name: "CALLtail", argLength: -1, reg: regInfo{clobbers: callerSave}, aux: "CallOff", clobberFlags: true, call: true, tailCall: true},

		// Address formation and memory. The 64-bit forms are not here: the
		// decomposition has already split them into two 32-bit accesses.
		{name: "MOVWaddr", argLength: 1, reg: regInfo{inputs: []regMask{gpspsb}, outputs: []regMask{gp}}, aux: "SymOff", asm: "ADDI", rematerializeable: true, symEffect: "Addr"}, // arg0 + auxInt + aux.(*obj.LSym)
		{name: "MOVBload", argLength: 2, reg: gpload, aux: "SymOff", asm: "LBZ", typ: "Int8", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},
		{name: "MOVBUload", argLength: 2, reg: gpload, aux: "SymOff", asm: "LBZ", typ: "UInt8", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},
		{name: "MOVHload", argLength: 2, reg: gpload, aux: "SymOff", asm: "LHA", typ: "Int16", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},
		{name: "MOVHUload", argLength: 2, reg: gpload, aux: "SymOff", asm: "LHZ", typ: "UInt16", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},
		{name: "MOVWload", argLength: 2, reg: gpload, aux: "SymOff", asm: "LWZ", typ: "UInt32", faultOnNilArg0: true, symEffect: "Read", addrSinkArg0: true},
		{name: "MOVBstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "STB", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true},
		{name: "MOVHstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "STH", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true},
		{name: "MOVWstore", argLength: 3, reg: gpstore, aux: "SymOff", asm: "STW", typ: "Mem", faultOnNilArg0: true, symEffect: "Write", addrSinkArg0: true},
	}

	archs = append(archs, arch{
		name:    "PPC",
		pkg:     "cmd/internal/obj/ppc",
		genfile: "../../ppc/ssa.go",
		ops:     ops,
		blocks: []blockData{
			{name: "EQ", controls: 1}, {name: "NE", controls: 1},
			{name: "LT", controls: 1}, {name: "LE", controls: 1},
			{name: "GT", controls: 1}, {name: "GE", controls: 1},
			{name: "FLT", controls: 1}, {name: "FLE", controls: 1},
			{name: "FGT", controls: 1}, {name: "FGE", controls: 1},
		},
		ParamIntRegNames: "R3 R4 R5 R6 R7 R8 R9 R10",
		regnames:         regNamesPPC,
		gpregmask:        gp,
		fpregmask:        fp,
		specialregmask:   0,
		framepointerreg:  -1, // no hardware frame pointer on ppc32
		linkreg:          -1, // LR is not a GPR the allocator sees
	})
}
