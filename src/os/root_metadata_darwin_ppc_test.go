// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package os_test

// Metadata and tree-removal controls for os.Root on Leopard ppc32.
//
// Root.Chmod and Root.Chtimes have no *at call on this kernel. They are served by
// setattrlist(FSOPT_NOFOLLOW) selecting ATTR_CMN_ACCESSMASK or the
// MODTIME/ACCTIME timespecs, which is the same request libSystem's lchmod and
// lutimes are built on (measured and disassembled on the target,
// runs/root-at-metadata-20261023). Root.RemoveAll is served by the
// descriptor-based walk now that the anchored *at operations exist.
//
// What these tests can see that the Unix-layer tests cannot:
//
//   - that the call reached the root's entry and not the process working
//     directory, using a decoy of the same name in the cwd with a different mode
//     and different times;
//   - that a symlink's target is what Root.Chmod changes (Root.Chmod follows a
//     final symlink by contract, unlike lchmod), because the traversal resolves
//     it before the no-follow metadata call;
//   - that the metadata result matches os.Chmod/os.Chtimes applied to the same
//     tree — the Darwin semantic baseline on this machine — for mode bits, both
//     times, and one-omitted time;
//   - that an omitted time is omitted, not read back and rewritten: the other
//     time keeps a stamp set long before the call;
//   - that the confined walk of Root.RemoveAll keeps working when the path that
//     named the root no longer exists, which is exactly what a path-based
//     substitute cannot do.

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// metadataRoot builds a root tree holding a file, a directory, a symlink to the
// file, and a decoy of the same names in the process working directory with a
// different mode. It returns the Root, the root directory and the decoy.
func metadataRoot(t *testing.T) (root *os.Root, dir, cwd string) {
	t.Helper()
	dir, cwd = t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("f", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cwd, "f"), []byte("cwd"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("f", filepath.Join(cwd, "link")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root, dir, cwd
}

// modeOf, timesOf and ctimeOf read the entry without following a final symlink.
func modeOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat %s: %v", path, err)
	}
	return fi.Mode()
}

func timesOf(t *testing.T, path string) (atime, mtime time.Time) {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat %s: %v", path, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("no syscall.Stat_t for %s", path)
	}
	return time.Unix(int64(st.Atimespec.Sec), int64(st.Atimespec.Nsec)),
		time.Unix(int64(st.Mtimespec.Sec), int64(st.Mtimespec.Nsec))
}

func ctimeOf(t *testing.T, path string) time.Time {
	t.Helper()
	fi, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat %s: %v", path, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatalf("no syscall.Stat_t for %s", path)
	}
	return time.Unix(int64(st.Ctimespec.Sec), int64(st.Ctimespec.Nsec))
}

// TestRootChmodDarwinPPC requires Root.Chmod to change the root's entry (or the
// target of the root's symlink, as the call documents) and nothing else.
func TestRootChmodDarwinPPC(t *testing.T) {
	root, dir, cwd := metadataRoot(t)

	if err := root.Chmod("f", 0o640); err != nil {
		t.Fatalf("Root.Chmod: %v", err)
	}
	if got := modeOf(t, filepath.Join(dir, "f")).Perm(); got != 0o640 {
		t.Errorf("root file mode = %o; want 640", got)
	}
	if got := modeOf(t, filepath.Join(cwd, "f")).Perm(); got != 0o600 {
		t.Errorf("decoy file mode = %o; want the untouched 600: the mode change must not reach the working directory", got)
	}

	// Root.Chmod on a symlink changes the link's target: the traversal
	// (doInRoot) resolves the final component before the no-follow call, which is
	// what makes Root.Chmod differ from os.Root.Lchown-style no-follow calls.
	if err := root.Chmod("link", 0o604); err != nil {
		t.Fatalf("Root.Chmod(link): %v", err)
	}
	if got := modeOf(t, filepath.Join(dir, "f")).Perm(); got != 0o604 {
		t.Errorf("after Root.Chmod(link) the target's mode = %o; want 604: the target is what a follow-form chmod changes", got)
	}
	if got := modeOf(t, filepath.Join(dir, "link")).Perm(); got == 0o604 {
		t.Errorf("Root.Chmod(link) changed the link itself; want the target")
	}

	if err := root.Chmod("no-such-entry", 0o600); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Root.Chmod(missing) = %v; want a not-exist error", err)
	}

	// Parity with os.Chmod on the same object: the same bits, the same result.
	alt := filepath.Join(dir, "alt")
	if err := os.WriteFile(alt, []byte("alt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := root.Chmod("alt", 0o754); err != nil {
		t.Fatalf("Root.Chmod(alt): %v", err)
	}
	if err := os.Chmod(alt, 0o754); err != nil {
		t.Fatalf("os.Chmod(alt): %v", err)
	}
	if got, want := modeOf(t, alt).Perm(), os.FileMode(0o754); got != want {
		t.Errorf("Root.Chmod(alt) = %o; os.Chmod(alt) = %o; want the same", got, want)
	}
}

