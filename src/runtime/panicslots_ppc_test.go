// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"fmt"
	"internal/runtime/startlinetest"
	"testing"
)

// TestPPC32PanicBoundsSlots checks the RUNTIME half of the bounds-save protocol:
// runtime/panic_ppc.s stores R3..R9 into save-array cells 0..6 and
// panicBounds32/panicBounds32X index them by the slot numbers in
// PCDATA_PanicBounds (reg - REG_R3; panicExtend packs hi<<2|lo), through the real
// linked pcdata table and the real boundsError text.
//
// SCOPE LIMIT, deliberately stated: this harness writes its own PCDATA metadata
// and its own register contents, so it does NOT observe what
// cmd/compile/internal/ppc/ssa.go emits. A wrong base or a reversed pack in the
// compiler would leave this gate green; the compiler side is covered only for
// slots 0..2 by test/fixedbugs/issue30116*.go. See the bead that drives this gate
// from compiler-emitted PCDATA. Every slot is used as x
// and as y at least once; compiled code uses all seven (a std census finds
// slots 0..6 in runtime, math/big, encoding/json), but
// test/fixedbugs/issue30116*.go only reaches slots 0..2.
func TestPPC32PanicBoundsSlots(t *testing.T) {
	type slotCase struct {
		site uint32
		want string
	}
	var cases []slotCase
	for x := 0; x < 7; x++ {
		y := (x + 1) % 7
		cases = append(cases, slotCase{uint32(x), fmt.Sprintf("runtime error: index out of range [%d] with length %d", 1000+x, 1000+y)})
	}
	for i, e := range [][3]int{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}, {3, 0, 5}} {
		x := int64(1000+e[0])<<32 + int64(1000+e[1])
		cases = append(cases, slotCase{uint32(7 + i), fmt.Sprintf("runtime error: index out of range [%d] with length %d", x, 1000+e[2])})
	}
	for _, c := range cases {
		got := func() (msg string) {
			defer func() {
				if r := recover(); r != nil {
					msg = fmt.Sprint(r)
				}
			}()
			startlinetest.PanicSlot(c.site)
			return "no panic"
		}()
		if got != c.want {
			t.Errorf("site %d: got %q, want %q", c.site, got, c.want)
		}
	}
}
