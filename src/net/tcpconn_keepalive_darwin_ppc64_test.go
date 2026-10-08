// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc64

package net

import (
	"errors"
	"internal/poll"
	"io"
	"syscall"
	"testing"
	"time"
)

func getCurrentKeepAliveSettings(fd fdType) (cfg KeepAliveConfig, err error) {
	keepAlive, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE)
	if err != nil {
		return
	}
	idle, err := syscall.GetsockoptInt(fd, syscall.IPPROTO_TCP, syscall_TCP_KEEPIDLE)
	if err != nil {
		return
	}
	cfg.Enable = keepAlive != 0
	cfg.Idle = time.Duration(idle) * time.Second
	return
}

func verifyKeepAliveSettings(t *testing.T, fd fdType, oldCfg, cfg KeepAliveConfig) {
	t.Helper()
	if cfg.Idle == 0 {
		cfg.Idle = defaultTCPKeepAliveIdle
	} else if cfg.Idle == -1 {
		cfg.Idle = oldCfg.Idle
	}
	keepAlive, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE)
	if err != nil {
		t.Fatal(err)
	}
	if (keepAlive != 0) != cfg.Enable {
		t.Fatalf("SO_KEEPALIVE: got %t; want %t", keepAlive != 0, cfg.Enable)
	}
	idle, err := syscall.GetsockoptInt(fd, syscall.IPPROTO_TCP, syscall_TCP_KEEPIDLE)
	if err != nil {
		t.Fatal(err)
	}
	if time.Duration(idle)*time.Second != cfg.Idle {
		t.Fatalf("TCP_KEEPALIVE: got %ds; want %v", idle, cfg.Idle)
	}
	// Leopard does not expose per-socket interval/count options. Their requested
	// values are ignored and the kernel's system-wide defaults remain in effect.
	t.Log("TCP_KEEPINTVL and TCP_KEEPCNT are not settable per socket on Leopard; system defaults apply")
}

func TestDarwinKeepAliveBestEffortRoundTrip(t *testing.T) {
	ln, err := Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	serverErr := make(chan error, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer c.Close()
		var b [1]byte
		if _, err := io.ReadFull(c, b[:]); err == nil {
			_, err = c.Write(b[:])
		}
		serverErr <- err
	}()

	d := Dialer{KeepAliveConfig: KeepAliveConfig{Enable: true, Idle: time.Second, Interval: time.Second, Count: 2}}
	c, err := d.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Dial with unsupported interval/count: %v", err)
	}
	defer c.Close()
	if _, err := c.Write([]byte{'K'}); err != nil {
		t.Fatal(err)
	}
	var got [1]byte
	if _, err := io.ReadFull(c, got[:]); err != nil {
		t.Fatalf("read reply after successful dial: %v", err)
	}
	if got[0] != 'K' {
		t.Fatalf("round-trip byte = %q; want K", got[0])
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("server round trip: %v", err)
	}
}

func TestDarwinKeepAliveOptionErrorPropagates(t *testing.T) {
	// Negative control: setsockopt on an invalid descriptor fails with EBADF,
	// which must survive the same path that ignores ENOPROTOOPT.
	fd := &netFD{pfd: poll.FD{Sysfd: -1}}
	if err := setKeepAliveInterval(fd, time.Second); err == nil {
		t.Fatal("setKeepAliveInterval with invalid descriptor succeeded; want EBADF")
	} else if !errors.Is(err, syscall.EBADF) {
		t.Fatalf("setKeepAliveInterval with invalid descriptor: got %v; want EBADF", err)
	}
}
