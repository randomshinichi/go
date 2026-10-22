// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall_test

import (
	"syscall"
	"testing"
)

// TestAtCompoundLayoutDarwinPPC is the drift control for the request block that
// at_compound_darwin_ppc.s addresses by fixed offset. A disagreement between the
// struct and either side's constants silently corrupts a request instead of
// failing to build, so the offsets are asserted rather than trusted.
func TestAtCompoundLayoutDarwinPPC(t *testing.T) {
	names, assembly, actual := syscall.AtCompoundReqLayoutForTest()
	if len(names) != len(assembly) || len(names) != len(actual) {
		t.Fatalf("layout export is inconsistent: %d names, %d assembly offsets, %d actual offsets", len(names), len(assembly), len(actual))
	}
	for i, name := range names {
		if assembly[i] != actual[i] {
			t.Errorf("atCompoundReq.%s is at %d; the assembly says %d", name, actual[i], assembly[i])
		}
	}
	size, need := syscall.AtCompoundReqSizeForTest()
	if size < need {
		t.Errorf("atCompoundReq is %d bytes; want at least %d", size, need)
	}

	// Every field the routine reads with a 32-bit load is one 32-bit word.
	word, status, bytePtr, statPtr := syscall.WordSizesForTest()
	for _, tc := range []struct {
		name string
		got  int
	}{
		{"uintptr word", word}, {"int32 status", status},
		{"*byte", bytePtr}, {"*Stat_t", statPtr},
	} {
		if tc.got != 4 {
			t.Errorf("%s is %d bytes on this ABI; the native routine assumes 4", tc.name, tc.got)
		}
	}

	// The identity check compares these fields by offset, before any Go code
	// runs, and compares Ino as exactly two words.
	devOff, inoOff, devSize, inoSize, assemblyDev, assemblyIno := syscall.StatIdentityLayoutForTest()
	if devOff != assemblyDev || inoOff != assemblyIno {
		t.Errorf("Stat_t.Dev/Ino are at %d/%d; the assembly compares %d/%d", devOff, inoOff, assemblyDev, assemblyIno)
	}
	if devSize != 4 {
		t.Errorf("Stat_t.Dev is %d bytes; the assembly reads one word", devSize)
	}
	if inoSize != 8 {
		t.Errorf("Stat_t.Ino is %d bytes; the assembly compares exactly two words", inoSize)
	}

	// The pathname the identity check stats must be a NUL-terminated ".".
	if dot := syscall.AtDotForTest(); len(dot) < 2 || dot[0] != '.' || dot[1] != 0 {
		t.Errorf("atDot = %q; want a NUL-terminated \".\"", dot)
	}
}

// TestOpenFromLayoutDarwinPPC pins the ppc32 struct fopenfrom: unsigned flags,
// 16-bit mode_t, two bytes of padding, 32-bit pathname pointer, 12 bytes. The
// kernel's user_fopenfrom conversion is what makes 12 the right size here
// (bsd/kern/kern_descrip.c:1276-1283); the numbers were measured on the target
// in runs/root-at-design-20261022T221551Z.
func TestOpenFromLayoutDarwinPPC(t *testing.T) {
	size, flagsOff, modeOff, pathOff := syscall.OpenFromLayoutForTest()
	if size != 12 {
		t.Errorf("fopenfrom is %d bytes; want the measured 12", size)
	}
	if flagsOff != 0 {
		t.Errorf("fopenfrom.o_flags is at %d; want 0", flagsOff)
	}
	if modeOff != 4 {
		t.Errorf("fopenfrom.o_mode is at %d; want 4", modeOff)
	}
	if pathOff != 8 {
		t.Errorf("fopenfrom.o_pathname is at %d; want 8", pathOff)
	}
}
