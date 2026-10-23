// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall

import (
	"runtime"
	"unsafe"
	_ "unsafe" // for linkname
)

// The anchored primitives are reached from internal/syscall/unix, which owns the
// dispatch. These are push declarations: they are what makes the names
// linkname-able from another package.
//
//go:linkname atOpenFrom
//go:linkname atUnlinkFrom
//go:linkname atOpenatCompound
//go:linkname atRmdirCompound
//go:linkname atReadlinkatCompound
//go:linkname atMkdiratCompound
//go:linkname atSymlinkatCompound
//go:linkname atStatatCompound
//go:linkname atChownatCompound
//go:linkname atChmodatCompound
//go:linkname atChmodNofollowCompound
//go:linkname atUtimensatCompound
//go:linkname atRenameatCompound

// Leopard has no *at family, but xnu-1228 has two kernel interfaces that make
// most of os.Root expressible without reconstructing a pathname:
//
//   - fcntl F_OPENFROM(56): open a pathname relative to a held directory
//     descriptor (bsd/kern/kern_descrip.c:1245-1300).
//   - fcntl F_UNLINKFROM(57): the same directory anchor for unlink(2), and
//     only unlink: it is not a remove-directory primitive (kern_descrip.c:1305).
//
// Neither covers a second anchored operand, so Linkat and a Rename between two
// different anchors stay refused; see internal/syscall/unix/at_darwin_ppc.go
// for the complete dispatch table and the reasons.
//
// Every other representable operation needs the *thread-private* working
// directory that syscall 349 (__pthread_fchdir) installs for one thread
// (bsd/vfs/vfs_syscalls.c:1799-1909). Doing that from Go code is not safe: the
// cwd is per-thread state, and a goroutine can migrate between the syscall that
// sets it and the syscall that uses it, leaving the private cwd on a thread
// that later runs unrelated application code. So one native routine does the
// whole compound in a single entersyscall/exitsyscall scope, on one M, with no
// Go code, allocation, callback or preemption point inside:
//
//	entersyscall
//	raw 349(anchor)                       install the anchor as private cwd
//	stat64(".", st1); fstat64(anchor, st2) identity of the anchor it really is
//	the selected raw operation        (the kernel resolves "." against the
//	                                   private cwd, so this is anchored I/O)
//	raw 349(-1)                           clear the private cwd
//	exitsyscall
//
// The identity comparison is load-bearing. fchdir follows a covering mount
// point (vfs_syscalls.c:1851-1862) whereas F_OPENFROM starts at the held vnode,
// so on a mount-covered anchor the private cwd is a different object than the
// descriptor this call is anchored to. A dev/ino mismatch clears the private
// cwd and returns ENOTSUP before any operation runs.
//
// A failed clear after a successful install is an internal invariant failure,
// not a filesystem error: the M is left with a private cwd it must not carry
// into application code. The routine therefore exits the process with a raw
// exit(127) *before* exitsyscall, rather than reporting the failure to Go,
// where the next scheduling point could run another goroutine on that M.
//
// Everything the native routine reads is reachable from the request through
// typed Go pointers, so the garbage collector cannot free it while the call is
// in flight; the w words are copies of those addresses and are not traced.
// The request itself is always heap-allocated, and the routine runs on a
// transient helper goroutine, so a caller that is locked to its thread (and may
// hold its own private cwd) never has its thread state inspected or changed.

// Layout of atCompoundReq. The five w words are handed to the kernel in
// register order R3..R7, so an operation assigns them positionally.
//
// at_compound_darwin_ppc.s carries the same numbers as assembler constants;
// TestAtCompoundLayoutDarwinPPC checks them against this struct, because a
// disagreement would silently corrupt the request instead of failing to build.
const (
	atReqNr       = 0  // int32: raw syscall number of the selected operation
	atReqAnchor   = 4  // int32: directory descriptor to anchor the operation
	atReqW0       = 8  // uintptr: R3
	atReqW1       = 12 // uintptr: R4
	atReqW2       = 16 // uintptr: R5
	atReqW3       = 20 // uintptr: R6
	atReqW4       = 24 // uintptr: R7
	atReqDot      = 28 // *byte: "." used to identify the private cwd
	atReqSt1      = 32 // *Stat_t: stat64(".") output
	atReqSt2      = 36 // *Stat_t: fstat64(anchor) output
	atReqENOTSUP  = 40 // int32: errno reported for an anchor identity mismatch
	atReqR0       = 44 // int32: operation result word
	atReqEnterErr = 48 // int32: errno from the private-cwd install
	atReqIdentErr = 52 // int32: errno from the identity check, or ENOTSUP
	atReqOpErr    = 56 // int32: errno from the selected operation
)

