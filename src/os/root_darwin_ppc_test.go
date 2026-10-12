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
//  1. The refusal in statat_darwin_ppc.go for a File opened in a Root. Every
//     Root operation is ENOTSUP here, so an ordinary program cannot construct
//     such a File, and no test reached that guard: deleting it would have left
//     the suite green. MarkInRoot (export_darwin_ppc_test.go) fabricates the
//     state so the guard can be exercised.
//
//  2. That a confined operation does not change the process working directory
//     even transiently. Comparing Getwd before and after an operation is blind
//     to a swap that is restored before the operation returns -- which is the
//     shape a cwd-based fallback would have, and the opposite of what the
//     comment on TestRootNotSupportedDarwinPPC promises. TestRootConfinedOps-
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

// rootOpsDarwinPPC is the Root surface a confined operation is exposed through.
// It is the union of the sets pinned by root_darwin_ppc64_test.go and
// removeall_darwin_ppc_test.go, so a ppc32 divergence shows up as a difference
// rather than as silence. Every one of them must be refused: the *at calls they
// need take a real directory descriptor, which Leopard does not have
// (internal/syscall/unix/at_darwin_ppc.go), and refusing is the whole of this
// platform's Root support.
//
// The errno differs for a path with more than one component, and that is
// recorded rather than smoothed over: the intermediate openat is refused with
// ENOTSUP, and rootOpenDir (root_unix.go, upstream code) deliberately reports an
// intermediate component it cannot open as ENOTDIR. So "m/n" is refused with
// ENOTDIR. Both are refusals of the same call; nothing acts on the working
// directory either way, which is what this test is for.
func rootOpsDarwinPPC(r *os.Root) map[string]func() error {
	now := time.Now()
	return map[string]func() error{
		"Open": func() error {
			f, err := r.Open("f")
			if f != nil {
				f.Close()
			}
			return err
		},
		"OpenFile": func() error {
			f, err := r.OpenFile("f", os.O_RDONLY, 0)
			if f != nil {
				f.Close()
			}
			return err
		},
		"Create": func() error {
			f, err := r.Create("new")
			if f != nil {
				f.Close()
			}
			return err
		},
		"OpenRoot": func() error {
			sub, err := r.OpenRoot("sub")
			if sub != nil {
				sub.Close()
			}
			return err
		},
		"Mkdir":     func() error { return r.Mkdir("m", 0o755) },
		"MkdirAll":  func() error { return r.MkdirAll("m/n", 0o755) },
		"Remove":    func() error { return r.Remove("f") },
		"RemoveAll": func() error { return r.RemoveAll("sub") },
		"Stat":      func() error { _, err := r.Stat("f"); return err },
		"Lstat":     func() error { _, err := r.Lstat("f"); return err },
		"Readlink":  func() error { _, err := r.Readlink("link"); return err },
		"ReadFile":  func() error { _, err := r.ReadFile("f"); return err },
		"WriteFile": func() error { return r.WriteFile("w", []byte("x"), 0o644) },
		"Rename":    func() error { return r.Rename("f", "g") },
		"Link":      func() error { return r.Link("f", "g") },
		"Symlink":   func() error { return r.Symlink("f", "s") },
		"Chmod":     func() error { return r.Chmod("f", 0o600) },
		"Chown":     func() error { return r.Chown("f", os.Getuid(), os.Getgid()) },
		"Chtimes":   func() error { return r.Chtimes("f", now, now) },
	}
}

// rootRefusalDarwinPPC is the errno Root."%s" is refused with. Paths of one
// component stop at the *at substitution and get ENOTSUP; "m/n" reaches
// rootOpenDir for the intermediate "m", which reports what it cannot open as
// ENOTDIR.
func rootRefusalDarwinPPC(op string) error {
	if op == "MkdirAll" {
		return syscall.ENOTDIR
	}
	return syscall.ENOTSUP
}

