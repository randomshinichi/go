// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// A field is one operand position of a PowerPC instruction word. The names are
// the architectural field names; "rd" and "rs" (and their floating-point
// spellings frd/frs) are the same 5-bit field, because PowerPC puts the
// destination of a logical operation in the RA position and the source in the
// RS position — the operand spec below, not the encoder, decides which name a
// form uses, so `or ra,rs,rb` and `add rd,ra,rb` both encode correctly.
type field int8

const (
	fInvalid field = iota
	fRD            // rd / rs / frd / frs : bits 6-10
	fRA            // ra / fra              : bits 11-15
	fRB            // rb / frb / sh         : bits 16-20
	fMB            // mb / frc              : bits 21-25
	fME            // me                    : bits 26-30
	fSI            // signed 16-bit immediate, bits 16-31
	fUI            // unsigned 16-bit immediate, bits 16-31
	fD             // signed 16-bit displacement, bits 16-31
	fLI            // signed 26-bit displacement scaled by 4, bits 6-29
	fBD            // signed 16-bit displacement scaled by 4, bits 16-29
	fBO            // branch condition operand, bits 6-10
	fBI            // branch condition bit, bits 11-15
	fBH            // branch hint, bits 16-17
	fBF            // condition register field, bits 6-8
	fCRM           // condition register mask, bits 12-19
)

var fieldNames = map[string]field{
	"rd": fRD, "rs": fRD, "frd": fRD, "frs": fRD,
	"ra": fRA, "fra": fRA,
	"rb": fRB, "frb": fRB, "sh": fRB,
	"mb": fMB, "frc": fMB,
	"me": fME,
	"si": fSI, "ui": fUI, "d": fD,
	"li": fLI, "bd": fBD,
	"bo": fBO, "bi": fBI, "bh": fBH,
	"bf": fBF, "crm": fCRM,
}

// A form is one encodable instruction: the opcode, its assembler mnemonic, the
// fixed bits of the word, and the operand fields in source order. spec names the
// fields and is parsed at init, so the table below stays readable.
type form struct {
	op     Inst
	mnem   string
	base   uint32
	spec   string
	fields []field
}

