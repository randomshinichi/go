// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os_test

// This file is the ppc32 counterpart of root_darwin_ppc64_test.go: the Root
// semantics that were exercised on ppc64 only. removeall_darwin_ppc_test.go
// already pins the refusals themselves and the working directory's endpoints;
// what it cannot see is the two things below.
//
//  1. The confined branch in statat_darwin_ppc.go for a File opened in a Root.
//     Before the anchored backend existed every Root operation was ENOTSUP here,
//     so an ordinary program could not construct such a File and no test reached
//     that branch: deleting it would have left the suite green. MarkInRoot
//     (export_darwin_ppc_test.go) still fabricates the state, because a File that
//     is in a root but was opened elsewhere cannot be produced through os.Root.
//
//  2. That a confined operation does not change the process working directory
//     even transiently. Comparing Getwd before and after an operation is blind
//     to a swap that is restored before the operation returns -- which is the
//     shape a cwd-based fallback would have. TestRootConfinedOps-
//     NoChdirDarwinPPC records the chdir where package os performs it, and
//     TestRootConfinedOpsNoChdirSyscallDarwinPPC records it one layer lower, at
//     the libc functions syscall.Chdir/Fchdir are imported from, so neither the
//     os calls nor a raw syscall can hide a swap. There is no polling: a poller
//     would race with a swap this short, and a race is not a detector.
//
// Neither test claims Root works, or makes it work. Both are instruments for
// refusals that are believed correct today.

import (
	"bytes"
	"errors"
	"internal/testlog"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// chdirRecorder is a testlog.Interface that keeps the directories package os
// moved into. os.Chdir and (*File).Chdir both call Logger().Chdir with the
// directory they arrived at, immediately after the move succeeded, so a chdir
// that is undone before a confined operation returns is still recorded -- the
// two endpoints of a swap-and-restore are identical, the swap is not.
//
// The other three methods are noisy and irrelevant to confinement, so they are
// no-ops: this is installed process-wide and must not slow the suite down.
type chdirRecorder struct {
	mu   sync.Mutex
	dirs []string
}

func (*chdirRecorder) Getenv(string) {}
func (*chdirRecorder) Stat(string)   {}
func (*chdirRecorder) Open(string)   {}

func (r *chdirRecorder) Chdir(dir string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dirs = append(r.dirs, dir)
}

func (r *chdirRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dirs = nil
}

func (r *chdirRecorder) since() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.dirs...)
}

var (
	chdirLog     chdirRecorder
	chdirLogOnce sync.Once
)

// installChdirRecorder turns on the chdir record for this process.
func installChdirRecorder() {
	// testlog.SetLogger may be called only once per process, hence the Once.
	chdirLogOnce.Do(func() { testlog.SetLogger(&chdirLog) })
}

