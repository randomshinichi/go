// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objfile

import (
	"debug/elf"
	"encoding/binary"
	"testing"
)

// TestElfGOARCH checks the ELF machine to GOARCH mapping.
//
// EM_PPC (20, 32-bit PowerPC) and EM_PPC64 (21) are distinct machine values.
// An unmatched machine yields the empty string, which cmd/internal/disasm
// rejects as `unsupported architecture ""`; that is what happened to linux/ppc
// ELF files before EM_PPC was handled.
func TestElfGOARCH(t *testing.T) {
	for _, tt := range []struct {
		machine elf.Machine
		order   binary.ByteOrder
		want    string
	}{
		{elf.EM_386, binary.LittleEndian, "386"},
		{elf.EM_X86_64, binary.LittleEndian, "amd64"},
		{elf.EM_PPC, binary.BigEndian, "ppc"},
		{elf.EM_PPC64, binary.BigEndian, "ppc64"},
		{elf.EM_PPC64, binary.LittleEndian, "ppc64le"},
		{elf.EM_NONE, binary.BigEndian, ""},
	} {
		f := &elfFile{elf: &elf.File{FileHeader: elf.FileHeader{Machine: tt.machine, ByteOrder: tt.order}}}
		if got := f.goarch(); got != tt.want {
			t.Errorf("goarch() for machine %s = %q, want %q", tt.machine, got, tt.want)
		}
	}
}