// atCompoundReq is one anchored operation. A zero status in all three result
// words means the operation succeeded; atReqR0 then holds its result.
type atCompoundReq struct {
	nr       int32
	anchor   int32
	w0       uintptr
	w1       uintptr
	w2       uintptr
	w3       uintptr
	w4       uintptr
	dot      *byte
	st1      *Stat_t
	st2      *Stat_t
	enotsup  int32
	r0       int32
	enterErr int32
	identErr int32
	opErr    int32

	// Not read by the native routine. These keep the objects whose addresses
	// appear in the w words (and in dot, st1 and st2) reachable from the
	// request, which is what stops the garbage collector from freeing a
	// buffer while its address is only held in a uintptr word.
	keepPath *byte
	keepBuf  *byte

	// out is where the routine writes a Stat_t result. It is request memory,
	// not the caller's: the caller is parked on a channel while the helper
	// runs, and a parked goroutine's stack may be moved or shrunk.
	out *Stat_t

	// keepAlist and keepAttr hold the two buffers a setattrlist(2) request
	// points at, for the same reason. They are after atReqOpErr, so they do not
	// move any offset the native routine uses.
	keepAlist *atAttrlist
	keepAttr  *byte
}

// The two private (SPI) fcntl commands this file uses, from bsd/sys/fcntl.h:235.
const (
	_F_OPENFROM   = 56
	_F_UNLINKFROM = 57
)

// Offsets of the two Stat_t fields the identity check compares. The same
// numbers appear in at_compound_darwin_ppc.s; TestAtCompoundLayoutDarwinPPC
// checks them against the struct definition.
const (
	atStatDev = 0 // Stat_t.Dev, int32
	atStatIno = 8 // Stat_t.Ino, uint64
)

// atDot is the pathname the identity check stats. It is a package variable so
// that its address is stable and never on a goroutine stack.
var atDot = [2]byte{'.', 0}

// atAttrlist is the ppc32 struct attrlist that setattrlist(2) copies in
// (bsd/sys/attr.h:65-86 for the 32-bit field widths; the shape is fixed-width,
// so it is the same 24 bytes on both ABIs). Measured on the target in
// runs/root-at-metadata-20261023: sizeof 24, bitmapcount at 0, reserved at 2,
// commonattr/volattr/dirattr/fileattr/forkattr at 4/8/12/16/20.
type atAttrlist struct {
	bitmapcount uint16
	reserved    uint16
	commonattr  uint32
	volattr     uint32
	dirattr     uint32
	fileattr    uint32
	forkattr    uint32
}

// The attrlist values the no-follow metadata forms use, from bsd/sys/attr.h.
// ATTR_BIT_MAP_COUNT is the only accepted bitmapcount
// (bsd/vfs/vfs_attrlist.c:1578-1581), FSOPT_NOFOLLOW keeps the final component
// unresolved (vfs_attrlist.c:1571-1574), and ACCESSMASK/MODTIME/ACCTIME are the
// three attributes Root.Chmod and Root.Chtimes need.
const (
	atAttrBitMapCount   = 5
	atFsOptNoFollow     = 0x00000001
	atAttrCmnModTime    = 0x00000400
	atAttrCmnAcctime    = 0x00001000
	atAttrCmnAccessMask = 0x00020000
)

// atCompound is implemented in at_compound_darwin_ppc.s.
func atCompound(req *atCompoundReq)

// atCompoundResult is the helper goroutine's answer. It holds no pointers, so
// publishing it cannot extend the life of any request memory.
type atCompoundResult struct {
	r0  int32
	err error
}