// TestRootChtimesDarwinPPC requires Root.Chtimes to set the times it was given on
// the root's entry, to leave the other tree alone, and to omit — not rewrite — a
// time it was not given.
func TestRootChtimesDarwinPPC(t *testing.T) {
	root, dir, cwd := metadataRoot(t)
	stamp := time.Unix(1200000000, 0)
	rootFile, cwdFile := filepath.Join(dir, "f"), filepath.Join(cwd, "f")
	for _, p := range []string{rootFile, cwdFile} {
		if err := os.Chtimes(p, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}

	newAcc, newMod := time.Unix(1300000000, 0), time.Unix(1300000001, 0)
	if err := root.Chtimes("f", newAcc, newMod); err != nil {
		t.Fatalf("Root.Chtimes: %v", err)
	}
	acc, mod := timesOf(t, rootFile)
	if acc.Unix() != newAcc.Unix() || mod.Unix() != newMod.Unix() {
		t.Errorf("root file times = %d/%d; want %d/%d", acc.Unix(), mod.Unix(), newAcc.Unix(), newMod.Unix())
	}
	acc, mod = timesOf(t, cwdFile)
	if acc.Unix() != stamp.Unix() || mod.Unix() != stamp.Unix() {
		t.Errorf("decoy file times = %d/%d; want the untouched %d: the time change must not reach the working directory", acc.Unix(), mod.Unix(), stamp.Unix())
	}

	// A zero time.Time is UTIME_OMIT: the modification time is set, the access
	// time keeps its old stamp rather than being read and written back.
	if err := os.Chtimes(rootFile, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	onlyMod := time.Unix(1400000000, 0)
	if err := root.Chtimes("f", time.Time{}, onlyMod); err != nil {
		t.Fatalf("Root.Chtimes(zero atime): %v", err)
	}
	acc, mod = timesOf(t, rootFile)
	if acc.Unix() != stamp.Unix() {
		t.Errorf("atime = %d; want the untouched %d: an omitted time must not be rewritten", acc.Unix(), stamp.Unix())
	}
	if mod.Unix() != onlyMod.Unix() {
		t.Errorf("mtime = %d; want %d", mod.Unix(), onlyMod.Unix())
	}

	// An access-time-ONLY request is the form HFS would turn into a lazy
	// access-time touch if the request selected nothing but the access time
	// (bsd/hfs/hfs_vnops.c:776-780): the supplied value would be discarded and the
	// stored access time would become "now". The confined call therefore also
	// selects the change time on HFS, with a value the filesystem ignores, so that
	// the ordinary timestamp block runs. All three parts of that claim are
	// asserted here: the requested access time is stored, the omitted modification
	// time is not touched, and the change time does not move, which is what shows
	// the extra attribute was ignored rather than applied. The decoy in the
	// working directory is checked again.
	if err := os.Chtimes(rootFile, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	onlyAcc := time.Unix(1500000000, 0)
	beforeAcc, beforeMod := timesOf(t, rootFile)
	beforeCtime := ctimeOf(t, rootFile)
	if err := root.Chtimes("f", onlyAcc, time.Time{}); err != nil {
		t.Fatalf("Root.Chtimes(zero mtime): %v", err)
	}
	acc, mod = timesOf(t, rootFile)
	if acc.Unix() != onlyAcc.Unix() {
		t.Errorf("atime = %d; want the requested %d: an access-time-only request must store the value it was given", acc.Unix(), onlyAcc.Unix())
	}
	if mod.Unix() != beforeMod.Unix() {
		t.Errorf("mtime = %d; want the untouched %d: an omitted time must not be sent", mod.Unix(), beforeMod.Unix())
	}
	if got := ctimeOf(t, rootFile); !got.Equal(beforeCtime) {
		t.Errorf("ctime = %v; want the unchanged %v: the change time sent as part of the HFS workaround must be ignored by HFS", got, beforeCtime)
	}
	if onlyAcc.Unix() != beforeAcc.Unix() && acc.Unix() == beforeAcc.Unix() {
		t.Errorf("atime is still %d; the request did nothing", acc.Unix())
	}
	if gotAcc, gotMod := timesOf(t, cwdFile); gotAcc.Unix() != stamp.Unix() || gotMod.Unix() != stamp.Unix() {
		t.Errorf("decoy file times = %d/%d; want the untouched %d: the time change must not reach the working directory", gotAcc.Unix(), gotMod.Unix(), stamp.Unix())
	}

	// Parity with os.Chtimes on the same object: the both-times form is the one
	// the path-based call can express exactly, so it is the one compared.
	alt := filepath.Join(dir, "alt")
	if err := os.WriteFile(alt, []byte("alt"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := root.Chtimes("alt", newAcc, newMod); err != nil {
		t.Fatal(err)
	}
	rootAcc, rootMod := timesOf(t, alt)
	if err := os.Chtimes(alt, newAcc, newMod); err != nil {
		t.Fatal(err)
	}
	altAcc, altMod := timesOf(t, alt)
	if rootAcc.Unix() != altAcc.Unix() || rootMod.Unix() != altMod.Unix() {
		t.Errorf("Root.Chtimes gave %d/%d, os.Chtimes gave %d/%d; want the same result on HFS+",
			rootAcc.Unix(), rootMod.Unix(), altAcc.Unix(), altMod.Unix())
	}

	if err := root.Chtimes("no-such-entry", newAcc, newMod); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Root.Chtimes(missing) = %v; want a not-exist error", err)
	}
	// Both times omitted still resolves the name, so a missing one is still
	// reported as missing rather than silently succeeding.
	if err := root.Chtimes("no-such-entry", time.Time{}, time.Time{}); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Root.Chtimes(missing, both omitted) = %v; want a not-exist error", err)
	}
}

// TestRootMetadataPermissionParityDarwinPPC compares the errno of the confined
// metadata calls with the path-based ones on an object the caller does not own.
// They must agree: the metadata call is served by a different primitive
// (setattrlist, which reports EACCES) than os.Chmod/os.Chtimes (which report
// EPERM), and the port maps the authorization failure so that a program cannot
// tell from the errno which form it reached.
//
// Skipped when the caller is root, because then there is no non-owner case to
// measure, and when /usr/bin/true is not owned by someone else.
func TestRootMetadataPermissionParityDarwinPPC(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: no non-owner metadata case exists")
	}
	const victim = "/usr/bin/true"
	fi, err := os.Lstat(victim)
	if err != nil {
		t.Skipf("%s: %v", victim, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skipf("no syscall.Stat_t for %s", victim)
	}
	if int(st.Uid) == os.Geteuid() {
		t.Skipf("%s is owned by the caller, so a permission failure cannot be produced", victim)
	}
	mode := fi.Mode()
	root, err := os.OpenRoot("/usr/bin")
	if err != nil {
		t.Skipf("OpenRoot(/usr/bin): %v", err)
	}
	defer root.Close()

	errRoot := root.Chmod("true", 0o600)
	errPath := os.Chmod(victim, 0o600)
	if errRoot == nil || errPath == nil {
		t.Fatalf("a mode change on someone else's file succeeded: Root.Chmod = %v, os.Chmod = %v", errRoot, errPath)
	}
	if !errors.Is(errRoot, os.ErrPermission) || !errors.Is(errPath, os.ErrPermission) {
		t.Errorf("permission errors disagree in kind: Root.Chmod = %v, os.Chmod = %v; want both to be permission errors", errRoot, errPath)
	}
	if errors.Is(errRoot, syscall.EPERM) != errors.Is(errPath, syscall.EPERM) {
		t.Errorf("errno parity broken: Root.Chmod = %v (EPERM: %v), os.Chmod = %v (EPERM: %v); the libc-equivalent calls report the same errno on this kernel",
			errRoot, errors.Is(errRoot, syscall.EPERM), errPath, errors.Is(errPath, syscall.EPERM))
	}

	past := time.Unix(1300000000, 0)
	errRoot = root.Chtimes("true", past, past)
	errPath = os.Chtimes(victim, past, past)
	if errRoot == nil || errPath == nil {
		t.Fatalf("an explicit time change on someone else's file succeeded: Root.Chtimes = %v, os.Chtimes = %v", errRoot, errPath)
	}
	if errors.Is(errRoot, syscall.EPERM) != errors.Is(errPath, syscall.EPERM) {
		t.Errorf("errno parity broken: Root.Chtimes = %v, os.Chtimes = %v; want the same errno class", errRoot, errPath)
	}

	// None of the refusals changed the file.
	if after := modeOf(t, victim); after != mode {
		if err := os.Chmod(victim, mode); err != nil && after.Perm() != mode.Perm() {
			t.Errorf("%s mode changed to %v; want %v (and the restore failed: %v)", victim, after, mode, err)
		} else {
			t.Errorf("%s mode changed to %v; want %v unchanged", victim, after, mode)
		}
	}
}

// TestRootRemoveAllConfinedDarwinPPC requires Root.RemoveAll to remove the root's
// subtree without following a symlink out of it, and to keep working after the
// path that named the root is gone — which only the descriptor-based walk can do
// and a path-based substitute cannot.
func TestRootRemoveAllConfinedDarwinPPC(t *testing.T) {
	base, cwd := t.TempDir(), t.TempDir()
	rootDir := filepath.Join(base, "root")
	if err := os.Mkdir(rootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tree := filepath.Join(rootDir, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "a", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"x", "a/y", "a/b/z"} {
		if err := os.WriteFile(filepath.Join(tree, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	outside := filepath.Join(base, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(outside, "keep")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(tree, "link-out")); err != nil {
		t.Fatal(err)
	}
	// A decoy of the same name in the working directory, so a path-based
	// substitute that resolved against the cwd would remove the wrong tree.
	if err := os.MkdirAll(filepath.Join(cwd, "tree", "decoy"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)

	root, err := os.OpenRoot(rootDir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if err := root.RemoveAll("tree"); err != nil {
		t.Fatalf("Root.RemoveAll: %v", err)
	}
	if _, err := os.Lstat(tree); !os.IsNotExist(err) {
		t.Errorf("%s: %v; want the subtree removed inside the root", tree, err)
	}
	if b, err := os.ReadFile(keep); err != nil || string(b) != "keep" {
		t.Errorf("%s = %q, %v; want the symlink's target untouched", keep, b, err)
	}
	if fi, err := os.Lstat(filepath.Join(cwd, "tree", "decoy")); err != nil || !fi.IsDir() {
		t.Errorf("the working directory's decoy tree was touched: %v, %v", fi, err)
	}

	// The descriptor-relative control: after the root's own path is renamed, a
	// walk that reconstructed paths from it would fail. The confined walk does
	// not use the path at all.
	if err := os.MkdirAll(filepath.Join(rootDir, "second", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	mapSource := filepath.Join(rootDir, "second", "map")
	if err := os.WriteFile(mapSource, []byte("m"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved := rootDir + "-moved"
	if err := os.Rename(rootDir, moved); err != nil {
		t.Fatal(err)
	}
	if err := root.RemoveAll("second"); err != nil {
		t.Fatalf("Root.RemoveAll after the root's path was renamed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(moved, "second")); !os.IsNotExist(err) {
		t.Errorf("%s/second: %v; want the subtree removed through the held descriptor", filepath.Join(moved), err)
	}
}