// confinedTree creates the entries a Root operation can name: a file, a
// directory and a symlink to the file. Every operation below is non-destructive
// even if a future platform makes it succeed.
func confinedTree(t *testing.T, dir, fileContent string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte(fileContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("f", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
}

// rootOpDarwinPPC is one entry of the Root surface a confined operation is
// exposed through. It is the union of the sets pinned by
// root_darwin_ppc64_test.go and removeall_darwin_ppc_test.go, so a ppc32
// divergence shows up as a difference rather than as silence.
//
// Each operation runs against its own fresh tree. Several of them mutate the
// tree (Remove, Rename, Mkdir, Create), and a shared tree made the expectations
// depend on iteration order: a Refused-everything version could share one
// because nothing changed, a version that succeeds cannot.
//
// want is nil for an operation the anchored backend serves on Leopard 10.5.8,
// and the refusal errno for the one that stays refused by design: Link, because
// link(2) follows the source symlink and this kernel has no two-anchor
// alternative. A refusal is listed explicitly so that changing it is a visible
// edit. RemoveAll, Chmod and Chtimes used to be listed here as refusals; they are
// served now (the confined directory walk and the setattrlist(FSOPT_NOFOLLOW)
// metadata forms) and are asserted as operations, with their own effect controls
// in root_metadata_darwin_ppc_test.go and removeall_darwin_ppc_test.go.
type rootOpDarwinPPC struct {
	name string
	run  func(r *os.Root) error
	want error
}

func rootOpsDarwinPPC() []rootOpDarwinPPC {
	now := time.Now()
	closeErr := func(f *os.File, err error) error {
		if f != nil {
			f.Close()
		}
		return err
	}
	return []rootOpDarwinPPC{
		{"Open", func(r *os.Root) error { return closeErr(r.Open("f")) }, nil},
		{"OpenFile", func(r *os.Root) error { return closeErr(r.OpenFile("f", os.O_RDONLY, 0)) }, nil},
		{"Create", func(r *os.Root) error { return closeErr(r.Create("new")) }, nil},
		{"OpenRoot", func(r *os.Root) error {
			sub, err := r.OpenRoot("sub")
			if sub != nil {
				sub.Close()
			}
			return err
		}, nil},
		{"Mkdir", func(r *os.Root) error { return r.Mkdir("m", 0o755) }, nil},
		{"MkdirAll", func(r *os.Root) error { return r.MkdirAll("m/n", 0o755) }, nil},
		{"Remove", func(r *os.Root) error { return r.Remove("f") }, nil},
		{"RemoveEmptyDir", func(r *os.Root) error { return r.Remove("sub") }, nil},
		{"RemoveAll", func(r *os.Root) error { return r.RemoveAll("sub") }, nil},
		{"Stat", func(r *os.Root) error { _, err := r.Stat("f"); return err }, nil},
		{"Lstat", func(r *os.Root) error { _, err := r.Lstat("f"); return err }, nil},
		{"Readlink", func(r *os.Root) error { _, err := r.Readlink("link"); return err }, nil},
		{"ReadFile", func(r *os.Root) error { _, err := r.ReadFile("f"); return err }, nil},
		{"WriteFile", func(r *os.Root) error { return r.WriteFile("w", []byte("x"), 0o644) }, nil},
		{"Rename", func(r *os.Root) error { return r.Rename("f", "g") }, nil},
		{"Link", func(r *os.Root) error { return r.Link("f", "g") }, syscall.ENOTSUP},
		{"Symlink", func(r *os.Root) error { return r.Symlink("f", "s") }, nil},
		{"Chmod", func(r *os.Root) error { return r.Chmod("f", 0o600) }, nil},
		{"Chown", func(r *os.Root) error { return r.Chown("f", os.Getuid(), os.Getgid()) }, nil},
		{"Chtimes", func(r *os.Root) error { return r.Chtimes("f", now, now) }, nil},
	}
}

// check applies the expected outcome for this operation.
func (op rootOpDarwinPPC) check(t *testing.T, err error) {
	t.Helper()
	if op.want == nil {
		if err != nil {
			t.Errorf("Root.%s = %v; want success (the anchored backend serves this operation on this kernel)", op.name, err)
		}
		return
	}
	// A refusal must not be reached sideways: an ENOTSUP produced before the
	// operation got to its own refusal (a missing parent, say) would look the
	// same through errors.Is. The operations below all name existing entries in
	// a fresh tree, so a "not exist" error is reported as such.
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Root.%s = %v; want %v, and the entry it names exists in the fresh tree", op.name, err, op.want)
		return
	}
	if !errors.Is(err, op.want) {
		t.Errorf("Root.%s = %v; want an error wrapping %v", op.name, err, op.want)
	}
}

// freshConfinedRoot creates a fresh confined tree and opens a Root on it. Each
// operation gets its own tree: the mutating ones must not consume another
// operation's fixtures.
func freshConfinedRoot(t *testing.T) (*os.Root, string) {
	t.Helper()
	rootDir := t.TempDir()
	confinedTree(t, rootDir, "root")
	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", rootDir, err)
	}
	t.Cleanup(func() { r.Close() })
	return r, rootDir
}

// TestRootConfinedOpsNoChdirDarwinPPC requires the confined operations to be
// refused on their own terms. A refusal that is reached by changing the process
// working directory and changing it back satisfies the endpoints the other test
// compares, and would operate on the wrong directory for the length of the
// operation, so it is a defect the endpoint comparison cannot express.
func TestRootConfinedOpsNoChdirDarwinPPC(t *testing.T) {
	installChdirRecorder()

	cwd := t.TempDir()
	confinedTree(t, cwd, "cwd")
	t.Chdir(cwd)

	for _, op := range rootOpsDarwinPPC() {
		r, _ := freshConfinedRoot(t)
		chdirLog.reset()
		err := op.run(r)
		moved := chdirLog.since()

		op.check(t, err)
		if len(moved) != 0 {
			t.Errorf("Root.%s moved the process working directory to %q; a confined operation must not need to, whether or not it moves back", op.name, moved)
		}
	}

	// A recorder that cannot see a chdir would pass the loop above no matter
	// what the operations did, so make one through the same path a cwd-based
	// fallback would use and require it to be recorded. The recorded path is
	// the one Getwd reported, so on this system /tmp arrives as /private/tmp;
	// compare the directories, not the spellings.
	probe := t.TempDir()
	chdirLog.reset()
	if err := os.Chdir(probe); err != nil {
		t.Fatalf("os.Chdir(%q): %v", probe, err)
	}
	if moved := chdirLog.since(); len(moved) != 1 || !sameDir(t, moved[0], probe) {
		t.Fatalf("the chdir recorder is not live: os.Chdir(%q) recorded %q; a confined operation could change the working directory unobserved", probe, moved)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("os.Chdir(%q): %v", cwd, err)
	}
}

