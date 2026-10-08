// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

var relocLine = regexp.MustCompile(`^\trel ([0-9]+)\+([0-9]+) t=(R_[A-Z]+) (.+)\+(-?[0-9]+)$`)

func emittedRelocs(out, fn string) []string {
	var relocs []string
	in := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, fn+" STEXT") {
			in = true
			continue
		}
		if !in {
			continue
		}
		if !strings.HasPrefix(line, "\t") {
			break
		}
		if m := relocLine.FindStringSubmatch(line); m != nil {
			relocs = append(relocs, fmt.Sprintf("%s+%s %s %s+%s", m[1], m[2], m[3], m[4], m[5]))
		}
	}
	return relocs
}

// Go's R_ADDRPOWER names an entire addis/addi pair; ELF names the two
// immediate fields separately (HA16 and LO16). Translate clang's records to
// the pair representation, checking offsets, symbol identity and addends.
// This compares object intent, not Darwin linker application.
func clangRelocs(t *testing.T, path string) []string {
	t.Helper()
	f, err := elf.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s := f.Section(".rela.text")
	if s == nil {
		return nil
	}
	data, err := s.Data()
	if err != nil {
		t.Fatal(err)
	}
	syms, err := f.Symbols()
	if err != nil {
		t.Fatal(err)
	}
	if len(data)%12 != 0 {
		t.Fatalf("malformed ELF32 RELA section: %d bytes", len(data))
	}
	type relocation struct {
		off uint32
		typ elf.R_PPC
		sym string
		add int32
	}
	var records []relocation
	for i := 0; i < len(data); i += 12 {
		info := binary.BigEndian.Uint32(data[i+4:])
		index := int(info >> 8)
		if index < 1 || index > len(syms) {
			t.Fatalf("bad relocation symbol index %d", index)
		}
		records = append(records, relocation{binary.BigEndian.Uint32(data[i:]), elf.R_PPC(info & 255), syms[index-1].Name, int32(binary.BigEndian.Uint32(data[i+8:]))})
	}
	var relocs []string
	for i := 0; i < len(records); i++ {
		r := records[i]
		switch r.typ {
		case elf.R_PPC_ADDR16_HA:
			if i+1 == len(records) {
				t.Fatal("HA16 relocation missing LO16 pair")
			}
			lo := records[i+1]
			if r.off%4 != 2 || lo.off != r.off+4 || lo.typ != elf.R_PPC_ADDR16_LO || lo.sym != r.sym || lo.add != r.add {
				t.Fatalf("invalid HA16/LO16 pair: %+v %+v", r, lo)
			}
			relocs = append(relocs, fmt.Sprintf("%d+8 R_ADDRPOWER %s+%d", r.off-2, r.sym, r.add))
			i++
		case elf.R_PPC_REL24:
			relocs = append(relocs, fmt.Sprintf("%d+4 R_CALLPOWER %s+%d", r.off, r.sym, r.add))
		default:
			t.Fatalf("unexpected clang text relocation: %+v", r)
		}
	}
	return relocs
}