// atCompoundDispatch runs one operation on a transient helper goroutine and
// waits for it.
//
// The helper is an ordinary unlocked goroutine, not a locked one, and it is
// created per call rather than pooled: an operation can block for as long as
// the filesystem takes (a FIFO open, a hung NFS mount), and a pool would make
// one such call prevent another from making progress. A caller that used
// LockOSThread keeps its own M, and therefore its own private cwd, because the
// compound never runs there.
//
// req must be heap-allocated. The caller keeps it (and so everything it
// references) alive until this returns; runtime.KeepAlive below is that
// statement for the compiler.
func atCompoundDispatch(req *atCompoundReq) (int32, error) {
	done := make(chan atCompoundResult, 1)
	go func() {
		atCompound(req)
		res := atCompoundResult{r0: req.r0}
		// A private-cwd install failure changed nothing. An identity or
		// operation failure was followed by a successful clear. A clear
		// failure never returns here: the native routine exits the process.
		switch {
		case req.enterErr != 0:
			res.err = errnoErr(Errno(req.enterErr))
		case req.identErr != 0:
			res.err = errnoErr(Errno(req.identErr))
		case req.opErr != 0:
			res.err = errnoErr(Errno(req.opErr))
		}
		done <- res
	}()
	res := <-done
	runtime.KeepAlive(req)
	return res.r0, res.err
}

// atNewRequest allocates a request for anchor and path.
func atNewRequest(nr int, anchor int, path string) (*atCompoundReq, error) {
	p, err := ByteSliceFromString(path)
	if err != nil {
		return nil, err
	}
	req := &atCompoundReq{
		nr:       int32(nr),
		anchor:   int32(anchor),
		dot:      &atDot[0],
		st1:      new(Stat_t),
		st2:      new(Stat_t),
		enotsup:  int32(ENOTSUP),
		w0:       uintptr(unsafe.Pointer(&p[0])),
		keepPath: &p[0],
	}
	return req, nil
}

// atAddPath adds a second pathname operand to a request. A path with an
// interior NUL is refused the same way the first operand was.
func atAddPath(req *atCompoundReq, path string, word int) error {
	p, err := ByteSliceFromString(path)
	if err != nil {
		return err
	}
	req.keepBuf = &p[0]
	w := uintptr(unsafe.Pointer(&p[0]))
	switch word {
	case 1:
		req.w1 = w
	case 2:
		req.w2 = w
	default:
		panic("syscall: bad atCompoundReq word")
	}
	return nil
}

// atOpenFrom is Openat for a real directory descriptor through F_OPENFROM.
// Only the forms the kernel serves correctly are accepted here: with O_CREAT
// and without O_EXCL the kernel's creation retry clears ni_dvp and can return
// a spurious ENOENT (bsd/vfs/vfs_vnops.c:294-305), so that case is routed to
// the private-cwd compound by the caller instead.
func atOpenFrom(anchor int, path string, flags int, mode uint32) (int, error) {
	p, err := ByteSliceFromString(path)
	if err != nil {
		return -1, err
	}
	var arg struct {
		flags    uint32
		mode     uint16
		_        uint16
		pathname *byte
	}
	arg.flags = uint32(flags)
	arg.mode = uint16(mode)
	arg.pathname = &p[0]
	r0, err := fcntlPtr(anchor, _F_OPENFROM, unsafe.Pointer(&arg))
	runtime.KeepAlive(p)
	if err != nil {
		return -1, err
	}
	return r0, nil
}

// atUnlinkFrom is Unlinkat(flags 0) for a real directory descriptor through
// F_UNLINKFROM. There is deliberately no remove-directory form: the kernel
// interface has none, and rmdir has no anchored equivalent other than the
// private-cwd compound.
func atUnlinkFrom(anchor int, path string) error {
	p, err := ByteSliceFromString(path)
	if err != nil {
		return err
	}
	_, err = fcntlPtr(anchor, _F_UNLINKFROM, unsafe.Pointer(&p[0]))
	runtime.KeepAlive(p)
	return err
}

// atOpenatCompound is Openat for a real directory descriptor through the
// private-cwd compound: open(2) relative to the anchor, for the O_CREAT form
// F_OPENFROM cannot serve.
func atOpenatCompound(anchor int, path string, flags int, mode uint32) (int, error) {
	req, err := atNewRequest(SYS_OPEN, anchor, path)
	if err != nil {
		return -1, err
	}
	req.w1 = uintptr(flags)
	req.w2 = uintptr(mode)
	r0, err := atCompoundDispatch(req)
	if err != nil {
		return -1, err
	}
	return int(r0), nil
}