// sameDir reports whether two paths name the same directory.
func sameDir(t *testing.T, a, b string) bool {
	t.Helper()
	ai, err := os.Stat(a)
	if err != nil {
		t.Errorf("stat %q: %v", a, err)
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		t.Errorf("stat %q: %v", b, err)
		return false
	}
	return os.SameFile(ai, bi)
}

// chdirInterposerLogEnv names the file the libc/syscall-boundary interposer
// writes to (probes/ppc-mac/chdir-interpose.c). TestRootConfinedOpsNoChdir-
// SyscallDarwinPPC skips without it, like the close-on-exec child helper in
// sys_darwin_ppc_test.go skips without its own variable.
const chdirInterposerLogEnv = "PPC32_CHDIR_LOG"

// chdirInterposerEnv returns the interposer's log path, or "" when it is not
// loaded.
func chdirInterposerLog() string {
	return os.Getenv(chdirInterposerLogEnv)
}

// TestRootConfinedOpsNoChdirSyscallDarwinPPC is the requirement of
// TestRootConfinedOpsNoChdirDarwinPPC seen from below package os. The recorder
// there is internal/testlog, which sees the calls package os makes; an operation
// that changed directory with syscall.Chdir or syscall.Fchdir directly never
// reaches it. This one interposes the libc functions those calls are imported
// from, so every caller in the process is covered, and it asserts the attempt's
// own path rather than a cwd read before and after.
//
// Probe the instrument before believing it: the liveness control at the end uses
// a raw syscall.Chdir/Fchdir pair precisely because the other instrument cannot
// see it. If those do not appear here, the interposer is not loaded and this test
// would pass for the wrong reason.
func TestRootConfinedOpsNoChdirSyscallDarwinPPC(t *testing.T) {
	logPath := chdirInterposerLog()
	if logPath == "" {
		t.Skipf("needs the chdir interposer: build probes/ppc-mac/chdir-interpose.c on the target and run with DYLD_INSERT_LIBRARIES=<dylib> %s=<file>", chdirInterposerLogEnv)
	}
	clearLog := func() {
		t.Helper()
		if err := os.WriteFile(logPath, nil, 0o644); err != nil {
			t.Fatalf("clearing %s: %v", logPath, err)
		}
	}
	showsDir := func(want string) bool {
		t.Helper()
		b, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("reading %s: %v", logPath, err)
		}
		return bytes.Contains(b, []byte("chdir "+want+"\n"))
	}

	cwd := t.TempDir()
	confinedTree(t, cwd, "cwd")
	t.Chdir(cwd)

	for _, op := range rootOpsDarwinPPC() {
		r, _ := freshConfinedRoot(t)
		clearLog()
		err := op.run(r)
		b, rerr := os.ReadFile(logPath)
		if rerr != nil {
			t.Fatalf("reading %s: %v", logPath, rerr)
		}
		op.check(t, err)
		if len(bytes.TrimSpace(b)) != 0 {
			t.Errorf("Root.%s reached the working directory through the syscall boundary: %q; a confined operation must not need to, whether or not it changes back", op.name, bytes.TrimSpace(b))
		}
	}

	// Liveness: the two calls the os-level recorder cannot see. Both must be
	// recorded, and the chdir must be recorded as the path asked for.
	probe := t.TempDir()
	clearLog()
	if err := syscall.Chdir(probe); err != nil {
		t.Fatalf("syscall.Chdir(%q): %v", probe, err)
	}
	f, err := os.Open(probe)
	if err != nil {
		t.Fatal(err)
	}
	err = syscall.Fchdir(int(f.Fd()))
	f.Close()
	if err != nil {
		t.Fatalf("syscall.Fchdir: %v", err)
	}
	if !showsDir(probe) {
		t.Fatalf("the chdir interposer is not live: syscall.Chdir(%q) was not recorded in %s", probe, logPath)
	}
	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("fchdir ")) {
		t.Fatalf("the chdir interposer did not record syscall.Fchdir: %q", b)
	}
	if err := syscall.Chdir(cwd); err != nil {
		t.Fatalf("syscall.Chdir(%q): %v", cwd, err)
	}
}

