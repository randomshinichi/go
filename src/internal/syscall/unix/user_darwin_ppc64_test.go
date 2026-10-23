// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package unix_test

import (
	"testing"
	"unsafe"

	"internal/syscall/unix"
)

const (
	ppc64PasswdSize   = unsafe.Sizeof(unix.Passwd{})
	ppc64PasswdUid    = unsafe.Offsetof(unix.Passwd{}.Uid)
	ppc64PasswdGid    = unsafe.Offsetof(unix.Passwd{}.Gid)
	ppc64PasswdChange = unsafe.Offsetof(unix.Passwd{}.Change)
	ppc64PasswdClass  = unsafe.Offsetof(unix.Passwd{}.Class)
	ppc64PasswdGecos  = unsafe.Offsetof(unix.Passwd{}.Gecos)
	ppc64PasswdDir    = unsafe.Offsetof(unix.Passwd{}.Dir)
	ppc64PasswdShell  = unsafe.Offsetof(unix.Passwd{}.Shell)
	ppc64PasswdExpire = unsafe.Offsetof(unix.Passwd{}.Expire)
)

var (
	_ [72 - ppc64PasswdSize]byte
	_ [ppc64PasswdSize - 72]byte
	_ [16 - ppc64PasswdUid]byte
	_ [ppc64PasswdUid - 16]byte
	_ [20 - ppc64PasswdGid]byte
	_ [ppc64PasswdGid - 20]byte
	_ [24 - ppc64PasswdChange]byte
	_ [ppc64PasswdChange - 24]byte
	_ [32 - ppc64PasswdClass]byte
	_ [ppc64PasswdClass - 32]byte
	_ [40 - ppc64PasswdGecos]byte
	_ [ppc64PasswdGecos - 40]byte
	_ [48 - ppc64PasswdDir]byte
	_ [ppc64PasswdDir - 48]byte
	_ [56 - ppc64PasswdShell]byte
	_ [ppc64PasswdShell - 56]byte
	_ [64 - ppc64PasswdExpire]byte
	_ [ppc64PasswdExpire - 64]byte
)

func TestPasswdLayout(t *testing.T) {
	fields := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"size", ppc64PasswdSize, 72},
		{"Uid", ppc64PasswdUid, 16},
		{"Gid", ppc64PasswdGid, 20},
		{"Change", ppc64PasswdChange, 24},
		{"Class", ppc64PasswdClass, 32},
		{"Gecos", ppc64PasswdGecos, 40},
		{"Dir", ppc64PasswdDir, 48},
		{"Shell", ppc64PasswdShell, 56},
		{"Expire", ppc64PasswdExpire, 64},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("Passwd.%s offset/size = %d, want %d", field.name, field.got, field.want)
		}
	}
}
