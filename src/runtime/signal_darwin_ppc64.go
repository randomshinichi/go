// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) regs() *regs64 { return &(*ucontext)(c.ctxt).uc_mcontext.ss }

func (c *sigctxt) r0() uint64  { return c.regs().gpr[0].get() }
func (c *sigctxt) r1() uint64  { return c.regs().gpr[1].get() }
func (c *sigctxt) r2() uint64  { return c.regs().gpr[2].get() }
func (c *sigctxt) r3() uint64  { return c.regs().gpr[3].get() }
func (c *sigctxt) r4() uint64  { return c.regs().gpr[4].get() }
func (c *sigctxt) r5() uint64  { return c.regs().gpr[5].get() }
func (c *sigctxt) r6() uint64  { return c.regs().gpr[6].get() }
func (c *sigctxt) r7() uint64  { return c.regs().gpr[7].get() }
func (c *sigctxt) r8() uint64  { return c.regs().gpr[8].get() }
func (c *sigctxt) r9() uint64  { return c.regs().gpr[9].get() }
func (c *sigctxt) r10() uint64 { return c.regs().gpr[10].get() }
func (c *sigctxt) r11() uint64 { return c.regs().gpr[11].get() }
func (c *sigctxt) r12() uint64 { return c.regs().gpr[12].get() }
func (c *sigctxt) r13() uint64 { return c.regs().gpr[13].get() }
func (c *sigctxt) r14() uint64 { return c.regs().gpr[14].get() }
func (c *sigctxt) r15() uint64 { return c.regs().gpr[15].get() }
func (c *sigctxt) r16() uint64 { return c.regs().gpr[16].get() }
func (c *sigctxt) r17() uint64 { return c.regs().gpr[17].get() }
func (c *sigctxt) r18() uint64 { return c.regs().gpr[18].get() }
func (c *sigctxt) r19() uint64 { return c.regs().gpr[19].get() }
func (c *sigctxt) r20() uint64 { return c.regs().gpr[20].get() }
func (c *sigctxt) r21() uint64 { return c.regs().gpr[21].get() }
func (c *sigctxt) r22() uint64 { return c.regs().gpr[22].get() }
func (c *sigctxt) r23() uint64 { return c.regs().gpr[23].get() }
func (c *sigctxt) r24() uint64 { return c.regs().gpr[24].get() }
func (c *sigctxt) r25() uint64 { return c.regs().gpr[25].get() }
func (c *sigctxt) r26() uint64 { return c.regs().gpr[26].get() }
func (c *sigctxt) r27() uint64 { return c.regs().gpr[27].get() }
func (c *sigctxt) r28() uint64 { return c.regs().gpr[28].get() }
func (c *sigctxt) r29() uint64 { return c.regs().gpr[29].get() }
func (c *sigctxt) r30() uint64 { return c.regs().gpr[30].get() }
func (c *sigctxt) r31() uint64 { return c.regs().gpr[31].get() }
func (c *sigctxt) sp() uint64  { return c.r1() }

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) pc() uint64 { return c.regs().srr0.get() }

func (c *sigctxt) ctr() uint64    { return c.regs().ctr.get() }
func (c *sigctxt) link() uint64   { return c.regs().lr.get() }
func (c *sigctxt) xer() uint64    { return uint64(c.regs().xer) }
func (c *sigctxt) ccr() uint64    { return uint64(c.regs().cr) }
func (c *sigctxt) vrsave() uint32 { return c.regs().vrsave }

func (c *sigctxt) sigcode() uint32 { return uint32(c.info.si_code) }
func (c *sigctxt) sigaddr() uint64 { return uint64(uintptr(unsafe.Pointer(c.info.si_addr))) }
func (c *sigctxt) fault() uintptr  { return uintptr(c.sigaddr()) }

func (c *sigctxt) set_r0(x uint64)   { c.regs().gpr[0].set(x) }
func (c *sigctxt) set_r12(x uint64)  { c.regs().gpr[12].set(x) }
func (c *sigctxt) set_r30(x uint64)  { c.regs().gpr[30].set(x) }
func (c *sigctxt) set_pc(x uint64)   { c.regs().srr0.set(x) }
func (c *sigctxt) set_sp(x uint64)   { c.regs().gpr[1].set(x) }
func (c *sigctxt) set_link(x uint64) { c.regs().lr.set(x) }

func (c *sigctxt) set_sigcode(x uint32) { c.info.si_code = int32(x) }
func (c *sigctxt) set_sigaddr(x uint64) { c.info.si_addr = (*byte)(unsafe.Pointer(uintptr(x))) }

// No PPC64 signal-code correction has been established on Leopard. In
// particular, the amd64/arm64 breakpoint-PC workarounds are not PPC64 facts.
//
//go:nosplit
func (c *sigctxt) fixsigcode(sig uint32) {}

func dumpregs(c *sigctxt) {
	for i := range c.regs().gpr {
		print("r", i, " ", hex(c.regs().gpr[i].get()), "\n")
	}
	print("pc   ", hex(c.pc()), "\t")
	print("ctr  ", hex(c.ctr()), "\n")
	print("link ", hex(c.link()), "\t")
	print("xer  ", hex(c.xer()), "\n")
	print("ccr  ", hex(c.ccr()), "\t")
	print("vrsave ", hex(c.vrsave()), "\n")
}

// The following helpers follow signal_ppc64x.go's Go-to-Go stack conventions.
// They do not describe or implement the Apple C signal trampoline ABI.

//go:nosplit
//go:nowritebarrierrec
func (c *sigctxt) sigpc() uintptr    { return uintptr(c.pc()) }
func (c *sigctxt) setsigpc(x uint64) { c.set_pc(x) }

func (c *sigctxt) sigsp() uintptr { return uintptr(c.sp()) }
func (c *sigctxt) siglr() uintptr { return uintptr(c.link()) }

func (c *sigctxt) preparePanic(sig uint32, gp *g) {
	// Save LR even in leaf functions; execution will resume in sigpanic,
	// not in this overwritten frame.
	sp := c.sp() - sys.MinFrameSize
	c.set_sp(sp)
	*(*uint64)(unsafe.Pointer(uintptr(sp))) = c.link()

	pc := gp.sigpc
	if shouldPushSigpanic(gp, pc, uintptr(c.link())) {
		c.set_link(uint64(pc))
	}

	// Restore Go register invariants when panicking from external code.
	c.set_r0(0)
	c.set_r30(uint64(uintptr(unsafe.Pointer(gp))))
	c.set_r12(uint64(abi.FuncPCABIInternal(sigpanic)))
	c.set_pc(uint64(abi.FuncPCABIInternal(sigpanic)))
}

func (c *sigctxt) pushCall(targetPC, resumePC uintptr) {
	// asyncPreempt restores LR/SP and the saved R2/R12 from this frame.
	sp := c.sp() - sys.MinFrameSize
	c.set_sp(sp)
	*(*uint64)(unsafe.Pointer(uintptr(sp))) = c.link()
	*(*uint64)(unsafe.Pointer(uintptr(sp) + 8)) = c.r2()
	*(*uint64)(unsafe.Pointer(uintptr(sp) + 16)) = c.r12()
	c.set_link(uint64(resumePC))
	c.set_r12(uint64(targetPC))
	c.set_pc(uint64(targetPC))
}
