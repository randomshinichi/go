// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package pprof

// machVMRegionBasicInfoData matches vm_region_basic_info_data_64_t. On the
// PowerPC Leopard target, a C probe including <mach/vm_region.h> measured size
// 36 and offsets protection=0, offset=20, behavior=28, user_wired_count=32.
type machVMRegionBasicInfoData struct {
	Protection       int32
	Max_protection   int32
	Inheritance      uint32
	Shared           uint32
	Reserved         uint32
	Offset           [8]byte
	Behavior         int32
	User_wired_count uint16
	Pad_cgo_1        [2]byte
}

// VM_PROT flags from /usr/include/mach/vm_prot.h, read on the PowerPC Leopard
// target (Mac OS X 10.5.8): READ=0x01, WRITE=0x02, EXECUTE=0x04.
const (
	_VM_PROT_READ    = 0x1
	_VM_PROT_WRITE   = 0x2
	_VM_PROT_EXECUTE = 0x4

	_MACH_SEND_INVALID_DEST = 0x10000003
	_MAXPATHLEN             = 0x400
)
