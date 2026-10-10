// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objfile

import (
	"debug/macho"
	"testing"
)

// TestMachoGOARCH checks the Mach-O cpu type to GOARCH mapping.
//
// CpuPpc (0x12) and CpuPpc64 (0x12|0x01000000) are distinct values and both
// have to be listed: goarch is the name that cmd/internal/disasm looks up in
// its disassembler and byte-order tables, and an unmatched cpu type yields the
// empty string, which fails there as `unsupported architecture ""`. That is
// what happened to ppc32 Mach-O files before CpuPpc was handled.
func TestMachoGOARCH(t *testing.T) {
	for _, tt := range []struct {
		cpu  macho.Cpu
		want string
	}{
		{macho.Cpu386, "386"},
		{macho.CpuAmd64, "amd64"},
		{macho.CpuArm, "arm"},
		{macho.CpuArm64, "arm64"},
		{macho.CpuPpc, "ppc"},
		{macho.CpuPpc64, "ppc64"},
	} {
		f := &machoFile{macho: &macho.File{FileHeader: macho.FileHeader{Cpu: tt.cpu}}}
		if got := f.goarch(); got != tt.want {
			t.Errorf("goarch() for cpu %s = %q, want %q", tt.cpu, got, tt.want)
		}
	}
}
