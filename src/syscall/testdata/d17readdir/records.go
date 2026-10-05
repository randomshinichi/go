// Package d17readdir is a test-only module: it is the "third party" the task
// describes -- a program that reaches libc's directory reader only through
// //go:linkname of syscall.readdir_r, because nothing in the standard library
// calls it on darwin/ppc64 (D16).
//
// records.go holds the platform-independent half: the field offsets of the two
// Darwin struct dirent variants, a parser, and the comparison the test uses.
// Keeping this half free of Darwin-only code is deliberate: it lets the same
// comparison run on the host (records_host_test.go), where a fabricated legacy
// record shows what the checker accepts and what it rejects. A checker whose
// discriminating power has never been demonstrated is not yet evidence.
package d17readdir

import (
	"fmt"
	"sort"
	"strings"
)

// direntLayout is the field layout of one Darwin struct dirent variant. The
// two variants are selected in C by __DARWIN_USE_64_BIT_INODE, and Leopard
// exports a separate libc function for each: the record a call fills is decided
// by which function is called, not by which struct the caller declares.
type direntLayout struct {
	Name      string // human name of the variant, used in failure messages
	InoOff    int
	InoLen    int
	ReclenOff int
	NamlenOff int
	NamlenLen int // the legacy record's d_namlen is one byte, the 64-bit one two
	TypeOff   int
	NameOff   int
	NameLen   int // bytes the record has for the name
	Size      int // sizeof(struct dirent)
}

// layout64 is the record the Go type syscall.Dirent describes, measured on the
// G5 through SYS_getdirentries64 (344):
//
//	ino@0(8) seekoff@8(8) reclen@16(2) namlen@18(2) type@20(1) name@21
//	sizeof(struct dirent) = 1048
//
// Source: out/stage2-20261003/D7-syscall/g5-request2/out-d7r2/C-dirent-test.txt
// (first block, "SYS_getdirentries64 (344) ..., 64-bit-inode record"). The
// darwin/ppc64 test asserts these offsets equal unsafe.Offsetof of the fields of
// syscall.Dirent, so this table cannot silently drift from the Go type.
var layout64 = direntLayout{
	Name:      "64-bit-inode",
	InoOff:    0,
	InoLen:    8,
	ReclenOff: 16,
	NamlenOff: 18,
	NamlenLen: 2,
	TypeOff:   20,
	NameOff:   21,
	NameLen:   1024,
	Size:      1048,
}

// layoutLegacy is the record Leopard's plain readdir_r fills, measured on the
// G5 in the same file (second block, "SYS_getdirentries (196) legacy record",
// and the C probe's "plain: ino=... reclen=12 namlen=1 type=4 name=." rows):
//
//	fileno@0(4) reclen@4(2) type@6(1) namlen@7(1) name@8
//	sizeof(struct dirent) = 264
//
// This test parses the same bytes with both layouts, which is how a legacy
// record can be named as the cause instead of only being detected as garbage.
var layoutLegacy = direntLayout{
	Name:      "legacy-32-bit-inode",
	InoOff:    0,
	InoLen:    4,
	ReclenOff: 4,
	NamlenOff: 7,
	NamlenLen: 1,
	TypeOff:   6,
	NameOff:   8,
	NameLen:   256,
	Size:      264,
}

// The directory the darwin/ppc64 test creates. It is fixed, and it is declared
// here rather than in the Darwin-only test file, so that the host control
// (records_host_test.go) can fabricate the same record set and show what the
// checker does with it. Names of several lengths are used because the legacy and
// 64-bit headers are different sizes: a short name leaves the 64-bit namlen field
// reading zero, a long name leaves it reading bytes of the name instead, and
// either way the record is silently wrong.
var d17Files = []string{"a", "bb", "gamma", "delta-longer-name-0123456789"}

const (
	d17Dir     = "subdir-entry"
	d17DirType = 4 // DT_DIR, measured on the G5
	d17RegType = 8 // DT_REG, measured on the G5
)

// d17AllNames is the whole entry set the test's directory must contain, in the
// order D7 measured the kernel returning it ('.' then '..').
var d17AllNames = append(append([]string{".", ".."}, d17Files...), d17Dir)

// entry is one directory entry as read out of a raw record by a given layout.
type entry struct {
	Name   string // Namlen bytes of the record's name field, byte for byte
	Ino    uint64
	Reclen uint16
	Namlen uint16
	Type   uint8
	Valid  bool // Namlen fits the layout's name field and the buffer is long enough
}

