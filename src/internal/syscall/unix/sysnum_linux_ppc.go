// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && ppc

package unix

const (
	getrandomTrap       uintptr = 359
	copyFileRangeTrap   uintptr = 379
	pidfdSendSignalTrap uintptr = 424
	pidfdOpenTrap       uintptr = 434
	openat2Trap         uintptr = 437
)
