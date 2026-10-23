// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall_test

// Layout and dispatch controls for the timestamp workaround: the request that
// sets an access time on HFS also selects the change time, whose value the
// filesystem ignores. Two things must hold, and neither is visible from the
// os-level tests:
//
//   - the attrlist bitmap and the attribute buffer agree with the kernel's
//     unpack order (MODTIME, CHGTIME, ACCTIME; bsd/vfs/vfs_attrlist.c:1663-1674),
//     because a mismatch would shift a timespec into the wrong attribute rather
//     than fail;
//   - the filesystem is identified from the descriptor (fstatfs64) and compared
//     with the entry's own device, so the dummy change time cannot reach a
//     filesystem that honours it.

import (
	"errors"
	"syscall"
	"testing"
	"unsafe"
)

// TestAtTimesPayloadDarwinPPC pins the bitmap bits and the buffer order for every
// selection the dispatcher can build.
func TestAtTimesPayloadDarwinPPC(t *testing.T) {
	bitMapCount, modBit, chgBit, accBit, _ := syscall.AtAttrBitsForTest()
	// The attrlist constants are ABI: bsd/sys/attr.h:309-316 and
	// ATTR_BIT_MAP_COUNT (bsd/sys/attr.h:303).
	if bitMapCount != 5 || modBit != 0x400 || chgBit != 0x800 || accBit != 0x1000 {
		t.Errorf("attrlist constants = count %d, mod %#x, chg %#x, acc %#x; want 5, 0x400, 0x800, 0x1000",
			bitMapCount, modBit, chgBit, accBit)
	}
	if sz := unsafe.Sizeof(syscall.Timespec{}); sz != 8 {
		t.Errorf("sizeof(Timespec) = %d; want 8, the 32-bit size the kernel unpacks (ATTR_UNPACK_TIME)", sz)
	}

	mod := syscall.Timespec{Sec: 1300000001, Nsec: 11}
	chg := syscall.Timespec{Sec: 1300000002, Nsec: 22}
	acc := syscall.Timespec{Sec: 1300000003, Nsec: 33}

	for _, tc := range []struct {
		name     string
		mod, chg *syscall.Timespec
		acc      *syscall.Timespec
		wantAttr uint32
		want     []syscall.Timespec
	}{{
		name: "modtime only", mod: &mod,
		wantAttr: modBit,
		want:     []syscall.Timespec{mod},
	}, {
		name: "access time only", acc: &acc,
		wantAttr: accBit,
		want:     []syscall.Timespec{acc},
	}, {
		// The HFS workaround's shape: change time then access time, and no
		// modification time, which is the attribute the caller omitted.
		name: "change and access time", chg: &chg, acc: &acc,
		wantAttr: chgBit | accBit,
		want:     []syscall.Timespec{chg, acc},
	}, {
		name: "all three", mod: &mod, chg: &chg, acc: &acc,
		wantAttr: modBit | chgBit | accBit,
		want:     []syscall.Timespec{mod, chg, acc},
	}, {
		name:     "none",
		wantAttr: 0,
		want:     nil,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			count, attr, buf, n := syscall.AtTimesPayloadForTest(tc.mod, tc.chg, tc.acc)
			if count != 5 {
				t.Errorf("bitmapcount = %d; want 5", count)
			}
			if attr != tc.wantAttr {
				t.Errorf("commonattr = %#x; want %#x", attr, tc.wantAttr)
			}
			if n != len(tc.want) {
				t.Fatalf("payload length = %d; want %d", n, len(tc.want))
			}
			for i, want := range tc.want {
				if buf[i] != want {
					t.Errorf("payload[%d] = %+v; want %+v (the kernel unpacks in bitmap order)", i, buf[i], want)
				}
			}
		})
	}
}

// TestAtDummyChgTimeDarwinPPC pins the sentinel: it is not a value any caller can
// have asked for, so if a filesystem ever stored it the change time would be the
// epoch and unmistakable.
func TestAtDummyChgTimeDarwinPPC(t *testing.T) {
	if got := syscall.AtDummyChgTimeForTest(); got.Sec != 0 || got.Nsec != 0 {
		t.Errorf("dummy change time = %+v; want the epoch", got)
	}
}