var forms = []form{
	{op: AADD, mnem: "add", base: 0x7C000214, spec: "rd,ra,rb"},
	{op: AADDC, mnem: "addc", base: 0x7C000014, spec: "rd,ra,rb"},
	{op: AADDE, mnem: "adde", base: 0x7C000114, spec: "rd,ra,rb"},
	{op: AADDI, mnem: "addi", base: 0x38000000, spec: "rd,ra,si"},
	{op: AADDIC, mnem: "addic", base: 0x30000000, spec: "rd,ra,si"},
	{op: AADDIC_, mnem: "addic.", base: 0x34000000, spec: "rd,ra,si"},
	{op: AADDIS, mnem: "addis", base: 0x3C000000, spec: "rd,ra,si"},
	{op: ASUBF, mnem: "subf", base: 0x7C000050, spec: "rd,ra,rb"},
	{op: ASUBFC, mnem: "subfc", base: 0x7C000010, spec: "rd,ra,rb"},
	{op: ASUBFE, mnem: "subfe", base: 0x7C000110, spec: "rd,ra,rb"},
	{op: ASUBFIC, mnem: "subfic", base: 0x20000000, spec: "rd,ra,si"},
	{op: AMULLW, mnem: "mullw", base: 0x7C0001D6, spec: "rd,ra,rb"},
	{op: AMULHW, mnem: "mulhw", base: 0x7C000096, spec: "rd,ra,rb"},
	{op: AMULHWU, mnem: "mulhwu", base: 0x7C000016, spec: "rd,ra,rb"},
	{op: ADIVW, mnem: "divw", base: 0x7C0003D6, spec: "rd,ra,rb"},
	{op: ADIVWU, mnem: "divwu", base: 0x7C000396, spec: "rd,ra,rb"},
	{op: ANEG, mnem: "neg", base: 0x7C0000D0, spec: "rd,ra"},
	{op: AAND, mnem: "and", base: 0x7C000038, spec: "ra,rs,rb"},
	{op: AANDC, mnem: "andc", base: 0x7C000078, spec: "ra,rs,rb"},
	{op: AANDI_, mnem: "andi.", base: 0x70000000, spec: "ra,rs,ui"},
	{op: AANDIS_, mnem: "andis.", base: 0x74000000, spec: "ra,rs,ui"},
	{op: AOR, mnem: "or", base: 0x7C000378, spec: "ra,rs,rb"},
	{op: AORC, mnem: "orc", base: 0x7C000338, spec: "ra,rs,rb"},
	{op: AORI, mnem: "ori", base: 0x60000000, spec: "ra,rs,ui"},
	{op: AORIS, mnem: "oris", base: 0x64000000, spec: "ra,rs,ui"},
	{op: AXOR, mnem: "xor", base: 0x7C000278, spec: "ra,rs,rb"},
	{op: AXORI, mnem: "xori", base: 0x68000000, spec: "ra,rs,ui"},
	{op: AXORIS, mnem: "xoris", base: 0x6C000000, spec: "ra,rs,ui"},
	{op: ANAND, mnem: "nand", base: 0x7C0003B8, spec: "ra,rs,rb"},
	{op: ANOR, mnem: "nor", base: 0x7C0000F8, spec: "ra,rs,rb"},
	{op: AEQV, mnem: "eqv", base: 0x7C000238, spec: "ra,rs,rb"},
	{op: ASLW, mnem: "slw", base: 0x7C000030, spec: "ra,rs,rb"},
	{op: ASRW, mnem: "srw", base: 0x7C000430, spec: "ra,rs,rb"},
	{op: ASRAW, mnem: "sraw", base: 0x7C000630, spec: "ra,rs,rb"},
	{op: ASRAWI, mnem: "srawi", base: 0x7C000670, spec: "ra,rs,sh"},
	{op: ARLWINM, mnem: "rlwinm", base: 0x54000000, spec: "ra,rs,sh,mb,me"},
	{op: ARLWIMI, mnem: "rlwimi", base: 0x50000000, spec: "ra,rs,sh,mb,me"},
	{op: ARLWNM, mnem: "rlwnm", base: 0x5C000000, spec: "ra,rs,rb,mb,me"},
	{op: AROTLW, mnem: "rotlw", base: 0x5C00003E, spec: "ra,rs,rb"},
	{op: ACNTLZW, mnem: "cntlzw", base: 0x7C000034, spec: "ra,rs"},
	{op: AEXTSB, mnem: "extsb", base: 0x7C000774, spec: "ra,rs"},
	{op: AEXTSH, mnem: "extsh", base: 0x7C000734, spec: "ra,rs"},
	{op: ACMPW, mnem: "cmpw", base: 0x7C000000, spec: "bf,ra,rb"},
	{op: ACMPLW, mnem: "cmplw", base: 0x7C000040, spec: "bf,ra,rb"},
	{op: ACMPWI, mnem: "cmpwi", base: 0x2C000000, spec: "bf,ra,si"},
	{op: ACMPLWI, mnem: "cmplwi", base: 0x28000000, spec: "bf,ra,ui"},
	{op: ALWZ, mnem: "lwz", base: 0x80000000, spec: "rd,d,ra"},
	{op: ALWZU, mnem: "lwzu", base: 0x84000000, spec: "rd,d,ra"},
	{op: ALWZX, mnem: "lwzx", base: 0x7C00002E, spec: "rd,ra,rb"},
	{op: ALWZUX, mnem: "lwzux", base: 0x7C00006E, spec: "rd,ra,rb"},
	{op: ALBZ, mnem: "lbz", base: 0x88000000, spec: "rd,d,ra"},
	{op: ALBZU, mnem: "lbzu", base: 0x8C000000, spec: "rd,d,ra"},
	{op: ALBZX, mnem: "lbzx", base: 0x7C0000AE, spec: "rd,ra,rb"},
	{op: ALBZUX, mnem: "lbzux", base: 0x7C0000EE, spec: "rd,ra,rb"},
	{op: ALHZ, mnem: "lhz", base: 0xA0000000, spec: "rd,d,ra"},
	{op: ALHZU, mnem: "lhzu", base: 0xA4000000, spec: "rd,d,ra"},
	{op: ALHZX, mnem: "lhzx", base: 0x7C00022E, spec: "rd,ra,rb"},
	{op: ALHZUX, mnem: "lhzux", base: 0x7C00026E, spec: "rd,ra,rb"},
	{op: ALHA, mnem: "lha", base: 0xA8000000, spec: "rd,d,ra"},
	{op: ALHAU, mnem: "lhau", base: 0xAC000000, spec: "rd,d,ra"},
	{op: ALHAX, mnem: "lhax", base: 0x7C0002AE, spec: "rd,ra,rb"},
	{op: ALHAUX, mnem: "lhaux", base: 0x7C0002EE, spec: "rd,ra,rb"},
	{op: ALMW, mnem: "lmw", base: 0xB8000000, spec: "rd,d,ra"},
	{op: ALWARX, mnem: "lwarx", base: 0x7C000028, spec: "rd,ra,rb"},
	{op: ALWBRX, mnem: "lwbrx", base: 0x7C00042C, spec: "rd,ra,rb"},
	{op: ALHBRX, mnem: "lhbrx", base: 0x7C00062C, spec: "rd,ra,rb"},
	{op: ASTW, mnem: "stw", base: 0x90000000, spec: "rs,d,ra"},
	{op: ASTWU, mnem: "stwu", base: 0x94000000, spec: "rs,d,ra"},
	{op: ASTWX, mnem: "stwx", base: 0x7C00012E, spec: "rs,ra,rb"},
	{op: ASTWUX, mnem: "stwux", base: 0x7C00016E, spec: "rs,ra,rb"},
	{op: ASTB, mnem: "stb", base: 0x98000000, spec: "rs,d,ra"},
	{op: ASTBU, mnem: "stbu", base: 0x9C000000, spec: "rs,d,ra"},
	{op: ASTBX, mnem: "stbx", base: 0x7C0001AE, spec: "rs,ra,rb"},
	{op: ASTBUX, mnem: "stbux", base: 0x7C0001EE, spec: "rs,ra,rb"},
	{op: ASTH, mnem: "sth", base: 0xB0000000, spec: "rs,d,ra"},
	{op: ASTHU, mnem: "sthu", base: 0xB4000000, spec: "rs,d,ra"},
	{op: ASTHX, mnem: "sthx", base: 0x7C00032E, spec: "rs,ra,rb"},
	{op: ASTHUX, mnem: "sthux", base: 0x7C00036E, spec: "rs,ra,rb"},
	{op: ASTMW, mnem: "stmw", base: 0xBC000000, spec: "rs,d,ra"},
	{op: ASTWCX_, mnem: "stwcx.", base: 0x7C00012D, spec: "rs,ra,rb"},
	{op: ASTWBRX, mnem: "stwbrx", base: 0x7C00052C, spec: "rs,ra,rb"},
	{op: ASTHBRX, mnem: "sthbrx", base: 0x7C00072C, spec: "rs,ra,rb"},
	{op: AB, mnem: "b", base: 0x48000000, spec: "li"},
	{op: ABL, mnem: "bl", base: 0x48000001, spec: "li"},
	{op: ABC, mnem: "bc", base: 0x40000000, spec: "bo,bi,bd"},
	{op: ABLR, mnem: "blr", base: 0x4E800020, spec: ""},
	{op: ABLRL, mnem: "blrl", base: 0x4E800021, spec: ""},
	{op: ABCTR, mnem: "bctr", base: 0x4E800420, spec: ""},
	{op: ABCTRL, mnem: "bctrl", base: 0x4E800421, spec: ""},
	{op: ABCLR, mnem: "bclr", base: 0x4C000020, spec: "bo,bi,bh"},
	{op: ABCLRL, mnem: "bclrl", base: 0x4C000021, spec: "bo,bi,bh"},
	{op: ABCCTR, mnem: "bcctr", base: 0x4C000420, spec: "bo,bi,bh"},
	{op: ABCCTRL, mnem: "bcctrl", base: 0x4C000421, spec: "bo,bi,bh"},
	{op: AMFCR, mnem: "mfcr", base: 0x7C000026, spec: "rd"},
	{op: AMFLR, mnem: "mflr", base: 0x7C0802A6, spec: "rd"},
	{op: AMFCTR, mnem: "mfctr", base: 0x7C0902A6, spec: "rd"},
	{op: AMTCRF, mnem: "mtcrf", base: 0x7C000120, spec: "crm,rs"},
	{op: AMTLR, mnem: "mtlr", base: 0x7C0803A6, spec: "rs"},
	{op: AMTCTR, mnem: "mtctr", base: 0x7C0903A6, spec: "rs"},
	{op: ALFS, mnem: "lfs", base: 0xC0000000, spec: "frd,d,ra"},
	{op: ALFSU, mnem: "lfsu", base: 0xC4000000, spec: "frd,d,ra"},
	{op: ALFSX, mnem: "lfsx", base: 0x7C00042E, spec: "frd,ra,rb"},
	{op: ALFSUX, mnem: "lfsux", base: 0x7C00046E, spec: "frd,ra,rb"},
	{op: ALFD, mnem: "lfd", base: 0xC8000000, spec: "frd,d,ra"},
	{op: ALFDU, mnem: "lfdu", base: 0xCC000000, spec: "frd,d,ra"},
	{op: ALFDX, mnem: "lfdx", base: 0x7C0004AE, spec: "frd,ra,rb"},
	{op: ALFDUX, mnem: "lfdux", base: 0x7C0004EE, spec: "frd,ra,rb"},
	{op: ASTFS, mnem: "stfs", base: 0xD0000000, spec: "frs,d,ra"},
	{op: ASTFSU, mnem: "stfsu", base: 0xD4000000, spec: "frs,d,ra"},
	{op: ASTFSX, mnem: "stfsx", base: 0x7C00052E, spec: "frs,ra,rb"},
	{op: ASTFSUX, mnem: "stfsux", base: 0x7C00056E, spec: "frs,ra,rb"},
	{op: ASTFD, mnem: "stfd", base: 0xD8000000, spec: "frs,d,ra"},
	{op: ASTFDU, mnem: "stfdu", base: 0xDC000000, spec: "frs,d,ra"},
	{op: ASTFDX, mnem: "stfdx", base: 0x7C0005AE, spec: "frs,ra,rb"},
	{op: ASTFDUX, mnem: "stfdux", base: 0x7C0005EE, spec: "frs,ra,rb"},
	{op: AFADD, mnem: "fadd", base: 0xFC00002A, spec: "frd,fra,frb"},
	{op: AFADDS, mnem: "fadds", base: 0xEC00002A, spec: "frd,fra,frb"},
	{op: AFSUB, mnem: "fsub", base: 0xFC000028, spec: "frd,fra,frb"},
	{op: AFSUBS, mnem: "fsubs", base: 0xEC000028, spec: "frd,fra,frb"},
	{op: AFMUL, mnem: "fmul", base: 0xFC000032, spec: "frd,fra,frc"},
	{op: AFMULS, mnem: "fmuls", base: 0xEC000032, spec: "frd,fra,frc"},
	{op: AFDIV, mnem: "fdiv", base: 0xFC000024, spec: "frd,fra,frb"},
	{op: AFDIVS, mnem: "fdivs", base: 0xEC000024, spec: "frd,fra,frb"},
	{op: AFNEG, mnem: "fneg", base: 0xFC000050, spec: "frd,frb"},
	{op: AFABS, mnem: "fabs", base: 0xFC000210, spec: "frd,frb"},
	{op: AFMR, mnem: "fmr", base: 0xFC000090, spec: "frd,frb"},
	{op: AFRSP, mnem: "frsp", base: 0xFC000018, spec: "frd,frb"},
	{op: AFCTIWZ, mnem: "fctiwz", base: 0xFC00001E, spec: "frd,frb"},
	{op: AFCMPU, mnem: "fcmpu", base: 0xFC000000, spec: "bf,fra,frb"},
	{op: AFCMPO, mnem: "fcmpo", base: 0xFC000040, spec: "bf,fra,frb"},
	{op: ASYNC, mnem: "sync", base: 0x7C0004AC, spec: ""},
	{op: AISYNC, mnem: "isync", base: 0x4C00012C, spec: ""},
}

