// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package runtime

import (
	"internal/abi"
	"unsafe"
)

// Mach thread profiling helpers. These are libSystem calls, not cgo calls;
// their ppc32 Mach-O import/call path must be confirmed on Leopard hardware.

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
	return int32(ret)
}
func thread_info_trampoline()

//go:cgo_import_dynamic libc_pthread_mach_thread_np pthread_mach_thread_np "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_thread_info thread_info "/usr/lib/libSystem.B.dylib"
