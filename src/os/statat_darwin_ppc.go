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
//   - For an ordinary directory File the same confined form is now the primary
//     answer too, which retires the older substitute for the kernels where the
//     backend exists: it used to resolve the descriptor to a path with F_GETPATH
//     and lstat the entry by that path (unix.LstatatByPath), losing the
//     descriptor-based lookup's symlink-race protection and failing with
//     ENAMETOOLONG on a path longer than MAXPATHLEN. The substitute is kept only
//     as the fallback for a kernel where the anchored backend is disabled, since
//     there the descriptor form cannot run at all; it is a read either way, never
//     a write.
//
// Per-arch file (darwin_ppc.go / darwin_ppc64.go), not merged: the same xnu-1228
// kernel serves both ABIs so this logic is arch-neutral, but keeping them
// separate lets each ABI be audited alone and stops a ppc64-only symbol from
// silently compiling into ppc32.
func (f *File) lstatatNolog(name string) (FileInfo, error) {
	var fs fileStat
	var err error
	inRoot := f.inRoot
	confined := inRoot || unix.AnchoredAvailable()
	if rerr := f.pfd.RawControl(func(fd uintptr) {
		if confined {
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
