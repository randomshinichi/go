// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBootstrapHostPPC32Baseline is the ppc32 counterpart of
// TestBootstrapHostPPC64Floor. It drives the same real xinit and the same
// platform switches cmdbootstrap makes (switchToHostPlatform and
// restoreTargetPlatform), with the host modelled as a Darwin/ppc machine, and
// pins what dist itself decides for a ppc32 build.
//
// There is deliberately no ppc32 ISA-floor case, because ppc32 has no
// build-floor setting that dist could get wrong. $GOPPC64 is a ppc64-only knob:
// for GOARCH=ppc it is validated and then not consulted, so GOPPC64=ppc970 and
// GOPPC64=power8 produce byte-identical darwin/ppc binaries
// (docs/plans/ppc-mac-plan.md). The ppc32 floor is the G3/750 scalar baseline,
// declared in the ppc assembler's refusal table and censused from the emitted
// object, not selected here. What dist does own for GOARCH=ppc is the platform
// each phase builds for and what that platform admits; that is what this test
// pins.
//
// Scope, so that a green here is not over-read: simulateDist supplies the
// modelled host identity (gohostos/gohostarch) as package globals, exactly as
// the ppc64 test does, and xinit never detects the host on this path (the
// nativeGOARCH fallback runs only from main). This is therefore a model of
// dist's phase behaviour, not evidence of host detection on a real machine.
// Host detection is controlled separately, by running the built target
// `dist env` as a standalone process in an empty environment on the machine.
func TestBootstrapHostPPC32Baseline(t *testing.T) {
	const pkg = "crypto/internal/fips140/sha256"
	dir := filepath.Join("..", "..", "crypto", "internal", "fips140", "sha256")
	power8SHA := filepath.Join(dir, "sha256block_ppc64x.go")
	genericSHA := filepath.Join(dir, "sha256block_noasm.go")

	type testCase struct {
		name       string
		targetOS   string
		targetArch string
		setting    string // $GOPPC64 for the target; "" is unset
		wantFloor  string // the target's floor, for ppc64 targets only
	}
	cases := []testCase{
		{name: "native darwin/ppc", targetOS: "darwin", targetArch: "ppc"},
		{name: "darwin/ppc host, target linux/amd64", targetOS: "linux", targetArch: "amd64"},
		{name: "darwin/ppc host, target darwin/ppc64", targetOS: "darwin", targetArch: "ppc64", wantFloor: "ppc970"},
		{name: "darwin/ppc host, target linux/ppc64/GOPPC64=power10", targetOS: "linux", targetArch: "ppc64", setting: "power10", wantFloor: "power10"},
	}

	// checkPlatform names the field it failed on, so that a broken phase says
	// which of GOOS and GOARCH was wrong, not merely that something was.
	// t is the subtest's, so that FailNow ends the phase under test and not the
	// test above it.
	checkPlatform := func(t *testing.T, phase, wantOS, wantArch string) {
		t.Helper()
		if goos != wantOS {
			t.Fatalf("%s: GOOS=%q, want %q", phase, goos, wantOS)
		}
		if goarch != wantArch {
			t.Fatalf("%s: GOARCH=%q, want %q", phase, goarch, wantArch)
		}
	}

	// check asserts what a phase building for the current goos/goarch may use.
	// wantFloor is the floor the target must get, and is set only for ppc64
	// targets: a ppc32 build has no floor to select.
	check := func(t *testing.T, phase, wantFloor string) {
		t.Helper()
		if goarch != "ppc" {
			if wantFloor == "" {
				return
			}
			if got := ppc64Floor(); got != wantFloor {
				t.Errorf("%s: ppc64 floor = %q, want %q", phase, got, wantFloor)
			}
			if got := strings.Join(ppc64AsmDefinesForArch(goarch), " "); !strings.Contains(got, "GOPPC64_") {
				t.Errorf("%s: ppc64 assembler arguments = %q, want a PPC64 ISA level", phase, got)
			}
			return
		}
		// GOARCH=ppc may not admit anything that assumes a POWER ISA level.
		if tags := archTags(); len(tags) != 0 {
			t.Errorf("%s: GOARCH=ppc build tags = %v, want none", phase, tags)
		}
		if defs := ppc64AsmDefinesForArch(goarch); len(defs) != 0 {
			t.Errorf("%s: GOARCH=ppc assembler arguments = %q, want none ($GOPPC64 is a ppc64-only setting)", phase, defs)
		}
		if shouldbuild(power8SHA, pkg) {
			t.Errorf("%s: GOARCH=ppc admits the POWER8-only source %s", phase, filepath.Base(power8SHA))
		}
		if !shouldbuild(genericSHA, pkg) {
			t.Errorf("%s: GOARCH=ppc does not admit the generic source %s", phase, filepath.Base(genericSHA))
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"GOOS": tc.targetOS, "GOARCH": tc.targetArch}
			if tc.setting != "" {
				env["GOPPC64"] = tc.setting
			}
			simulateDist(t, "darwin", "ppc", nil, env)

			// xinit exports the platform it was handed. This pins that export;
			// it is not host detection, because the modelled values came from
			// the call above.
			if got := os.Getenv("GOHOSTOS"); got != "darwin" {
				t.Errorf("GOHOSTOS exported by xinit = %q, want darwin", got)
			}
			if got := os.Getenv("GOHOSTARCH"); got != "ppc" {
				t.Errorf("GOHOSTARCH exported by xinit = %q, want ppc", got)
			}

			const targetPhase = "target phase (after xinit)"
			checkPlatform(t, targetPhase, tc.targetOS, tc.targetArch)
			check(t, targetPhase, tc.wantFloor)

			// cmdbootstrap's switch to the build's own tools. Whatever the
			// target is, the host phase builds darwin/ppc binaries for this
			// toolchain to run.
			const hostPhase = "host phase (go_bootstrap, host tools)"
			switchToHostPlatform()
			checkPlatform(t, hostPhase, "darwin", "ppc")
			check(t, hostPhase, "")

			const restoredPhase = "target phase restored"
			restoreTargetPlatform()
			checkPlatform(t, restoredPhase, tc.targetOS, tc.targetArch)
			check(t, restoredPhase, tc.wantFloor)
		})
	}
}
