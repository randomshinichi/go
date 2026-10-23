// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc

package unix_test

import (
	"testing"
	"unsafe"

	"internal/syscall/unix"
)

const (
	ppcPasswdSize   = unsafe.Sizeof(unix.Passwd{})
	ppcPasswdUid    = unsafe.Offsetof(unix.Passwd{}.Uid)
	ppcPasswdGid    = unsafe.Offsetof(unix.Passwd{}.Gid)
	ppcPasswdChange = unsafe.Offsetof(unix.Passwd{}.Change)
	ppcPasswdClass  = unsafe.Offsetof(unix.Passwd{}.Class)
	ppcPasswdGecos  = unsafe.Offsetof(unix.Passwd{}.Gecos)
	ppcPasswdDir    = unsafe.Offsetof(unix.Passwd{}.Dir)
	ppcPasswdShell  = unsafe.Offsetof(unix.Passwd{}.Shell)
	ppcPasswdExpire = unsafe.Offsetof(unix.Passwd{}.Expire)
)

var (
	_ [40 - ppcPasswdSize]byte
	_ [ppcPasswdSize - 40]byte
	_ [8 - ppcPasswdUid]byte
	_ [ppcPasswdUid - 8]byte
	_ [12 - ppcPasswdGid]byte
	_ [ppcPasswdGid - 12]byte
	_ [16 - ppcPasswdChange]byte
	_ [ppcPasswdChange - 16]byte
	_ [20 - ppcPasswdClass]byte
	_ [ppcPasswdClass - 20]byte
	_ [24 - ppcPasswdGecos]byte
	_ [ppcPasswdGecos - 24]byte
	_ [28 - ppcPasswdDir]byte
	_ [ppcPasswdDir - 28]byte
	_ [32 - ppcPasswdShell]byte
	_ [ppcPasswdShell - 32]byte
	_ [36 - ppcPasswdExpire]byte
	_ [ppcPasswdExpire - 36]byte
)

func TestPasswdLayout(t *testing.T) {
	fields := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"size", ppcPasswdSize, 40},
		{"Uid", ppcPasswdUid, 8},
		{"Gid", ppcPasswdGid, 12},
		{"Change", ppcPasswdChange, 16},
		{"Class", ppcPasswdClass, 20},
		{"Gecos", ppcPasswdGecos, 24},
		{"Dir", ppcPasswdDir, 28},
		{"Shell", ppcPasswdShell, 32},
		{"Expire", ppcPasswdExpire, 36},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("Passwd.%s offset/size = %d, want %d", field.name, field.got, field.want)
		}
	}
}