// parseRecord reads one record out of b using l. A record whose Namlen does not
// fit the layout's name field is returned with Valid=false: that is exactly the
// shape a mismatch produces, so it must be reportable rather than a panic or a
// truncated name that looks plausible.
func parseRecord(b []byte, l direntLayout) entry {
	var e entry
	if l.InoOff+l.InoLen > len(b) || l.ReclenOff+2 > len(b) ||
		l.NamlenOff+l.NamlenLen > len(b) || l.TypeOff >= len(b) || l.NameOff > len(b) {
		return e
	}
	for i := 0; i < l.InoLen; i++ {
		e.Ino |= uint64(b[l.InoOff+i]) << (8 * uint(i))
	}
	e.Reclen = uint16(b[l.ReclenOff]) | uint16(b[l.ReclenOff+1])<<8
	e.Namlen = uint16(b[l.NamlenOff])
	if l.NamlenLen == 2 {
		e.Namlen |= uint16(b[l.NamlenOff+1]) << 8
	}
	e.Type = b[l.TypeOff]
	if int(e.Namlen) > l.NameLen {
		return e
	}
	if l.NameOff+int(e.Namlen) > len(b) {
		return e
	}
	e.Valid = true
	e.Name = string(b[l.NameOff : l.NameOff+int(e.Namlen)])
	return e
}

// fieldDiff describes where got differs from want, in one line each.
func fieldDiff(want, got entry) []string {
	var out []string
	if got.Ino != want.Ino {
		out = append(out, fmt.Sprintf("ino %d, kernel says %d", got.Ino, want.Ino))
	}
	if got.Namlen != want.Namlen {
		out = append(out, fmt.Sprintf("namlen %d, kernel says %d", got.Namlen, want.Namlen))
	}
	if got.Type != want.Type {
		out = append(out, fmt.Sprintf("type %d, kernel says %d", got.Type, want.Type))
	}
	if got.Name != want.Name {
		out = append(out, fmt.Sprintf("name %s, kernel says %s", quote(got.Name), quote(want.Name)))
	}
	if got.Reclen != want.Reclen {
		out = append(out, fmt.Sprintf("record length %d, kernel says %d (the record Go reads is not the kernel's 64-bit record)", got.Reclen, want.Reclen))
	}
	return out
}

// compareRecords compares the entries the subject (the linknamed readdir_r)
// returned with the entries the kernel returned through the independent 64-bit
// path (syscall.Getdirentries on darwin/ppc64 = raw SYS_getdirentries64), keyed
// by name. It returns human-readable differences, sorted, so a failure names
// every entry that disagrees rather than stopping at the first.
func compareRecords(kernel, subject map[string]entry) []string {
	var out []string
	for name, want := range kernel {
		got, ok := subject[name]
		if !ok {
			out = append(out, fmt.Sprintf("entry %s: the subject did not return it (names the subject returned: %s)",
				quote(name), quotedNames(subject)))
			continue
		}
		if !got.Valid {
			out = append(out, fmt.Sprintf("entry %s: namlen %d does not fit the %s name field (%d bytes), so the name cannot be read at all",
				quote(name), got.Namlen, layout64.Name, layout64.NameLen))
			continue
		}
		for _, d := range fieldDiff(want, got) {
			out = append(out, fmt.Sprintf("entry %s: %s", quote(name), d))
		}
	}
	for name := range subject {
		if _, ok := kernel[name]; !ok {
			out = append(out, fmt.Sprintf("entry %s: the subject returned it but the kernel did not (record %d bytes, namlen %d, ino %d)",
				quote(name), subject[name].Reclen, subject[name].Namlen, subject[name].Ino))
		}
	}
	sort.Strings(out)
	return out
}

func quotedNames(m map[string]entry) string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, quote(n))
	}
	sort.Strings(names)
	if len(names) > 8 {
		names = append(names[:8], fmt.Sprintf("... (%d more)", len(names)-8))
	}
	return strings.Join(names, " ")
}

// quote makes a byte string that came out of a record field printable: names
// read through the wrong layout can hold arbitrary bytes, including NUL.
func quote(s string) string {
	if strings.ToValidUTF8(s, "") == s && !strings.ContainsAny(s, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f") {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%q (hex % x)", s, s)
}