var formByOp = map[Inst]*form{}
var formByName = map[string]*form{}
var anames []string

func init() {
	for i := range forms {
		f := &forms[i]
		if int(f.op) != int(AXXX)+1+i {
			panic(fmt.Sprintf("ppc: opcode table out of step with the Inst constants at %q", f.mnem))
		}
		for _, name := range strings.Split(f.spec, ",") {
			if name == "" {
				continue
			}
			fl, ok := fieldNames[name]
			if !ok {
				panic(fmt.Sprintf("ppc: %s: unknown operand field %q", f.mnem, name))
			}
			f.fields = append(f.fields, fl)
		}
		formByOp[f.op] = f
		formByName[f.mnem] = f
	}
	anames = make([]string, len(forms)+1)
	anames[0] = "XXX"
	for i := range forms {
		anames[i+1] = strings.ToUpper(forms[i].mnem)
	}
}

// Anames are the assembler mnemonics, indexed by op-AXXX, so that obj.As
// printing and the assembler front end can name instructions. Index 0 is AXXX.
func Anames() []string { return anames }

// Mnemonic returns the assembler mnemonic for op, or "" if op is not one of the
// ppc32 baseline instructions.
func Mnemonic(op Inst) string {
	if f, ok := formByOp[op]; ok {
		return f.mnem
	}
	return ""
}

