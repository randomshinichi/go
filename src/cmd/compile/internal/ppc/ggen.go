// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/objw"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/ppc"
)

// zerorange zeroes cnt bytes of stack starting at SP+off. R0 reads as the
// literal zero in the RA position of a store, so no scratch register is needed
// and the r0 argument is unused, exactly as on ppc64.
func zerorange(pp *objw.Progs, p *obj.Prog, off, cnt int64, _ *uint32) *obj.Prog {
	if cnt == 0 {
		return p
	}
	for i := int64(0); i < cnt; i += int64(types.PtrSize) {
		p = pp.Append(p, ppc.ASTW, obj.TYPE_REG, ppc.REG_R0, 0, obj.TYPE_MEM, ppc.REGSP, base.Ctxt.Arch.FixedFrameSize+off+i)
	}
	return p
}

func ginsnop(pp *objw.Progs) *obj.Prog {
	// The preferred hardware no-op: ori r0,r0,0.
	p := pp.Prog(ppc.AORI)
	p.From = obj.Addr{Type: obj.TYPE_CONST, Offset: 0}
	p.Reg = ppc.REG_R0
	p.To = obj.Addr{Type: obj.TYPE_REG, Reg: ppc.REG_R0}
	return p
}
