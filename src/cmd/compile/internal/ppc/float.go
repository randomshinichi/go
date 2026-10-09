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

// emitIntToFloat builds 2^52 + x in memory, then subtracts the bias.
// Signed inputs flip their sign bit first and subtract 2^52 + 2^31.
// Every word is exact in binary64, including unsigned values above 2^31.
func emitIntToFloat(s *ssagen.State, dst, src int16, signed bool) {
	scratch := s.ScratchAddr(ppc.REGSP)
	low, bias := src, uint64(0x4330000000000000)
	if signed {
		p := s.Prog(ppc.AXORIS)
		p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0x8000}
		p.Reg = src
		p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R31}
		low, bias = ppc.REG_R31, 0x4330000080000000
	}
	memoryStore(s, ppc.ASTW, low, displaced(scratch, 4))
	emitConst32(s, ppc.REG_R31, 0x43300000)
	memoryStore(s, ppc.ASTW, ppc.REG_R31, scratch)
	memoryLoad(s, ppc.ALFD, dst, scratch)
	emitFloatBits(s, ppc.REG_F31, bias, displaced(scratch, 8))
	reg3(s, ppc.AFSUB, dst, dst, ppc.REG_F31)
}

// fctiwz is G3-legal and truncates toward zero. Its integer result is in the
// low word of the FPR image, at +4 in big-endian memory. Out-of-range/NaN
// results are deliberately not given a portable saturation contract.
func emitFloatToInt(s *ssagen.State, dst, src int16) {
	scratch := s.ScratchAddr(ppc.REGSP)
	reg2(s, ppc.AFCTIWZ, ppc.REG_F31, src)
	memoryStore(s, ppc.ASTFD, ppc.REG_F31, scratch)
	memoryLoad(s, ppc.ALWZ, dst, displaced(scratch, 4))
}

// fctiwz only has a signed result. Partition at 2^31; the upper range
// subtracts that exact bias, converts, and restores the unsigned high bit.
func emitFloatToUint(s *ssagen.State, dst, src int16) {
	scratch := s.ScratchAddr(ppc.REGSP)
	emitFloatBits(s, ppc.REG_F31, 0x41e0000000000000, scratch)
	p := s.Prog(ppc.AFCMPU)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: src}
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_F31}
	low := localBranch(s, 12, 0)
	reg3(s, ppc.AFSUB, ppc.REG_F30, src, ppc.REG_F31)
	reg2(s, ppc.AFCTIWZ, ppc.REG_F30, ppc.REG_F30)
	memoryStore(s, ppc.ASTFD, ppc.REG_F30, scratch)
	memoryLoad(s, ppc.ALWZ, dst, displaced(scratch, 4))
	p = s.Prog(ppc.AXORIS)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0x8000}
	p.Reg = dst
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: dst}
	over := s.Prog(ppc.AB)
	over.To.Type = obj.TYPE_BRANCH
	p = s.Prog(ppc.AFCTIWZ)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: src}
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_F30}
	low.To.SetTarget(p)
	memoryStore(s, ppc.ASTFD, ppc.REG_F30, scratch)
	memoryLoad(s, ppc.ALWZ, dst, displaced(scratch, 4))
	p = s.Prog(ppc.AOR)
	p.From = obj.Addr{Type: obj.TYPE_REG, Reg: dst}
	p.Reg = dst
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: dst}
	over.To.SetTarget(p)
}

func emitFloatBits(s *ssagen.State, r int16, bits uint64, scratch obj.Addr) {
	emitConst32(s, ppc.REG_R31, int32(bits>>32))
	memoryStore(s, ppc.ASTW, ppc.REG_R31, scratch)
	emitConst32(s, ppc.REG_R31, int32(bits))
	memoryStore(s, ppc.ASTW, ppc.REG_R31, displaced(scratch, 4))
	memoryLoad(s, ppc.ALFD, r, scratch)
}