// Lookup finds the instruction for an assembler mnemonic (case-insensitive).
// The error distinguishes a mnemonic that is deliberately outside the G3
// baseline (Refused, with the reason) from one that is simply unknown.
func Lookup(mnemonic string) (Inst, error) {
	name := strings.ToLower(strings.TrimSpace(mnemonic))
	if f, ok := formByName[name]; ok {
		return f.op, nil
	}
	if reason, ok := Refused(name); ok {
		return AXXX, fmt.Errorf("ppc: %s is outside the G3 scalar baseline: %s", mnemonic, reason)
	}
	return AXXX, fmt.Errorf("ppc: unknown instruction %q", mnemonic)
}

// Encode returns the big-endian instruction word for op with the given operands,
// in the operand order the form documents (registers by architectural number,
// GPR 0-31 and FPR 0-31).
//
// Branch operands are byte displacements relative to the branch instruction
// itself, already resolved and a multiple of four; symbolic targets are the
// linker's business and are not accepted here.
func Encode(op Inst, args ...int32) (uint32, error) {
	f, ok := formByOp[op]
	if !ok {
		return 0, fmt.Errorf("ppc: 0x%x is not a ppc32 baseline instruction", uint32(op))
	}
	return f.encode(args)
}

// Emit encodes the instruction and appends its four big-endian bytes to dst.
func Emit(dst []byte, op Inst, args ...int32) ([]byte, error) {
	w, err := Encode(op, args...)
	if err != nil {
		return dst, err
	}
	return binary.BigEndian.AppendUint32(dst, w), nil
}

