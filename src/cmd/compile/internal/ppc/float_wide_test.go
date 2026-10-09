// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"fmt"
	"internal/testenv"
	"math/bits"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Independent integer-only round-to-nearest-even oracle. No host integer-to-
// float conversion (possibly itself double-rounded) determines the expected bits.
func integerFloatBits(x uint64, fraction, bias int) uint64 {
	if x == 0 {
		return 0
	}
	exp := bits.Len64(x) - 1
	shift := exp - fraction
	var q uint64
	if shift > 0 {
		q = x >> shift
		rem := x & ((uint64(1) << shift) - 1)
		half := uint64(1) << (shift - 1)
		if rem > half || rem == half && q&1 != 0 {
			q++
		}
	} else {
		q = x << (-shift)
	}
	if q == uint64(1)<<(fraction+1) {
		q >>= 1
		exp++
	}
	return uint64(exp+bias)<<fraction | q&((uint64(1)<<fraction)-1)
}

func widePopulationBody(vectors []uint64) (string, int) {
	population := append([]uint64{}, vectors...)
	for e := 0; e < 64; e++ {
		base := uint64(1) << e
		for _, delta := range []uint64{0, 1, 4095, 4096, 4097} {
			population = append(population, base+delta, base-delta)
		}
		if e >= 24 {
			half := uint64(1) << (e - 24)
			for _, k := range []uint64{1, 3, 5, 7} {
				for _, d := range []uint64{k*half - 1, k * half, k*half + 1} {
					population = append(population, base+d, -base-d)
				}
			}
		}
	}
	x := uint64(0x9e3779b97f4a7c15)
	for i := 0; i < 2048; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		population = append(population, x)
	}
	var loop strings.Builder
	// R13 is excluded from SSA allocation. The table pointer is in the
	// caller's private area: calls can clobber every allocatable GPR.
	loop.WriteString("li 13,0\nbl table_pc\ntable_pc:\nmflr 3\naddi 3,3,vector_table-table_pc\nstw 3,352(1)\nb vector_loop\nvector_table:\n")
	for _, x := range population {
		mag, sign32, sign64 := x, uint64(0), uint64(0)
		if int64(x) < 0 {
			mag = -x
			sign32 = 1 << 31
			sign64 = 1 << 63
		}
		u64 := integerFloatBits(x, 52, 1023)
		s64 := integerFloatBits(mag, 52, 1023) | sign64
		for _, w := range []uint32{uint32(x >> 32), uint32(x), uint32(integerFloatBits(x, 23, 127)), uint32(u64 >> 32), uint32(u64), uint32(integerFloatBits(mag, 23, 127) | sign32), uint32(s64 >> 32), uint32(s64)} {
			fmt.Fprintf(&loop, ".long 0x%08x\n", w)
		}
	}
	loop.WriteString("vector_loop:\n")
	for i, fn := range []string{"wideU32", "wideU64", "wideS32", "wideS64"} {
		loop.WriteString("lwz 3,352(1)\nlwz 4,0(3)\nstw 4,24(1)\nlwz 4,4(3)\nstw 4,28(1)\n")
		fmt.Fprintf(&loop, "bl fn_%s\ncmpw 1,2\nbne stack_bad\nlwz 5,352(1)\n", fn)
		off, size := []int{8, 12, 20, 24}[i], 4
		if i%2 == 1 {
			size = 8
		}
		for d := 0; d < size; d += 4 {
			fmt.Fprintf(&loop, "lwz 3,%d(1)\nlwz 4,%d(5)\ncmpw 3,4\nbne result_bad\n", 32+d, off+d)
		}
	}
	fmt.Fprintf(&loop, "lwz 3,352(1)\naddi 3,3,32\nstw 3,352(1)\naddi 13,13,1\ncmpwi 13,%d\nbne vector_loop\n", len(population))
	return loop.String(), len(population)
}

func TestFloatWideConversionsQEMU(t *testing.T) {
	helper, err := os.ReadFile(filepath.Join(testenv.GOROOT(t), "src", "runtime", "float_wide.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(helper) + `
//go:noinline
//go:nosplit
func wideU32(x uint64) float32 {return float32(x)}
//go:noinline
//go:nosplit
func wideU64(x uint64) float64 {return float64(x)}
//go:noinline
//go:nosplit
func wideS32(x int64) float32 {return float32(x)}
//go:noinline
//go:nosplit
func wideS64(x int64) float64 {return float64(x)}
`
	functions := []string{"uint64tofloat32", "uint64tofloat64", "int64tofloat32", "int64tofloat64", "wideU32", "wideU64", "wideS32", "wideS64"}
	dir, dump := compileFloatFixture(t, source, functions, "")
	vectors := []uint64{0, 1, 1<<24 - 1, 1 << 24, 1<<24 + 1, 1<<32 - 1, 1 << 32, 1<<38 + 1, 1<<46 - 1, 1<<46 + 1, 1<<46 + (1 << 22) + 1, 1<<53 + 1, 1<<63 - 1, 1 << 63, 1<<63 + 1, 1<<63 + (1 << 39) - 1, 1<<63 + (1 << 39), 1<<63 + (1 << 39) + 1, 1<<63 + 3*(1<<39) - 1, ^uint64(0)}
	var body strings.Builder
	for _, x := range vectors {
		single := integerFloatBits(x, 23, 127)
		double := integerFloatBits(x, 52, 1023)
		t.Logf("unsigned x=%016x binary32=%08x binary64=%016x", x, single, double)
		body.WriteString(setDouble(24, x) + "bl fn_wideU32\n" + checkWord(32, uint32(single)))
		body.WriteString(setDouble(24, x) + "bl fn_wideU64\n" + checkDouble(32, double))
		magnitude := x
		sign32, sign64 := uint64(0), uint64(0)
		if int64(x) < 0 {
			magnitude = -x
			sign32 = 1 << 31
			sign64 = 1 << 63
		}
		body.WriteString(setDouble(24, x) + "bl fn_wideS32\n" + checkWord(32, uint32(integerFloatBits(magnitude, 23, 127)|sign32)))
		body.WriteString(setDouble(24, x) + "bl fn_wideS64\n" + checkDouble(32, integerFloatBits(magnitude, 52, 1023)|sign64))
	}
	t.Run("exact-and-adversarial-rounding", func(t *testing.T) { runFloatELF(t, dir, dump, functions, body.String(), nil) })
	t.Run("edge-and-random-population", func(t *testing.T) {
		loop, n := widePopulationBody(vectors)
		runFloatELF(t, dir, dump, functions, loop, nil)
		t.Logf("PASS: %d vectors, %d complete Go conversions", n, n*4)
	})
	t.Run("negative-sticky-bit", func(t *testing.T) {
		one := -1
		runFloatELF(t, dir, dump, functions, body.String(), func(fn string, w uint32) uint32 {
			if fn != "uint64tofloat32" {
				return w
			}
			if w>>26 == 14 && w&0x001fffff == 1 {
				one = int(w>>21) & 31
			}
			if w>>26 == 31 && (w>>1)&1023 == 444 && int(w>>11)&31 == one {
				src := (w >> 21) & 31
				return w&^0x0000f800 | src<<11 // sticky OR becomes an FPR-independent integer copy
			}
			return w
		})
	})
}
