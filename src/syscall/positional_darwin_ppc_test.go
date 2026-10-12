// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package syscall_test

import (
	"bytes"
	"internal/poll"
	"io"
	"os"
	"syscall"
	"testing"
)

func positionalFile(t *testing.T, data []byte) *os.File {
	t.Helper()
	f, err := os.CreateTemp("", "positional-ppc-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f.Close()
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	})
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	return f
}

// Offsets must be two consecutive big-endian words, not a truncated uintptr.
// Both readers must leave the descriptor position unchanged, including at EOF.
func TestPositionalIODarwinPPC(t *testing.T) {
	data := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKL")
	f := positionalFile(t, data)
	fd, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		t.Fatal(err)
	}
	pfd := &poll.FD{Sysfd: fd, IsStream: true, ZeroReadIsEOF: true}
	if err := pfd.Init("file", false); err != nil {
		t.Fatal(err)
	}
	defer pfd.Close()
	for _, reader := range []struct {
		name string
		read func([]byte, int64) (int, error)
	}{{"os.File.ReadAt", f.ReadAt}, {"poll.FD.Pread", pfd.Pread}} {
		t.Run(reader.name, func(t *testing.T) {
			for _, off := range []int64{0, 16, 42, 16, 64} {
				buf := make([]byte, 8)
				n, err := reader.read(buf, off)
				if off == 64 {
					if n != 0 || err != io.EOF {
						t.Fatalf("offset %d: n=%d err=%v, want EOF", off, n, err)
					}
				} else if n != 8 || err != nil || !bytes.Equal(buf, data[off:off+8]) {
					t.Fatalf("offset %d: n=%d err=%v got=%q want=%q", off, n, err, buf, data[off:off+8])
				}
			}
		})
	}
	if off, err := f.Seek(0, io.SeekCurrent); err != nil || off != 64 {
		t.Fatalf("position after ReadAt: %d, %v, want 64", off, err)
	}
	for _, off := range []int64{16, 1<<32 + 16} {
		if got, err := f.Seek(off, io.SeekStart); err != nil || got != off {
			t.Fatalf("Seek(%d): %d, %v", off, got, err)
		}
	}
	if _, err := f.Seek(16, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 8)
	if n, err := f.Read(buf); n != 8 || err != nil || !bytes.Equal(buf, data[16:24]) {
		t.Fatalf("sequential read: %d %v %q", n, err, buf)
	}
	for _, writer := range []struct {
		name  string
		write func([]byte, int64) (int, error)
		data  string
	}{{"os.File.WriteAt", f.WriteAt, "12345678"}, {"poll.FD.Pwrite", pfd.Pwrite, "87654321"}} {
		t.Run(writer.name, func(t *testing.T) {
			if n, err := writer.write([]byte(writer.data), 16); n != 8 || err != nil {
				t.Fatalf("WriteAt: %d, %v", n, err)
			}
			if n, err := f.ReadAt(buf, 16); n != 8 || err != nil || string(buf) != writer.data {
				t.Fatalf("readback: %d %v %q", n, err, buf)
			}
		})
	}
}

func TestGeneratedWordArgsDarwinPPC(t *testing.T) {
	page := syscall.Getpagesize()
	data := make([]byte, 2*page)
	copy(data[page:], "mapped at nonzero offset")
	f := positionalFile(t, data)
	fd := int(f.Fd())
	b, err := syscall.Mmap(fd, int64(page), page, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b[:23], data[page:page+23]) {
		t.Errorf("mmap offset: got %q want %q", b[:23], data[page:page+23])
	}
	if err := syscall.Munmap(b); err != nil {
		t.Fatal(err)
	}
	if _, err := syscall.Mmap(-1, 0, page, syscall.PROT_READ, syscall.MAP_PRIVATE); err == nil {
		t.Fatal("mmap invalid fd succeeded")
	}
	if err := syscall.Ftruncate(fd, 128); err != nil {
		t.Fatal(err)
	}
	if st, err := f.Stat(); err != nil || st.Size() != 128 {
		t.Fatalf("Ftruncate size: %v, %v", st, err)
	}
	if err := syscall.Truncate(f.Name(), 64); err != nil {
		t.Fatal(err)
	}
	if st, err := f.Stat(); err != nil || st.Size() != 64 {
		t.Fatalf("Truncate size: %v, %v", st, err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	dup, err := syscall.Dup(fd)
	if err != nil {
		t.Fatal(err)
	}
	pfd := &poll.FD{Sysfd: dup, IsStream: true}
	if err := pfd.Init("file", false); err != nil {
		t.Fatal(err)
	}
	defer pfd.Close()
	v := [][]byte{[]byte("word"), []byte("args")}
	if n, err := pfd.Writev(&v); n != 8 || err != nil {
		t.Fatalf("writev: %d, %v", n, err)
	}
	buf := make([]byte, 8)
	if n, err := f.ReadAt(buf, 0); n != 8 || err != nil || string(buf) != "wordargs" {
		t.Fatalf("writev readback: %d, %v, %q", n, err, buf)
	}
}
