//go:build darwin && ppc64

// TestReaddirR64 is the honest example for the darwin/ppc64 readdir_r fix: it
// is the third-party caller that the fix exists for, and it fails without the
// fix.
//
// Nothing in the standard library calls readdir_r on darwin/ppc64 (D16: the
// libc-directory implementation in syscall and os is built only for !ppc64, and
// syscall.Getdirentries is a raw SYS_getdirentries64 there), so the only way to
// reach the function is the push-linkname that syscall exports for os:
//
//	src/syscall/linkname_darwin.go:11   //go:linkname readdir_r
//	src/os/dir_darwin.go:149            //go:linkname readdir_r syscall.readdir_r
//
// which is exactly what this file does. The call therefore goes through the same
// symbol a third-party user would reach, typed with the same exported
// syscall.Dirent, and the record it fills is compared field by field with the
// record the kernel returns through the independent 64-bit path
// (syscall.Getdirentries = raw SYS_getdirentries64) for the same directory.
//
// What each path proves:
//   - syscall.Getdirentries is the reference: its record layout and its agreement
//     with the machine's C output were measured in D7 (C-dirent-test.txt: raw
//     SYS_getdirentries64 records match the 64-bit-inode record Go reads).
//   - the linknamed readdir_r is the subject. Without the fix it imports plain
//     readdir_r, which fills the legacy 32-bit-inode record (264 bytes, name at
//     offset 8) into a buffer the Go type reads as the 64-bit record (name at
//     offset 21). Silent: no error, count right, every field wrong.
//   - a third instrument, syscall.Stat (stat64, also D7-verified), supplies each
//     file's inode independently of both directory readers.
//
// Why this is sufficient: the defect is "the record the call fills is not the
// record the Go type describes". A field-by-field agreement with the kernel's
// own 64-bit record on the same directory, plus inode agreement with stat64, is
// a direct measurement of that; the two layouts differ in enough fields (name
// offset 8 vs 21, namlen offset 7 vs 18, header 8 vs 24 bytes) that a legacy
// record cannot pass it. The one thing it does not measure is which libc symbol
// the binary bound to -- that is a static property of the image, and the launch
// gate output for both binaries (fixed and pre-fix) is delivered beside this
// test so the symbol itself is visible too.
package d17readdir

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// The directory this test creates (d17Files, d17Dir, d17AllNames) is declared in
// records.go, so that the host control can fabricate the same record set.

//go:cgo_import_dynamic libc_opendir_inode64 opendir$INODE64 "/usr/lib/libSystem.B.dylib"
var libc_opendir_inode64_trampoline_addr uintptr

// //go:linkname of a syscall function that std itself exports for x/sys/unix
// (syscall_darwin.go: "golang.org/x/sys linknames the following syscalls"), so
// the call below is the same libc call path the generated wrappers use: the C
// ABI, errno from __error (which is what errnoPtr reads), and entersyscall
// around a call that may block.
//
//go:linkname syscall_syscallPtr syscall.syscallPtr
func syscall_syscallPtr(fn, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)

//go:linkname readdir_r syscall.readdir_r
func readdir_r(dir uintptr, entry *syscall.Dirent, result **syscall.Dirent) (res syscall.Errno)

//go:linkname closedir syscall.closedir
func closedir(dir uintptr) (err error)

// opendirInode64 opens d through libc's opendir$INODE64 and returns the DIR*.
func opendirInode64(d string) (uintptr, error) {
	p, err := syscall.BytePtrFromString(d)
	if err != nil {
		return 0, err
	}
	r0, _, e1 := syscall_syscallPtr(libc_opendir_inode64_trampoline_addr, uintptr(unsafe.Pointer(p)), 0, 0)
	if r0 == 0 {
		if e1 != 0 {
			return 0, e1
		}
		return 0, syscall.Errno(0) // reported by the caller as "opendir returned NULL"
	}
	return r0, nil
}

// checkLayoutMatchesGoType ties the offsets this test parses with to the Go type
// the linknamed call is typed with. If syscall.Dirent ever changes shape, this
// fails before any comparison below is believed.
func checkLayoutMatchesGoType(t *testing.T) {
	t.Helper()
	var d syscall.Dirent
	got := direntLayout{
		Name:      layout64.Name,
		InoOff:    int(unsafe.Offsetof(d.Ino)),
		InoLen:    int(unsafe.Sizeof(d.Ino)),
		ReclenOff: int(unsafe.Offsetof(d.Reclen)),
		NamlenOff: int(unsafe.Offsetof(d.Namlen)),
		NamlenLen: int(unsafe.Sizeof(d.Namlen)),
		TypeOff:   int(unsafe.Offsetof(d.Type)),
		NameOff:   int(unsafe.Offsetof(d.Name)),
		NameLen:   len(d.Name),
		Size:      int(unsafe.Sizeof(d)),
	}
	if got != layout64 {
		t.Fatalf("syscall.Dirent does not match the measured 64-bit-inode record this test parses with:\n got  %+v\n want %+v", got, layout64)
	}
}

// sample is one record as the subject returned it, kept under both
// interpretations so that a legacy-filled record can be named rather than only
// detected.
type sample struct {
	as64     entry
	asLegacy entry
}

