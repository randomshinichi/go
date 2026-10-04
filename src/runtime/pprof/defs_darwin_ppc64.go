// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file has the layout cgo -godefs would emit for
// vm_region_basic_info_data_64_t from the Mac OS X 10.5.8 SDK headers.
// cgo is unavailable in this build, so every value here was MEASURED ON THE
// TARGET (Power Mac G5 Quad, Darwin 9.8.0 / 10.5.8, ppc64) with
// out/stage2-20261003/parent-native-sched/probe-pprof/vminfo.c, which printed:
//
//	VM_PROT_READ 1   VM_PROT_WRITE 2   VM_PROT_EXECUTE 4
//	MACH_SEND_INVALID_DEST 0x10000003   MAXPATHLEN 1024
//	sizeof_data64 36   off_offset 20   off_user_wired_count 32
//	off_protection 0  off_max_protection 4  off_inheritance 8
//	off_shared 12     off_reserved 16       off_behavior 28
//
// The measured size (36) and every offset agree with defs_darwin_amd64.go, so the
// 64-bit Darwin layout is shared; the values are not copied on trust.
// As on amd64, Offset is [8]byte rather than uint64 because the field is not
// naturally aligned (it sits at 20).

package pprof

type machVMRegionBasicInfoData struct {
	Protection       int32
	Max_protection   int32
	Inheritance      uint32
	Shared           uint32
	Reserved         uint32
	Offset           [8]byte // measured at offset 20; cannot use uint64 due to alignment
	Behavior         int32
	User_wired_count uint16
	Pad_cgo_1        [2]byte
}

const (
	_VM_PROT_READ    = 0x1
	_VM_PROT_WRITE   = 0x2
	_VM_PROT_EXECUTE = 0x4

	_MACH_SEND_INVALID_DEST = 0x10000003

	_MAXPATHLEN = 0x400
)
