// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc64

import (
	"bytes"
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/ld"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"encoding/binary"
	"os"
	"sort"
	"testing"
)

// These inputs transcribe testdata/{probe,carry}.s, not the expected
// relocation words. The golden bytes come directly from Apple's objects.
func TestMachoreloc1Apple(t *testing.T) {
	type reference struct {
		off  int32
		typ  objabi.RelocType
		add  int64
		name string
		id   int32
	}
	for _, tt := range []struct {
		name           string
		reloff, nreloc int
		words          []uint32 // instructions before the writer applies addends
		refs           []reference
	}{
		{
			name: "probe", reloff: 312, nreloc: 9,
			words: []uint32{0x48000001, 0x3c600000, 0x38630000, 0x3c800000, 0xe8840000, 0x4e800020},
			refs: []reference{
				{0, objabi.R_CALLPOWER, 0, "_ext", 1},
				{4, objabi.R_ADDRPOWER, 8, "_global", 2},
				{12, objabi.R_ADDRPOWER_DS, 8, "_global", 2},
			},
		},
		{
			name: "carry", reloff: 324, nreloc: 16,
			words: []uint32{0x3c600000, 0x38630000, 0x3c800000, 0x38840000, 0x3ca00000, 0x38a50000, 0x3cc00000, 0xe8c60000, 0x4e800020},
			refs: []reference{
				{0, objabi.R_ADDRPOWER, 0x18000, "_g", 1},
				{8, objabi.R_ADDRPOWER, -8, "_g", 1},
				{16, objabi.R_ADDRPOWER, 0x8000, "_g", 1},
				{24, objabi.R_ADDRPOWER_DS, 0x18000, "_g", 1},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oracle, err := os.ReadFile("testdata/" + tt.name + ".macho")
			if err != nil {
				t.Fatal(err)
			}
			end := tt.reloff + 8*tt.nreloc
			if len(oracle) < end || tt.reloff != 288+4*len(tt.words) {
				t.Fatal("unexpected Apple fixture layout")
			}
			// The full 72-byte probe.o[312:384] (or 128-byte carry.o
			// [324:452]) includes every primary AND every PAIR word.
			want := normalizeAppleReferences(t, oracle[tt.reloff:end])
			ldr := loader.NewLoader(0, &loader.ErrorReporter{AfterErrorAction: func() { t.Fatal("loader error") }})
			s := ldr.CreateSymForUpdate("_probe", 0)
			s.SetType(sym.STEXT)
			target := &ld.Target{Arch: sys.ArchPPC64, HeadType: objabi.Hdarwin, LinkMode: ld.LinkExternal}
			text := make([]byte, 4*len(tt.words))
			for i, word := range tt.words {
				binary.BigEndian.PutUint32(text[4*i:], word)
			}
			out := ld.NewOutBuf(sys.ArchPPC64)
			for _, ref := range tt.refs {
				rs := ldr.LookupOrCreateSym(ref.name, 0)
				ldr.SetSymDynid(rs, ref.id) // indices in Apple's nlist table
				r, _ := s.AddRel(ref.typ)
				r.SetOff(ref.off)
				r.SetAdd(ref.add)
				r.SetSym(rs)
				size, count := uint8(8), 4
				if ref.typ == objabi.R_CALLPOWER {
					size, count = 4, 1
				}
				r.SetSiz(size)
				var val int64
				if size == 4 {
					val = int64(binary.BigEndian.Uint32(text[ref.off:]))
				} else {
					val = int64(binary.BigEndian.Uint64(text[ref.off:]))
				}
				val, n, ok := archrelocmacho(target, ldr, r, s.Sym(), val)
				if !ok || n != count {
					t.Fatalf("archrelocmacho at %#x: ok=%v count=%d, want %d", ref.off, ok, n, count)
				}
				if size == 4 {
					binary.BigEndian.PutUint32(text[ref.off:], uint32(val))
				} else {
					binary.BigEndian.PutUint64(text[ref.off:], uint64(val))
				}
				ext := loader.ExtReloc{Xsym: rs, Xadd: ref.add, Type: ref.typ, Size: size}
				if !machoreloc1(sys.ArchPPC64, out, ldr, s.Sym(), ext, int64(ref.off)) {
					t.Fatalf("machoreloc1 rejected reference at %#x", ref.off)
				}
			}
			got := normalizeAppleReferences(t, out.Data())
			if !bytes.Equal(got, want) {
				t.Errorf("%d-byte relocation span differs:\n got %x\nwant %x", len(want), got, want)
			} else {
				t.Logf("all %d relocation bytes match, including PAIR words", len(want))
			}
			if !bytes.Equal(text, oracle[288:tt.reloff]) {
				t.Errorf("pre-relocation instruction bytes differ:\n got %x\nwant %x", text, oracle[288:tt.reloff])
			}

			t.Run("PAIR-word-negative-control", func(t *testing.T) {
				// First reference is paired after normalisation. Mutate
				// ONLY its PAIR's bitfield word: one symbolnum bit, leaving
				// type/length/extern intact. The same normalisation and
				// full-byte comparison must reject it in BOTH directions.
				mutant := bytes.Clone(want)
				word := binary.BigEndian.Uint32(mutant[12:16])
				binary.BigEndian.PutUint32(mutant[12:16], word^0x100)
				mutant = normalizeAppleReferences(t, mutant)
				if bytes.Equal(mutant, want) || bytes.Equal(want, mutant) {
					t.Fatal("comparison accepted a mutated PAIR word")
				}
				t.Log("single PAIR-word mutation rejected: mutant vs oracle and oracle vs mutant")
			})
		})
	}
}

// normalizeAppleReferences states the ONLY order normalisation used above:
// treat each instruction reference as an indivisible primary+PAIR (16 bytes),
// or a lone BR24 (8 bytes), then sort references by descending primary
// r_address, as Apple does. Go's ascending instruction-pair traversal is not
// Apple's global order. Never sort PAIRs by their r_address (it is an addend
// half), swap words, mask fields, or reconstruct records. All input bytes are
// retained exactly, including PAIR r_symbolnum, before bytes.Equal is applied.
func normalizeAppleReferences(t *testing.T, records []byte) []byte {
	t.Helper()
	var refs [][]byte
	for off := 0; off < len(records); {
		if len(records)-off < 8 {
			t.Fatal("partial relocation record")
		}
		word := binary.BigEndian.Uint32(records[off+4:])
		n := 8
		// Literal Apple PPC relocation types, deliberately independent of
		// the writer's constants: BR24=3, LO16=5, HA16=6, LO14=7, PAIR=1.
		switch word & 15 {
		case 3:
		case 5, 6, 7:
			n = 16
			if len(records)-off < n || binary.BigEndian.Uint32(records[off+12:])&15 != 1 {
				t.Fatal("missing PAIR immediately after primary")
			}
		default:
			t.Fatalf("unexpected primary relocation word %#x", word)
		}
		refs = append(refs, records[off:off+n])
		off += n
	}
	sort.Slice(refs, func(i, j int) bool {
		return binary.BigEndian.Uint32(refs[i]) > binary.BigEndian.Uint32(refs[j])
	})
	normalized := make([]byte, 0, len(records))
	for _, ref := range refs {
		normalized = append(normalized, ref...)
	}
	return normalized
}
