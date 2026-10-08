// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

// libSystem functions that exist on Leopard and stand in, inside the
// darwin/ppc64 trampolines of sys_darwin_ppc64.s, for modern functions that
// Leopard lacks: gettimeofday for clock_gettime, and arc4random for
// arc4random_buf.

//go:cgo_import_dynamic libc_gettimeofday gettimeofday "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_arc4random arc4random "/usr/lib/libSystem.B.dylib"

// libc function wrappers. Must run on system stack.

//go:nosplit
//go:cgo_unsafe_args
func g0_pthread_key_create(k *pthreadkey, destructor uintptr) int32 {
	ret := asmcgocall(unsafe.Pointer(abi.FuncPCABI0(pthread_key_create_trampoline)), unsafe.Pointer(&k))
	KeepAlive(k)
	return ret
}
func pthread_key_create_trampoline()

//go:nosplit
//go:cgo_unsafe_args
func g0_pthread_setspecific(k pthreadkey, value uintptr) int32 {
	return asmcgocall(unsafe.Pointer(abi.FuncPCABI0(pthread_setspecific_trampoline)), unsafe.Pointer(&k))
}
func pthread_setspecific_trampoline()

//go:cgo_import_dynamic libc_pthread_key_create pthread_key_create "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_setspecific pthread_setspecific "/usr/lib/libSystem.B.dylib"

// osyield is called from signal handlers: cpuprof.add spins on
// prof.signalLock with it when two SIGPROF handlers run at once. Elsewhere
// on darwin it is usleep(1), but Leopard's usleep is not async-signal-safe:
// nanosleep calls _pthread_testcancel, which takes the calling thread's own
// pthread spinlock. A handler that interrupted that locked region and then
// calls usleep spins forever (measured on the G5: hung 3 of 3 times within
// 2 s). sched_yield is the bare swtch_pri Mach trap there and takes no lock.

//go:nosplit
func osyield() {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sched_yield_trampoline)), nil)
}

//go:nosplit
func osyield_no_g() {
	asmcgocall_no_g(unsafe.Pointer(abi.FuncPCABI0(sched_yield_trampoline)), nil)
}
func sched_yield_trampoline()

//go:cgo_import_dynamic libc_sched_yield sched_yield "/usr/lib/libSystem.B.dylib"

// Used by the CPU profiler's thread sampler (cpuprof_darwin_ppc64.go).

//go:nosplit
//go:cgo_unsafe_args
func pthread_mach_thread_np(t pthread) (port uint32) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(pthread_mach_thread_np_trampoline)), unsafe.Pointer(&t))
	return
}
func pthread_mach_thread_np_trampoline()

//go:nosplit
//go:cgo_unsafe_args
func thread_info(thread uint32, flavor uint32, info *threadBasicInfo, count *uint32) int32 {
	ret := libcCall(unsafe.Pointer(abi.FuncPCABI0(thread_info_trampoline)), unsafe.Pointer(&thread))
	KeepAlive(info)
	KeepAlive(count)
	return ret
}
func thread_info_trampoline()

//go:cgo_import_dynamic libc_pthread_mach_thread_np pthread_mach_thread_np "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_thread_info thread_info "/usr/lib/libSystem.B.dylib"

// tlsinit allocates a thread-local storage slot for g and records its byte
// offset from the thread pointer (R13) in runtime.tls_g.
//
// The offset is computed from the layout libSystem itself uses and then
// checked, not assumed: a magic value stored with pthread_setspecific must be
// visible at tlsbase+offset, or the layout is not what the formula says and
// the runtime aborts.
//
// This runs at startup on the OS stack before g is set, so it must not split
// the stack (transitively). g is expected to be nil, so things (e.g.
// asmcgocall) will skip saving or reading g.
//
//go:nosplit
func tlsinit(tlsg *uintptr, tlsbase uintptr) {
	var k pthreadkey
	err := g0_pthread_key_create(&k, 0)
	if err != 0 {
		abort()
	}

	const magic = 0xc476c475c47957
	err = g0_pthread_setspecific(k, magic)
	if err != 0 {
		abort()
	}

	off := uintptr(_PTHREAD_TSD_OFFSET) + uintptr(k)*goarch.PtrSize
	if *(*uintptr)(unsafe.Pointer(tlsbase + off)) != magic {
		abort()
	}
	*tlsg = off
	g0_pthread_setspecific(k, 0)
}
