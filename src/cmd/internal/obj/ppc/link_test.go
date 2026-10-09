// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"encoding/binary"
	"reflect"
	"testing"
)

// Words independently assembled with clang --target=powerpc-unknown-linux-gnu:
// b .+12; lis 3,(target+12)@ha; addi 3,3,(target+12)@l; bl callee+4; blr.
// A local branch crosses the expanded address pair, so a one-word PC estimate
// would produce the wrong displacement even if the pair itself were correct.
func TestAssembleSymbolAddressAndCall(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	target := ctxt.Lookup("target")
	callee := ctxt.Lookup("callee")
	text := &obj.Prog{As: obj.ATEXT}
	branch := &obj.Prog{As: AB, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	addr := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_EXTERN, Sym: target, Offset: 12}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	call := &obj.Prog{As: obj.ACALL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: callee, Offset: 4}}
	ret := &obj.Prog{As: ABLR}
	text.Link, branch.Link, addr.Link, call.Link = branch, addr, call, ret
	branch.To.SetTarget(call)
	s := ctxt.Lookup("fixture")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	want := []uint32{0x4800000c, 0x3c600000, 0x38630000, 0x48000001, 0x4e800020}
	if s.Size != int64(4*len(want)) {
		t.Fatalf("text size %d, want %d", s.Size, 4*len(want))
	}
	for i, w := range want {
		if got := binary.BigEndian.Uint32(s.P[4*i:]); got != w {
			t.Errorf("word %d: got %08x, clang %08x", i, got, w)
		}
	}
	wantRelocs := []obj.Reloc{
		{Off: 4, Siz: 8, Type: objabi.R_ADDRPOWER, Sym: target, Add: 12},
		{Off: 12, Siz: 4, Type: objabi.R_CALLPOWER, Sym: callee, Add: 4},
	}
	if !reflect.DeepEqual(s.R, wantRelocs) {
		t.Errorf("relocations = %+v, want %+v", s.R, wantRelocs)
	}
}

// clang: cmpw 3,4; cmplw 3,4; cmpwi 3,-1; cmplwi 3,65535.
func TestAssembleSymbolJMP(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	target := ctxt.Lookup("target")
	text := &obj.Prog{As: obj.ATEXT}
	jump := &obj.Prog{As: obj.AJMP, From: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: target}}
	text.Link = jump
	s := ctxt.Lookup("fixture")
	s.NewFuncInfo().Text = text
	assemble(ctxt, s, nil)
	if s.Size != 4 || binary.BigEndian.Uint32(s.P) != 0x48000000 {
		t.Fatalf("JMP encoding = % x, want 48 00 00 00 (b target)", s.P)
	}
	want := []obj.Reloc{{Off: 0, Siz: 4, Type: objabi.R_CALLPOWER, Sym: target}}
	if !reflect.DeepEqual(s.R, want) {
		t.Fatalf("JMP relocations = %+v, want %+v", s.R, want)
	}
}

func TestAssembleRegisterTargetJMP(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		target int16
		want   uint32
	}{
		{REG_CTR, 0x4e800420},
		{REG_LR, 0x4e800020},
	} {
		p := &obj.Prog{As: obj.AJMP, From: obj.Addr{Type: obj.TYPE_REG, Reg: tc.target}}
		words, ok := encodeProg(ctxt, p, 0)
		if !ok || len(words) != 1 || words[0] != tc.want {
			t.Errorf("%v = %08x, want %08x", p, words, tc.want)
		}
	}
}

func TestEncodeIndirectBranchForms(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		as   obj.As
		want uint32
	}{
		{ABLR, 0x4e800020},
		{ABLRL, 0x4e800021},
		{ABCTR, 0x4e800420},
		{ABCTRL, 0x4e800421},
		{ABCLR, 0x4e000020},
		{ABCLRL, 0x4e000021},
		{ABCCTR, 0x4e000420},
		{ABCCTRL, 0x4e000421},
	} {
		p := &obj.Prog{As: tc.as, From: obj.Addr{Offset: 16}, Reg: REG_R0, To: obj.Addr{Offset: 0}}
		words, ok := encodeProg(ctxt, p, 0)
		if !ok || len(words) != 1 || words[0] != tc.want {
			t.Errorf("%v = %08x, want %08x", p, words, tc.want)
		}
	}
}

func TestEncodeCompareProgs(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	for _, tc := range []struct {
		as   obj.As
		to   obj.Addr
		word uint32
	}{
		{ACMPW, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R4}, 0x7c032000},
		{ACMPLW, obj.Addr{Type: obj.TYPE_REG, Reg: REG_R4}, 0x7c032040},
		{ACMPWI, obj.Addr{Type: obj.TYPE_CONST, Offset: -1}, 0x2c03ffff},
		{ACMPLWI, obj.Addr{Type: obj.TYPE_CONST, Offset: 65535}, 0x2803ffff},
	} {
		p := &obj.Prog{As: tc.as, From: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}, To: tc.to}
		words, ok := encodeProg(ctxt, p, 0)
		if !ok || len(words) != 1 || words[0] != tc.word {
			t.Errorf("%v = %08x, want clang %08x", p, words, tc.word)
		}
	}
}

func TestRejectUnsupportedSymbolAddress(t *testing.T) {
	ctxt := obj.Linknew(&Linkppc)
	diagnosed := false
	ctxt.DiagFunc = func(string, ...interface{}) { diagnosed = true }
	p := &obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_GOTREF, Sym: ctxt.Lookup("target")}, To: obj.Addr{Type: obj.TYPE_REG, Reg: REG_R3}}
	if _, ok := encodeProg(ctxt, p, 0); ok || !diagnosed {
		t.Fatal("GOT address silently accepted without a supported relocation path")
	}
}
