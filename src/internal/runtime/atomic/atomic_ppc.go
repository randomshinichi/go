// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package atomic

import (
	"internal/cpu"
	"unsafe"
)

//go:noescape
func casRaw(ptr *uint32, old, new uint32) bool

//go:noescape
func loadRaw(ptr *uint32) uint32

//go:noescape
func storeRaw(ptr *uint32, val uint32)

// Export atomic operations used by sync and sync/atomic assembly.
//go:linkname And32
//go:linkname And64
//go:linkname Anduintptr
//go:linkname Cas
//go:linkname Cas64
//go:linkname Casuintptr
//go:linkname Load
//go:linkname Load64
//go:linkname LoadAcquintptr
//go:linkname Loadp
//go:linkname Loaduintptr
//go:linkname Or32
//go:linkname Or64
//go:linkname Oruintptr
//go:linkname Store
//go:linkname Store64
//go:linkname StoreReluintptr
//go:linkname Storeuintptr
//go:linkname Xadd
//go:linkname Xadd64
//go:linkname Xadduintptr
//go:linkname Xchg
//go:linkname Xchg64
//go:linkname Xchguintptr

//go:nosplit
func Cas(ptr *uint32, old, new uint32) bool { return casRaw(ptr, old, new) }

//go:nosplit
func CasRel(ptr *uint32, old, new uint32) bool { return casRaw(ptr, old, new) }

//go:nosplit
func Load(ptr *uint32) uint32 { return loadRaw(ptr) }

//go:nosplit
func LoadAcq(ptr *uint32) uint32 { return loadRaw(ptr) }

//go:nosplit
func Load8(ptr *uint8) uint8 {
	word, shift := byteLane(ptr)
	return uint8(loadRaw(word) >> shift)
}

//go:nosplit
func Store(ptr *uint32, val uint32) { storeRaw(ptr, val) }

//go:nosplit
func StoreRel(ptr *uint32, val uint32) { storeRaw(ptr, val) }

//go:nosplit
func Store8(ptr *uint8, val uint8) {
	word, shift := byteLane(ptr)
	mask := uint32(0xff) << shift
	for {
		old := loadRaw(word)
		if casRaw(word, old, old&^mask|uint32(val)<<shift) {
			return
		}
	}
}

//go:nosplit
func Xchg(ptr *uint32, val uint32) (old uint32) {
	for {
		old = loadRaw(ptr)
		if casRaw(ptr, old, val) {
			return
		}
	}
}

//go:nosplit
func Xchg8(ptr *uint8, val uint8) (old uint8) {
	word, shift := byteLane(ptr)
	mask := uint32(0xff) << shift
	for {
		w := loadRaw(word)
		old = uint8(w >> shift)
		if casRaw(word, w, w&^mask|uint32(val)<<shift) {
			return
		}
	}
}

//go:noescape
// Xadd is implemented in assembly to keep the reservation loop frameless.
func Xadd(ptr *uint32, delta int32) (new uint32)

//go:nosplit
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr {
	return uintptr(Xadd((*uint32)(unsafe.Pointer(ptr)), int32(delta)))
}

//go:nosplit
func And(ptr *uint32, val uint32) { andOr(ptr, val, false) }

//go:nosplit
func Or(ptr *uint32, val uint32) { andOr(ptr, val, true) }

//go:nosplit
func And8(ptr *uint8, val uint8) { byteAndOr(ptr, val, false) }

//go:nosplit
func Or8(ptr *uint8, val uint8) { byteAndOr(ptr, val, true) }

//go:nosplit
func And32(ptr *uint32, val uint32) (old uint32) { return andOrResult(ptr, val, false) }

//go:nosplit
func Or32(ptr *uint32, val uint32) (old uint32) { return andOrResult(ptr, val, true) }

//go:nosplit
func andOr(ptr *uint32, val uint32, or bool) { _ = andOrResult(ptr, val, or) }

//go:nosplit
func andOrResult(ptr *uint32, val uint32, or bool) (old uint32) {
	for {
		old = loadRaw(ptr)
		new := old & val
		if or {
			new = old | val
		}
		if casRaw(ptr, old, new) {
			return
		}
	}
}

