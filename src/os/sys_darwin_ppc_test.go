// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os_test

import (
	"fmt"
	"internal/syscall/unix"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// sys_darwin_ppc.go sets supportsCloseOnExec to false, because Mac OS X 10.5
// (Leopard) has no O_CLOEXEC: syscall.O_CLOEXEC is 0 there, so open is never
// asked for close-on-exec, and file_unix.go (openFileNolog, openDirNolog) and
// root_unix.go (newRoot) instead follow each open with syscall.CloseOnExec, an
// explicit fcntl(F_SETFD, FD_CLOEXEC).
//
// These tests observe the descriptor flag and its effect across exec rather
// than the constant, because the constant alone is no evidence: reverting it to
// true still compiles, and anything that asserts the constant directly passes
// either way.

// TestCloseOnExecDarwinPPC checks that the three kinds of descriptor os hands
// out through the open path — a file, a directory and a Root handle — carry
// FD_CLOEXEC on Leopard, where the open flag cannot do it.
func TestCloseOnExecDarwinPPC(t *testing.T) {
	dir := t.TempDir()

	f, err := os.Create(filepath.Join(dir, "f"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer f.Close()

	d, err := os.Open(dir)
	if err != nil {
		t.Fatalf("Open(dir): %v", err)
	}
	defer d.Close()

	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	for _, h := range []struct {
		name string
		fd   int
	}{
		{"os.Create(file)", int(f.Fd())},
		{"os.Open(dir)", int(d.Fd())},
		{"os.OpenRoot(dir)", os.RootFd(r)},
	} {
		flags, err := unix.Fcntl(h.fd, syscall.F_GETFD, 0)
		if err != nil {
			t.Errorf("%s: fcntl(fd %d, F_GETFD): %v", h.name, h.fd, err)
			continue
		}
		if flags&syscall.FD_CLOEXEC == 0 {
			t.Errorf("%s: fd %d has F_GETFD flags %#x; FD_CLOEXEC (%#x) is not set, so the handle would survive exec",
				h.name, h.fd, flags, syscall.FD_CLOEXEC)
			continue
		}
		t.Logf("%s: fd %d F_GETFD flags %#x (FD_CLOEXEC set)", h.name, h.fd, flags)
	}
}

// closeOnExecChildEnv carries the handle identities the child process must look
// for; its value is a ";"-separated list of "kind|name|dev|ino".
const closeOnExecChildEnv = "GO_TEST_CLOEXEC_DARWIN_PPC"

// handleID identifies an open file by the object it refers to. A (device,
// inode) pair, unlike a descriptor number, survives exec: exec renumbers
// descriptors and the child's own runtime reuses freed numbers.
type handleID struct {
	kind string // "closed" (must not be inherited) or "leaked" (control)
	name string
	dev  int64
	ino  uint64
}

func newHandleID(t *testing.T, kind, name string, fd int) handleID {
	t.Helper()
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err != nil {
		t.Fatalf("%s: fstat(fd %d): %v", name, fd, err)
	}
	return handleID{kind, name, int64(st.Dev), st.Ino}
}

func encodeHandleIDs(ids []handleID) string {
	var parts []string
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%s|%s|%d|%d", id.kind, id.name, id.dev, id.ino))
	}
	return strings.Join(parts, ";")
}

// TestScriptCloseOnExecDarwinPPCChild is the helper process for
// TestCloseOnExecAcrossExecDarwinPPC: it Fstats every descriptor it can see and
// reports each one that still refers to a handle the parent had open. It is a
// no-op without closeOnExecChildEnv, so an ordinary `go test os` skips it.
func TestScriptCloseOnExecDarwinPPCChild(t *testing.T) {
	spec := os.Getenv(closeOnExecChildEnv)
	if spec == "" {
		t.Skip("helper process for TestCloseOnExecAcrossExecDarwinPPC")
	}

	var wants []handleID
	for _, entry := range strings.Split(spec, ";") {
		field := strings.Split(entry, "|")
		if len(field) != 4 {
			t.Fatalf("malformed spec entry %q", entry)
		}
		dev, err := strconv.ParseInt(field[2], 10, 64)
		if err != nil {
			t.Fatalf("malformed spec entry %q: %v", entry, err)
		}
		ino, err := strconv.ParseUint(field[3], 10, 64)
		if err != nil {
			t.Fatalf("malformed spec entry %q: %v", entry, err)
		}
		wants = append(wants, handleID{field[0], field[1], dev, ino})
	}

	seen, controlSeen := 0, false
	for fd := 0; fd < 256; fd++ {
		var st syscall.Stat_t
		if err := syscall.Fstat(fd, &st); err != nil {
			continue
		}
		seen++
		for _, w := range wants {
			if int64(st.Dev) != w.dev || st.Ino != w.ino {
				continue
			}
			if w.kind == "leaked" {
				controlSeen = true
				t.Logf("fd %d is the parent's %s, as intended for the control", fd, w.name)
			} else {
				t.Errorf("fd %d still refers to the parent's %s: it survived exec, so FD_CLOEXEC was not set on it",
					fd, w.name)
			}
		}
	}

	// A probe that can see no descriptor at all reports "nothing leaked" no
	// matter what the parent did, so a blind run must not be read as success.
	// The control handle was deliberately left without FD_CLOEXEC and must be
	// found; otherwise the leak check proves nothing.
	if seen == 0 {
		t.Errorf("no descriptor was visible to Fstat: the probe is blind and this run proves nothing")
	}
	if !controlSeen {
		t.Errorf("the control handle, deliberately left without FD_CLOEXEC, was not seen: the probe cannot detect a leak")
	}
}

// TestCloseOnExecAcrossExecDarwinPPC checks the property that matters and the
// descriptor flag alone does not: handles opened by os do not reach a child
// process across exec. The child is this test binary re-executed in helper mode.
func TestCloseOnExecAcrossExecDarwinPPC(t *testing.T) {
	dir := t.TempDir()

	f, err := os.Create(filepath.Join(dir, "f"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer f.Close()

	d, err := os.Open(dir)
	if err != nil {
		t.Fatalf("Open(dir): %v", err)
	}
	defer d.Close()

	r, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	// The control: an ordinary open whose FD_CLOEXEC is cleared, so it must
	// reach the child. Without it, a child that could see nothing would pass.
	control, err := os.Create(filepath.Join(dir, "control"))
	if err != nil {
		t.Fatalf("Create(control): %v", err)
	}
	defer control.Close()
	if _, err := unix.Fcntl(int(control.Fd()), syscall.F_SETFD, 0); err != nil {
		t.Fatalf("clearing FD_CLOEXEC on the control handle: %v", err)
	}

	ids := []handleID{
		newHandleID(t, "closed", "Create(file)", int(f.Fd())),
		newHandleID(t, "closed", "Open(dir)", int(d.Fd())),
		newHandleID(t, "closed", "OpenRoot(dir)", os.RootFd(r)),
		newHandleID(t, "leaked", "control", int(control.Fd())),
	}

	out, err := os.Create(filepath.Join(dir, "child.out"))
	if err != nil {
		t.Fatalf("Create(child.out): %v", err)
	}
	defer out.Close()
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("Open(%s): %v", os.DevNull, err)
	}
	defer devNull.Close()

	exe := os.Args[0]
	if !filepath.IsAbs(exe) {
		if exe, err = filepath.Abs(exe); err != nil {
			t.Fatalf("filepath.Abs(%q): %v", os.Args[0], err)
		}
	}
	attr := &os.ProcAttr{
		Dir:   dir,
		Env:   append(os.Environ(), closeOnExecChildEnv+"="+encodeHandleIDs(ids)),
		Files: []*os.File{devNull, out, out},
	}
	argv := []string{exe, "-test.run=^TestScriptCloseOnExecDarwinPPCChild$", "-test.v"}
	p, err := os.StartProcess(exe, argv, attr)
	if err != nil {
		t.Fatalf("StartProcess: %v", err)
	}
	st, werr := p.Wait()
	childOut, _ := os.ReadFile(filepath.Join(dir, "child.out"))
	if werr != nil || !st.Success() {
		t.Errorf("child %q exited with %v (%v); a handle that survived exec is a leak.\nchild output:\n%s",
			exe, st, werr, childOut)
		return
	}
	t.Logf("child %q exited with %v; output:\n%s", exe, st, childOut)
}
