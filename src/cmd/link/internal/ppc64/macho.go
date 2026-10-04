// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc64

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
)

// archrelocmacho prepares instruction addends in an external Mach-O object.
// PPC has no separate addend field: HA16's PAIR holds the low half, and
// LO16/LO14's PAIR holds the unadjusted high half. machoreloc1 emits those
// records. ELF relocation counts and TOC/TLS rules do not apply here.
func archrelocmacho(target *ld.Target, ldr *loader.Loader, r loader.Reloc, s loader.Sym, val int64) (int64, int, bool) {
	switch r.Type() {
	case objabi.R_CALLPOWER:
		if r.Siz() != 4 {
			return val, 0, false
		}
		t := r.Add()
		if ldr.SymSect(r.Sym()) != nil {
			// A section-targeted BR24 contains S+A-P in the instruction.
			t += ldr.SymValue(r.Sym()) - (ldr.SymValue(s) + int64(r.Off()))
		}
		if t&3 != 0 || int64(int32(t<<6)>>6) != t {
			ldr.Errorf(s, "Mach-O BR24 displacement out of range or unaligned: %d", t)
		}
		// Preserve AA and LK (low two bits), not the old displacement.
		return val&^0x03fffffc | int64(uint32(t)&0x03fffffc), 1, true

	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS, objabi.R_ADDRPOWER_PCREL:
		if r.Siz() != 8 {
			return val, 0, false
		}
		_, off := ld.FoldSubSymbolOffset(ldr, r.Sym())
		t := r.Add() + off
		if r.Type() == objabi.R_ADDRPOWER_PCREL {
			if ldr.SymSect(r.Sym()) == nil {
				ldr.Errorf(s, "Mach-O section difference requires a defined target: %s", ldr.SymName(r.Sym()))
			}
			t = ldr.SymValue(r.Sym()) + r.Add() - (ldr.SymValue(s) + int64(r.Off()))
		}
		if t < -1<<31 || t >= 1<<31 {
			ldr.Errorf(s, "Mach-O instruction relocation addend out of signed 32-bit range: %d", t)
		}
		hi, lo := unpackInstPair(target, val)
		hi = hi&^0xffff | computeHA(int32(t))
		if r.Type() == objabi.R_ADDRPOWER_DS {
			if t&3 != 0 {
				ldr.Errorf(s, "Mach-O LO14 addend unaligned: %d", t)
			}
			lo = lo&^0xfffc | uint32(t)&0xfffc
		} else {
			lo = lo&^0xffff | computeLO(int32(t))
		}
		return packInstPair(target, hi, lo), 4, true
	}
	return val, 0, false
}

// machoRelocWord packs Leopard's big-endian relocation_info bitfields.
// Unlike the little-endian architectures, symbolnum is in bits 31..8.
func machoRelocWord(symbol uint32, pcrel bool, length uint32, external bool, typ uint32) uint32 {
	v := symbol<<8 | length<<5 | typ
	if pcrel {
		v |= 1 << 7
	}
	if external {
		v |= 1 << 4
	}
	return v
}

// machoScatteredWord packs scattered_relocation_info. Both endian variants
// put R_SCATTERED in bit 31 and the section offset in the low 24 bits.
func machoScatteredWord(address uint32, pcrel bool, length uint32, typ uint32) uint32 {
	v := uint32(1<<31) | length<<28 | typ<<24 | address
	if pcrel {
		v |= 1 << 30
	}
	return v
}

