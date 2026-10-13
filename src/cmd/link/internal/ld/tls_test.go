// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"fmt"
	"internal/buildcfg"
	"os"
	"strings"
	"testing"
)

// darwinPPCLink builds the darwin/ppc Link that the driver would build.
// linknew insists that buildcfg.GOARCH agree with the architecture, hence the
// temporary override, as in TestAddGotSym.
func darwinPPCLink(t *testing.T) *Link {
	t.Helper()
	origArch := buildcfg.GOARCH
	buildcfg.GOARCH = sys.ArchPPC.Name
	t.Cleanup(func() { buildcfg.GOARCH = origArch })
	return setUpContext(sys.ArchPPC, false, objabi.Hdarwin, "exe", "internal")
}

// TestComputeTLSOffsetDarwinPPC pins the "invalid state" that computeTLSOffset
// records for darwin/ppc, whose TLS mechanism the port has not measured. The
// linker never consumes the value -- relocsym rejects darwin/ppc TLS
// relocations outright, see TestRelocsymRejectsDarwinPPCTLS -- so it only has
// to stay impossible to mistake for a real TLS offset. Real ones are small:
// -PtrSize on ELF, 0x30 on darwin/amd64, 0 on darwin/arm64 and darwin/ppc64.
//
// This is the control for the encoding: the value that used to be written here,
// -1<<63, did not compile at all where int is 32 bits, so it silently broke
// cmd/link/internal/ld for linux/386, linux/arm, linux/mips and for any native
// ppc32 toolchain building cmd/link for GOARCH=ppc (goppc-kig0).
func TestComputeTLSOffsetDarwinPPC(t *testing.T) {
	ctxt := darwinPPCLink(t)
	ctxt.computeTLSOffset()
	if off := ctxt.Tlsoffset; off > -4096 {
		t.Fatalf("darwin/ppc Tlsoffset = %d, want a value far outside the range of real TLS offsets", off)
	}
}

// TestRelocsymRejectsDarwinPPCTLS checks the half of the darwin/ppc TLS story
// that actually refuses work: relocsym must report an error for a TLS
// relocation and must not apply it. Without this, an "invalid" TLS offset would
// not keep the linker honest -- it would be a plausible-looking number that
// nothing rejects.
func TestRelocsymRejectsDarwinPPCTLS(t *testing.T) {
	ctxt := darwinPPCLink(t)

	// relocsym reports through ctxt.ErrorReporter, which writes the message to
	// os.Stderr; this reporter is the test's own so that no package-level
	// 'nerrors' counting or atexit output-file cleanup is involved.
	var nerr int
	er := loader.ErrorReporter{AfterErrorAction: func() { nerr++ }}
	ldr := loader.NewLoader(0, &er)
	st := &relocSymState{
		target: &ctxt.Target,
		ldr:    ldr,
		err:    &ErrorReporter{ErrorReporter: er},
		syms:   &ctxt.ArchSyms,
	}

	tls := ldr.CreateSymForUpdate("main.tls", 0)
	tls.SetType(sym.STLSBSS)

	type relocCase struct {
		name string
		rt   objabi.RelocType
		sym  loader.Sym
		data []byte
	}
	var cases []relocCase
	for i, rt := range []objabi.RelocType{objabi.R_TLS_LE, objabi.R_TLS_IE} {
		sb := ldr.CreateSymForUpdate(fmt.Sprintf("main.reloc%d", i), 0)
		sb.SetType(sym.STEXT)
		sb.SetData(make([]byte, 8))
		rel, _ := sb.AddRel(rt)
		rel.SetSym(tls.Sym())
		rel.SetSiz(4)
		cases = append(cases, relocCase{sb.Name(), rt, sb.Sym(), sb.Data()})
	}

	got := captureStderr(t, func() {
		for _, c := range cases {
			st.relocsym(c.sym, c.data)
		}
	})

	if nerr != len(cases) {
		t.Fatalf("relocsym reported %d errors, want %d: darwin/ppc TLS relocations must be rejected", nerr, len(cases))
	}
	for _, c := range cases {
		want := fmt.Sprintf("%s: darwin/ppc TLS relocation %s for main.tls unsupported", c.name, c.rt)
		if !strings.Contains(got, want) {
			t.Errorf("stderr does not mention %q; got:\n%s", want, got)
		}
		for i, b := range c.data {
			if b != 0 {
				t.Errorf("%s: rejected relocation was applied anyway: byte %d = %#x, want 0", c.name, i, b)
			}
		}
	}
}

// captureStderr runs f with os.Stderr redirected to a file and returns what was
// written. relocsym reports errors with fmt.Fprintf(os.Stderr, ...).
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	orig := os.Stderr
	os.Stderr = file
	f()
	os.Stderr = orig
	out, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