// readWithReaddirR walks d with the linknamed readdir_r.
func readWithReaddirR(t *testing.T, d string) []sample {
	t.Helper()
	dir, err := opendirInode64(d)
	if err != nil {
		t.Fatalf("opendir$INODE64(%q): %v", d, err)
	}
	defer func() {
		if err := closedir(dir); err != nil {
			t.Errorf("closedir: %v", err)
		}
	}()

	var out []sample
	for i := 0; i < 4096; i++ {
		var e syscall.Dirent
		var ep *syscall.Dirent
		if errno := readdir_r(dir, &e, &ep); errno != 0 {
			t.Fatalf("readdir_r(%q): returned errno %d: %v", d, int(errno), error(errno))
		}
		if ep == nil {
			return out
		}
		buf := unsafe.Slice((*byte)(unsafe.Pointer(&e)), int(unsafe.Sizeof(e)))
		out = append(out, sample{as64: parseRecord(buf, layout64), asLegacy: parseRecord(buf, layoutLegacy)})
	}
	t.Fatalf("readdir_r(%q) did not end after 4096 entries", d)
	return nil
}

// readWithGetdirentries collects the kernel's own 64-bit-inode records through
// syscall.Getdirentries (raw SYS_getdirentries64 on darwin/ppc64).
func readWithGetdirentries(t *testing.T, d string) map[string]entry {
	t.Helper()
	fd, err := syscall.Open(d, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("open(%q): %v", d, err)
	}
	defer syscall.Close(fd)

	out := map[string]entry{}
	var buf [4096]byte
	var base uintptr
	for {
		n, err := syscall.Getdirentries(fd, buf[:], &base)
		if err != nil {
			t.Fatalf("Getdirentries(%q): %v", d, err)
		}
		if n == 0 {
			return out
		}
		if n > len(buf) {
			t.Fatalf("Getdirentries returned n=%d for a %d-byte buffer", n, len(buf))
		}
		for off := 0; off < n; {
			rec := parseRecord(buf[off:n], layout64)
			if int(rec.Reclen) < layout64.NameOff || off+int(rec.Reclen) > n {
				t.Fatalf("Getdirentries(%q): record at offset %d has reclen %d, %d bytes left in the buffer: %+v",
					d, off, rec.Reclen, n-off, rec)
			}
			if !rec.Valid {
				t.Fatalf("Getdirentries(%q): record %+v does not fit the 64-bit-inode name field", d, rec)
			}
			out[rec.Name] = rec
			off += int(rec.Reclen)
		}
	}
}

// statIno is the third instrument: the inode of one name, through stat64.
func statIno(t *testing.T, path string) uint64 {
	t.Helper()
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		t.Fatalf("stat(%q): %v", path, err)
	}
	return st.Ino
}

func TestReaddirR64(t *testing.T) {
	checkLayoutMatchesGoType(t)

	dir := t.TempDir()
	for _, n := range d17Files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("d17"), 0600); err != nil {
			t.Fatalf("WriteFile(%q): %v", n, err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, d17Dir), 0700); err != nil {
		t.Fatalf("Mkdir(%q): %v", d17Dir, err)
	}

	kernel := readWithGetdirentries(t, dir)
	subject := readWithReaddirR(t, dir)

	// The reference must be sound before it can judge anything: it must contain
	// exactly the names this test created, plus "." and "..".
	var refNames []string
	for n := range kernel {
		refNames = append(refNames, n)
	}
	sort.Strings(refNames)
	sortedWant := append([]string{}, d17AllNames...)
	sort.Strings(sortedWant)
	if strings.Join(refNames, "\x00") != strings.Join(sortedWant, "\x00") {
		t.Fatalf("the reference (Getdirentries) returned %q, want %q", refNames, sortedWant)
	}

	// Log what each path read, so the transcript shows the record fields rather
	// than only pass/fail. This is the evidence a reviewer reads.
	for _, s := range subject {
		t.Logf("readdir_r: name=%s ino=%d reclen=%d namlen=%d type=%d | same bytes as legacy record: name=%s ino=%d reclen=%d namlen=%d type=%d",
			quote(s.as64.Name), s.as64.Ino, s.as64.Reclen, s.as64.Namlen, s.as64.Type,
			quote(s.asLegacy.Name), s.asLegacy.Ino, s.asLegacy.Reclen, s.asLegacy.Namlen, s.asLegacy.Type)
	}

	got := map[string]entry{}
	for _, s := range subject {
		if s.as64.Name == "" {
			t.Errorf("readdir_r returned a record that reads as namlen=%d ino=%d reclen=%d through the 64-bit-inode layout; "+
				"the same bytes read as the legacy record name=%s namlen=%d ino=%d. A zero or absurd namlen with a name that "+
				"makes sense only in the legacy layout is the signature of plain readdir_r filling a legacy 32-bit-inode record",
				s.as64.Namlen, s.as64.Ino, s.as64.Reclen, quote(s.asLegacy.Name), s.asLegacy.Namlen, s.asLegacy.Ino)
			continue
		}
		if _, dup := got[s.as64.Name]; dup {
			t.Errorf("readdir_r returned %s twice", quote(s.as64.Name))
			continue
		}
		got[s.as64.Name] = s.as64
	}

	for _, d := range compareRecords(kernel, got) {
		t.Errorf("readdir_r vs the kernel's 64-bit records: %s", d)
	}

	// Third instrument: the inode each name must have, from stat64, and the type
	// the record must report for it.
	created := append([]string{}, d17Files...)
	created = append(created, d17Dir)
	for _, n := range created {
		wantType := uint8(d17RegType)
		if n == d17Dir {
			wantType = d17DirType
		}
		if want := statIno(t, filepath.Join(dir, n)); got[n].Ino != want {
			t.Errorf("readdir_r returned ino %d for %s, stat64 says %d", got[n].Ino, quote(n), want)
		}
		if got[n].Type != wantType {
			t.Errorf("readdir_r returned type %d for %s, want %d", got[n].Type, quote(n), wantType)
		}
	}
}
