// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ppc32 implements the Darwin PowerPC 32-bit linker path.
package ppc32

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
)

func Init() (*sys.Arch, ld.Arch) {
	return sys.ArchPPC, ld.Arch{
		Funcalign: 4, Maxalign: 32, Minalign: 4, TrampLimit: 0x1c00000,
		Adddynrel: adddynrel, Archinit: archinit, Archreloc: archreloc,
		Extreloc: extreloc, Gentext: gentext,
		Machoreloc1: machoreloc1, MachorelocSize: 8,
	}
}

func archinit(ctxt *ld.Link) {
	if ctxt.HeadType != objabi.Hdarwin {
		ld.Exitf("ppc32 linker supports only darwin/ppc Mach-O")
	}
	ld.HEADR = ld.INITIAL_MACHO_HEADR
	if *ld.FlagRound == -1 {
		*ld.FlagRound = 4096
	}
	if *ld.FlagTextAddr == -1 {
		*ld.FlagTextAddr = ld.Rnd(0x1000000, *ld.FlagRound) + int64(ld.HEADR)
	}
}

func archreloc(target *ld.Target, ldr *loader.Loader, _ *ld.ArchSyms, r loader.Reloc, s loader.Sym, val int64) (int64, int, bool) {
	switch r.Type() {
	case objabi.R_CALLPOWER:
		delta := ldr.SymValue(r.Sym()) + r.Add() - (ldr.SymValue(s) + int64(r.Off()))
		if delta&3 != 0 || int64(int32(delta<<6)>>6) != delta {
			ldr.Errorf(s, "ppc32 branch relocation out of range or unaligned: %d", delta)
		}
		return val&^0x03fffffc | int64(uint32(delta)&0x03fffffc), 0, true
	case objabi.R_ADDRPOWER:
		if r.Siz() != 8 {
			return val, 0, false
		}
		t := ldr.SymValue(r.Sym()) + r.Add()
		hi := uint32(val >> 32)
		lo := uint32(val)
		hi = hi&^0xffff | uint32((t+0x8000)>>16)&0xffff
		lo = lo&^0xffff | uint32(t)&0xffff
		return int64(hi)<<32 | int64(lo), 0, true
	default:
		return val, 0, false
	}
}

// adddynrel rejects dynamic relocations until the ppc32 dynamic relocation
// and call-stub machinery is implemented.
func adddynrel(_ *ld.Target, ldr *loader.Loader, _ *ld.ArchSyms, s loader.Sym, _ loader.Reloc, _ int) bool {
	ldr.Errorf(s, "ppc32 dynamic relocations are not supported")
	return false
}

// extreloc rejects external linking until ppc32 Mach-O external relocations
// are implemented. Returning false alone would silently omit the relocation.
func extreloc(_ *ld.Target, ldr *loader.Loader, r loader.Reloc, s loader.Sym) (loader.ExtReloc, bool) {
	ldr.Errorf(s, "ppc32 external relocations are not supported (type %d)", r.Type())
	return loader.ExtReloc{}, false
}

// gentext is intentionally empty for the initial ppc32 internal-link path:
// it has no dynamic call stubs or other generated text yet. Unsupported
// external relocations still fail in machoreloc1 rather than being emitted
// without the required stub machinery.
func gentext(_ *ld.Link, _ *loader.Loader) {}

func machoreloc1(_ *sys.Arch, _ *ld.OutBuf, _ *loader.Loader, _ loader.Sym, _ loader.ExtReloc, _ int64) bool {
	// This milestone links internally; external Mach-O relocations are not yet
	// supported and must fail rather than silently produce incomplete objects.
	return false
}
