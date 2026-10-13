// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic_test

import (
	. "sync/atomic"
	"testing"
	"unsafe"
)

func TestPPC32Raw64OperandPlacement(t *testing.T) {
	// Raw 64-bit operands must escape on ppc32, as on the other 32-bit
	// ports. This does not change the ABI alignment of the int64 type.
	var signed struct{ before, i, after int64 }
	if p := uintptr(unsafe.Pointer(&signed.i)); p&7 != 0 {
		t.Fatalf("raw int64 operand at %#x is not 8-aligned", p)
	}
	if old := SwapInt64(&signed.i, 42); old != 0 || signed.i != 42 {
		t.Fatalf("SwapInt64: old=%d, new=%d", old, signed.i)
	}

	var unsigned struct{ before, i, after uint64 }
	if p := uintptr(unsafe.Pointer(&unsigned.i)); p&7 != 0 {
		t.Fatalf("raw uint64 operand at %#x is not 8-aligned", p)
	}
	if old := SwapUint64(&unsigned.i, 42); old != 0 || unsigned.i != 42 {
		t.Fatalf("SwapUint64: old=%d, new=%d", old, unsigned.i)
	}

	var typed struct {
		pad int32
		i   Int64
		u   Uint64
	}
	if off := unsafe.Offsetof(typed.i); off != 8 {
		t.Fatalf("typed Int64 offset=%d, want 8", off)
	}
	if uintptr(unsafe.Pointer(&typed.i))&7 != 0 || uintptr(unsafe.Pointer(&typed.u))&7 != 0 {
		t.Fatal("typed atomic operands are not 8-aligned")
	}
	typed.i.Store(42)
	typed.u.Store(42)
	if typed.i.Swap(7) != 42 || typed.i.Load() != 7 || typed.u.Swap(7) != 42 || typed.u.Load() != 7 {
		t.Fatal("typed atomic swap/load gave incorrect values")
	}
}

func TestPPC32Raw64FieldLayout(t *testing.T) {
	var signed struct {
		pad int32
		i   int64
	}
	if a := unsafe.Alignof(signed.i); a != 4 {
		t.Fatalf("int64 ABI alignment=%d, want 4", a)
	}
	if off := unsafe.Offsetof(signed.i); off != 4 {
		t.Fatalf("raw int64 offset=%d, want 4", off)
	}
	if uintptr(unsafe.Pointer(&signed))&7 != 0 || uintptr(unsafe.Pointer(&signed.i))&7 != 4 {
		t.Fatal("expected 8-aligned allocation with a 4-mod-8 raw int64 field")
	}
	// An aligned allocation base does not fix a raw field at offset 4.
	// The runtime must still reject these operands, rather than silently
	// accepting the caller's invalid placement.
	shouldPanic(t, "SwapInt64 raw field", func() { SwapInt64(&signed.i, 42) })

	var unsigned struct {
		pad int32
		i   uint64
	}
	if off := unsafe.Offsetof(unsigned.i); off != 4 {
		t.Fatalf("raw uint64 offset=%d, want 4", off)
	}
	if uintptr(unsafe.Pointer(&unsigned))&7 != 0 || uintptr(unsafe.Pointer(&unsigned.i))&7 != 4 {
		t.Fatal("expected 8-aligned allocation with a 4-mod-8 raw uint64 field")
	}
	shouldPanic(t, "SwapUint64 raw field", func() { SwapUint64(&unsigned.i, 42) })
}