func machoreloc1(arch *sys.Arch, out *ld.OutBuf, ldr *loader.Loader, s loader.Sym, r loader.ExtReloc, sectoff int64) bool {
	if arch != sys.ArchPPC64 || sectoff < 0 || sectoff >= 1<<31 {
		return false
	}
	rs := r.Xsym
	sect := ldr.SymSect(rs)
	var symbol uint32
	external := true
	if r.Type == objabi.R_ADDRPOWER_PCREL {
		// Scattered records identify addresses, not symbol-table indices.
	} else if ldr.SymType(s).IsDWARF() || r.Type == objabi.R_CALLPOWER && sect != nil {
		if sect == nil || sect.Extnum <= 0 {
			return false
		}
		symbol = uint32(sect.Extnum)
		external = false
	} else {
		id := ldr.SymDynid(rs)
		if id < 0 || id >= 1<<24 {
			ldr.Errorf(s, "Mach-O relocation target has no valid symbol index: %s", ldr.SymName(rs))
			return false
		}
		symbol = uint32(id)
	}

	emit := func(address int64, pcrel bool, length uint32, ext bool, typ uint32, symnum uint32) {
		out.Write32(uint32(address))
		out.Write32(machoRelocWord(symnum, pcrel, length, ext, typ))
	}
	emitPair := func(half uint32) {
		emit(int64(half), false, 2, false, ld.MACHO_PPC_RELOC_PAIR, 0)
	}

	switch r.Type {
	case objabi.R_ADDR:
		var length uint32
		switch r.Size {
		case 1:
			length = 0
		case 2:
			length = 1
		case 4:
			length = 2
		case 8:
			length = 3
		default:
			return false
		}
		emit(sectoff, false, length, external, ld.MACHO_PPC_RELOC_VANILLA, symbol)
		return true

	case objabi.R_CALLPOWER:
		if r.Size != 4 {
			return false
		}
		emit(sectoff, true, 2, external, ld.MACHO_PPC_RELOC_BR24, symbol)
		return true

	case objabi.R_ADDRPOWER, objabi.R_ADDRPOWER_DS:
		if r.Size != 8 || sectoff+4 >= 1<<31 || r.Xadd < -1<<31 || r.Xadd >= 1<<31 {
			return false
		}
		if r.Type == objabi.R_ADDRPOWER_DS && r.Xadd&3 != 0 {
			return false
		}
		loType := uint32(ld.MACHO_PPC_RELOC_LO16)
		if r.Type == objabi.R_ADDRPOWER_DS {
			loType = ld.MACHO_PPC_RELOC_LO14
		}
		t := uint32(r.Xadd)
		// Low instruction first, as in Apple's assembler output. PAIR's
		// high half is HI, not HA: carry correction belongs to HA16 only.
		emit(sectoff+4, false, 2, external, loType, symbol)
		emitPair(t >> 16)
		emit(sectoff, false, 2, external, ld.MACHO_PPC_RELOC_HA16, symbol)
		emitPair(t & 0xffff)
		return true

	case objabi.R_ADDRPOWER_PCREL:
		if r.Size != 8 || sect == nil || sectoff+4 >= 1<<24 {
			ldr.Errorf(s, "Mach-O scattered relocation requires a defined target and a section offset below 16MB")
			return false
		}
		// Go's paired PC-relative relocation subtracts the address of
		// the first instruction, for BOTH halves. PAIR r_value is that
		// address; the high/low remainder is carried in PAIR r_address.
		if ldr.SymSect(s) == nil {
			return false
		}
		p := int64(ldr.SymSect(s).Vaddr) + sectoff
		v := ldr.SymValue(rs)
		if p < 0 || p > 1<<32-1 || v < 0 || v > 1<<32-1 {
			ldr.Errorf(s, "Mach-O scattered relocation values exceed 32-bit object addresses")
			return false
		}
		t := uint32(v + r.Xadd - p)
		emitScattered := func(address uint32, typ uint32, value uint32) {
			out.Write32(machoScatteredWord(address, false, 2, typ))
			out.Write32(value)
		}
		emitScattered(uint32(sectoff+4), ld.MACHO_PPC_RELOC_LO16_SECTDIFF, uint32(v))
		emitScattered(t>>16, ld.MACHO_PPC_RELOC_PAIR, uint32(p))
		emitScattered(uint32(sectoff), ld.MACHO_PPC_RELOC_HA16_SECTDIFF, uint32(v))
		emitScattered(t&0xffff, ld.MACHO_PPC_RELOC_PAIR, uint32(p))
		return true
	}
	return false
}
