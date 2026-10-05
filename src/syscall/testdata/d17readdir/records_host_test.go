// Control for the checker in records.go: it runs on the host (linux/amd64), so
// the discriminating power of TestReaddirR64 is demonstrated without a G5.
//
// It fabricates two byte buffers of exactly the size of a Go zero-value
// syscall.Dirent (1048 bytes):
//
//   - the bytes Leopard's *plain* readdir_r writes into such a buffer, laid out
//     as the legacy 32-bit-inode record (fileno@0(4) reclen@4(2) type@6(1)
//     namlen@7(1) name@8, reclen 12 for "." per D7's C-dirent-test.txt);
//   - the bytes readdir_r$INODE64 writes, laid out as the 64-bit-inode record
//     (ino@0(8) seekoff@8(8) reclen@16(2) namlen@18(2) type@20(1) name@21).
//
// and then runs the same parseRecord + compareRecords the darwin/ppc64 test runs.
// The legacy buffer must be rejected on every entry of the real test's file set,
// and the 64-bit buffer accepted: a checker that accepted both would prove
// nothing on hardware.
package d17readdir

import (
	"sort"
	"strings"
	"testing"
)

// d17AllNames and the file set it is built from live in records.go, so the host
// control and the darwin/ppc64 test cannot drift apart.

// inos for the fabrication are D7's measured inodes from the G5
// (C-dirent-test.txt); only the *offsets* matter for what this control shows,
// but using measured values keeps the buffers plausible.
func fabricatedInos() map[string]uint64 {
	m := map[string]uint64{".": 504223, "..": 501161}
	for i, n := range d17Files {
		m[n] = uint64(514150 + i)
	}
	m[d17Dir] = 514200
	return m
}

func putU16(b []byte, off int, v uint16) {
	b[off], b[off+1] = byte(v), byte(v>>8)
}

func putU32(b []byte, off int, v uint32) {
	b[off], b[off+1], b[off+2], b[off+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
}

func putU64(b []byte, off int, v uint64) {
	for i := 0; i < 8; i++ {
		b[off+i] = byte(v >> (8 * uint(i)))
	}
}

func roundup4(n int) int { return (n + 3) &^ 3 }

// legacyBytes is the record Leopard's plain readdir_r writes: header 8 bytes,
// name from offset 8, item length 8+roundup(namlen+1,4) -- the relation every
// legacy row of D7's C-dirent-test.txt satisfies (namlen 1 -> 12, 5 -> 16,
// 9 -> 20, 14 -> 24). The rest of the 1048-byte buffer stays zero, which is what
// a Go zero-value syscall.Dirent gives the call.
func legacyBytes(ino uint32, name string, typ byte) []byte {
	buf := make([]byte, layout64.Size)
	reclen := 8 + roundup4(len(name)+1)
	putU32(buf, 0, ino)
	putU16(buf, 4, uint16(reclen))
	buf[6] = typ
	buf[7] = byte(len(name))
	copy(buf[8:], name)
	return buf
}

// measuredReclen64 is the record length D7 measured for a 64-bit-inode record
// with this namlen (SYS_getdirentries64 rows: 1 -> 24, 2 -> 24, 5 -> 32,
// 6 -> 32, 7 -> 32, 8 -> 36, 9 -> 36, 10 -> 36, 14 -> 40). For lengths D7 did
// not measure it uses 24+roundup(namlen+1,4), the relation that fits all of the
// measured rows except "." and "..". The real test never depends on this: it
// takes reclen from the kernel, and only compares.
func measuredReclen64(namlen int) uint16 {
	switch namlen {
	case 1, 2:
		return 24
	case 5, 6, 7:
		return 32
	case 8, 9, 10:
		return 36
	case 14:
		return 40
	}
	return uint16(24 + roundup4(namlen+1))
}

func inode64Bytes(ino uint64, name string, typ byte) []byte {
	buf := make([]byte, layout64.Size)
	putU64(buf, 0, ino)
	putU64(buf, 8, 1) // d_seekoff: the kernel cookie, not compared by the test
	putU16(buf, 16, measuredReclen64(len(name)))
	putU16(buf, 18, uint16(len(name)))
	buf[20] = typ
	copy(buf[21:], name)
	return buf
}

func typOf(name string) byte {
	if name == d17Dir {
		return d17DirType
	}
	return d17RegType
}

func buildSet(as64 bool) map[string]entry {
	inos := fabricatedInos()
	out := map[string]entry{}
	for _, n := range d17AllNames {
		var buf []byte
		if as64 {
			buf = inode64Bytes(inos[n], n, typOf(n))
		} else {
			buf = legacyBytes(uint32(inos[n]), n, typOf(n))
		}
		out[n] = parseRecord(buf, layout64)
	}
	return out
}

func TestCheckerAcceptsInode64Records(t *testing.T) {
	kernel := buildSet(true)
	diffs := compareRecords(kernel, buildSet(true))
	if len(diffs) != 0 {
		t.Errorf("the checker rejects records it must accept: %s", strings.Join(diffs, "; "))
	}
	for _, n := range d17AllNames {
		got := kernel[n]
		if !got.Valid || got.Name != n || got.Namlen != uint16(len(n)) {
			t.Errorf("64-bit-inode record for %q parsed as %+v", n, got)
		}
	}
}

// TestCheckerRejectsLegacyRecords is the control that matters: for every entry
// of the file set the real test uses, a buffer holding the legacy record (what
// the unfixed binary fills) must be reported as different from the kernel's
// 64-bit record. It also prints what the 64-bit layout reads out of those bytes,
// which is what the G5 transcript of the pre-fix binary is expected to show.
func TestCheckerRejectsLegacyRecords(t *testing.T) {
	kernel := buildSet(true)
	legacy := buildSet(false)

	for _, n := range d17AllNames {
		got := legacy[n]
		if got.Valid && got.Name == n && got.Ino == kernel[n].Ino {
			t.Errorf("a legacy-filled buffer parsed as the correct 64-bit record for %q: the checker has no power here", n)
		}
		// The legacy interpretation of the same bytes must still be the entry
		// the call meant to return, i.e. the fabrication really is a legacy
		// record and the two layouts really are distinguishable.
		inos := fabricatedInos()
		asLegacy := parseRecord(legacyBytes(uint32(inos[n]), n, typOf(n)), layoutLegacy)
		if !asLegacy.Valid || asLegacy.Name != n || asLegacy.Namlen != uint16(len(n)) || asLegacy.Type != typOf(n) {
			t.Errorf("legacy record for %q does not read back as a legacy record: %+v", n, asLegacy)
		}
		t.Logf("legacy-filled buffer for %-30q reads through the 64-bit layout as name=%s namlen=%d ino=%d reclen=%d (legacy layout: name=%s namlen=%d)",
			n, quote(got.Name), got.Namlen, got.Ino, got.Reclen, quote(asLegacy.Name), asLegacy.Namlen)
	}

	diffs := compareRecords(kernel, legacy)
	if len(diffs) == 0 {
		t.Fatalf("the checker accepted a legacy-filled record set: it cannot detect the defect it was written for")
	}
	if len(diffs) < len(d17AllNames) {
		t.Errorf("only %d differences for %d entries; every entry should differ: %s", len(diffs), len(d17AllNames), strings.Join(diffs, "; "))
	}
	sort.Strings(diffs)
	t.Logf("the checker reports %d differences for a legacy-filled record set; first three:\n  %s", len(diffs), strings.Join(diffs[:min(3, len(diffs))], "\n  "))
}
