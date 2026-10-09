// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm || 386 || mips || mipsle

package runtime

//go:noescape
func _mul64by32(lo64 *uint64, a uint64, b uint32) (hi32 uint32)
