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
	"cmd/link/internal/sym"
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
		// addis plus addi consumes a signed low half, so carry into the
		// high half when bit 15 of the low half is set (@ha/@l semantics).
		hi = hi&^0xffff | uint32((t+0x8000)>>16)&0xffff
		lo = lo&^0xffff | uint32(t)&0xffff
		return int64(hi)<<32 | int64(lo), 0, true
	default:
		return val, 0, false
	}
}

// adddynrel redirects the address in a generated import-call stub through
// the Mach-O non-lazy symbol pointer table. Other dynamic relocations remain
// errors rather than being silently dropped.
func adddynrel(target *ld.Target, ldr *loader.Loader, syms *ld.ArchSyms, s loader.Sym, r loader.Reloc, rIdx int) bool {
	if !target.IsExternal() && ldr.SymType(r.Sym()) == sym.SDYNIMPORT && r.Type() == objabi.R_ADDRPOWER {
		if r.Add() != 0 {
			ldr.Errorf(s, "Mach-O dynamic call requires a zero addend: %s%+d", ldr.SymName(r.Sym()), r.Add())
			return false
		}
		if r.Siz() != 8 {
			ldr.Errorf(s, "Mach-O dynamic call has invalid relocation size %d", r.Siz())
			return false
		}
		su := ldr.MakeSymbolUpdater(s)
		data := su.Data()[r.Off():]
		if target.Arch.ByteOrder.Uint32(data) != opLisR12 || target.Arch.ByteOrder.Uint32(data[4:]) != opLwzR12R12 {
			ldr.Errorf(s, "Mach-O dynamic call stub is not lis;lwz for %s", ldr.SymName(r.Sym()))
			return false
		}
		targ := r.Sym()
		ld.AddGotSym(target, ldr, syms, targ, 0)
		su.SetRelocSym(rIdx, syms.GOT)
		su.SetRelocAdd(rIdx, int64(ldr.SymGot(targ)))
		return true
	}
	ldr.Errorf(s, "ppc32 dynamic relocations are not supported")
	return false
}

// extreloc rejects external linking until ppc32 Mach-O external relocations
// are implemented. Returning false alone would silently omit the relocation.
func extreloc(_ *ld.Target, ldr *loader.Loader, r loader.Reloc, s loader.Sym) (loader.ExtReloc, bool) {
	ldr.Errorf(s, "ppc32 external relocations are not supported (type %d)", r.Type())
	return loader.ExtReloc{}, false
}

// gentext creates a 32-bit Mach-O call stub for each call to an imported
// function. The stub loads the function pointer from the non-lazy symbol
// pointer table, then branches through CTR. dyld fills that table at launch.
func gentext(ctxt *ld.Link, ldr *loader.Loader) {
	if ctxt.LinkMode != ld.LinkInternal || !ctxt.IsDarwin() {
		return
	}

	var stubs []loader.Sym
	for _, s := range ctxt.Textp {
		relocs := ldr.Relocs(s)
		for i := 0; i < relocs.Count(); i++ {
			r := relocs.At(i)
			if r.Type() != objabi.R_CALLPOWER || ldr.SymType(r.Sym()) != sym.SDYNIMPORT {
				continue
			}
			if r.Add() != 0 {
				ldr.Errorf(s, "Mach-O dynamic call requires a zero addend: %s%+d", ldr.SymName(r.Sym()), r.Add())
				continue
			}

			targ := r.Sym()
			stub := ldr.CreateSymForUpdate("_macho_callstub."+ldr.SymName(targ), 0)
			if stub.Size() == 0 {
				stub.SetType(sym.STEXT)
				stub.SetReachable(true)
				// The dynamic relocation pass allocates the non-lazy GOT
				// slot after Mach-O symbol IDs have been assigned.
				stub.AddSymRef(ctxt.Arch, targ, 0, objabi.R_ADDRPOWER, 8)
				stub.SetUint32(ctxt.Arch, 0, opLisR12)
				stub.SetUint32(ctxt.Arch, 4, opLwzR12R12)
				stub.AddUint32(ctxt.Arch, opMtctrR12)
				stub.AddUint32(ctxt.Arch, opBctr)
				stubs = append(stubs, stub.Sym())
			}
			ldr.MakeSymbolUpdater(s).SetRelocSym(i, stub.Sym())
		}
	}
	ctxt.Textp = append(stubs, ctxt.Textp...)
}

const (
	opLisR12    = 0x3d800000 // lis r12,0
	opLwzR12R12 = 0x818c0000 // lwz r12,0(r12)
	opMtctrR12  = 0x7d8903a6 // mtctr r12
	opBctr      = 0x4e800420 // bctr
)

func machoreloc1(_ *sys.Arch, _ *ld.OutBuf, _ *loader.Loader, _ loader.Sym, _ loader.ExtReloc, _ int64) bool {
	// This milestone links internally; external Mach-O relocations are not yet
	// supported and must fail rather than silently produce incomplete objects.
	return false
}
