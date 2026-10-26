// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcount_test

package syscall

import (
	"sync/atomic"
	_ "unsafe" // for linkname
)

// A per-subject trap counter, incremented without locks by the tagged native
// copies. Tests read only at quiescent boundaries; no thread state is changed.
var at349Count uint32

//go:linkname at349CountRead
func at349CountRead() uint32 { return atomic.LoadUint32(&at349Count) }
