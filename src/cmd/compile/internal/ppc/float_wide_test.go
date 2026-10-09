// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
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