func (f *form) encode(args []int32) (uint32, error) {
	if len(args) != len(f.fields) {
		return 0, fmt.Errorf("ppc: %s takes %d operands (%s), got %d", f.mnem, len(f.fields),
			f.operandNames(), len(args))
	}
	word := f.base
	for i, fl := range f.fields {
		v, err := fl.encode(args[i])
		if err != nil {
			return 0, fmt.Errorf("ppc: %s operand %d: %v", f.mnem, i+1, err)
		}
		word |= v
	}
	if err := f.checkOperandCombination(args); err != nil {
		return 0, err
	}
	return word, nil
}

func (f *form) operandNames() string { return f.spec }

func (fl field) String() string {
	for name, k := range fieldNames {
		if k == fl {
			return name
		}
	}
	return "?"
}

// encode masks and shifts one validated operand into its field position.
func (fl field) encode(v int32) (uint32, error) {
	switch fl {
	case fRD, fRA, fRB, fMB, fME, fBO, fBI:
		if v < 0 || v > 31 {
			return 0, fmt.Errorf("%s = %d out of range 0-31", fl, v)
		}
	case fBH:
		if v < 0 || v > 3 {
			return 0, fmt.Errorf("branch hint %d out of range 0-3", v)
		}
	case fBF:
		if v < 0 || v > 7 {
			return 0, fmt.Errorf("condition register field %d out of range 0-7", v)
		}
	case fCRM:
		if v < 0 || v > 0xff {
			return 0, fmt.Errorf("condition register mask 0x%x out of range 0x00-0xff", v)
		}
	case fSI, fD:
		if v < -32768 || v > 32767 {
			return 0, fmt.Errorf("immediate %d out of signed 16-bit range", v)
		}
	case fUI:
		if v < 0 || v > 0xffff {
			return 0, fmt.Errorf("immediate %d out of unsigned 16-bit range", v)
		}
	case fLI:
		if v < -(1<<25) || v > (1<<25)-4 {
			return 0, fmt.Errorf("branch displacement %d out of signed 26-bit byte range", v)
		}
		if v%4 != 0 {
			return 0, fmt.Errorf("branch displacement %d is not a multiple of 4", v)
		}
	case fBD:
		if v < -32768 || v > 32764 {
			return 0, fmt.Errorf("branch displacement %d out of signed 16-bit byte range", v)
		}
		if v%4 != 0 {
			return 0, fmt.Errorf("branch displacement %d is not a multiple of 4", v)
		}
	default:
		return 0, fmt.Errorf("internal error: unencodable field %d", fl)
	}
	shift, mask := fl.position()
	return (uint32(v) & mask) << shift, nil
}

