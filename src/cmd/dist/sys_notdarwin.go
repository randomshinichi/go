// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin

package main

import "errors"

// sysHostMachCPU has nothing to read off Darwin.
func sysHostMachCPU() (cputype, cpusubtype uint32, err error) {
	return 0, 0, errors.New("hw.cputype is only available on Darwin")
}
