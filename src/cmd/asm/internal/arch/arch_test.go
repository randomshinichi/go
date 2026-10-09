// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arch

import (
	"cmd/internal/obj/ppc"
	"testing"
)

func TestPPCRegistered(t *testing.T) {
	a := Set("ppc", false)
	if a == nil {
		t.Fatal(`Set("ppc") = nil`)
	}
	if a.LinkArch != &ppc.Linkppc {
		t.Fatalf(`Set("ppc").LinkArch = %p, want ppc.Linkppc`, a.LinkArch)
	}
	if a.Instructions["LWARX"] != ppc.ALWARX || a.Instructions["STWCX."] != ppc.ASTWCX_ {
		t.Fatal("PPC reservation instructions are not registered")
	}
	if _, ok := a.RegisterNumber("R", 31); !ok {
		t.Fatal("PPC assembler did not register R31")
	}
	if _, ok := a.RegisterNumber("CR", 7); !ok || a.Register["LR"] == 0 || a.Register["CTR"] == 0 {
		t.Fatal("PPC assembler did not register the baseline special registers")
	}
	if _, ok := a.RegisterNumber("V", 0); ok {
		t.Fatal("PPC G3 assembler registered a vector register")
	}
	for _, name := range []string{"LDAR", "STDCCC", "ISEL", "LWSYNC"} {
		if _, ok := a.Instructions[name]; ok {
			t.Errorf("PPC G3 assembler registered unsupported instruction %s", name)
		}
	}
}