// position returns the shift and width mask of a field in a big-endian
// instruction word, with bit numbers as the PowerPC manual gives them: field
// bits are named 0-31 from the most significant bit, so manual bit n is at
// shift 31-n here. Branch displacements are the exception: their two low-order
// bits are implied zero, so masking a byte displacement into 0x03fffffc or
// 0x0000fffc stores displacement/4 in the field, which is the encoding clang and
// the PowerPC manual agree on (checked against the expected-words fixture).
func (fl field) position() (shift uint, mask uint32) {
	switch fl {
	case fRD, fBO:
		return 21, 0x1f // bits 6-10
	case fRA, fBI:
		return 16, 0x1f // bits 11-15
	case fRB:
		return 11, 0x1f // bits 16-20
	case fMB:
		return 6, 0x1f // bits 21-25
	case fME:
		return 1, 0x1f // bits 26-30
	case fSI, fUI, fD:
		return 0, 0xffff // bits 16-31
	case fBH:
		return 14, 0x3 // bits 16-17
	case fBF:
		return 23, 0x7 // bits 6-8
	case fCRM:
		return 12, 0xff // bits 12-19
	case fLI:
		// Bits 6-29, a byte displacement whose two low-order bits are implied
		// zero, so the encoded field is displacement/4 and masking the (aligned)
		// byte displacement into those bits gives exactly that.
		return 0, 0x03fffffc
	case fBD:
		// Bits 16-29, same convention as LI.
		return 0, 0x0000fffc
	}
	return 0, 0
}

// checkOperandCombination rejects operand combinations whose *encoding* is
// reserved even though each field is individually in range. The one rule in the
// baseline is the branch-conditional BO field of bclr/bcctr: isa_check.py
// classifies BO values with bits 2 and 4 set and either of bits 0/1 set (for
// example BO=23, the corpus's "branch-bad" row) as an illegal form, and emitting
// one would be exactly the kind of word this port must not produce.
func (f *form) checkOperandCombination(args []int32) error {
	switch f.mnem {
	case "bclr", "bclrl", "bcctr", "bcctrl":
		bo := args[0]
		if bo&20 == 20 && bo&11 != 0 {
			return fmt.Errorf("ppc: %s: reserved BO encoding %d", f.mnem, bo)
		}
	}
	return nil
}

