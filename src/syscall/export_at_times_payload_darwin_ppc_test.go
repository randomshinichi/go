// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall

// Test-only exports for the timestamp workaround: the payload builder, the
// attrlist bits it selects, the sentinel change time and the HFS guard. They live
// in package syscall for the same reason export_at_compound_darwin_ppc_test.go
// does (the constants and the guard are unexported, and this package's own test
// files cannot import testing).

// AtTimesPayloadForTest returns what atTimesPayload builds for a selection.
func AtTimesPayloadForTest(mod, chg, acc *Timespec) (bitmapcount uint16, commonattr uint32, buf [3]Timespec, n int) {
	al, buf, n := atTimesPayload(mod, chg, acc)
	return al.bitmapcount, al.commonattr, buf, n
}

// AtAttrBitsForTest returns the attrlist constants the metadata requests use.
func AtAttrBitsForTest() (bitMapCount, modTime, chgTime, accTime, accessMask uint32) {
	return atAttrBitMapCount, atAttrCmnModTime, atAttrCmnChgTime, atAttrCmnAcctime, atAttrCmnAccessMask
}

// AtDummyChgTimeForTest returns the change time sent with an access-time-only
// request on HFS.
func AtDummyChgTimeForTest() Timespec { return atDummyChgTime }

// AtHFSNameForTest returns the f_fstypename the workaround keys on.
func AtHFSNameForTest() string { return atHFSName }

// AtFstypenameIsForTest exposes the comparison the HFS guard rests on.
func AtFstypenameIsForTest(field []int8, name string) bool { return atFstypenameIs(field, name) }

// AtChgTimeForAccessTimeForTest exposes the guard itself.
func AtChgTimeForAccessTimeForTest(anchor int, path string) (Timespec, bool, error) {
	return atChgTimeForAccessTime(anchor, path)
}
