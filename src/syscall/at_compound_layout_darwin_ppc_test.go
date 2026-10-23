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

// TestAttrlistLayoutDarwinPPC pins the request struct the no-follow metadata
// forms hand to setattrlist(2). The numbers were measured on the target
// (runs/root-at-metadata-20261023: sizeof 24, fields at 0/2/4/8/12/16/20); they
// are fixed-width, so a struct drift is a silent kernel-side misread rather than
// a build failure.
func TestAttrlistLayoutDarwinPPC(t *testing.T) {
	size, bitmap, reserved, common, vol, dir, file, fork := syscall.AttrlistLayoutForTest()
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"sizeof(attrlist)", size, 24},
		{"bitmapcount", bitmap, 0},
		{"reserved", reserved, 2},
		{"commonattr", common, 4},
		{"volattr", vol, 8},
		{"dirattr", dir, 12},
		{"fileattr", file, 16},
		{"forkattr", fork, 20},
	} {
		if tc.got != tc.want {
			t.Errorf("attrlist.%s is at/size %d; want the measured %d", tc.name, tc.got, tc.want)
		}
	}

	bitmapCount, noFollow, accessMask, modTime, acctime := syscall.AttrlistConstForTest()
	if bitmapCount != 5 {
		t.Errorf("ATTR_BIT_MAP_COUNT = %d; want 5 (the only accepted bitmapcount)", bitmapCount)
	}
	if noFollow != 1 {
		t.Errorf("FSOPT_NOFOLLOW = %d; want 1", noFollow)
	}
	if accessMask != 0x00020000 {
		t.Errorf("ATTR_CMN_ACCESSMASK = %#x; want 0x20000", accessMask)
	}
	if modTime != 0x00000400 {
		t.Errorf("ATTR_CMN_MODTIME = %#x; want 0x400", modTime)
	}
	if acctime != 0x00001000 {
		t.Errorf("ATTR_CMN_ACCTIME = %#x; want 0x1000", acctime)
	}
	// 32-bit caller: the kernel unpacks an 8-byte timespec, not the 16-byte
	// struct a 64-bit process would send.
	if got := syscall.TimespecSizeForTest(); got != 8 {
		t.Errorf("timespec in the attribute buffer is %d bytes; the kernel unpacks 8 for a 32-bit caller", got)
	}
}
