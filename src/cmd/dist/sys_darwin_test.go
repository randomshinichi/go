// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package main

import (
	"runtime"
	"testing"
)

// TestSysHostMachCPU reads the real hw.cputype and hw.cpusubtype of the
// machine running the test. On a Power Mac G5 they were measured as 18 and
// 100, which must map to ppc970; any other Darwin machine only has to answer.
func TestSysHostMachCPU(t *testing.T) {
	cputype, cpusubtype, err := sysHostMachCPU()
	if err != nil {
		t.Fatalf("sysHostMachCPU: %v", err)
	}
	t.Logf("hw.cputype = %d, hw.cpusubtype = %d (GOARCH %s)", cputype, cpusubtype, runtime.GOARCH)
	if runtime.GOARCH != "ppc64" {
		return
	}
	saved := hostMachCPU
	savedOS, savedArch := gohostos, gohostarch
	defer func() { hostMachCPU, gohostos, gohostarch = saved, savedOS, savedArch }()
	gohostos, gohostarch = "darwin", "ppc64"
	hostMachCPU = sysHostMachCPU
	if got, ok := probeHostPPC64Floor(); !ok || got != "ppc970" {
		t.Errorf("probeHostPPC64Floor() = %q, %v on a darwin/ppc64 machine reporting cputype %d, cpusubtype %d; want ppc970, true (the G5 reports 18, 100)",
			got, ok, cputype, cpusubtype)
	}
}
