// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "os"

// The ISA floor (GOPPC64 level) of a compilation is a property of the machine
// that will RUN the result. A compilation for the target takes the target's
// floor; a compilation of the build's own tools, which this build then
// executes (toolchain1, go_bootstrap, the host commands), takes the host's.
//
// cmd/dist cannot share this table with internal/buildcfg: make.bash builds
// dist with the bootstrap toolchain against the bootstrap GOROOT's standard
// library, in which no new internal package exists. The table in
// defaultPPC64 therefore mirrors internal/buildcfg.defaultGOPPC64, and
// TestBuildcfgAgreement compares the two.
//
// Nothing here is computed once at startup. Every caller derives the floor
// from the platform being built at that moment.

// Mach-O CPU identifiers, as reported by the Darwin sysctls hw.cputype and
// hw.cpusubtype. Only values that were measured on real hardware are listed;
// the 750 (G3) and 74xx (G4) subtypes have not been measured on any machine
// available to this port and are deliberately absent.
const (
	machCPUTypePowerPC       = 18  // CPU_TYPE_POWERPC, measured on the G5
	machCPUSubtypePowerPC970 = 100 // CPU_SUBTYPE_POWERPC_970, measured on the G5
)

// hostMachCPU reads hw.cputype and hw.cpusubtype of the machine this dist
// binary is running on. It is a variable so that tests can supply a different
// machine. sysHostMachCPU is defined per OS (sys_darwin.go, sys_notdarwin.go).
var hostMachCPU = sysHostMachCPU

// ppc64FloorFromMachCPU maps a measured Mach-O cputype/cpusubtype pair to a
// GOPPC64 level. It reports false for every pair it has no measurement for;
// the caller must then fall back to the table default, never invent a level.
func ppc64FloorFromMachCPU(cputype, cpusubtype uint32) (string, bool) {
	if cputype == machCPUTypePowerPC && cpusubtype == machCPUSubtypePowerPC970 {
		return "ppc970", true
	}
	return "", false
}

// probeHostPPC64Floor measures the host's floor. It reports false when the
// probe fails or when the machine's identifiers are not in the measured table.
// hostPPC64Floor only calls it for a Darwin/ppc64 host.
func probeHostPPC64Floor() (string, bool) {
	cputype, cpusubtype, err := hostMachCPU()
	if err != nil {
		return "", false
	}
	return ppc64FloorFromMachCPU(cputype, cpusubtype)
}

// defaultPPC64 is the table default GOPPC64 for a platform. It mirrors
// internal/buildcfg.defaultGOPPC64.
func defaultPPC64(osName, arch string) string {
	if osName == "darwin" && arch == "ppc64" {
		return "ppc970"
	}
	return "power8"
}

// buildingForHost reports whether dist is currently building for the host
// platform while the target is a different one (cmdbootstrap's go_bootstrap
// and host-commands phases). It is false before cmdbootstrap switches
// platforms, after it switches back, and whenever host and target are the
// same platform.
func buildingForHost() bool {
	return oldgoos != "" && (goos != oldgoos || goarch != oldgoarch)
}

// hostPPC64Floor is the floor for code that runs on the host.
//
// On a Darwin/ppc64 host the machine is measured. A measurement outranks
// $GOPPC64, because in a cross build $GOPPC64 describes the target. If the
// measurement fails or names a CPU that has not been measured, the answer is the
// table default for the host, never the target's floor and never $GOPPC64: an
// unknown machine gets the conservative level, not an invented one.
//
// Everywhere else the host cannot be measured, so an explicit $GOPPC64 is the
// operator's only statement about the machine and keeps applying to the host
// tools, as it always did (this is the operator's responsibility, exactly as for
// the other GO$GOARCH settings). Without one, the table default for the host
// applies.
func hostPPC64Floor() string {
	if gohostos == "darwin" && gohostarch == "ppc64" {
		if floor, ok := probeHostPPC64Floor(); ok {
			return floor
		}
		return defaultPPC64(gohostos, gohostarch)
	}
	if goppc64Env != "" {
		return goppc64Env
	}
	return defaultPPC64(gohostos, gohostarch)
}

// ppc64Floor returns the GOPPC64 level for the platform currently being built
// (goos/goarch). While building for the host in a cross build that is the
// host's floor; otherwise it is the target's: $GOPPC64 if set, else the table
// default for goos/goarch.
func ppc64Floor() string {
	if buildingForHost() {
		return hostPPC64Floor()
	}
	if goppc64Env != "" {
		return goppc64Env
	}
	return defaultPPC64(goos, goarch)
}

// setBuildPlatform makes the build produce code for os/arch and exports the
// settings that follow from the platform, so that the tools dist runs and the
// go command see the same floor that dist's own build rules use.
func setBuildPlatform(osName, arch string) {
	goos = osName
	goarch = arch
	os.Setenv("GOARCH", goarch)
	os.Setenv("GOOS", goos)
	os.Setenv("GOPPC64", ppc64Floor())
}

// switchToHostPlatform is cmdbootstrap's switch from building for the target
// to building for the host. The target is remembered in oldgoos/oldgoarch.
func switchToHostPlatform() {
	oldgoos = goos
	oldgoarch = goarch
	os.Setenv("GOHOSTARCH", gohostarch)
	os.Setenv("GOHOSTOS", gohostos)
	setBuildPlatform(gohostos, gohostarch)
}

// restoreTargetPlatform undoes switchToHostPlatform. The target's floor comes
// back because ppc64Floor derives it from the restored goos/goarch.
func restoreTargetPlatform() {
	setBuildPlatform(oldgoos, oldgoarch)
}

// exportHostPPC64Floor exports the host's floor as $GOPPC64 for tools that
// build for the host while dist's own goos/goarch still name the target:
// toolchain1, which the bootstrap go command builds with GOOS and GOARCH
// emptied. The returned function restores the previous value.
func exportHostPPC64Floor() (restore func()) {
	old, wasSet := os.LookupEnv("GOPPC64")
	os.Setenv("GOPPC64", hostPPC64Floor())
	return func() {
		if wasSet {
			os.Setenv("GOPPC64", old)
		} else {
			os.Unsetenv("GOPPC64")
		}
	}
}

// ppc64AsmDefines returns the assembler arguments that tell cmd/asm which
// POWER ISA level the code being assembled may use.
func ppc64AsmDefines(floor string) []string {
	var args []string
	// We treat each powerpc version as a superset of functionality.
	switch floor {
	case "ppc970":
		args = append(args, "-D", "GOPPC64_ppc970")
	case "power10":
		args = append(args, "-D", "GOPPC64_power10")
		fallthrough
	case "power9":
		args = append(args, "-D", "GOPPC64_power9")
		fallthrough
	default: // This should always be power8.
		args = append(args, "-D", "GOPPC64_power8")
	}
	if floor != "ppc970" {
		args = append(args, "-D", "GOPPC64_vsx")
	}
	return args
}
