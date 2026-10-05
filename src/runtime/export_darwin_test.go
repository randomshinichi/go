// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"unsafe"
)

var SetNonblock = setNonblock

// MachAbsoluteTime reads mach_absolute_time and the cached mach_timebase_info
// pair through nanotime_trampoline, the same call nanotime1 makes, but returns
// the raw values without converting them. It exists so tests can bracket
// Nanotime with independently converted readings; the struct layout must match
// nanotime1's.
func MachAbsoluteTime() (ticks uint64, numer, denom uint32) {
	var r struct {
		t            int64
		numer, denom uint32
	}
	libcCall(unsafe.Pointer(abi.FuncPCABI0(nanotime_trampoline)), unsafe.Pointer(&r))
	return uint64(r.t), r.numer, r.denom
}
