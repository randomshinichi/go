// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"internal/platform"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestMustLinkExternal verifies that the mustLinkExternal helper
// function matches internal/platform.MustLinkExternal.
func TestMustLinkExternal(t *testing.T) {
	for _, goos := range okgoos {
		for _, goarch := range okgoarch {
			for _, cgoEnabled := range []bool{true, false} {
				got := mustLinkExternal(goos, goarch, cgoEnabled)
				want := platform.MustLinkExternal(goos, goarch, cgoEnabled)
				if got != want {
					t.Errorf("mustLinkExternal(%q, %q, %v) = %v; want %v", goos, goarch, cgoEnabled, got, want)
				}
			}
		}
	}
}

func TestRequiredBootstrapVersion(t *testing.T) {
	testCases := map[string]string{
		"1.22": "1.20",
		"1.23": "1.20",
		"1.24": "1.22",
		"1.25": "1.22",
		"1.26": "1.24",
		"1.27": "1.24",
	}

	for v, want := range testCases {
		if got := requiredBootstrapVersion(v); got != want {
			t.Errorf("requiredBootstrapVersion(%v): got %v, want %v", v, got, want)
		}
	}
}

// TestDistEnvGOPPC64Default builds cmd/dist and runs the resulting binary
// directly, as src/make.bash does, with GOPPC64 absent from its environment.
// It must not go through "go tool dist": the go command exports an explicit
// GOPPC64 to the tools it runs, which bypasses the default selected by xinit.
func TestDistEnvGOPPC64Default(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if testing.Short() {
		t.Skip("skipping building cmd/dist in short mode")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	goroot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))

	dist := filepath.Join(t.TempDir(), "dist")
	build := exec.Command(testenv.GoToolPath(t), "build", "-o", dist, ".")
	build.Dir = wd
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build cmd/dist: %v\n%s", err, out)
	}

	// Unset GOOS falls back to the host OS.
	hostOSDefault := "power8"
	if runtime.GOOS == "darwin" {
		hostOSDefault = "ppc970"
	}
	// Unset GOARCH falls back to GOHOSTARCH, so the native darwin/ppc64
	// case is simulated with GOHOSTARCH=ppc64. Want "" means dist env
	// does not print GOPPC64 because the target is not a ppc64 architecture.
	tests := []struct {
		name string
		env  []string
		want string
	}{
		{"GOOS unset", []string{"GOARCH=ppc64"}, hostOSDefault},
		{"darwin/ppc64", []string{"GOOS=darwin", "GOARCH=ppc64"}, "ppc970"},
		{"darwin/ppc64 GOARCH unset", []string{"GOOS=darwin", "GOHOSTARCH=ppc64"}, "ppc970"},
		{"darwin/ppc64 explicit power8", []string{"GOOS=darwin", "GOARCH=ppc64", "GOPPC64=power8"}, "power8"},
		{"darwin/ppc64 explicit power9", []string{"GOOS=darwin", "GOARCH=ppc64", "GOPPC64=power9"}, "power9"},
		{"darwin/ppc64 empty GOPPC64", []string{"GOOS=darwin", "GOARCH=ppc64", "GOPPC64="}, "ppc970"},
		{"linux/ppc64", []string{"GOOS=linux", "GOARCH=ppc64"}, "power8"},
		{"linux/ppc64le", []string{"GOOS=linux", "GOARCH=ppc64le"}, "power8"},
		{"darwin/amd64", []string{"GOOS=darwin", "GOARCH=amd64"}, ""},
	}
	line := regexp.MustCompile(`(?m)^GOPPC64="([^"]*)"`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(dist, "env")
			for _, kv := range os.Environ() {
				switch strings.SplitN(kv, "=", 2)[0] {
				case "GOROOT", "GOOS", "GOARCH", "GOHOSTOS", "GOHOSTARCH", "GOPPC64", "GOFLAGS":
				default:
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Env = append(cmd.Env, "GOROOT="+goroot)
			cmd.Env = append(cmd.Env, tt.env...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("dist env: %v\n%s", err, out)
			}
			m := line.FindSubmatch(out)
			if tt.want == "" {
				if m != nil {
					t.Errorf("unexpected GOPPC64 line in dist env output:\n%s", out)
				}
				return
			}
			if m == nil {
				t.Fatalf("no GOPPC64 line in dist env output:\n%s", out)
			}
			if got := string(m[1]); got != tt.want {
				t.Errorf("GOPPC64 = %q, want %q (env %q)", got, tt.want, tt.env)
			}
		})
	}
}
