// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// distEnvKeys lists every environment variable that xinit reads, writes or
// unsets, so that simulateDist can isolate a test from the real environment
// and restore it afterwards.
var distEnvKeys = []string{
	"GO386", "GOAMD64", "GOARCH", "GOARM", "GOARM64", "GOHOSTARCH", "GOHOSTOS",
	"GOOS", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOROOT", "GOFIPS140",
	"GOBIN", "GO_EXTLINK_ENABLED", "GOEXPERIMENT", "BOOT_GO_GCFLAGS",
	"BOOT_GO_LDFLAGS", "CC", "CXX", "PKG_CONFIG", "GO_LDSO", "GO111MODULE",
	"GOENV", "GOFLAGS", "GOWORK", "LANG", "LANGUAGE", "GOTMPDIR",
}

// saveDistGlobals restores, when the test ends, every package global that
// xinit and the platform switches write.
func saveDistGlobals(t *testing.T) {
	t.Helper()
	strs := []*string{
		&goos, &goarch, &gohostos, &gohostarch, &goppc64Env, &oldgoos, &oldgoarch,
		&goroot, &gorootBin, &gorootBinGo, &goarm, &goarm64, &go386, &goamd64,
		&gomips, &gomips64, &goriscv64, &gofips140, &goextlinkenabled,
		&goexperiment, &gogcflags, &goldflags, &defaultpkgconfig, &defaultldso,
		&workdir, &tooldir,
	}
	saved := make([]string, len(strs))
	for i, p := range strs {
		saved[i] = *p
	}
	savedCC, savedCXX := defaultcc, defaultcxx
	savedRelease, savedAtexits, savedProbe := isRelease, atexits, hostMachCPU
	t.Cleanup(func() {
		for i, p := range strs {
			*p = saved[i]
		}
		defaultcc, defaultcxx = savedCC, savedCXX
		isRelease, atexits, hostMachCPU = savedRelease, savedAtexits, savedProbe
	})
}

// simulateDist runs the REAL xinit as dist would run it on a host of
// hostOS/hostArch whose CPU the probe reports as cpu (nil: the probe fails),
// with env as the process environment ($GOOS and $GOARCH select the target).
// Nothing about the floor is injected: xinit decides it. On return the package
// is in the state cmdbootstrap starts from.
func simulateDist(t *testing.T, hostOS, hostArch string, cpu func() (uint32, uint32, error), env map[string]string) {
	t.Helper()
	saveDistGlobals(t)
	for _, k := range distEnvKeys {
		t.Setenv(k, os.Getenv(k)) // registers the restore
		os.Unsetenv(k)
	}
	for k, v := range env {
		os.Setenv(k, v)
	}
	os.Setenv("GOROOT", repoGOROOT(t))
	os.Setenv("GOTMPDIR", t.TempDir()) // xinit creates a work directory
	if cpu == nil {
		cpu = func() (uint32, uint32, error) { return 0, 0, errors.New("simulated probe failure") }
	}
	hostMachCPU = cpu
	gohostos, gohostarch = hostOS, hostArch
	oldgoos, oldgoarch = "", ""
	xinit()
}

func repoGOROOT(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", "..", ".."))
}

func g5CPU() (uint32, uint32, error) {
	return machCPUTypePowerPC, machCPUSubtypePowerPC970, nil
}

// TestPPC64FloorFromMachCPU pins the measured part of the CPU table: one entry,
// measured on the target machine. Nothing is claimed for CPUs nobody measured.
func TestPPC64FloorFromMachCPU(t *testing.T) {
	if got, ok := ppc64FloorFromMachCPU(18, 100); !ok || got != "ppc970" {
		t.Errorf("ppc64FloorFromMachCPU(18, 100) = %q, %v; want ppc970, true (measured on the G5)", got, ok)
	}
	// Unmeasured pairs must report "unknown", not a level. The values are
	// arbitrary; they are not claims about any real CPU.
	for _, c := range [][2]uint32{{18, 0}, {18, 1}, {18, 101}, {18, 0x80000064}, {0, 100}, {7, 100}, {0x01000012, 100}} {
		if got, ok := ppc64FloorFromMachCPU(c[0], c[1]); ok {
			t.Errorf("ppc64FloorFromMachCPU(%d, %d) = %q, true; want unknown", c[0], c[1], got)
		}
	}
}

