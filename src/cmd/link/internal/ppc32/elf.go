// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc32

import (
	"debug/elf"

	"cmd/internal/objabi"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
)

func elfreloc1(ctxt *ld.Link, out *ld.OutBuf, _ *loader.Loader, _ loader.Sym, r loader.ExtReloc, _ int, sectoff int64) bool {
	elfsym := uint32(ld.ElfSymForReloc(ctxt, r.Xsym))
	writeRela := func(offset int64, typ elf.R_PPC, addend int64) {
		out.Write32(uint32(offset))
		out.Write32(elf.R_INFO32(elfsym, uint32(typ)))
		out.Write32(uint32(addend))
	}

	switch r.Type {
	case objabi.R_ADDR, objabi.R_DWARFSECREF:
		if r.Size != 4 {
			return false
		}
		writeRela(sectoff, elf.R_PPC_ADDR32, r.Xadd)
	case objabi.R_ADDRPOWER:
		if r.Size != 8 {
			return false
		}
		writeRela(sectoff+2, elf.R_PPC_ADDR16_HA, r.Xadd)
		writeRela(sectoff+6, elf.R_PPC_ADDR16_LO, r.Xadd)
	case objabi.R_CALLPOWER:
		if r.Size != 4 {
			return false
		}
		writeRela(sectoff, elf.R_PPC_REL24, r.Xadd)
	default:
		return false
	}
	return true
}

func elfsetupplt(ctxt *ld.Link, _ *loader.Loader, _ *loader.SymbolBuilder, _ *loader.SymbolBuilder, _ loader.Sym) {
	ld.Exitf("ppc32 ELF PLT linking is not supported")
}
