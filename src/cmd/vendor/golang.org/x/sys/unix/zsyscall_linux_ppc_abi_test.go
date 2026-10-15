//go:build linux && ppc

package unix

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func abiFixture(t *testing.T, name string) string {
	t.Helper()
	root := os.Getenv("PPC32_ABI_FIXTURE")
	if root == "" {
		t.Fatal("PPC32_ABI_FIXTURE is required; fixtures are prepared and checked by the host")
	}
	return filepath.Join(root, name)
}

func TestLinuxPPC32VendoredPositionedIOAndTruncate(t *testing.T) {
	fd, err := Open(abiFixture(t, "vendor-positioned"), O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(fd)
	for _, tc := range []struct {
		off  int64
		want string
	}{
		{32, "low-read"}, {1<<33 + 16, "high-read"},
	} {
		buf := make([]byte, len(tc.want))
		n, err := Pread(fd, buf, tc.off)
		if err != nil || n != len(buf) || string(buf) != tc.want {
			t.Fatalf("Pread(%d)=(%d,%q,%v)", tc.off, n, buf, err)
		}
		w := "low-write"
		if tc.off > 32 {
			w = "high-write"
		}
		n, err = Pwrite(fd, []byte(w), tc.off+8)
		if err != nil || n != len(w) {
			t.Fatalf("Pwrite(%d)=(%d,%v)", tc.off+8, n, err)
		}
	}
	if err := Truncate(abiFixture(t, "vendor-truncate"), 5); err != nil {
		t.Fatal(err)
	}
	if err := Truncate(abiFixture(t, "vendor-truncate-high"), 1<<33+7); err != nil {
		t.Fatal(err)
	}
	fd, err = Open(abiFixture(t, "vendor-ftruncate"), O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(fd)
	if err := Ftruncate(fd, 5); err != nil {
		t.Fatal(err)
	}
	fdHigh, err := Open(abiFixture(t, "vendor-ftruncate-high"), O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer Close(fdHigh)
	if err := Ftruncate(fdHigh, 1<<33+7); err != nil {
		t.Fatal(err)
	}
}

func TestLinuxPPC32TeeSpliceOneWordCounts(t *testing.T) {
	var in, teeOut, spliceOut [2]int
	if err := Pipe(in[:]); err != nil {
		t.Fatal(err)
	}
	defer Close(in[0])
	defer Close(in[1])
	if err := Pipe(teeOut[:]); err != nil {
		t.Fatal(err)
	}
	defer Close(teeOut[0])
	defer Close(teeOut[1])
	if err := Pipe(spliceOut[:]); err != nil {
		t.Fatal(err)
	}
	defer Close(spliceOut[0])
	defer Close(spliceOut[1])
	if _, err := Write(in[1], []byte("X")); err != nil {
		t.Fatal(err)
	}
	n, err := Tee(in[0], teeOut[1], 1, 0)
	if err != nil || n != 1 {
		t.Fatalf("Tee count=%d err=%v, want 1,nil", n, err)
	}
	buf := []byte{0}
	if nr, err := Read(teeOut[0], buf); err != nil || nr != 1 || buf[0] != 'X' {
		t.Fatalf("Tee payload=(%d,%q,%v)", nr, buf, err)
	}
	nSplice, err := Splice(in[0], nil, spliceOut[1], nil, 1, 0)
	if err != nil || nSplice != 1 {
		t.Fatalf("Splice count=%d err=%v, want 1,nil", nSplice, err)
	}
	buf[0] = 0
	if nr, err := Read(spliceOut[0], buf); err != nil || nr != 1 || buf[0] != 'X' {
		t.Fatalf("Splice payload=(%d,%q,%v)", nr, buf, err)
	}
	Close(in[1])
	nSplice, err = Splice(in[0], nil, spliceOut[1], nil, 1, 0)
	if err != nil || nSplice != 0 {
		t.Fatalf("Splice EOF=(%d,%v), want 0,nil", nSplice, err)
	}
	if _, err := Tee(-1, teeOut[1], 1, 0); !errors.Is(err, EBADF) {
		t.Fatalf("Tee errno=%v, want EBADF", err)
	}
}