// TestBootstrapHostPPC64Floor is the regression test for B4. The ISA floor of
// a compilation is a property of the machine that will run the result.
// cmdbootstrap builds, for the HOST, the runtime and go command that it then
// executes (go_bootstrap), and the toolchain, before building the target.
// Before the fix the floor was chosen once from the target, so a Power Mac G5
// cross-building for any target other than darwin/ppc64, with GOPPC64 unset,
// compiled its own host tools at power8 and then died with SIGILL running them.
//
// Each case runs the real xinit for a simulated host and target, then the real
// platform switches cmdbootstrap makes (switchToHostPlatform and
// restoreTargetPlatform), and checks at each point the floor dist's build rules
// use, the GOPPC64 exported to the tools it runs, the assembler arguments, and
// which crypto/internal/fips140/sha256 sources the real shouldbuild selects.
func TestBootstrapHostPPC64Floor(t *testing.T) {
	const pkg = "crypto/internal/fips140/sha256"
	dir := filepath.Join("..", "..", "crypto", "internal", "fips140", "sha256")
	power8SHA := filepath.Join(dir, "sha256block_ppc64x.go")
	genericSHA := filepath.Join(dir, "sha256block_noasm.go")

	type target struct{ goos, goarch string }
	unknownCPU := func() (uint32, uint32, error) { return machCPUTypePowerPC, 101, nil }
	failedProbe := (func() (uint32, uint32, error))(nil)

	type testCase struct {
		name       string
		hostOS     string
		hostArch   string
		cpu        func() (uint32, uint32, error)
		target     target
		setting    string // $GOPPC64; "" is unset
		wantTarget string // floor for target compilations
		wantHost   string // floor for the build's own host tools
	}
	var cases []testCase

	// A Power Mac G5 as the build host. Whatever the target and however
	// GOPPC64 is set, the host tools must be compiled at the 970's floor.
	for _, tgt := range []target{{"darwin", "ppc64"}, {"linux", "amd64"}, {"linux", "ppc64"}, {"linux", "s390x"}} {
		for _, setting := range []string{"", "ppc970"} {
			want := "ppc970"
			if setting == "" && tgt != (target{"darwin", "ppc64"}) {
				want = "power8" // the target's own table default
			}
			cases = append(cases, testCase{
				name:   fmt.Sprintf("G5 host/%s_%s/GOPPC64=%q", tgt.goos, tgt.goarch, setting),
				hostOS: "darwin", hostArch: "ppc64", cpu: g5CPU, target: tgt, setting: setting,
				wantTarget: want, wantHost: "ppc970",
			})
		}
	}
	cases = append(cases,
		// An explicit target setting is the TARGET's. The measured host floor
		// outranks it for the host tools.
		testCase{"G5 host/linux_ppc64/explicit power8", "darwin", "ppc64", g5CPU, target{"linux", "ppc64"}, "power8", "power8", "ppc970"},
		testCase{"G5 host/linux_ppc64/explicit power10", "darwin", "ppc64", g5CPU, target{"linux", "ppc64"}, "power10", "power10", "ppc970"},
		testCase{"G5 host/linux_amd64/explicit power8", "darwin", "ppc64", g5CPU, target{"linux", "amd64"}, "power8", "power8", "ppc970"},
		// A probe that fails, or names a CPU nobody measured, gives the
		// conservative table default for a Darwin/ppc64 host, never the
		// target's setting.
		testCase{"G5 host, probe fails/linux_ppc64/explicit power8", "darwin", "ppc64", failedProbe, target{"linux", "ppc64"}, "power8", "power8", "ppc970"},
		testCase{"G5 host, probe fails/linux_amd64/unset", "darwin", "ppc64", failedProbe, target{"linux", "amd64"}, "", "power8", "ppc970"},
		testCase{"G5 host, unmeasured subtype/linux_ppc64/explicit power8", "darwin", "ppc64", unknownCPU, target{"linux", "ppc64"}, "power8", "power8", "ppc970"},
		// Native build: host and target are the same platform, so there is no
		// host phase. A native build keeps the target's setting.
		testCase{"G5 native/darwin_ppc64/explicit power8", "darwin", "ppc64", g5CPU, target{"darwin", "ppc64"}, "power8", "power8", "power8"},

		// Controls: genuine POWER8-class hosts must still get power8 for
		// their own tools.
		testCase{"POWER8 linux/ppc64 host/linux_amd64/unset", "linux", "ppc64", nil, target{"linux", "amd64"}, "", "power8", "power8"},
		testCase{"POWER8 linux/ppc64 host/linux_ppc64/unset", "linux", "ppc64", nil, target{"linux", "ppc64"}, "", "power8", "power8"},
		testCase{"POWER8 linux/ppc64 host/darwin_ppc64/unset", "linux", "ppc64", nil, target{"darwin", "ppc64"}, "", "ppc970", "power8"},
		// The target's floor must not leak into a POWER8 host's tools, and
		// the host's must not leak into the target.
		testCase{"POWER8 linux/ppc64 host/darwin_ppc64/explicit ppc970", "linux", "ppc64", nil, target{"darwin", "ppc64"}, "ppc970", "ppc970", "ppc970"},
		// A host that cannot be measured takes an explicit setting as its
		// own: a Linux G5 cross-building with GOPPC64=ppc970 keeps working.
		testCase{"Linux G5 host (explicit)/linux_amd64/explicit ppc970", "linux", "ppc64", nil, target{"linux", "amd64"}, "ppc970", "ppc970", "ppc970"},
		// The common cross-build on the development host.
		testCase{"amd64 host/darwin_ppc64/unset", "linux", "amd64", nil, target{"darwin", "ppc64"}, "", "ppc970", "power8"},
	)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"GOOS": tc.target.goos, "GOARCH": tc.target.goarch}
			if tc.setting != "" {
				env["GOPPC64"] = tc.setting
			}
			simulateDist(t, tc.hostOS, tc.hostArch, tc.cpu, env)

			// check asserts the state of one phase. arch/floor describe the
			// platform being built in that phase and the floor it must get.
			check := func(phase, osName, arch, floor string) {
				t.Helper()
				if goos != osName || goarch != arch {
					t.Fatalf("%s: building for %s/%s, want %s/%s", phase, goos, goarch, osName, arch)
				}
				if got := ppc64Floor(); got != floor {
					t.Errorf("%s: floor = %q, want %q", phase, got, floor)
				}
				if got := os.Getenv("GOPPC64"); got != floor {
					t.Errorf("%s: $GOPPC64 exported to the tools dist runs = %q, want %q", phase, got, floor)
				}
				if arch != "ppc64" && arch != "ppc64le" {
					for _, tag := range archTags() {
						if strings.HasPrefix(tag, "ppc64") {
							t.Errorf("%s: %s build has ppc64 tag %q", phase, arch, tag)
						}
					}
					return
				}
				defs := strings.Join(ppc64AsmDefines(ppc64Floor()), " ")
				hasVSX := strings.Contains(defs, "GOPPC64_vsx")
				has970 := strings.Contains(defs, "GOPPC64_ppc970")
				if (floor == "ppc970") != has970 || (floor != "ppc970") != hasVSX {
					t.Errorf("%s: assembler arguments %q do not match floor %q", phase, defs, floor)
				}
				if arch != "ppc64" {
					return // the SHA files are selected by ppc64 and ppc64le alike, but the 970 tag exists for ppc64 only
				}
				wantPOWER8 := floor != "ppc970"
				if got := shouldbuild(power8SHA, pkg); got != wantPOWER8 {
					t.Errorf("%s: shouldbuild(POWER8 SHA-2 assembly) = %v, want %v at floor %s", phase, got, wantPOWER8, floor)
				}
				if got := shouldbuild(genericSHA, pkg); got == wantPOWER8 {
					t.Errorf("%s: shouldbuild(generic SHA-2) = %v, want %v at floor %s", phase, got, !wantPOWER8, floor)
				}
			}

			check("target phase (after xinit)", tc.target.goos, tc.target.goarch, tc.wantTarget)

			switchToHostPlatform()
			if tc.hostArch == "ppc64" {
				// (ii) the build's own tools must not contain POWER8 code.
				wantAdmitted := tc.wantHost != "ppc970"
				if got := shouldbuild(power8SHA, pkg); got != wantAdmitted {
					t.Errorf("host phase: POWER8 SHA-2 admitted = %v, want %v", got, wantAdmitted)
				}
			}
			check("host phase (go_bootstrap, host tools)", tc.hostOS, tc.hostArch, tc.wantHost)

			restoreTargetPlatform()
			check("target phase restored", tc.target.goos, tc.target.goarch, tc.wantTarget)
		})
	}
}