// TestRootConfinedOpsNoChdirDarwinPPC requires the confined operations to be
// refused on their own terms. A refusal that is reached by changing the process
// working directory and changing it back satisfies the endpoints the other test
// compares, and would operate on the wrong directory for the length of the
// operation, so it is a defect the endpoint comparison cannot express.
func TestRootConfinedOpsNoChdirDarwinPPC(t *testing.T) {
	installChdirRecorder()

	rootDir, cwd := t.TempDir(), t.TempDir()
	confinedTree(t, rootDir, "root")
	confinedTree(t, cwd, "cwd")
	t.Chdir(cwd)

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	for name, op := range rootOpsDarwinPPC(r) {
		chdirLog.reset()
		err := op()
		moved := chdirLog.since()

		if !errors.Is(err, rootRefusalDarwinPPC(name)) {
			t.Errorf("Root.%s = %v; want an error wrapping %v", name, err, rootRefusalDarwinPPC(name))
		}
		if len(moved) != 0 {
			t.Errorf("Root.%s moved the process working directory to %q; a confined operation must be refused without acting on it, whether or not it moves back", name, moved)
		}
	}

	// A recorder that cannot see a chdir would pass the loop above no matter
	// what the operations did, so make one through the same path a cwd-based
	// fallback would use and require it to be recorded. The recorded path is
	// the one Getwd reported, so on this system /tmp arrives as /private/tmp;
	// compare the directories, not the spellings.
	chdirLog.reset()
	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("os.Chdir(%q): %v", rootDir, err)
	}
	if moved := chdirLog.since(); len(moved) != 1 || !sameDir(t, moved[0], rootDir) {
		t.Fatalf("the chdir recorder is not live: os.Chdir(%q) recorded %q; a confined operation could change the working directory unobserved", rootDir, moved)
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

	rootDir, cwd := t.TempDir(), t.TempDir()
	confinedTree(t, rootDir, "root")
	confinedTree(t, cwd, "cwd")
	t.Chdir(cwd)

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	defer r.Close()

	for name, op := range rootOpsDarwinPPC(r) {
		clearLog()
		err := op()
		b, rerr := os.ReadFile(logPath)
		if rerr != nil {
			t.Fatalf("reading %s: %v", logPath, rerr)
		}
		if !errors.Is(err, rootRefusalDarwinPPC(name)) {
			t.Errorf("Root.%s = %v; want an error wrapping %v", name, err, rootRefusalDarwinPPC(name))
		}
		if len(bytes.TrimSpace(b)) != 0 {
			t.Errorf("Root.%s reached the working directory through the syscall boundary: %q; a confined operation must be refused without changing directory, whether or not it changes back", name, bytes.TrimSpace(b))
		}
	}

	// Liveness: the two calls the os-level recorder cannot see. Both must be
	// recorded, and the chdir must be recorded as the path asked for.
	clearLog()
	if err := syscall.Chdir(rootDir); err != nil {
		t.Fatalf("syscall.Chdir(%q): %v", rootDir, err)
	}
	f, err := os.Open(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	err = syscall.Fchdir(int(f.Fd()))
	f.Close()
	if err != nil {
		t.Fatalf("syscall.Fchdir: %v", err)
	}
	if !showsDir(rootDir) {
		t.Fatalf("the chdir interposer is not live: syscall.Chdir(%q) was not recorded in %s", rootDir, logPath)
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

// checkInRootRefusal requires exactly the refusal statat_darwin_ppc.go makes for
// a File opened in a Root. The operation is asserted, not only the errno: an
// ENOTSUP from the readdir syscall itself, or from the *at substitution on the
// way somewhere else, would be a different fault and must not be read as this
// refusal.
func checkInRootRefusal(t *testing.T, name string, err error) {
	t.Helper()
	if !errors.Is(err, syscall.ENOTSUP) {
		t.Errorf("%s on a File marked inRoot = %v; want an error wrapping ENOTSUP", name, err)
		return
	}
	var pe *fs.PathError
	if !errors.As(err, &pe) || pe.Op != "lstat" {
		t.Errorf("%s on a File marked inRoot = %v; want the refusal from lstatat, a *fs.PathError with Op \"lstat\"", name, err)
	}
}

// TestInRootEntryStatRefusedDarwinPPC exercises the dormant guard at
// statat_darwin_ppc.go:39. Root's operations are all ENOTSUP here, so the only
// way to reach a File whose entries would be stat'ed through the substituted
// path lookup is to mark an ordinary directory File. That is what MarkInRoot
// does: an instrument for the guard, not evidence that Root works, and not a
// second way to open a File in a Root.
//
// The un-marked File is the control. Without it, a failure of the marked File
// would be indistinguishable from a broken descriptor, an empty directory or a
// directory listing that was never reached.
func TestInRootEntryStatRefusedDarwinPPC(t *testing.T) {
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

	for _, tc := range []struct {
		name string
		call func(*os.File) error
	}{
		{"File.Readdir", func(f *os.File) error { _, err := f.Readdir(-1); return err }},
		{"File.ReadDir", func(f *os.File) error { _, err := f.ReadDir(-1); return err }},
	} {
		f := confinedDir()
		open = append(open, f)
		statts = 0
		err := tc.call(f)
		checkInRootRefusal(t, tc.name, err)
		if statts != 1 {
			t.Errorf("%s stat'ed %d entries before returning; want 1 -- the refusal must come from the first entry, or the assertion above did not test it", tc.name, statts)
		}
	}

	// The refusal is scoped to entry statting, not to the directory: listing
	// the same File's names needs no statat and must keep working, as it does
	// for a Root (removeall_darwin_ppc_test.go relies on exactly that).
	listed := confinedDir()
	open = append(open, listed)
	names, err := listed.Readdirnames(-1)
	if err != nil || len(names) != 3 {
		t.Errorf("File.Readdirnames on a File marked inRoot = %q, %v; want the 3 names (the refusal must not spread to operations that need no statat)", names, err)
	}
}
