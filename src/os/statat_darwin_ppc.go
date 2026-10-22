// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import "internal/syscall/unix"

// lstatatNolog on darwin/ppc (and its ppc64 twin) needs its own dispatch,
// because Leopard's *at substitution answers two different questions for a
// directory descriptor:
//
//   - For a File opened through an os.Root, the descriptor-relative form is the
//     whole point, and it is available: unix.Fstatat takes a real descriptor and
//     runs lstat64 under the thread-private cwd installed from that descriptor
//     (internal/syscall/unix/at_darwin_ppc.go). This is the confined answer and
//     it is what a Root's entries get. On a kernel where the anchored backend
//     is not enabled it returns ENOTSUP, which is the honest answer rather than
//     a pathname lookup that could describe a replaced directory.
//
//   - For an ordinary directory File, the confined form would work too, but the
//     listing path it belongs to still routes through the older substitute: the
//     descriptor is resolved to a path with F_GETPATH and the entry is lstat'ed
//     by that path (unix.LstatatByPath). That loses the descriptor-based
//     lookup's symlink-race (TOCTOU) protection -- if the directory is moved or
//     replaced between the lookup and the lstat, the FileInfo describes the
//     wrong file -- and it is a read, never a write. Retiring it in favour of
//     unix.Fstatat is a separate, reviewable step with its own controls
//     (docs/leopard-at-compat-design.md, "Integration").
//
// Per-arch file (darwin_ppc.go / darwin_ppc64.go), not merged: the same xnu-1228
// kernel serves both ABIs so this logic is arch-neutral, but keeping them
// separate lets each ABI be audited alone and stops a ppc64-only symbol from
// silently compiling into ppc32.
func (f *File) lstatatNolog(name string) (FileInfo, error) {
	var fs fileStat
	var err error
	inRoot := f.inRoot
	if rerr := f.pfd.RawControl(func(fd uintptr) {
		if inRoot {
			err = unix.Fstatat(int(fd), name, &fs.sys, unix.AT_SYMLINK_NOFOLLOW)
			return
		}
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