// TestBootstrapToolchain1HostFloor covers the one phase that does not go
// through switchToHostPlatform: toolchain1 is built by the bootstrap go command
// with GOOS and GOARCH emptied, so it builds for the host while dist's own
// goos/goarch still name the target. exportHostPPC64Floor is what
// bootstrapBuildTools exports for it.
func TestBootstrapToolchain1HostFloor(t *testing.T) {
	for _, tc := range []struct {
		name, hostOS, hostArch, targetOS, targetArch, setting, want string
		cpu                                                         func() (uint32, uint32, error)
	}{
		{"G5 host, target linux/amd64, unset", "darwin", "ppc64", "linux", "amd64", "", "ppc970", g5CPU},
		{"G5 host, target linux/ppc64, explicit power8", "darwin", "ppc64", "linux", "ppc64", "power8", "ppc970", g5CPU},
		{"POWER8 host, target darwin/ppc64, unset", "linux", "ppc64", "darwin", "ppc64", "", "power8", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"GOOS": tc.targetOS, "GOARCH": tc.targetArch}
			if tc.setting != "" {
				env["GOPPC64"] = tc.setting
			}
			simulateDist(t, tc.hostOS, tc.hostArch, tc.cpu, env)
			target := os.Getenv("GOPPC64")
			restore := exportHostPPC64Floor()
			if got := os.Getenv("GOPPC64"); got != tc.want {
				t.Errorf("$GOPPC64 for toolchain1 = %q, want %q", got, tc.want)
			}
			restore()
			if got := os.Getenv("GOPPC64"); got != target {
				t.Errorf("$GOPPC64 after toolchain1 = %q, want the target's %q restored", got, target)
			}
		})
	}
}