//go:nosplit
func byteAndOr(ptr *uint8, val uint8, or bool) {
	word, shift := byteLane(ptr)
	mask := uint32(0xff) << shift
	for {
		old := loadRaw(word)
		b := old >> shift & 0xff
		if or {
			b |= uint32(val)
		} else {
			b &= uint32(val)
		}
		if casRaw(word, old, old&^mask|b<<shift) {
			return
		}
	}
}

//go:nosplit
func byteLane(ptr *uint8) (*uint32, uint) {
	a := uintptr(unsafe.Pointer(ptr))
	return (*uint32)(unsafe.Pointer(a &^ 3)), uint((a&3)^3) * 8
}

// The 32-bit target has no native 64-bit reservation instruction (lwarx and
// stwcx. cover one word), so every 64-bit operation runs under a spinlock built
// from them. Following the 32-bit ARM fallback (atomic_arm.go), the lock is
// chosen by the operand's address from a table of 57 stripes:
//
//	stripe = (uintptr(unsafe.Pointer(addr)) >> 3) % 57
//
// 57 is the upstream ARM table size; upstream records no reason for it (it is
// 3*19, not prime, and an address stride of 8*k reaches 57/gcd(k,57) stripes).
// The operand stays where it is and any number of operands share the table.
// Distinct operands that hash together only wait for each other. Every 64-bit
// operation on one operand (load, store, CAS, exchange, add, and, or) derives
// the same stripe from the same address, so they exclude one another; an
// operation on an operand with a different stripe is independent.
//
// Each stripe is one word padded to cpu.CacheLinePadSize (128 bytes here). An
// lwarx reservation covers a granule of at least a cache line on the 750 and
// 74xx (32-byte lines), and a store by another processor to the granule
// cancels it, so two stripes in one line would make their spinners cancel each
// other's stwcx. 128 bytes is the 970's line; ARM's 32 would be too small if
// this ppc32 code runs on a G5, and no smaller value has been established for
// the G4's 64-byte L2/L3 line. This costs 57*128 = 7296 bytes of BSS.
//
// Locking uses only lwarx/stwcx. and sync, which exist on the G3 (750) floor.
// casRaw brackets its reservation loop with sync, as storeRaw does its store:
// acquire and release are therefore full barriers, and the plain 32-bit loads
// and stores of the operand between them are ordered with every other
// operation that takes the same stripe. 64-bit accesses are never made outside
// the stripe, so no operation sees a torn value.
//
// A stripe is a plain spinlock and is NOT reentrant. A signal handler that
// runs on a thread holding a stripe, and then needs that stripe for a 64-bit
// atomic of its own, spins forever (goppc-452y). The SIGPROF handler does that
// through profBuf.write, so sigprof first asks Atomic64AnyStripeHeld and drops
// the sample when the answer is yes. See that function for why the test is
// "any stripe" and not "the stripe the handler needs".
//
// Operations nest no stripes: a critical section contains only plain loads and
// stores of the operand, so a thread holds at most one stripe at a time and
// there is no lock-order deadlock.
//
// atomic64StripeAcquire, atomic64StripeRelease, atomic64StripeAcquireAll,
// atomic64StripeReleaseAll and atomic64StripeIndex are linknamed (push) only so
// that runtime tests can hold stripes across a self-directed signal; see
// runtime/proflock64_ppc_test.go.
//
//go:linkname atomic64StripeAcquire
//go:linkname atomic64StripeRelease
//go:linkname atomic64StripeAcquireAll
//go:linkname atomic64StripeReleaseAll
//go:linkname atomic64StripeIndex

const atomic64Stripes = 57

type atomic64Stripe struct {
	l   uint32
	pad [cpu.CacheLinePadSize - 4]byte
}

var atomic64Table [atomic64Stripes]atomic64Stripe

// atomic64StripeIndex is the address-to-stripe mapping, and the only place it
// is written down.
//
//go:nosplit
func atomic64StripeIndex(ptr *uint64) int {
	return int((uintptr(unsafe.Pointer(ptr)) >> 3) % atomic64Stripes)
}

//go:nosplit
func atomic64Lock(ptr *uint64) *uint32 {
	l := &atomic64Table[atomic64StripeIndex(ptr)].l
	for !casRaw(l, 0, 1) {
	}
	return l
}

//go:nosplit
func atomic64Unlock(l *uint32) { storeRaw(l, 0) }

//go:nosplit
func atomic64StripeAcquire(ptr *uint64) { atomic64Lock(ptr) }

