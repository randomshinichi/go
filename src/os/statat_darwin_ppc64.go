// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/unix"
	"syscall"
)

// lstatatNolog on darwin/ppc64 is a SUBSTITUTED mechanism. The shared
// implementation (statat_unix.go) calls fstatat with the directory's own
// descriptor, which Leopard cannot do: it has no *at calls, and the
// substitution in internal/syscall/unix refuses any real descriptor with
// ENOTSUP. That would make (*File).Readdir, io/ioutil.ReadDir and net/http's
// directory listings fail. Instead the descriptor is resolved to a path with
// F_GETPATH and the entry is lstat'ed by that path (unix.LstatatByPath, which
// documents the G5 measurement and the cost).
//
// The cost, stated here because it is a divergence from every other Darwin
// target (and the same trade as RemoveAll's, removeall_darwin_ppc64.go): it
// looks the path up and then uses it, so it loses the descriptor-based call's
// symlink-race (TOCTOU) protection. If the directory is moved or replaced
// between the lookup and the lstat, the FileInfo describes the wrong file. It
// reads; it never writes.
//
// A File opened through an os.Root is refused: Root exists to give the
// race-free guarantee this form lacks. (Root's own operations are ENOTSUP on
// this platform, so such a File cannot normally exist; the check keeps it that
// way if that changes.)
func (f *File) lstatatNolog(name string) (FileInfo, error) {
	if f.inRoot {
		return nil, f.wrapErr("lstat", syscall.ENOTSUP)
	}
	var fs fileStat
	var err error
	if rerr := f.pfd.RawControl(func(fd uintptr) {
		err = unix.LstatatByPath(int(fd), name, &fs.sys)
	}); rerr != nil {
		return nil, f.wrapErr("lstat", rerr)
	}
	if err != nil {
		return nil, f.wrapErr("lstat", err)
	}
	fillFileStatFromSys(&fs, name)
	return &fs, nil
}
