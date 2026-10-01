// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildcfg

import (
	"os"
	"slices"
	"testing"
)

func TestConfigFlags(t *testing.T) {
	os.Setenv("GOAMD64", "v1")
	if goamd64() != 1 {
		t.Errorf("Wrong parsing of GOAMD64=v1")
	}
	os.Setenv("GOAMD64", "v4")
	if goamd64() != 4 {
		t.Errorf("Wrong parsing of GOAMD64=v4")
	}
	Error = nil
	os.Setenv("GOAMD64", "1")
	if goamd64(); Error == nil {
		t.Errorf("Wrong parsing of GOAMD64=1")
	}

	os.Setenv("GORISCV64", "rva20u64")
	if goriscv64() != 20 {
		t.Errorf("Wrong parsing of RISCV64=rva20u64")
	}
	os.Setenv("GORISCV64", "rva22u64")
	if goriscv64() != 22 {
		t.Errorf("Wrong parsing of RISCV64=rva22u64")
	}
	os.Setenv("GORISCV64", "rva23u64")
	if goriscv64() != 23 {
		t.Errorf("Wrong parsing of RISCV64=rva23u64")
	}
	Error = nil
	os.Setenv("GORISCV64", "rva22")
	if _ = goriscv64(); Error == nil {
		t.Errorf("Wrong parsing of RISCV64=rva22")
	}
	Error = nil
	os.Setenv("GOARM64", "v7.0")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=7.0")
	}
	Error = nil
	os.Setenv("GOARM64", "8.0")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=8.0")
	}
	Error = nil
	os.Setenv("GOARM64", "v8.0,lsb")
	if _ = goarm64(); Error == nil {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lsb")
	}
	os.Setenv("GOARM64", "v8.0,lse")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != false {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lse")
	}
	os.Setenv("GOARM64", "v8.0,crypto")
	if goarm64().Version != "v8.0" || goarm64().LSE != false || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,crypto")
	}
	os.Setenv("GOARM64", "v8.0,crypto,lse")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,crypto,lse")
	}
	os.Setenv("GOARM64", "v8.0,lse,crypto")
	if goarm64().Version != "v8.0" || goarm64().LSE != true || goarm64().Crypto != true {
		t.Errorf("Wrong parsing of GOARM64=v8.0,lse,crypto")
	}
	os.Setenv("GOARM64", "v9.0")
	if goarm64().Version != "v9.0" || goarm64().LSE != true || goarm64().Crypto != false {
		t.Errorf("Wrong parsing of GOARM64=v9.0")
	}
}

