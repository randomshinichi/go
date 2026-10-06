// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package main

import "syscall"

// sysHostMachCPU reads hw.cputype and hw.cpusubtype. Both are 4-byte integers;
// SysctlUint32 reads them with the width checked, whereas syscall.Sysctl
// returns a byte string that is NUL-trimmed, which misreads a little-endian
// value such as 0x12 00 00 00.
func sysHostMachCPU() (cputype, cpusubtype uint32, err error) {
	cputype, err = syscall.SysctlUint32("hw.cputype")
	if err != nil {
		return 0, 0, err
	}
	cpusubtype, err = syscall.SysctlUint32("hw.cpusubtype")
	if err != nil {
		return 0, 0, err
	}
	return cputype, cpusubtype, nil
}
