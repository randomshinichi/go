// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

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

func (c *sigctxt) regs() *sigcontext { return &(*ucontext)(c.ctxt).uc_mcontext }

func (c *sigctxt) r0() uint32  { return c.regs().mc_gregs[0] }
func (c *sigctxt) r1() uint32  { return c.regs().mc_gregs[1] }
func (c *sigctxt) r2() uint32  { return c.regs().mc_gregs[2] }
func (c *sigctxt) r3() uint32  { return c.regs().mc_gregs[3] }
func (c *sigctxt) r4() uint32  { return c.regs().mc_gregs[4] }
func (c *sigctxt) r5() uint32  { return c.regs().mc_gregs[5] }
func (c *sigctxt) r6() uint32  { return c.regs().mc_gregs[6] }
func (c *sigctxt) r7() uint32  { return c.regs().mc_gregs[7] }
func (c *sigctxt) r8() uint32  { return c.regs().mc_gregs[8] }
func (c *sigctxt) r9() uint32  { return c.regs().mc_gregs[9] }
func (c *sigctxt) r10() uint32 { return c.regs().mc_gregs[10] }
func (c *sigctxt) r11() uint32 { return c.regs().mc_gregs[11] }
func (c *sigctxt) r12() uint32 { return c.regs().mc_gregs[12] }
func (c *sigctxt) r13() uint32 { return c.regs().mc_gregs[13] }
func (c *sigctxt) r14() uint32 { return c.regs().mc_gregs[14] }
func (c *sigctxt) r15() uint32 { return c.regs().mc_gregs[15] }
func (c *sigctxt) r16() uint32 { return c.regs().mc_gregs[16] }
func (c *sigctxt) r17() uint32 { return c.regs().mc_gregs[17] }
func (c *sigctxt) r18() uint32 { return c.regs().mc_gregs[18] }
func (c *sigctxt) r19() uint32 { return c.regs().mc_gregs[19] }
func (c *sigctxt) r20() uint32 { return c.regs().mc_gregs[20] }
func (c *sigctxt) r21() uint32 { return c.regs().mc_gregs[21] }
func (c *sigctxt) r22() uint32 { return c.regs().mc_gregs[22] }
func (c *sigctxt) r23() uint32 { return c.regs().mc_gregs[23] }
func (c *sigctxt) r24() uint32 { return c.regs().mc_gregs[24] }
func (c *sigctxt) r25() uint32 { return c.regs().mc_gregs[25] }
func (c *sigctxt) r26() uint32 { return c.regs().mc_gregs[26] }
func (c *sigctxt) r27() uint32 { return c.regs().mc_gregs[27] }
func (c *sigctxt) r28() uint32 { return c.regs().mc_gregs[28] }
func (c *sigctxt) r29() uint32 { return c.regs().mc_gregs[29] }
func (c *sigctxt) r30() uint32 { return c.regs().mc_gregs[30] }
func (c *sigctxt) r31() uint32 { return c.regs().mc_gregs[31] }

func (c *sigctxt) pc() uint32   { return c.regs().mc_gregs[32] }
func (c *sigctxt) ctr() uint32  { return c.regs().mc_gregs[35] }
func (c *sigctxt) link() uint32 { return c.regs().mc_gregs[36] }
func (c *sigctxt) xer() uint32  { return c.regs().mc_gregs[37] }
func (c *sigctxt) ccr() uint32  { return c.regs().mc_gregs[38] }
func (c *sigctxt) trap() uint32 { return c.regs().mc_gregs[40] }
func (c *sigctxt) sp() uint32   { return c.regs().mc_gregs[1] }

