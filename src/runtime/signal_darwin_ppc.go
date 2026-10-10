// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime

import (
	"internal/abi"
	"internal/runtime/sys"
	"unsafe"
)

type sigctxt struct {
	info *siginfo
	ctxt unsafe.Pointer
}

func (c *sigctxt) regs() *regs32      { return &(*ucontext)(c.ctxt).uc_mcontext.ss }
func (c *sigctxt) pc() uintptr        { return uintptr(c.regs().srr0) }
func (c *sigctxt) sp() uintptr        { return uintptr(c.regs().gpr[1]) }
func (c *sigctxt) link() uintptr      { return uintptr(c.regs().lr) }
func (c *sigctxt) sigpc() uintptr     { return c.pc() }
func (c *sigctxt) sigsp() uintptr     { return c.sp() }
func (c *sigctxt) siglr() uintptr     { return c.link() }
func (c *sigctxt) setsigpc(x uint64)  { c.regs().srr0 = uint32(x) }
func (c *sigctxt) sigcode() uint32    { return uint32(c.info.si_code) }
func (c *sigctxt) sigaddr() uintptr   { return uintptr(unsafe.Pointer(c.info.si_addr)) }
func (c *sigctxt) fault() uintptr     { return c.sigaddr() }
func (c *sigctxt) set_pc(x uintptr)   { c.regs().srr0 = uint32(x) }
func (c *sigctxt) set_sp(x uintptr)   { c.regs().gpr[1] = uint32(x) }
func (c *sigctxt) set_link(x uintptr) { c.regs().lr = uint32(x) }
func (c *sigctxt) fixsigcode(uint32)  {}

func dumpregs(c *sigctxt) {
	for i, r := range c.regs().gpr {
		print("r", i, " ", hex(uintptr(r)), "\n")
	}
	print("pc   ", hex(c.pc()), "\tsp   ", hex(c.sp()), "\n")
	print("link ", hex(c.link()), "\n")
}

// preparePanic redirects execution to sigpanic and preserves the saved link
// register in a machine-word slot below the interrupted stack pointer.
func (c *sigctxt) preparePanic(sig uint32, gp *g) {
	// The faked call frame for sigpanic must be exactly the size the unwinder and
	// the GC scanner remove again: (*unwinder).next and scanframeworker both add
	// alignUp(MinFrameSize, StackAlign). ppc32's MinFrameSize is 24, so that is
	// 32; reserving the bare 24 left every signal-injected panic 8 bytes short of
	// the interrupted SP (and left SP 8-mod-16, breaking the frame alignment
	// invariant for everything below sigpanic). ppc64's MinFrameSize is already
	// 32, which is why only ppc32 was affected.
	sp := c.sp() - alignUp(sys.MinFrameSize, sys.StackAlign)
	c.set_sp(sp)
	*(*uint32)(unsafe.Pointer(sp)) = uint32(c.link())
	if shouldPushSigpanic(gp, gp.sigpc, c.link()) {
		c.set_link(gp.sigpc)
	}
	c.regs().gpr[0] = 0
	c.regs().gpr[30] = uint32(uintptr(unsafe.Pointer(gp)))
	c.regs().gpr[12] = uint32(abi.FuncPCABIInternal(sigpanic))
	c.set_pc(abi.FuncPCABIInternal(sigpanic))
}

func (c *sigctxt) pushCall(targetPC, resumePC uintptr) {
	// Reserve a 32-byte, 16-byte-aligned caller area. The first 24 bytes are
	// Darwin's fixed frame; the extra words preserve PIC state across the call.
	sp := c.sp() - (sys.MinFrameSize + 8)
	c.set_sp(sp)
	*(*uint32)(unsafe.Pointer(sp)) = uint32(c.link())
	*(*uint32)(unsafe.Pointer(sp + sys.MinFrameSize)) = c.regs().gpr[2]
	*(*uint32)(unsafe.Pointer(sp + sys.MinFrameSize + 4)) = c.regs().gpr[12]
	c.set_link(resumePC)
	c.regs().gpr[12] = uint32(targetPC)
	c.set_pc(targetPC)
}
