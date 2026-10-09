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
	"testing"
)

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