func (c *sigctxt) sigpc() uintptr { return uintptr(c.pc()) }
func (c *sigctxt) sigsp() uintptr { return uintptr(c.sp()) }
func (c *sigctxt) siglr() uintptr { return uintptr(c.link()) }
func (c *sigctxt) setsigpc(x uint64) { c.regs().mc_gregs[32] = uint32(x) }
func (c *sigctxt) sigcode() uint32 { return uint32(c.info.si_code) }
func (c *sigctxt) sigaddr() uintptr { return uintptr(c.info.si_addr) }
// fault is the address the faulting instruction touched. Read it from the kernel's
// siginfo, not from mc_gregs[PT_DAR]: on Linux ppc32 the kernel reports the data
// fault address in siginfo.si_addr, and PT_DAR (mc_gregs[41]) is not filled for a
// userspace data fault -- measured under qemu-ppc 11.1.1 -cpu 7450, where a store
// to 0x40000000 yielded si_addr=0x40000000 with si_code=SEGV_MAPERR but DAR=0
// (out/ppc-mac/qemu-signal-evidence/faultfatal-diag-7450.log). sigpanic's
// "gp.sigcode1 < 0x1000" gate then turned every wild-pointer fault into a
// recoverable nil-pointer panic. darwin/ppc32 reads si_addr the same way.
func (c *sigctxt) fault() uintptr { return c.sigaddr() }

func (c *sigctxt) set_r0(x uint32)  { c.regs().mc_gregs[0] = x }
func (c *sigctxt) set_r2(x uint32)  { c.regs().mc_gregs[2] = x }
func (c *sigctxt) set_r12(x uint32) { c.regs().mc_gregs[12] = x }
func (c *sigctxt) set_r30(x uint32) { c.regs().mc_gregs[30] = x }
func (c *sigctxt) set_pc(x uintptr) { c.regs().mc_gregs[32] = uint32(x) }
func (c *sigctxt) set_sp(x uintptr) { c.regs().mc_gregs[1] = uint32(x) }
func (c *sigctxt) set_link(x uintptr) { c.regs().mc_gregs[36] = uint32(x) }
func (c *sigctxt) set_sigcode(x uint32) { c.info.si_code = int32(x) }
func (c *sigctxt) set_sigaddr(x uintptr) { c.info.si_addr = uint32(x) }

func dumpregs(c *sigctxt) {
	for i, r := range c.regs().mc_gregs[:32] {
		print("r", i, " ", hex(uintptr(r)), "\n")
	}
	print("pc   ", hex(c.sigpc()), "\tsp   ", hex(c.sigsp()), "\n")
	print("ctr  ", hex(uintptr(c.ctr())), "\tlink ", hex(uintptr(c.link())), "\n")
	print("xer  ", hex(uintptr(c.xer())), "\tccr  ", hex(uintptr(c.ccr())), "\n")
	print("trap ", hex(uintptr(c.trap())), "\n")
}

// preparePanic builds a caller frame for sigpanic and preserves the interrupted
// function's link register for traceback.
func (c *sigctxt) preparePanic(sig uint32, gp *g) {
	sp := c.sigsp() - alignUp(sys.MinFrameSize, sys.StackAlign)
	c.set_sp(sp)
	*(*uint32)(unsafe.Pointer(sp)) = c.link()
	if shouldPushSigpanic(gp, gp.sigpc, uintptr(c.link())) {
		c.set_link(gp.sigpc)
	}
	c.set_r0(0)
	c.set_r30(uint32(uintptr(unsafe.Pointer(gp))))
	c.set_r12(uint32(abi.FuncPCABIInternal(sigpanic)))
	c.set_pc(abi.FuncPCABIInternal(sigpanic))
}

func (c *sigctxt) pushCall(targetPC, resumePC uintptr) {
	sp := c.sigsp() - (sys.MinFrameSize + 8)
	c.set_sp(sp)
	*(*uint32)(unsafe.Pointer(sp)) = c.link()
	*(*uint32)(unsafe.Pointer(sp + sys.MinFrameSize)) = c.r2()
	*(*uint32)(unsafe.Pointer(sp + sys.MinFrameSize + 4)) = c.r12()
	c.set_link(resumePC)
	c.set_r12(uint32(targetPC))
	c.set_pc(targetPC)
}
