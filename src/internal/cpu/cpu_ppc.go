// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package cpu

// CacheLinePadSize covers the 128-byte cache line documented for the 970.
// The 750 manual specifies a 32-byte line; a 7400-specific line size has not
// been established here. Use the larger 970 value for the G3/G4/G5 target
// range, avoiding under-padding on the established G5 target.
const CacheLinePadSize = 128

func doinit() {}
