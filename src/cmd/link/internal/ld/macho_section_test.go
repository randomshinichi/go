// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"testing"
)

func TestMachoshbitsPPC64PLT(t *testing.T) {
	// machoshbits and Errorf update linker globals; do not run in parallel.
	oldNsect, oldErrors, oldH := nsect, nerrors, *flagH
	defer func() { nsect, nerrors, *flagH = oldNsect, oldErrors, oldH }()
	*flagH = false
	for _, tt := range []struct {
		name        string
		arch        *sys.Arch
		length      uint64
		wantSection bool
		wantErrors  int
	}{
		{"ppc64-empty", sys.ArchPPC64, 0, false, 0},
		{"ppc64-populated", sys.ArchPPC64, 16, false, 1},
		{"amd64-empty", sys.ArchAMD64, 0, true, 0},
		{"amd64-populated", sys.ArchAMD64, 6, true, 0},
		{"arm64-empty", sys.ArchARM64, 0, true, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			nsect, nerrors = 0, 0
			ctxt := &Link{Target: Target{Arch: tt.arch}}
			seg := &sym.Segment{Filelen: tt.length}
			sect := &sym.Section{Name: ".plt", Length: tt.length, Seg: seg}
			mseg := &MachoSeg{msect: 1, sect: make([]MachoSect, 1)}
			machoshbits(ctxt, mseg, sect, "__TEXT")
			if nerrors != tt.wantErrors {
				t.Fatalf("errors=%d, want %d", nerrors, tt.wantErrors)
			}
			wantCount := uint32(0)
			if tt.wantSection {
				wantCount = 1
			}
			if mseg.nsect != wantCount || nsect != int(wantCount) {
				t.Fatalf("section counts=(%d,%d), want %d", mseg.nsect, nsect, wantCount)
			}
			if tt.wantSection {
				ms := mseg.sect[0]
				if ms.name != "__symbol_stub1" || ms.size != tt.length || ms.res2 != 6 {
					t.Fatalf("sibling stub section changed: %+v", ms)
				}
			}
		})
	}
}
