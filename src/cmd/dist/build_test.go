// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
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

func TestDistNativeDarwinPPCHostDetection(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if testing.Short() {
		t.Skip("skipping building cmd/dist in short mode")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	goroot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))

	// Model only the platform facts unavailable on this Linux test host. The
	// real dist startup, uname parsing, xinit validation and target default run.
	mainSrc, err := os.ReadFile(filepath.Join(wd, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	main := strings.Replace(string(mainSrc), "gohostos = runtime.GOOS", `gohostos = "darwin"`, 1)
	main = strings.Replace(main, "func nativeGOARCH() string {\n\treturn runtime.GOARCH\n}", "func nativeGOARCH() string {\n\treturn \"ppc64\"\n}", 1)
	if main == string(mainSrc) || strings.Contains(main, "return runtime.GOARCH") {
		t.Fatal("platform-fact overlay did not replace both intended seams")
	}
	mainOverlay := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(mainOverlay, []byte(main), 0600); err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	overlayJSON, err := json.Marshal(map[string]map[string]string{"Replace": {filepath.Join(wd, "main.go"): mainOverlay}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlay, overlayJSON, 0600); err != nil {
		t.Fatal(err)
	}

	dist := filepath.Join(t.TempDir(), "dist")
	build := exec.Command(testenv.GoToolPath(t), "build", "-overlay", overlay, "-o", dist, ".")
	build.Dir = wd
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build fixture dist: %v\n%s", err, out)
	}

	fakeBin := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(fakeBin, 0700); err != nil {
		t.Fatal(err)
	}
	uname := "#!/bin/sh\ncase $1 in\n-m) echo 'Power Macintosh' ;;\n-a) if [ \"$UNAME_ARM64\" = 1 ]; then echo 'Darwin armhost 21.1.0 xnu-8019/RELEASE_ARM64_T6000 x86_64'; else echo 'Darwin pmg5.lan 9.8.0 Darwin Kernel Version 9.8.0: Wed Jul 15 16:57:01 PDT 2009; root:xnu-1228.15.4~1/RELEASE_PPC Power Macintosh'; fi ;;\n-v) echo 'Darwin Kernel Version 9.8.0: root:xnu-1228.15.4~1/RELEASE_PPC' ;;\n*) exit 2 ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(fakeBin, "uname"), []byte(uname), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(dist, "env")
	cmd.Env = []string{"PATH=" + fakeBin + ":/usr/bin:/bin", "TMPDIR=" + os.TempDir(), "GOROOT=" + goroot}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("direct fixture dist env: %v\n%s", err, out)
	}
	for _, key := range []string{"GOHOSTARCH", "GOARCH", "GOPPC64"} {
		if strings.Contains(strings.Join(cmd.Env, "\n"), key+"=") {
			t.Fatalf("fixture exec environment unexpectedly contains %s", key)
		}
	}
	for _, want := range []string{`GOHOSTARCH="ppc64";`, `GOARCH="ppc64";`, `GOPPC64="ppc970";`} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("direct fixture output lacks %q:\n%s", want, out)
		}
	}

	// The uname RELEASE_ARM64 branch must continue to outrank the runtime
	// architecture fallback.
	arm := exec.Command(dist, "env")
	arm.Env = append(cmd.Env, "UNAME_ARM64=1")
	armOut, err := arm.CombinedOutput()
	if err != nil || !strings.Contains(string(armOut), `GOHOSTARCH="arm64";`) {
		t.Fatalf("Darwin translated ARM64 control: err=%v\n%s", err, armOut)
	}

	// An explicit host override is applied later by xinit and must win over
	// both uname and the modeled runtime fact.
	override := exec.Command(dist, "env")
	override.Env = append(cmd.Env, "GOHOSTARCH=amd64")
	overrideOut, err := override.CombinedOutput()
	if err != nil || !strings.Contains(string(overrideOut), `GOHOSTARCH="amd64";`) {
		t.Fatalf("explicit GOHOSTARCH control: err=%v\n%s", err, overrideOut)
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

// TestMatchtagArchTags checks that dist's own evaluation of //go:build lines
// agrees with the architecture sub-version tags that the real build system
// derives (internal/buildcfg.gogoarchTags). A disagreement makes a negated tag
// such as "!ppc64.ppc970" evaluate as true, which admits ISA-specific code into
// a binary built for an older machine.
func TestMatchtagArchTags(t *testing.T) {
	savedArch, savedPpc64, savedAmd64 := goarch, goppc64, goamd64
	defer func() {
		goarch, goppc64, goamd64 = savedArch, savedPpc64, savedAmd64
	}()

	tests := []struct {
		arch, ppc64, amd64, tag string
		want                    bool
	}{
		{"ppc64", "ppc970", "v1", "ppc64.ppc970", true},
		{"ppc64", "ppc970", "v1", "ppc64.power8", false},
		{"ppc64", "ppc970", "v1", "ppc64.power9", false},
		{"ppc64", "ppc970", "v1", "ppc64.power10", false},
		{"ppc64", "power8", "v1", "ppc64.power8", true},
		{"ppc64", "power8", "v1", "ppc64.ppc970", false},
		{"ppc64", "power8", "v1", "ppc64.power9", false},
		{"ppc64", "power9", "v1", "ppc64.power8", true},
		{"ppc64", "power9", "v1", "ppc64.power9", true},
		{"ppc64", "power9", "v1", "ppc64.power10", false},
		{"ppc64le", "power10", "v1", "ppc64le.power8", true},
		{"ppc64le", "ppc970", "v1", "ppc64le.ppc970", false}, // buildcfg emits nothing here; a divergent tag would flip a negated constraint
		{"amd64", "power8", "v1", "amd64.v1", true},
		{"amd64", "power8", "v1", "amd64.v2", false},
		{"amd64", "power8", "v3", "amd64.v2", true},
		{"amd64", "power8", "v3", "amd64.v4", false},
		{"arm64", "power8", "v1", "ppc64.ppc970", false},
	}
	for _, tt := range tests {
		goarch, goppc64, goamd64 = tt.arch, tt.ppc64, tt.amd64
		if got := matchtag(tt.tag); got != tt.want {
			t.Errorf("matchtag(%q) with goarch=%q goppc64=%q goamd64=%q = %v, want %v",
				tt.tag, tt.arch, tt.ppc64, tt.amd64, got, tt.want)
		}
	}
}

// TestShouldbuildPPC970Crypto is the regression test for a SIGILL in a native
// make.bash on a 970. dist builds the bootstrap go command itself, evaluating
// //go:build lines with matchtag/shouldbuild. Before those knew about the port's
// ppc64.ppc970 tag, "!ppc64.ppc970" evaluated as true, so the POWER8 AES/SHA-2
// assembly was compiled into go_bootstrap; it executed vshasigmaw on a 970 and
// died with "SIGILL" (and cascaded into "semasleep on Darwin signal stack").
// The POWER8 file must therefore be excluded, and the generic one included.
func TestShouldbuildPPC970Crypto(t *testing.T) {
	savedArch, savedPpc64 := goarch, goppc64
	defer func() {
		goarch, goppc64 = savedArch, savedPpc64
	}()

	const pkg = "crypto/internal/fips140/sha256"
	asmFile := filepath.Join("..", "..", "crypto", "internal", "fips140", "sha256", "sha256block_ppc64x.go")
	noasmFile := filepath.Join("..", "..", "crypto", "internal", "fips140", "sha256", "sha256block_noasm.go")

	goarch = "ppc64"
	goppc64 = "ppc970"
	if shouldbuild(asmFile, pkg) {
		t.Errorf("shouldbuild(%s) = true with GOPPC64=ppc970; the POWER8 implementation must be excluded", asmFile)
	}
	if !shouldbuild(noasmFile, pkg) {
		t.Errorf("shouldbuild(%s) = false with GOPPC64=ppc970; the generic implementation is required", noasmFile)
	}

	// Control: on a POWER8 (and later) target the assembly implementation is
	// the one that must be built.
	for _, setting := range []string{"power8", "power9", "power10"} {
		goppc64 = setting
		if !shouldbuild(asmFile, pkg) {
			t.Errorf("shouldbuild(%s) = false with GOPPC64=%s; the POWER8 implementation is required there", asmFile, setting)
		}
		if shouldbuild(noasmFile, pkg) {
			t.Errorf("shouldbuild(%s) = true with GOPPC64=%s; the generic implementation must be excluded there", noasmFile, setting)
		}
	}
}

// TestArchTagsFollowTheArchitectureBeingBuilt is the regression test for a
// stale tag set. cmd/dist builds for the target and then switches to build for
// the host (cmdbootstrap), so the sub-version tags must follow goarch, not a
// value computed once for the target.
//
// This is not hypothetical. With the tags derived once for a darwin/ppc64
// target, a cross-bootstrap from linux/amd64 still had ppc64.ppc970 "set" while
// building for amd64, and crypto/internal/fips140/subtle then defined xorBytes
// nowhere at all: xor_asm.go and xor_generic.go are both guarded by
// "!ppc64.ppc970", and xor_ppc970.go requires ppc64. The bootstrap failed with
// "undefined: xorBytes". The three files' guards make this package the sharpest
// available probe for the host/target switch.
func TestArchTagsFollowTheArchitectureBeingBuilt(t *testing.T) {
	savedArch, savedPpc64 := goarch, goppc64
	defer func() { goarch, goppc64 = savedArch, savedPpc64 }()

	const pkg = "crypto/internal/fips140/subtle"
	dir := filepath.Join("..", "..", "crypto", "internal", "fips140", "subtle")
	asmFile := filepath.Join(dir, "xor_asm.go")       // (amd64||arm64||ppc64||ppc64le||riscv64) && !purego && !ppc64.ppc970
	genFile := filepath.Join(dir, "xor_generic.go")   // (... || purego) && !ppc64.ppc970
	ppc970File := filepath.Join(dir, "xor_ppc970.go") // ppc64 && ppc64.ppc970

	check := func(when, arch, ppc64 string, wantAsm, wantGen, wantPPC970, wantProvider bool) {
		t.Helper()
		goarch, goppc64 = arch, ppc64
		if got := shouldbuild(asmFile, pkg); got != wantAsm {
			t.Errorf("%s: shouldbuild(xor_asm.go) = %v, want %v", when, got, wantAsm)
		}
		if got := shouldbuild(genFile, pkg); got != wantGen {
			t.Errorf("%s: shouldbuild(xor_generic.go) = %v, want %v", when, got, wantGen)
		}
		if got := shouldbuild(ppc970File, pkg); got != wantPPC970 {
			t.Errorf("%s: shouldbuild(xor_ppc970.go) = %v, want %v", when, got, wantPPC970)
		}
		// Where one of these three is the provider, exactly one must be
		// selected, or xorBytes is undefined. Other architectures (loong64)
		// have their own file, so this invariant does not apply there.
		n := 0
		for _, f := range []string{asmFile, genFile, ppc970File} {
			if shouldbuild(f, pkg) {
				n++
			}
		}
		if wantProvider && n != 1 {
			t.Errorf("%s: %d of the three xor implementations selected, want exactly 1 (else xorBytes is undefined)", when, n)
		}
		if !wantProvider && n != 0 {
			t.Errorf("%s: %d of the three xor implementations selected, want 0 (this architecture has its own provider)", when, n)
		}
	}

	// Host build for amd64 while the target is darwin/ppc64 tagged ppc970.
	// ppc64.ppc970 must NOT be set here.
	check("target darwin/ppc64 ppc970, building for the amd64 host", "amd64", "ppc970", true, false, false, true)
	// Native ppc64 build with the port's floor.
	check("ppc64 tagged ppc970", "ppc64", "ppc970", false, false, true, true)
	// An ordinary POWER8 target must still get the assembly implementation.
	check("ppc64 tagged power8", "ppc64", "power8", true, false, false, true)
	// A target with no ppc64 assembly at all falls back to the generic one.
	check("loong64 (its own provider)", "loong64", "power8", false, false, false, false)
}
