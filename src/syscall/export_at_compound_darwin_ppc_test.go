// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import "unsafe"

// Test-only exports for the anchored compound. They live in package syscall
// because the request block, the native routine and the ABI constants are
// unexported, and because this package's own test files cannot import testing
// here (testing pulls in a package that imports syscall). The external test
// package asserts against these values.

// AtCompoundDirectForTest runs one anchored lstat64 request on the calling
// goroutine, with no helper goroutine, and reports the three status words.
//
// It exists so a test can observe the native routine's own thread state: only
// the routine that installs the thread-private cwd can be asked whether it
// cleared it on the same M. Test-only; production callers go through
// atCompoundDispatch so that a locked caller's thread state is never touched.
func AtCompoundDirectForTest(anchor int, path string) (enterErr, identErr, opErr int32) {
	req, err := atNewRequest(SYS_LSTAT64, anchor, path)
	if err != nil {
		return int32(err.(Errno)), 0, 0
	}
	req.out = new(Stat_t)
	req.w1 = uintptr(unsafe.Pointer(req.out))
	atCompound(req)
	return req.enterErr, req.identErr, req.opErr
}

// AtCompoundReqLayoutForTest reports the request-block layout: field names, the
// offsets the assembly constants claim, and the offsets the compiler gives the
// struct.
func AtCompoundReqLayoutForTest() (names []string, assembly, actual []int) {
	var req atCompoundReq
	names = []string{"nr", "anchor", "w0", "w1", "w2", "w3", "w4", "dot", "st1",
		"st2", "enotsup", "r0", "enterErr", "identErr", "opErr"}
	assembly = []int{atReqNr, atReqAnchor, atReqW0, atReqW1, atReqW2, atReqW3,
		atReqW4, atReqDot, atReqSt1, atReqSt2, atReqENOTSUP, atReqR0,
		atReqEnterErr, atReqIdentErr, atReqOpErr}
	actual = []int{
		int(unsafe.Offsetof(req.nr)), int(unsafe.Offsetof(req.anchor)),
		int(unsafe.Offsetof(req.w0)), int(unsafe.Offsetof(req.w1)),
		int(unsafe.Offsetof(req.w2)), int(unsafe.Offsetof(req.w3)),
		int(unsafe.Offsetof(req.w4)), int(unsafe.Offsetof(req.dot)),
		int(unsafe.Offsetof(req.st1)), int(unsafe.Offsetof(req.st2)),
		int(unsafe.Offsetof(req.enotsup)), int(unsafe.Offsetof(req.r0)),
		int(unsafe.Offsetof(req.enterErr)), int(unsafe.Offsetof(req.identErr)),
		int(unsafe.Offsetof(req.opErr)),
	}
	return names, assembly, actual
}

// AtCompoundReqSizeForTest reports the request size and the last field's end.
func AtCompoundReqSizeForTest() (size, need int) {
	var req atCompoundReq
	return int(unsafe.Sizeof(req)), atReqOpErr + 4
}

// WordSizesForTest reports the sizes the native routine's 32-bit loads assume.
func WordSizesForTest() (word, status, bytePtr, statPtr int) {
	var req atCompoundReq
	return int(unsafe.Sizeof(req.w0)), int(unsafe.Sizeof(req.r0)),
		int(unsafe.Sizeof(req.keepPath)), int(unsafe.Sizeof(req.st1))
}

// StatIdentityLayoutForTest reports the Stat_t fields the identity check
// compares by offset.
func StatIdentityLayoutForTest() (devOff, inoOff, devSize, inoSize, assemblyDev, assemblyIno int) {
	var st Stat_t
	return int(unsafe.Offsetof(st.Dev)), int(unsafe.Offsetof(st.Ino)),
		int(unsafe.Sizeof(st.Dev)), int(unsafe.Sizeof(st.Ino)), atStatDev, atStatIno
}

// AtDotForTest reports the pathname the identity check stats.
func AtDotForTest() []byte { return atDot[:] }

// OpenFromLayoutForTest reports the ppc32 struct fopenfrom that F_OPENFROM
// copies in (bsd/sys/fcntl.h:450-458, measured on the target).
func OpenFromLayoutForTest() (size, flagsOff, modeOff, pathOff int) {
	var arg struct {
		flags    uint32
		mode     uint16
		_        uint16
		pathname *byte
	}
	return int(unsafe.Sizeof(arg)), int(unsafe.Offsetof(arg.flags)),
		int(unsafe.Offsetof(arg.mode)), int(unsafe.Offsetof(arg.pathname))
}
