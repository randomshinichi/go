// Copyright 2009,2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && !ppc64

// Getdirentries and fdopendir moved here, unchanged, from syscall_darwin.go so
// that darwin/ppc64 can provide its own Getdirentries
// (syscall_darwin_ppc64.go): Leopard's libSystem has no fdopendir.
// The libc_fdopendir_trampoline declaration and its cgo_import_dynamic stay in
// syscall_darwin.go, where mkasm.go reads them.

package syscall

import (
	"internal/abi"
	"unsafe"
)

// used by internal/poll
//go:linkname fdopendir
func fdopendir(fd int) (dir uintptr, err error) {
	r0, _, e1 := syscallPtr(abi.FuncPCABI0(libc_fdopendir_trampoline), uintptr(fd), 0, 0)
	dir = r0
	if e1 != 0 {
		err = errnoErr(e1)
	}
	return
}

func Getdirentries(fd int, buf []byte, basep *uintptr) (n int, err error) {
	// Simulate Getdirentries using fdopendir/readdir_r/closedir.
	// We store the number of entries to skip in the seek
	// offset of fd. See issue #31368.
	// It's not the full required semantics, but should handle the case
	// of calling Getdirentries or ReadDirent repeatedly.
	// It won't handle assigning the results of lseek to *basep, or handle
	// the directory being edited underfoot.
	skip, err := Seek(fd, 0, 1 /* SEEK_CUR */)
	if err != nil {
		return 0, err
	}

	// We need to duplicate the incoming file descriptor
	// because the caller expects to retain control of it, but
	// fdopendir expects to take control of its argument.
	// Just Dup'ing the file descriptor is not enough, as the
	// result shares underlying state. Use openat to make a really
	// new file descriptor referring to the same directory.
	fd2, err := openat(fd, ".", O_RDONLY, 0)
	if err != nil {
		return 0, err
	}
	d, err := fdopendir(fd2)
	if err != nil {
		Close(fd2)
		return 0, err
	}
	defer closedir(d)

	var cnt int64
	for {
		var entry Dirent
		var entryp *Dirent
		e := readdir_r(d, &entry, &entryp)
		if e != 0 {
			return n, errnoErr(e)
		}
		if entryp == nil {
			break
		}
		if skip > 0 {
			skip--
			cnt++
			continue
		}
		reclen := int(entry.Reclen)
		if reclen > len(buf) {
			// Not enough room. Return for now.
			// The counter will let us know where we should start up again.
			// Note: this strategy for suspending in the middle and
			// restarting is O(n^2) in the length of the directory. Oh well.
			break
		}
		// Copy entry into return buffer.
		copy(buf, unsafe.Slice((*byte)(unsafe.Pointer(&entry)), reclen))
		buf = buf[reclen:]
		n += reclen
		cnt++
	}
	// Set the seek offset of the input fd to record
	// how many files we've already returned.
	_, err = Seek(fd, cnt, 0 /* SEEK_SET */)
	if err != nil {
		return n, err
	}

	return n, nil
}