//go:nosplit
func atomic64StripeRelease(ptr *uint64) {
	atomic64Unlock(&atomic64Table[atomic64StripeIndex(ptr)].l)
}

// atomic64StripeAcquireAll holds every stripe, in ascending order. No
// operation holds one stripe while waiting for another, so this cannot
// deadlock against operations; it is for tests that need "no 64-bit atomic
// can proceed".
//
//go:nosplit
func atomic64StripeAcquireAll() {
	for i := range atomic64Table {
		for !casRaw(&atomic64Table[i].l, 0, 1) {
		}
	}
}

//go:nosplit
func atomic64StripeReleaseAll() {
	for i := range atomic64Table {
		atomic64Unlock(&atomic64Table[i].l)
	}
}

// Atomic64AnyStripeHeld reports whether any stripe is currently held. A signal
// handler that is about to do a 64-bit atomic calls it first and drops its
// work on a yes.
//
// Why this is sound. The handler runs on some thread T with signals blocked,
// and T makes no progress until the handler returns. If T was inside an
// operation it holds exactly one stripe (see above), and that stripe stays held
// for the whole handler: the word was written before the signal was taken,
// T reads its own writes in program order, and only T releases it. So the scan
// finds it set, whichever stripe it is and whichever operand the handler would
// have touched. If T held nothing, the handler may wait on another thread's
// stripe, but that thread's critical section is a few plain loads and stores
// that depend on nothing T holds, so the wait ends. The acquisition and release
// windows are covered by instruction order: the stripe is set once stwcx.
// succeeds, and cleared by the single stw in storeRaw; a signal before the first
// leaves T holding nothing (the kernel also cancels the reservation, so a
// signal inside the lwarx/stwcx. loop makes the loop retry), and a signal after
// the second leaves T holding nothing.
//
// Why it is not "the stripe the handler needs". That needs the handler's full
// operand set, which today is the fields of one profBuf but is whatever
// sigprof's callees reach. A missed operand is a hang, not a lost sample.
// "Any stripe" needs no such inventory. The price is the same one the global
// lock charged: it also answers yes when another thread holds an unrelated
// stripe, and the caller then drops a sample it could have taken. Those are
// counted drops, not hangs.
//
// The scan uses plain loads. It is exact for the calling thread's own stripe,
// which is the only one soundness needs; for other threads' stripes it is a
// racy snapshot, which can only add or omit a conservative drop.
//
//go:nosplit
func Atomic64AnyStripeHeld() bool {
	for i := range atomic64Table {
		if atomic64Table[i].l != 0 {
			return true
		}
	}
	return false
}

//go:nosplit
func check64(ptr *uint64) {
	if uintptr(unsafe.Pointer(ptr))&7 != 0 {
		panicUnaligned()
	}
	// Force the dereference before the caller takes a stripe. A nil (or otherwise
	// faulting) operand must panic while no stripe is held; otherwise the panic
	// unwinds out of a held stripe that nothing will release, and every later
	// 64-bit atomic that hashes there spins forever in atomic64Lock. Mirrors
	// atomic_mipsx.go's lockAndCheck ("force dereference before taking lock").
	// The loaded value is discarded; the real read/write happens under the stripe
	// in the caller.
	_ = *ptr
}

//go:nosplit
func Cas64(ptr *uint64, old, new uint64) bool {
	check64(ptr)
	l := atomic64Lock(ptr)
	ok := *ptr == old
	if ok {
		*ptr = new
	}
	atomic64Unlock(l)
	return ok
}

//go:nosplit
func Load64(ptr *uint64) uint64 {
	check64(ptr)
	l := atomic64Lock(ptr)
	v := *ptr
	atomic64Unlock(l)
	return v
}

//go:nosplit
func LoadAcq64(ptr *uint64) uint64 { return Load64(ptr) }

//go:nosplit
func Store64(ptr *uint64, val uint64) {
	check64(ptr)
	l := atomic64Lock(ptr)
	*ptr = val
	atomic64Unlock(l)
}

//go:nosplit
func StoreRel64(ptr *uint64, val uint64) { Store64(ptr, val) }

//go:nosplit
func Xchg64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	l := atomic64Lock(ptr)
	old = *ptr
	*ptr = val
	atomic64Unlock(l)
	return
}