// TestAtFstypenameDarwinPPC controls the comparison the HFS guard rests on: an
// exact match on the actual filesystem of a real directory, and rejection of a
// longer name or of a prefix.
func TestAtFstypenameDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	fd, err := syscall.Open(dir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fd)

	var fs syscall.Statfs_t
	if err := syscall.Fstatfs(fd, &fs); err != nil {
		t.Fatalf("Fstatfs: %v", err)
	}
	name := cStringDarwinPPC(fs.Fstypename[:])
	if name == "" {
		t.Fatalf("f_fstypename is empty; the guard cannot identify a filesystem")
	}
	if !syscall.AtFstypenameIsForTest(fs.Fstypename[:], name) {
		t.Errorf("the comparison does not match the filesystem's own name %q", name)
	}
	if syscall.AtFstypenameIsForTest(fs.Fstypename[:], name+"x") {
		t.Errorf("the comparison accepted %q for filesystem %q", name+"x", name)
	}
	if len(name) > 1 && syscall.AtFstypenameIsForTest(fs.Fstypename[:], name[:len(name)-1]) {
		t.Errorf("the comparison accepted the prefix %q for filesystem %q", name[:len(name)-1], name)
	}

	if hfs := syscall.AtHFSNameForTest(); name != hfs {
		t.Logf("warning: the temporary directory is on %q, not %q; the HFS access-time workaround is inert for tree-rooted tests on this machine", name, hfs)
	} else {
		t.Logf("temporary directory is on %q, the filesystem the workaround targets", name)
	}

	// A descriptor that is not open must fail rather than report a filesystem.
	if err := syscall.Fstatfs(fd+1000, &fs); err == nil {
		t.Errorf("Fstatfs(bad descriptor) = nil; want an error")
	} else if !errors.Is(err, syscall.EBADF) {
		t.Errorf("Fstatfs(bad descriptor) = %v; want EBADF", err)
	}
}

// TestAtChgTimeForAccessTimeDarwinPPC requires the guard to answer for a real
// directory — either "not HFS" (so no dummy change time) or the sentinel — and to
// change nothing while answering.
func TestAtChgTimeForAccessTimeDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	fd, err := syscall.Open(dir, syscall.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(fd)

	fixture := dir + "/file"
	wfd, err := syscall.Open(fixture, syscall.O_CREAT|syscall.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("creating %s: %v", fixture, err)
	}
	if err := syscall.Close(wfd); err != nil {
		t.Fatal(err)
	}

	var before syscall.Stat_t
	if err := syscall.Lstat(fixture, &before); err != nil {
		t.Fatal(err)
	}

	got, needed, err := syscall.AtChgTimeForAccessTimeForTest(fd, "file")
	if err != nil {
		t.Fatalf("the guard: %v", err)
	}
	var fs syscall.Statfs_t
	if err := syscall.Fstatfs(fd, &fs); err != nil {
		t.Fatal(err)
	}
	if name := cStringDarwinPPC(fs.Fstypename[:]); name == syscall.AtHFSNameForTest() {
		if want := syscall.AtDummyChgTimeForTest(); !needed || got != want {
			t.Errorf("on %q the guard = %+v, %v; want %+v, true", name, got, needed, want)
		}
	} else if needed {
		t.Errorf("on filesystem %q the guard returned needed=true; the dummy change time must never be sent there", name)
	}

	// Reading the filesystem and the entry's device changed nothing, including
	// the change time the request would carry.
	var after syscall.Stat_t
	if err := syscall.Lstat(fixture, &after); err != nil {
		t.Fatal(err)
	}
	if before.Ctimespec != after.Ctimespec || before.Atimespec != after.Atimespec ||
		before.Mtimespec != after.Mtimespec || before.Mode != after.Mode || before.Size != after.Size {
		t.Errorf("metadata changed while identifying the filesystem: before %+v, after %+v", before, after)
	}

	// A missing entry is not proven HFS: the guard reports "no dummy" so that the
	// operation's own ENOENT is what the caller sees, rather than a pre-emptive
	// error that could differ from it.
	if _, needed, err := syscall.AtChgTimeForAccessTimeForTest(fd, "no-such-entry"); err != nil || needed {
		t.Errorf("guard(missing) = _, %v, %v; want needed=false and no error", needed, err)
	}

	// A bad descriptor surfaces the filesystem error rather than a guess.
	if _, _, err := syscall.AtChgTimeForAccessTimeForTest(fd+1000, "file"); err == nil {
		t.Errorf("guard(bad descriptor) = nil error; want the query error, not a guess")
	}
}

// cStringDarwinPPC reads a NUL-terminated int8 field.
func cStringDarwinPPC(b []int8) string {
	var out []byte
	for _, c := range b {
		if c == 0 {
			break
		}
		out = append(out, byte(c))
	}
	return string(out)
}