// TestInRootEntryStatDarwinPPC exercises the in-root branch of
// statat_darwin_ppc.go: a File whose entries must be described through the
// descriptor it was opened from, not through a pathname that could name a
// different directory by the time it is read.
//
// Root's own operations now succeed on this kernel, so the branch is reachable
// through os.OpenRoot directly; MarkInRoot is kept for the case the Root test
// cannot express, a File that is in a root but was opened elsewhere, and it is
// still an instrument, not a second way to open a File in a Root.
//
// The un-marked File is the control: it takes the other branch
// (unix.LstatatByPath) and must describe the same entries. Without it, a
// failure of the marked File would be indistinguishable from a broken
// descriptor, an empty directory or a listing that was never reached.
func TestInRootEntryStatDarwinPPC(t *testing.T) {
	dir := t.TempDir()
	confinedTree(t, dir, "root")

	openDir := func() *os.File {
		t.Helper()
		f, err := os.Open(dir)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	confinedDir := func() *os.File {
		t.Helper()
		f := openDir()
		if os.InRoot(f) {
			t.Fatal("an ordinary os.Open reported inRoot; the fabrication below would prove nothing")
		}
		os.MarkInRoot(f, true)
		return f
	}
	var open []*os.File
	t.Cleanup(func() {
		for _, f := range open {
			os.MarkInRoot(f, false)
			f.Close()
		}
	})

	control := openDir()
	open = append(open, control)
	infos, err := control.Readdir(-1)
	if err != nil || len(infos) != 3 {
		t.Fatalf("control: Readdir = %v, %v; want the 3 entries of %s (nothing below can be read as a refusal if the control does not list them)", infos, err, dir)
	}

	// Entry statting is the operation under test, so record that it happens:
	// a run that never reached lstatatNolog must not be able to pass.
	var statts int
	os.SetStatHook(t, func(f *os.File, name string) (os.FileInfo, error) {
		statts++
		return nil, nil // let the real statat path run
	})

	// Entry statting through the confined branch must describe the same entries
	// as the path branch does, including lstat semantics for the symlink.
	for _, tc := range []struct {
		name string
		call func(*os.File) ([]os.FileInfo, error)
	}{
		{"File.Readdir", func(f *os.File) ([]os.FileInfo, error) { return f.Readdir(-1) }},
		{"File.ReadDirInfo", func(f *os.File) ([]os.FileInfo, error) {
			entries, err := f.ReadDir(-1)
			if err != nil {
				return nil, err
			}
			infos := make([]os.FileInfo, 0, len(entries))
			for _, e := range entries {
				info, err := e.Info()
				if err != nil {
					return nil, err
				}
				infos = append(infos, info)
			}
			return infos, nil
		}},
	} {
		f := confinedDir()
		open = append(open, f)
		statts = 0
		infos, err := tc.call(f)
		if err != nil {
			t.Fatalf("%s on a File marked inRoot: %v; want the confined statat to describe its entries", tc.name, err)
		}
		if len(infos) != 3 {
			t.Fatalf("%s on a File marked inRoot returned %d entries; want 3", tc.name, len(infos))
		}
		// The hook fires once per entry described, so a three-entry listing
		// must stat three. The old refusal-era expectation of 1 came from a
		// refusal that aborted on the first entry; a run that stats nothing at
		// all did not reach lstatatNolog and must not pass either.
		if statts != len(infos) {
			t.Errorf("%s stat'ed %d entries; want %d, one per entry in the listing", tc.name, statts, len(infos))
		}
		names := make([]string, 0, len(infos))
		for _, info := range infos {
			want, err := os.Lstat(filepath.Join(dir, info.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode() != want.Mode() || info.Size() != want.Size() || !os.SameFile(info, want) {
				t.Errorf("%s entry %s = %v; want Lstat %v", tc.name, info.Name(), info, want)
			}
			names = append(names, info.Name())
		}
	}

	// The confined branch is scoped to entry statting, not to the directory:
	// listing the same File's names needs no statat and must keep working.
	listed := confinedDir()
	open = append(open, listed)
	names, err := listed.Readdirnames(-1)
	if err != nil || len(names) != 3 {
		t.Errorf("File.Readdirnames on a File marked inRoot = %q, %v; want the 3 names", names, err)
	}
}