//go:nosplit
func Xadd64(ptr *uint64, delta int64) (new uint64) {
	check64(ptr)
	l := atomic64Lock(ptr)
	new = *ptr + uint64(delta)
	*ptr = new
	atomic64Unlock(l)
	return
}

//go:nosplit
func And64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	l := atomic64Lock(ptr)
	old = *ptr
	*ptr &= val
	atomic64Unlock(l)
	return
}

//go:nosplit
func Or64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	l := atomic64Lock(ptr)
	old = *ptr
	*ptr |= val
	atomic64Unlock(l)
	return
}

//go:nosplit
func Anduintptr(ptr *uintptr, val uintptr) uintptr {
	return uintptr(And32((*uint32)(unsafe.Pointer(ptr)), uint32(val)))
}

//go:nosplit
func Oruintptr(ptr *uintptr, val uintptr) uintptr {
	return uintptr(Or32((*uint32)(unsafe.Pointer(ptr)), uint32(val)))
}

//go:nosplit
func LoadAcquintptr(ptr *uintptr) uintptr { return uintptr(LoadAcq((*uint32)(unsafe.Pointer(ptr)))) }

//go:nosplit
func Loaduintptr(ptr *uintptr) uintptr { return uintptr(Load((*uint32)(unsafe.Pointer(ptr)))) }

//go:nosplit
func Loaduint(ptr *uint) uint { return uint(Load((*uint32)(unsafe.Pointer(ptr)))) }

//go:nosplit
func Loadint32(ptr *int32) int32 { return int32(Load((*uint32)(unsafe.Pointer(ptr)))) }

//go:nosplit
func Loadint64(ptr *int64) int64 { return int64(Load64((*uint64)(unsafe.Pointer(ptr)))) }

//go:nosplit
func Storeint32(ptr *int32, val int32) { Store((*uint32)(unsafe.Pointer(ptr)), uint32(val)) }

//go:nosplit
func Storeint64(ptr *int64, val int64) { Store64((*uint64)(unsafe.Pointer(ptr)), uint64(val)) }

//go:nosplit
func Storeuintptr(ptr *uintptr, val uintptr) { Store((*uint32)(unsafe.Pointer(ptr)), uint32(val)) }

//go:nosplit
func StoreReluintptr(ptr *uintptr, val uintptr) {
	StoreRel((*uint32)(unsafe.Pointer(ptr)), uint32(val))
}

//go:nosplit
func Xaddint32(ptr *int32, delta int32) int32 {
	return int32(Xadd((*uint32)(unsafe.Pointer(ptr)), delta))
}

//go:nosplit
func Xaddint64(ptr *int64, delta int64) int64 {
	return int64(Xadd64((*uint64)(unsafe.Pointer(ptr)), delta))
}

//go:nosplit
func Casint32(ptr *int32, old, new int32) bool {
	return Cas((*uint32)(unsafe.Pointer(ptr)), uint32(old), uint32(new))
}

//go:nosplit
func Casint64(ptr *int64, old, new int64) bool {
	return Cas64((*uint64)(unsafe.Pointer(ptr)), uint64(old), uint64(new))
}

//go:nosplit
func Casuintptr(ptr *uintptr, old, new uintptr) bool {
	return Cas((*uint32)(unsafe.Pointer(ptr)), uint32(old), uint32(new))
}

//go:nosplit
func Xchgint32(ptr *int32, val int32) int32 {
	return int32(Xchg((*uint32)(unsafe.Pointer(ptr)), uint32(val)))
}

//go:nosplit
func Xchgint64(ptr *int64, val int64) int64 {
	return int64(Xchg64((*uint64)(unsafe.Pointer(ptr)), uint64(val)))
}

//go:nosplit
func Xchguintptr(ptr *uintptr, val uintptr) uintptr {
	return uintptr(Xchg((*uint32)(unsafe.Pointer(ptr)), uint32(val)))
}

// Not noescape: a successful compare-and-swap installs new in *ptr.
func Casp1(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool

//go:nosplit
func Loadp(ptr unsafe.Pointer) unsafe.Pointer { return unsafe.Pointer(uintptr(Load((*uint32)(ptr)))) }

// Not noescape: the stored pointer must escape, even without a write barrier.
func StorepNoWB(ptr unsafe.Pointer, val unsafe.Pointer)
