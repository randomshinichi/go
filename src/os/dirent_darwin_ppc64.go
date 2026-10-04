// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"syscall"
	"unsafe"
)

// Directory reading on darwin/ppc64 is a SUBSTITUTED mechanism. Other Darwin
// targets read directories through libc fdopendir/readdir_r/closedir
// (dir_darwin.go, internal/poll/fd_opendir_darwin.go). Leopard's libSystem has
// no fdopendir, and plain readdir_r returns the legacy 32-bit-inode record
// (reclen 12 for ".") rather than the syscall.Dirent of ztypes_darwin_ppc64.go.
// So darwin/ppc64 uses the generic getdirentries loop in dir_unix.go over
// syscall.ReadDirent, which is syscall.Getdirentries over raw
// SYS_getdirentries64 (syscall_darwin_ppc64.go); that call returns the
// 64-bit-inode record described by syscall.Dirent: ino@0 seekoff@8 reclen@16
// namlen@18 type@20 name@21. Measured on the G5:
// out/stage2-20261003/D7-syscall/g5-request2/out-d7r2/C-dirent-test.txt.
// Dirent.Seekoff is an opaque kernel cookie and is never read here.

func direntIno(buf []byte) (uint64, bool) {
	return readInt(buf, unsafe.Offsetof(syscall.Dirent{}.Ino), unsafe.Sizeof(syscall.Dirent{}.Ino))
}

func direntReclen(buf []byte) (uint64, bool) {
	return readInt(buf, unsafe.Offsetof(syscall.Dirent{}.Reclen), unsafe.Sizeof(syscall.Dirent{}.Reclen))
}

func direntNamlen(buf []byte) (uint64, bool) {
	return readInt(buf, unsafe.Offsetof(syscall.Dirent{}.Namlen), unsafe.Sizeof(syscall.Dirent{}.Namlen))
}

func direntType(buf []byte) FileMode {
	off := unsafe.Offsetof(syscall.Dirent{}.Type)
	if off >= uintptr(len(buf)) {
		return ^FileMode(0) // unknown
	}
	typ := buf[off]
	switch typ {
	case syscall.DT_BLK:
		return ModeDevice
	case syscall.DT_CHR:
		return ModeDevice | ModeCharDevice
	case syscall.DT_DIR:
		return ModeDir
	case syscall.DT_FIFO:
		return ModeNamedPipe
	case syscall.DT_LNK:
		return ModeSymlink
	case syscall.DT_REG:
		return 0
	case syscall.DT_SOCK:
		return ModeSocket
	}
	return ^FileMode(0) // unknown
}
