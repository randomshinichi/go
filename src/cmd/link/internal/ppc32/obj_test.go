// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc32

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"fmt"
	"testing"
)

func TestAbsoluteAddressRelocation(t *testing.T) {
	const (
		lisR5  = uint32(0x3ca00000) // lis r5,0
		addiR5 = uint32(0x38a50000) // addi r5,r5,0
	)
	for _, addr := range []uint32{0x12347fff, 0x12348000, 0x1234ffff, 0xf2348000} {
		t.Run(fmt.Sprintf("%08x", addr), func(t *testing.T) {
			ldr := loader.NewLoader(0, &loader.ErrorReporter{})
			targetSym := ldr.CreateSymForUpdate("runtime.target", 0)
			targetSym.SetValue(int64(uint64(addr)))
			src := ldr.CreateSymForUpdate("runtime.address", 0)
			r, _ := src.AddRel(objabi.R_ADDRPOWER)
			r.SetSiz(8)
			r.SetSym(targetSym.Sym())
			pair := uint64(lisR5)<<32 | uint64(addiR5)
			got, _, ok := archreloc(nil, ldr, nil, r, src.Sym(), int64(pair))
			if !ok {
				t.Fatal("archreloc rejected R_ADDRPOWER absolute-address pair")
			}
			gotPair := uint64(got)
			hi, lo := uint32(gotPair>>32), uint32(gotPair)
			if hi&0xffff0000 != lisR5 || lo&0xffff0000 != addiR5 {
				t.Fatalf("instruction opcodes changed: got %#08x %#08x", hi, lo)
			}
			// addis uses a signed high immediate shifted left 16; addi adds
			// a signed low immediate. The high adjustment must carry exactly
			// when the low half has bit 15 set.
			materialized := uint32((int32(int16(uint16(hi))) << 16) + int32(int16(uint16(lo))))
			if materialized != addr {
				t.Fatalf("materialized address = %#08x, want %#08x (pair %#08x %#08x)", materialized, addr, hi, lo)
			}
		})
	}
}

func TestMachOImportCallStub(t *testing.T) {
	ldr := loader.NewLoader(0, &loader.ErrorReporter{})
	imp := ldr.CreateSymForUpdate("libc_write", 0)
	imp.SetType(sym.SDYNIMPORT)
	src := ldr.CreateSymForUpdate("runtime.write_trampoline", 0)
	src.SetType(sym.STEXT)
	src.SetReachable(true)
	r, _ := src.AddRel(objabi.R_CALLPOWER)
	r.SetOff(0)
	r.SetSiz(4)
	r.SetSym(imp.Sym())

	ctxt := &ld.Link{Target: ld.Target{Arch: sys.ArchPPC, HeadType: objabi.Hdarwin, LinkMode: ld.LinkInternal}, Textp: []loader.Sym{src.Sym()}}
	gentext(ctxt, ldr)

	srcRelocs := ldr.Relocs(src.Sym())
	got := srcRelocs.At(0)
	stubSym := got.Sym()
	if name := ldr.SymName(stubSym); name != "_macho_callstub.libc_write" {
		t.Fatalf("call target = %q, want generated Mach-O stub", name)
	}
	stubRelocs := ldr.Relocs(stubSym)
	if stubRelocs.Count() != 1 {
		t.Fatalf("stub relocations = %d, want 1", stubRelocs.Count())
	}
	stubReloc := stubRelocs.At(0)
	if stubReloc.Type() != objabi.R_ADDRPOWER || stubReloc.Sym() != imp.Sym() || stubReloc.Siz() != 8 {
		t.Fatalf("stub relocation = (%v, %q, %d), want R_ADDRPOWER libc_write size 8", stubReloc.Type(), ldr.SymName(stubReloc.Sym()), stubReloc.Siz())
	}
	data := ldr.Data(stubSym)
	want := []uint32{opLisR12, opLwzR12R12, opMtctrR12, opBctr}
	for i, insn := range want {
		if got := ctxt.Arch.ByteOrder.Uint32(data[i*4:]); got != insn {
			t.Errorf("stub instruction %d = %#08x, want %#08x", i, got, insn)
		}
	}
}
