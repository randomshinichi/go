// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/ssagen"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc"
)

func memoryLoad(s *ssagen.State, as obj.As, r int16, addr obj.Addr) {
	p := s.Prog(as)
	p.From = addr
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: r}
}
func memoryStore(s *ssagen.State, as obj.As, r int16, addr obj.Addr) {
	p := s.Prog(as)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: r}
	p.To = addr
}
func displaced(addr obj.Addr, off int64) obj.Addr {
	addr.Offset += off
	return addr
}

// Go aligns double slots to four bytes. Only the private scratch slot is
// provably eight-byte aligned (including the final allocated frame). Never
// issue lfd/stfd directly to a Go slot, even for spills or stack arguments.
func emitFloatLoad(s *ssagen.State, r int16, addr obj.Addr, size int64) {
	if size == 4 {
		memoryLoad(s, ppc.ALFS, r, addr)
		return
	}
	scratch := s.ScratchAddr(ppc.REGSP)
	for off := int64(0); off < 8; off += 4 {
		memoryLoad(s, ppc.ALWZ, ppc.REG_R31, displaced(addr, off))
		memoryStore(s, ppc.ASTW, ppc.REG_R31, displaced(scratch, off))
	}
	memoryLoad(s, ppc.ALFD, r, scratch)
}
func emitFloatStore(s *ssagen.State, r int16, addr obj.Addr, size int64) {
	if size == 4 {
		memoryStore(s, ppc.ASTFS, r, addr)
		return
	}
	scratch := s.ScratchAddr(ppc.REGSP)
	memoryStore(s, ppc.ASTFD, r, scratch)
	for off := int64(0); off < 8; off += 4 {
		memoryLoad(s, ppc.ALWZ, ppc.REG_R31, displaced(scratch, off))
		memoryStore(s, ppc.ASTW, ppc.REG_R31, displaced(addr, off))
	}
}
func emitFloatBits(s *ssagen.State, r int16, bits uint64, scratch obj.Addr) {
	emitConst32(s, ppc.REG_R31, int32(bits>>32))
	memoryStore(s, ppc.ASTW, ppc.REG_R31, scratch)
	emitConst32(s, ppc.REG_R31, int32(bits))
	memoryStore(s, ppc.ASTW, ppc.REG_R31, displaced(scratch, 4))
	memoryLoad(s, ppc.ALFD, r, scratch)
}
