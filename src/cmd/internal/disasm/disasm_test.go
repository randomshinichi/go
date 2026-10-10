// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package disasm

import (
	"encoding/binary"
	"strings"
	"testing"
)

// TestArchTables checks that the arch names accepted by disasms and byteOrders
// stay in agreement, and that the PowerPC entries a 32-bit Darwin Mach-O
// reports are present.
//
// These tables are keyed by the GOARCH string that an objfile container
// reports. cmd/internal/objfile's machoFile.goarch returns "ppc" for 32-bit
// PowerPC Mach-O files, so renaming the key here - or dropping it - brings back
// the failure this pair of tests exists to prevent: `go tool objdump` reporting
// `unsupported architecture "ppc"` for a file it can otherwise read.
func TestArchTables(t *testing.T) {
	for arch := range disasms {
		if byteOrders[arch] == nil {
			t.Errorf("disasms has %q but byteOrders does not", arch)
		}
	}
	for arch := range byteOrders {
		if disasms[arch] == nil {
			t.Errorf("byteOrders has %q but disasms does not", arch)
		}
	}
	for _, arch := range []string{"ppc", "ppc64", "ppc64le"} {
		if disasms[arch] == nil {
			t.Errorf("no disassembler for %q", arch)
		}
	}
	if got := byteOrders["ppc"]; got != binary.BigEndian {
		t.Errorf(`byteOrders["ppc"] = %v, want binary.BigEndian`, got)
	}
	if got := byteOrders["ppc64"]; got != binary.BigEndian {
		t.Errorf(`byteOrders["ppc64"] = %v, want binary.BigEndian`, got)
	}
}

// TestPPCDecoderHasNo32BitMode records a limitation that matters to anyone
// reading objdump output as evidence for a CPU floor.
//
// "ppc" and "ppc64" share one decoder (disasm_ppc64, backed by the ppc64asm
// table), and it is never told which of the two it is decoding. The encoding
// space is largely shared, so a 32-bit PowerPC object decodes correctly. But a
// ppc64-only encoding is not rejected either - it disassembles happily, because
// nothing here knows the target is 32-bit. Two of the words below (ld, std,
// rldicl) are illegal in 32-bit mode on the G5 this port targets.
//
// So objdump answers "can this tool read the file" and "what does each word
// mean"; it is NOT a CPU-floor check on its own. A census must filter the
// mnemonics or raw opcodes itself (or use a 32-bit-aware decoder) before
// concluding that a ppc32 binary contains no illegal instructions.
//
// If a real 32-bit PowerPC decoder replaces this one, this test should start
// failing; that is the signal that the paragraph above is out of date.
func TestPPCDecoderHasNo32BitMode(t *testing.T) {
	noLookup := func(uint64) (string, uint64) { return "", 0 }
	for _, c := range []struct {
		name string
		word []byte
	}{
		{"lwz r3,0(r4) (32-bit legal)", []byte{0x80, 0x64, 0x00, 0x00}},
		{"mullw r3,r4,r3 (32-bit legal)", []byte{0x7c, 0x64, 0x19, 0xd6}},
		{"ld r3,0(r4) (ppc64 only)", []byte{0xe8, 0x63, 0x00, 0x00}},
		{"std r3,0(r4) (ppc64 only)", []byte{0xf8, 0x64, 0x00, 0x00}},
		{"rldicl r3,r3,0,0 (ppc64 only)", []byte{0x78, 0x63, 0x00, 0x00}},
	} {
		text, size := disasm_ppc64(c.word, 0x1000, noLookup, binary.BigEndian, true)
		if size != 4 || strings.HasPrefix(strings.TrimSpace(text), "?") {
			t.Errorf("disasm_ppc64(%x) = %q size %d, want a decoded 4-byte instruction", c.word, text, size)
		}
	}
}