// refused maps a mnemonic outside the G3 scalar baseline to the reason. Named
// forms matter: a mnemonic that is merely absent could be added by accident,
// whereas a refusal is a recorded decision with a blocker.
var refused = map[string]string{
	// Not G3-legal: the 750 has no isel and no POWER5 rounding-to-integer forms.
	"isel": "POWER5+ conditional-move form; illegal on every isa-v1 profile",
	"friz": "POWER5 round-to-integer-form; not G3-legal",
	"frim": "POWER5 round-to-integer-form; not G3-legal",
	"frip": "POWER5 round-to-integer-form; not G3-legal",
	"frin": "POWER5 round-to-integer-form; not G3-legal",
	// 64-bit-only forms (design D-1: never emitted, even on a 970).
	"ld":      "64-bit-only load (D-1)",
	"ldu":     "64-bit-only load (D-1)",
	"lwa":     "64-bit-only load (D-1)",
	"std":     "64-bit-only store (D-1)",
	"stdu":    "64-bit-only store (D-1)",
	"ldarx":   "64-bit-only reservation load; 32-bit atomics use lwarx (D-4)",
	"stdcx.":  "64-bit-only reservation store; 32-bit atomics use stwcx. (D-4)",
	"mulld":   "64-bit-only multiply (D-1)",
	"mulhd":   "64-bit-only multiply (D-1)",
	"mulhdu":  "64-bit-only multiply (D-1)",
	"divd":    "64-bit-only divide (D-1)",
	"divdu":   "64-bit-only divide (D-1)",
	"sld":     "64-bit-only shift (D-1)",
	"srd":     "64-bit-only shift (D-1)",
	"srad":    "64-bit-only shift (D-1)",
	"sradi":   "64-bit-only shift (D-1)",
	"rldicl":  "64-bit-only rotate (D-1)",
	"rldicr":  "64-bit-only rotate (D-1)",
	"rldic":   "64-bit-only rotate (D-1)",
	"rldimi":  "64-bit-only rotate (D-1)",
	"rldcl":   "64-bit-only rotate (D-1)",
	"rldcr":   "64-bit-only rotate (D-1)",
	"cntlzd":  "64-bit-only bit count (D-1); ctz/clz use cntlzw",
	"extsd":   "64-bit-only sign extension (D-1)",
	"cmpd":    "64-bit-only compare (D-1)",
	"cmpld":   "64-bit-only compare (D-1)",
	"cmpdi":   "64-bit-only compare (D-1)",
	"cmpldi":  "64-bit-only compare (D-1)",
	"fctid":   "64-bit-only FP convert; the 750 has no fctid/fcfid",
	"fctidz":  "64-bit-only FP convert; illegal on 750/7450 (isa_check)",
	"fctiduz": "64-bit-only FP convert (D-1)",
	"fcfid":   "64-bit-only FP convert; the 750 has no fctid/fcfid",
	"fcfidu":  "64-bit-only FP convert (D-1)",
	"fcfids":  "64-bit-only FP convert (D-1)",
	"fcfidus": "64-bit-only FP convert (D-1)",
	// POWER5+/POWER6+ forms and instructions the 750 does not implement.
	"popcntd": "no popcntb/popcntw/popcntd on G3",
	"popcntw": "no popcntb/popcntw/popcntd on G3",
	"popcntb": "no popcntb/popcntw/popcntd on G3",
	"prtyd":   "not G3-legal",
	"prtyw":   "not G3-legal",
	"bpermd":  "not G3-legal",
	"lq":      "not G3-legal",
	"stq":     "not G3-legal",
	"mtfsb0":  "unestablished on 750 (open question Q-8)",
	"mtfsb1":  "unestablished on 750 (open question Q-8)",
	"mtfsfi":  "unestablished on 750 (open question Q-8)",
	"mtfsf":   "unestablished on 750 (open question Q-8)",
	"mffs":    "unestablished on 750 (open question Q-8)",
	// Privileged: the baseline is user-mode Go code, and isa_check classes the
	// corpus's mfmsr row execution_context=forbidden for privilege=user.
	"mfmsr":   "privileged (MSR) instruction; outside the user-mode baseline",
	"mtmsr":   "privileged (MSR) instruction; outside the user-mode baseline",
	"mfsr":    "privileged instruction",
	"mfsrin":  "privileged instruction",
	"mtsr":    "privileged instruction",
	"mtsrin":  "privileged instruction",
	"rfi":     "privileged instruction",
	"sc":      "supervisor call: the runtime lane owns the syscall path (3.4)",
	"tlbli":   "privileged instruction",
	"tlbie":   "privileged instruction",
	"tlbsync": "privileged instruction",
	"dcbf":    "privileged; G3 legality unestablished",
	"dcbi":    "privileged instruction",
	"dcbst":   "privileged; G3 legality unestablished",
	"icbi":    "privileged; G3 legality unestablished",
	"mfspr":   "generic SPR access is not in the baseline; use a named form",
	"mtspr":   "generic SPR access is not in the baseline; use a named form",
	// Open questions in the approved design: legality is not established, so the
	// baseline does not emit them.
	"lwsync":  "legality on 750/7400 unestablished (Q-9); use sync/isync",
	"stfiwx":  "legality on 750/7450/970 unestablished (Q-8)",
	"fsqrt":   "legality on 750/7450/970 unestablished (Q-8)",
	"fsqrts":  "legality on 750/7450/970 unestablished (Q-8)",
	"fres":    "legality on 750/7450/970 unestablished (Q-8)",
	"frsqrte": "legality on 750/7450/970 unestablished (Q-8)",
	"eciwx":   "optional-model form, unresolved on non-970 profiles (isa_check)",
	"ecowx":   "optional-model form, unresolved on non-970 profiles (isa_check)",
	// VMX/AltiVec and VSX are out of scope (design §10).
	"dst":      "VMX (AltiVec) is out of scope for the scalar baseline",
	"dstst":    "VMX (AltiVec) is out of scope for the scalar baseline",
	"mfsprg":   "not G3-legal",
	"mfvscr":   "VMX (AltiVec) is out of scope for the scalar baseline",
	"mtvscr":   "VMX (AltiVec) is out of scope for the scalar baseline",
	"lvsl":     "VMX (AltiVec) is out of scope for the scalar baseline",
	"lvsr":     "VMX (AltiVec) is out of scope for the scalar baseline",
	"lvx":      "VMX (AltiVec) is out of scope for the scalar baseline",
	"stvx":     "VMX (AltiVec) is out of scope for the scalar baseline",
	"vperm":    "VMX (AltiVec) is out of scope for the scalar baseline",
	"vor":      "VMX (AltiVec) is out of scope for the scalar baseline",
	"vxor":     "VMX (AltiVec) is out of scope for the scalar baseline",
	"vspltisb": "VMX (AltiVec) is out of scope for the scalar baseline",
}

// Refused reports whether a mnemonic is deliberately outside the baseline, and
// why. VMX and VSX mnemonics are refused by prefix, so a form nobody thought to
// list is still refused rather than silently unknown.
func Refused(mnemonic string) (string, bool) {
	m := strings.ToLower(strings.TrimSpace(mnemonic))
	if m == "" {
		return "", false
	}
	if reason, ok := refused[m]; ok {
		return reason, true
	}
	// Recording forms are spelled with a trailing dot ("stdcx.", "stwcx.").
	if reason, ok := refused[m+"."]; ok {
		return reason, true
	}
	switch {
	case strings.HasPrefix(m, "xx"), strings.HasPrefix(m, "xv"):
		return "VSX is out of scope and not G3-legal", true
	case strings.HasPrefix(m, "v"):
		// Every VMX (AltiVec) mnemonic starts with v, and no baseline mnemonic does.
		return "VMX (AltiVec) is out of scope for the scalar baseline", true
	}
	return "", false
}