// atRmdirCompound is Unlinkat(AT_REMOVEDIR) for a real directory descriptor.
func atRmdirCompound(anchor int, path string) error {
	req, err := atNewRequest(SYS_RMDIR, anchor, path)
	if err != nil {
		return err
	}
	_, err = atCompoundDispatch(req)
	return err
}

// atReadlinkatCompound is Readlinkat for a real directory descriptor.
func atReadlinkatCompound(anchor int, path string, buf []byte) (int, error) {
	var p *byte
	if len(buf) > 0 {
		p = &buf[0]
	} else {
		p = &atDot[0]
	}
	req, err := atNewRequest(SYS_READLINK, anchor, path)
	if err != nil {
		return 0, err
	}
	req.w1 = uintptr(unsafe.Pointer(p))
	req.w2 = uintptr(len(buf))
	req.keepBuf = p
	r0, err := atCompoundDispatch(req)
	runtime.KeepAlive(buf)
	if err != nil {
		return 0, err
	}
	return int(r0), nil
}

// atMkdiratCompound is Mkdirat for a real directory descriptor.
func atMkdiratCompound(anchor int, path string, mode uint32) error {
	req, err := atNewRequest(SYS_MKDIR, anchor, path)
	if err != nil {
		return err
	}
	req.w1 = uintptr(mode)
	_, err = atCompoundDispatch(req)
	return err
}

// atSymlinkatCompound is Symlinkat for a real directory descriptor: only the
// new name is anchored, the link contents are literal text.
func atSymlinkatCompound(target string, anchor int, newname string) error {
	req, err := atNewRequest(SYS_SYMLINK, anchor, target)
	if err != nil {
		return err
	}
	if err := atAddPath(req, newname, 1); err != nil {
		return err
	}
	_, err = atCompoundDispatch(req)
	return err
}

// atStatatCompound is Fstatat for a real directory descriptor. follow selects
// stat64 over lstat64.
//
// The result is written to request memory and copied into st after the helper
// returns, so st may point into the caller's stack frame.
func atStatatCompound(anchor int, path string, st *Stat_t, follow bool) error {
	nr := SYS_LSTAT64
	if follow {
		nr = SYS_STAT64
	}
	req, err := atNewRequest(nr, anchor, path)
	if err != nil {
		return err
	}
	req.out = new(Stat_t)
	req.w1 = uintptr(unsafe.Pointer(req.out))
	if _, err := atCompoundDispatch(req); err != nil {
		return err
	}
	*st = *req.out
	return nil
}

// atChownatCompound is Fchownat for a real directory descriptor. follow selects
// chown over lchown.
func atChownatCompound(anchor int, path string, uid, gid int, follow bool) error {
	nr := SYS_LCHOWN
	if follow {
		nr = SYS_CHOWN
	}
	req, err := atNewRequest(nr, anchor, path)
	if err != nil {
		return err
	}
	req.w1 = uintptr(uid)
	req.w2 = uintptr(gid)
	_, err = atCompoundDispatch(req)
	return err
}

// atChmodatCompound is Fchmodat for a real directory descriptor, following
// form only. The no-follow form needs setattrlist with FSOPT_NOFOLLOW, whose
// libc/permission/errno parity has not been established, so it stays refused
// (see internal/syscall/unix/at_darwin_ppc.go) rather than being approximated
// by a following chmod.
func atChmodatCompound(anchor int, path string, mode uint32) error {
	req, err := atNewRequest(SYS_CHMOD, anchor, path)
	if err != nil {
		return err
	}
	req.w1 = uintptr(mode)
	_, err = atCompoundDispatch(req)
	return err
}