// TestBuildcfgAgreement is plan §3.2's targeted differential check. dist's
// ISA-floor default and architecture tags mirror internal/buildcfg, and the
// mirror cannot be removed: make.bash builds dist with the bootstrap
// toolchain against the bootstrap GOROOT's library, where no shared internal
// package exists. So the real internal/buildcfg is run in a subprocess for
// each combination, and dist's answers for the same combination must match.
//
// The combinations are the shipping target and its nearest neighbours rather
// than the full cross product: darwin/ppc64 at the default and at power8 and
// ppc970, linux/ppc64 and ppc64le at their defaults and at power9/power10,
// amd64 at v1 and v3, arm64 at v8.0 and v9.5, and one each of the remaining
// architectures that have sub-version tags.
func TestBuildcfgAgreement(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if testing.Short() {
		t.Skip("skipping building buildcfgdump in short mode")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dump := filepath.Join(t.TempDir(), "buildcfgdump")
	build := exec.Command(testenv.GoToolPath(t), "build", "-o", dump, "./testdata/buildcfgdump")
	build.Dir = wd
	build.Env = append(os.Environ(), "GOOS=", "GOARCH=")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build buildcfgdump: %v\n%s", err, out)
	}

	type row struct {
		goos, goarch string
		env          map[string]string
	}
	rows := []row{
		{"darwin", "ppc64", nil},
		{"darwin", "ppc64", map[string]string{"GOPPC64": "power8"}},
		{"darwin", "ppc64", map[string]string{"GOPPC64": "ppc970"}},
		{"linux", "ppc64", nil},
		{"linux", "ppc64", map[string]string{"GOPPC64": "ppc970"}},
		{"linux", "ppc64", map[string]string{"GOPPC64": "power9"}},
		{"linux", "ppc64le", nil},
		{"linux", "ppc64le", map[string]string{"GOPPC64": "power10"}},
		{"linux", "amd64", nil},
		{"linux", "amd64", map[string]string{"GOAMD64": "v3"}},
		{"darwin", "amd64", nil},
		{"linux", "arm64", nil},
		{"linux", "arm64", map[string]string{"GOARM64": "v9.5"}},
		{"linux", "arm", map[string]string{"GOARM": "6"}},
		{"linux", "386", nil},
		{"linux", "mips64", map[string]string{"GOMIPS64": "softfloat"}},
		{"linux", "riscv64", map[string]string{"GORISCV64": "rva23u64"}},
		{"linux", "s390x", nil},
	}
	for _, r := range rows {
		name := fmt.Sprintf("%s_%s/%v", r.goos, r.goarch, r.env)
		t.Run(name, func(t *testing.T) {
			env := map[string]string{"GOOS": r.goos, "GOARCH": r.goarch}
			for k, v := range r.env {
				env[k] = v
			}

			// The real buildcfg, in its own process.
			cmd := exec.Command(dump)
			for _, kv := range os.Environ() {
				switch strings.SplitN(kv, "=", 2)[0] {
				case "GOOS", "GOARCH", "GO386", "GOAMD64", "GOARM", "GOARM64", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOEXPERIMENT", "GOROOT":
				default:
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Env = append(cmd.Env, "GOROOT="+repoGOROOT(t))
			for k, v := range env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("buildcfgdump: %v\n%s", err, out)
			}
			var wantTags []string
			var wantSetting string
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				switch {
				case strings.HasPrefix(line, "tag "):
					wantTags = append(wantTags, strings.TrimPrefix(line, "tag "))
				case strings.HasPrefix(line, "setting "):
					wantSetting = strings.TrimPrefix(line, "setting ")
				}
			}

			// dist, from the same environment, through the real xinit. The
			// host is a G5 for the ppc64 rows' sake: this is the target's
			// view, so the host must not influence it.
			simulateDist(t, "darwin", "ppc64", g5CPU, env)
			gotTags := archTags()
			slices.Sort(gotTags)
			slices.Sort(wantTags)
			if !slices.Equal(gotTags, wantTags) {
				t.Errorf("archTags() = %v, internal/buildcfg ToolTags = %v", gotTags, wantTags)
			}
			if r.goarch == "ppc64" || r.goarch == "ppc64le" {
				if got, want := "GOPPC64="+ppc64Floor(), wantSetting; got != want {
					t.Errorf("dist floor %q, internal/buildcfg setting %q", got, want)
				}
			}
		})
	}
}