func TestGOPPC64(t *testing.T) {
	oldGOPPC64, oldGOARCH, oldError := GOPPC64, GOARCH, Error
	defer func() {
		GOPPC64, GOARCH, Error = oldGOPPC64, oldGOARCH, oldError
	}()

	if DefaultGOPPC64 != "power8" {
		t.Fatalf("DefaultGOPPC64 = %q, want power8", DefaultGOPPC64)
	}

	for _, tc := range []struct {
		value string
		want  int
	}{
		{"ppc970", 5},
		{"power8", 8},
		{"power9", 9},
		{"power10", 10},
	} {
		t.Setenv("GOPPC64", tc.value)
		Error = nil
		GOPPC64 = goppc64()
		if GOPPC64 != tc.want || Error != nil {
			t.Errorf("GOPPC64 from %q = %d, error %v; want %d, nil", tc.value, GOPPC64, Error, tc.want)
		}
	}

	t.Setenv("GOPPC64", "")
	Error = nil
	GOPPC64 = goppc64()
	if GOPPC64 != 8 || Error != nil {
		t.Errorf("GOPPC64 from empty setting = %d, error %v; want default 8, nil", GOPPC64, Error)
	}

	for _, invalid := range []string{"power6"} {
		t.Setenv("GOPPC64", invalid)
		Error = nil
		GOPPC64 = goppc64()
		if GOPPC64 != 8 || Error == nil {
			t.Errorf("GOPPC64 from invalid %s = %d, error %v; want default 8 and validation error", invalid, GOPPC64, Error)
		} else if got, want := Error.Error(), "invalid GOPPC64: must be ppc970, power8, power9, power10"; got != want {
			t.Errorf("GOPPC64 error = %q, want %q", got, want)
		}
	}
	Error = nil

	GOARCH = "ppc64"
	GOPPC64 = 5
	if name, value := GOGOARCH(); name != "GOPPC64" || value != "ppc970" {
		t.Errorf("GOGOARCH() for GOPPC64=5 = (%q, %q), want (GOPPC64, ppc970)", name, value)
	}
	if tags := gogoarchTags(); !slices.Equal(tags, []string{"ppc64.ppc970"}) {
		t.Errorf("ppc64 build tags at GOPPC64=5 = %v, want [ppc64.ppc970]", tags)
	}

	for _, tc := range []struct {
		arch  string
		level int
		want  []string
	}{
		{"ppc64", 8, []string{"ppc64.power8"}},
		{"ppc64", 9, []string{"ppc64.power8", "ppc64.power9"}},
		{"ppc64", 10, []string{"ppc64.power8", "ppc64.power9", "ppc64.power10"}},
		{"ppc64le", 8, []string{"ppc64le.power8"}},
		{"ppc64le", 5, nil},
	} {
		GOARCH, GOPPC64 = tc.arch, tc.level
		if tags := gogoarchTags(); !slices.Equal(tags, tc.want) {
			t.Errorf("%s build tags at GOPPC64=%d = %v, want %v", tc.arch, tc.level, tags, tc.want)
		}
	}
	GOARCH, GOPPC64 = "ppc64", 8
	if name, value := GOGOARCH(); name != "GOPPC64" || value != "power8" {
		t.Errorf("GOGOARCH() for GOPPC64=8 = (%q, %q), want (GOPPC64, power8)", name, value)
	}
	Error = nil
}

func TestGoarm64FeaturesSupports(t *testing.T) {
	g, _ := ParseGoarm64("v9.3")

	if !g.Supports("v9.3") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v9.3")
	}

	if g.Supports("v9.4") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v9.4")
	}

	if !g.Supports("v8.8") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v8.8")
	}

	if g.Supports("v8.9") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, v8.9")
	}

	if g.Supports(",lse") {
		t.Errorf("Wrong goarm64Features.Supports for v9.3, ,lse")
	}
}

func TestGogoarchTags(t *testing.T) {
	old_goarch := GOARCH
	old_goarm64 := GOARM64

	GOARCH = "arm64"

	os.Setenv("GOARM64", "v9.5")
	GOARM64 = goarm64()
	tags := gogoarchTags()
	want := []string{"arm64.v9.0", "arm64.v9.1", "arm64.v9.2", "arm64.v9.3", "arm64.v9.4", "arm64.v9.5",
		"arm64.v8.0", "arm64.v8.1", "arm64.v8.2", "arm64.v8.3", "arm64.v8.4", "arm64.v8.5", "arm64.v8.6", "arm64.v8.7", "arm64.v8.8", "arm64.v8.9"}
	if len(tags) != len(want) {
		t.Errorf("Wrong number of tags for GOARM64=v9.5")
	} else {
		for i, v := range tags {
			if v != want[i] {
				t.Error("Wrong tags for GOARM64=v9.5")
				break
			}
		}
	}

	GOARCH = old_goarch
	GOARM64 = old_goarm64
}

var goodFIPS = []string{
	"v1.0.0",
	"v1.0.1",
	"v1.2.0",
	"v1.2.3",
}

var badFIPS = []string{
	"v1.0.0-fips",
	"v1.0.0+fips",
	"1.0.0",
	"x1.0.0",
}

func TestIsFIPSVersion(t *testing.T) {
	// good
	for _, s := range goodFIPS {
		if !isFIPSVersion(s) {
			t.Errorf("isFIPSVersion(%q) = false, want true", s)
		}
	}
	// truncated
	const v = "v1.2.3"
	for i := 0; i < len(v); i++ {
		if isFIPSVersion(v[:i]) {
			t.Errorf("isFIPSVersion(%q) = true, want false", v[:i])
		}
	}
	// bad
	for _, s := range badFIPS {
		if isFIPSVersion(s) {
			t.Errorf("isFIPSVersion(%q) = true, want false", s)
		}
	}
}
