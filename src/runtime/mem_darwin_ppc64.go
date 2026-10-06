// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//
//go:nosplit
func sysAllocOS(n uintptr, _ string) unsafe.Pointer {
	v, err := mmap(nil, n, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		return nil
	}
	return v
}

// sysUnusedOS uses Leopard's MS_KILLPAGES because its kernel rejects
// MADV_FREE with EINVAL. On xnu-1228 this moves eligible pages to the inactive
// queue while leaving the mapping valid, making them reclaimable by the kernel.
// The operation is conditional: xnu skips objects with ref_count != 1,
// needs_copy, or a shadow, and skips wired, busy, or gobbled pages even when
// msync succeeds. A successful call therefore does not mean immediate RSS reduction
// or guarantee that every page in the range was reclaimed. The G5 measurement
// found a 1 GiB anonymous mapping moved active-to-inactive with no pressure.
func sysUnusedOS(v unsafe.Pointer, n uintptr) {
	if msync(v, n, _MS_KILLPAGES) != 0 {
		print("runtime: msync(", v, ", ", n, ", MS_KILLPAGES) failed\n")
		throw("runtime: cannot mark unused pages reclaimable")
	}
}

func sysUsedOS(v unsafe.Pointer, n uintptr) {
}

func sysHugePageOS(v unsafe.Pointer, n uintptr) {
}

func sysNoHugePageOS(v unsafe.Pointer, n uintptr) {
}

func sysHugePageCollapseOS(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//
//go:nosplit
func sysFreeOS(v unsafe.Pointer, n uintptr) {
	munmap(v, n)
}

func sysFaultOS(v unsafe.Pointer, n uintptr) {
	mmap(v, n, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE|_MAP_FIXED, -1, 0)
}

func sysReserveOS(v unsafe.Pointer, n uintptr, _ string) unsafe.Pointer {
	p, err := mmap(v, n, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		return nil
	}
	return p
}

const _ENOMEM = 12

func sysMapOS(v unsafe.Pointer, n uintptr, _ string) {
	p, err := mmap(v, n, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_FIXED|_MAP_PRIVATE, -1, 0)
	if err == _ENOMEM {
		throw("runtime: out of memory")
	}
	if p != v || err != 0 {
		print("runtime: mmap(", v, ", ", n, ") returned ", p, ", ", err, "\n")
		throw("runtime: cannot map pages in arena address space")
	}
}

func needZeroAfterSysUnusedOS() bool {
	return true
}