// atSetattrCompound is one setattrlist(2) request relative to a held directory
// descriptor, run in the private-cwd compound. The kernel takes five arguments
// (path, attrlist, buffer, bufferSize, options), so they occupy R3..R7 — the
// request's five argument words — and no packing is involved.
//
// The attribute payload is passed as bytes, but it is built from typed values
// by the callers below, so its field order and width are the compiler's, not a
// hand-written byte sequence.
func atSetattrCompound(anchor int, path string, alist *atAttrlist, attr []byte) error {
	req, err := atNewRequest(SYS_SETATTRLIST, anchor, path)
	if err != nil {
		return err
	}
	var buf *byte
	if len(attr) > 0 {
		buf = &attr[0]
	} else {
		// An attribute-less request still has to resolve the pathname (that is
		// how utimensat(2) reports ENOENT when both times are UTIME_OMIT), but
		// a zero bufferSize is unusable: the kernel allocates bufferSize bytes
		// and fails with ENOMEM for a zero-size request (measured on the
		// target; runs/root-at-metadata-20261023). One harmless byte is passed
		// instead, and no attribute is unpacked from it.
		buf = &atDot[0]
		attr = atDot[:1]
	}
	req.keepAlist = alist
	req.keepAttr = buf
	req.w1 = uintptr(unsafe.Pointer(alist))
	req.w2 = uintptr(unsafe.Pointer(buf))
	req.w3 = uintptr(len(attr))
	req.w4 = uintptr(atFsOptNoFollow)
	_, err = atCompoundDispatch(req)
	runtime.KeepAlive(alist)
	runtime.KeepAlive(attr)
	return err
}

// atChmodNofollowCompound is Fchmodat(AT_SYMLINK_NOFOLLOW) for a real directory
// descriptor: setattrlist with FSOPT_NOFOLLOW selecting ATTR_CMN_ACCESSMASK.
//
// Leopard has no lchmod syscall, and libSystem's lchmod is built on exactly this
// request for a symlink and on plain chmod(2) otherwise (disassembled on the
// target: _lchmod lstat's, compares st_mode's type nibble with S_IFLNK, and
// calls setattrlist with bitmapcount 5, commonattr ATTR_CMN_ACCESSMASK, a 4-byte
// mode buffer and FSOPT_NOFOLLOW). os.Root.Chmod resolves symlinks before it
// gets here, so the no-follow form is the one that keeps a replaced component
// inside the root.
func atChmodNofollowCompound(anchor int, path string, mode uint32) error {
	al := &atAttrlist{bitmapcount: atAttrBitMapCount, commonattr: atAttrCmnAccessMask}
	attr := &struct{ mode uint32 }{mode: mode}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(attr)), unsafe.Sizeof(*attr))
	return atSetattrCompound(anchor, path, al, buf)
}

// atUtimensatCompound is Utimensat(AT_SYMLINK_NOFOLLOW) for a real directory
// descriptor. A nil mod or acc omits that attribute entirely; the kernel unpacks
// the selected timespec attributes in bitmap order, MODTIME before ACCTIME
// (bsd/vfs/vfs_attrlist.c:1663-1674), and an omitted one is never read back from
// the file and written out again.
//
// Each timespec is 8 bytes for a 32-bit caller: the kernel unpacks
// sizeof(struct timespec) for a non-64-bit process (ATTR_UNPACK_TIME,
// vfs_attrlist.c:1526-1536), and its timespec is int32 sec/nsec here.
func atUtimensatCompound(anchor int, path string, mod, acc *Timespec) error {
	var al atAttrlist
	al.bitmapcount = atAttrBitMapCount
	var buf [2]Timespec
	n := 0
	if mod != nil {
		al.commonattr |= atAttrCmnModTime
		buf[n] = *mod
		n++
	}
	if acc != nil {
		al.commonattr |= atAttrCmnAcctime
		buf[n] = *acc
		n++
	}
	attr := unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), n*int(unsafe.Sizeof(buf[0])))
	return atSetattrCompound(anchor, path, &al, attr)
}

// atRenameatCompound is Renameat when both operands share one anchor.
// rename(2) does not follow a final symlink on either operand
// (bsd/vfs/vfs_syscalls.c:4738,4751), which is what Root.Rename requires.
func atRenameatCompound(anchor int, oldpath, newpath string) error {
	req, err := atNewRequest(SYS_RENAME, anchor, oldpath)
	if err != nil {
		return err
	}
	if err := atAddPath(req, newpath, 1); err != nil {
		return err
	}
	_, err = atCompoundDispatch(req)
	return err
}
