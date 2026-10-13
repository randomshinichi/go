// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package atomic

import "unsafe"

var lock64 uint32

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

//go:nosplit
func Xadd(ptr *uint32, delta int32) (new uint32) {
	for {
		old := loadRaw(ptr)
		new = old + uint32(delta)
		if casRaw(ptr, old, new) {
			return
		}
	}
}

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

// The 32-bit target has no native 64-bit reservation instruction. All 64-bit
// operations serialize through lock64 and use ordinary 32-bit-pair accesses.
//
// lock64 is a plain spinlock and is NOT reentrant: a signal handler that runs
// on the thread that holds it, and then does a 64-bit atomic of its own,
// spins forever (goppc-452y). The SIGPROF handler does exactly that through
// profBuf.write, so sigprof asks Lock64Held first and drops the sample when
// the answer is yes. Every other 64-bit atomic reachable from a signal handler
// has the same hazard and needs the same care.
//
// lock64Acquire and lock64Release are linknamed (push) only so the runtime's
// test can hold the lock across a self-directed SIGPROF; see
// runtime/proflock64_ppc_test.go.
//
//go:linkname lock64Acquire
//go:linkname lock64Release
//go:nosplit
func lock64Acquire() {
	for !casRaw(&lock64, 0, 1) {
	}
}

//go:nosplit
func lock64Release() { storeRaw(&lock64, 0) }

// Lock64Held reports whether some thread currently holds lock64. A signal
// handler that is about to do a 64-bit atomic calls it first: if the answer is
// yes, the holder may be the very thread the signal interrupted, and waiting
// for the lock would wait for the handler's own caller. The test is exact in
// that direction: if the interrupted thread holds lock64, the word stays 1 for
// the whole handler, so a handler that sees 0 cannot be the lock's blocked
// owner. It can say yes when another thread holds the lock; the caller then
// drops work it could have done, which is a lost sample, not a hang.
//
//go:nosplit
func Lock64Held() bool { return loadRaw(&lock64) != 0 }

//go:nosplit
func check64(ptr *uint64) {
	if uintptr(unsafe.Pointer(ptr))&7 != 0 {
		panicUnaligned()
	}
}

//go:nosplit
func Cas64(ptr *uint64, old, new uint64) bool {
	check64(ptr)
	lock64Acquire()
	ok := *ptr == old
	if ok {
		*ptr = new
	}
	lock64Release()
	return ok
}

//go:nosplit
func Load64(ptr *uint64) uint64 { check64(ptr); lock64Acquire(); v := *ptr; lock64Release(); return v }

//go:nosplit
func LoadAcq64(ptr *uint64) uint64 { return Load64(ptr) }

//go:nosplit
func Store64(ptr *uint64, val uint64) { check64(ptr); lock64Acquire(); *ptr = val; lock64Release() }

//go:nosplit
func StoreRel64(ptr *uint64, val uint64) { Store64(ptr, val) }

//go:nosplit
func Xchg64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	lock64Acquire()
	old = *ptr
	*ptr = val
	lock64Release()
	return
}

//go:nosplit
func Xadd64(ptr *uint64, delta int64) (new uint64) {
	check64(ptr)
	lock64Acquire()
	new = *ptr + uint64(delta)
	*ptr = new
	lock64Release()
	return
}

//go:nosplit
func And64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	lock64Acquire()
	old = *ptr
	*ptr &= val
	lock64Release()
	return
}

//go:nosplit
func Or64(ptr *uint64, val uint64) (old uint64) {
	check64(ptr)
	lock64Acquire()
	old = *ptr
	*ptr |= val
	lock64Release()
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
