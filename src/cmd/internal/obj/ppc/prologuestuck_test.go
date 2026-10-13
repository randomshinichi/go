// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ppc

import (
	"cmd/internal/obj"
	"encoding/binary"
	"fmt"
	"internal/abi"
	"internal/testenv"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// This file is the gate for beads goppc-h890 and goppc-5vcc: at every PC of the
// inserted ppc32 prologue that has no frame yet, the link register must still be
// the CALLER's return address. It gates two different readers of that state.
//
// The unwinder reads LR from the register, not from memory, for the innermost
// frame when the function has not allocated its frame (PCSP delta 0). If LR
// there is the PC itself, runtime.(*unwinder).next sees frame.pc == frame.lr &&
// frame.sp == frame.fp and raises throw("traceback stuck"), unconditionally
// (src/runtime/traceback.go:488-492; not suppressed by unwindSilentErrors). Two
// independent runtime paths start an unwind at an arbitrary interrupted PC:
//
//   - SIGPROF. runtime.sighandler calls sigprof(c.sigpc(), c.sigsp(), c.siglr(),
//     ...) (signal_unix.go:706), which seeds unwinder.initAt with the raw
//     interrupted pc/sp/lr (proc.go:5815-5818). It NEVER asks whether the PC is a
//     safe point: PCDATA_UnsafePoint is not read, so no marking in the pc tables
//     can keep a profiling sample out of any PC. Only the register STATE can.
//   - Asynchronous preemption, whose later GC stack scan unwinds from the
//     preempted PC. Here the marking does matter: isAsyncSafePoint returns false
//     for PCDATA_UnsafePoint == Unsafe (preempt.go:443-449), so an Unsafe PC is
//     never the place a goroutine is stopped.
//
// So a gate that only asserts "no preemptible PC has the bad state" (the first
// version of this file, written for async preemption alone) is satisfied by a
// prologue that merely MARKS its bad PCs Unsafe -- while SIGPROF still walks
// straight into the throw. The checks below therefore assert the STATE (LR is the
// caller's at every frameless PC), reading the marking only for the async check
// and never to excuse the SIGPROF one.
//
// The state is computed by a dataflow over the prologue's control flow, not a
// linear walk. Three facts drive it:
//
//   - A branch-and-link (bl/blrl, DUFF) leaves LR = the following PC.
//   - runtime.morestack is reached by BL and, when it returns, it does so
//     through runtime.gogo: morestack saved R5 as gobuf_lr and its own LR as
//     gobuf_pc, and gogo does MTLR gobuf_lr; MTCTR gobuf_pc; JMP (CTR)
//     (src/runtime/asm_ppc.s, morestack and gogo). The prologue loads R5 from
//     the caller's linkage slot (8(SP)) just before the call, and that slot holds
//     the ENTRY LR. So after the BL to morestack returns, LR is the entry LR
//     again, not the next PC -- but only because of those two facts, which are
//     therefore checked against the real runtime sources (runtimeFacts below) and
//     fed to the model, so a runtime change that breaks them also breaks the gate.
//   - A load of R5 from the slot is the thing that makes the first fact true; any
//     other write to R5 spoils it.

const (
	externalLinkReg = int64(-1) // the caller's return address; never a PC of this function
	garbageLinkReg  = int64(-2) // some value that is neither the caller's nor a PC of this function
)

// The names of the checks, so a rejection says which one rejected.
const (
	checkSigprofState   = "sigprof-state"    // any frameless PC, marking ignored
	checkAsyncState     = "async-state"      // frameless PCs isAsyncSafePoint would accept
	checkNoStackcheckBL = "no-bl-stackcheck" // the BL form must not come back
	checkModelCoverage  = "model-coverage"   // the model must reach every prologue instruction
)

// prologueInsn is one real (byte-bearing) instruction of the prologue, with the
// facts the runtime would observe if it were interrupted just before it.
type prologueInsn struct {
	p         *obj.Prog
	pc        int64   // byte offset from the function's entry, 4 apart
	unsafeVal int64   // PCDATA_UnsafePoint in effect at pc
	spBefore  int64   // SP delta accumulated by earlier instructions
	lr        []int64 // every value LR may hold on arrival, over all paths (sorted)
	r5Slot    bool    // on every path, R5 holds the caller's linkage slot value
	reached   bool
}

// finding is one rejection, naming the check that produced it.
type finding struct {
	check  string
	pc     int64
	insn   string
	detail string
}

func (f finding) String() string {
	return fmt.Sprintf("[%s] pc=%d %s: %s", f.check, f.pc, f.insn, f.detail)
}

func checksOf(fs []finding) []string {
	var names []string
	for _, f := range fs {
		if !slices.Contains(names, f.check) {
			names = append(names, f.check)
		}
	}
	slices.Sort(names)
	return names
}

func findingsOf(fs []finding, check string) []finding {
	var out []finding
	for _, f := range fs {
		if f.check == check {
			out = append(out, f)
		}
	}
	return out
}

func describeFindings(fs []finding) string {
	var b strings.Builder
	for _, f := range fs {
		fmt.Fprintf(&b, "\t%s\n", f)
	}
	return b.String()
}

// runtimeFacts records whether the runtime sources still say what the model
// assumes about how runtime.morestack returns. failed names every anchor that did
// not hold; the model only trusts the return-through-gogo fact when it is empty.
type runtimeFacts struct {
	failed []string
}

func (f runtimeFacts) morestackResumesWithEntryLR() bool { return len(f.failed) == 0 }

var textStart = regexp.MustCompile(`(?m)^TEXT\s`)

// asmBody returns the body of the assembly function whose TEXT line contains
// header, with // comments removed, or "" if there is none.
func asmBody(src, header string) string {
	i := strings.Index(src, "TEXT "+header)
	if i < 0 {
		return ""
	}
	rest := src[i+len("TEXT "):]
	if loc := textStart.FindStringIndex(rest); loc != nil {
		rest = rest[:loc[0]]
	}
	var out []string
	for _, line := range strings.Split(rest, "\n") {
		if c := strings.Index(line, "//"); c >= 0 {
			line = line[:c]
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// goBody returns the text of top-level function name in Go source src.
func goBody(src, name string) string {
	i := strings.Index(src, "\nfunc "+name+"(")
	if i < 0 {
		return ""
	}
	rest := src[i+1:]
	if j := strings.Index(rest[1:], "\nfunc "); j >= 0 {
		rest = rest[:j+1]
	}
	return rest
}

// deriveRuntimeFacts checks the sources the model's "morestack returns through
// gogo with the entry LR" claim rests on. Each anchor is a separate name so a
// failure says which link of the chain broke.
func deriveRuntimeFacts(asm, stackGo, procGo string) runtimeFacts {
	var f runtimeFacts
	fail := func(name string) { f.failed = append(f.failed, name) }

	gogo := asmBody(asm, "runtime·gogo(")
	// gogo: LWZ Rl, gobuf_lr(Rb) ... MTLR Rl ; LWZ Rp, gobuf_pc(Rb) ... MTCTR Rp ; JMP (CTR)
	ok := false
	if m := regexp.MustCompile(`LWZ\s+(R\d+),\s*gobuf_lr\(R\d+\)`).FindStringSubmatch(gogo); m != nil {
		ok = regexp.MustCompile(`MTLR\s+` + m[1] + `\b`).MatchString(gogo)
	}
	if !ok {
		fail("gogo-restores-gobuf_lr")
	}
	ok = false
	if m := regexp.MustCompile(`LWZ\s+(R\d+),\s*gobuf_pc\(R\d+\)`).FindStringSubmatch(gogo); m != nil {
		ok = regexp.MustCompile(`MTCTR\s+`+m[1]+`\b`).MatchString(gogo) &&
			regexp.MustCompile(`JMP\s+\(CTR\)`).MatchString(gogo)
	}
	if !ok {
		fail("gogo-resumes-at-gobuf_pc")
	}

	more := asmBody(asm, "runtime·morestack(")
	// morestack: R5 (the slot value) is stored as gobuf_lr before anything writes R5.
	save := regexp.MustCompile(`STW\s+R5,\s*\(g_sched\+gobuf_lr\)\(R30\)`).FindStringIndex(more)
	if save == nil {
		fail("morestack-saves-R5-as-gobuf_lr")
	} else if regexp.MustCompile(`(?m)^\s*(?:L\w+|ADD\w*|SUB\w*|OR\w*|AND\w*|XOR\w*|NEG|MFLR|MFCTR|MR|SL\w+|SR\w+|RLW\w*|EXTS\w*|MUL\w*)\s+R5\b`).MatchString(more[:save[0]]) {
		// Go ppc32 assembly here is destination-first for loads and ALU ops (and
		// source-first for stores), so a write to R5 has R5 as the first operand.
		fail("morestack-keeps-R5-until-saved")
	}
	// morestack: its own LR (the address after the BL, i.e. the back edge) is gobuf_pc.
	ok = false
	if m := regexp.MustCompile(`MFLR\s+(R\d+)`).FindStringSubmatch(more); m != nil {
		ok = regexp.MustCompile(`STW\s+` + m[1] + `,\s*\(g_sched\+gobuf_pc\)\(R30\)`).MatchString(more)
	}
	if !ok {
		fail("morestack-saves-LR-as-gobuf_pc")
	}
	// morestack_noctxt is the entry the prologue uses when no closure context is
	// live: it must only set R11 and branch, never touch LR.
	noctxt := asmBody(asm, "runtime·morestack_noctxt(")
	if !regexp.MustCompile(`B\s+runtime·morestack\(SB\)`).MatchString(noctxt) ||
		regexp.MustCompile(`\b(MTLR|MFLR|BL)\b`).MatchString(noctxt) {
		fail("morestack_noctxt-leaves-LR-alone")
	}

	// Every exit from newstack resumes the goroutine through gogo(&gp.sched),
	// directly or (preempt/park paths) through the scheduler's execute.
	if strings.Count(goBody(stackGo, "newstack"), "gogo(&gp.sched)") < 2 {
		fail("newstack-resumes-via-gogo")
	}
	if !strings.Contains(goBody(procGo, "execute"), "gogo(&gp.sched)") {
		fail("execute-resumes-via-gogo")
	}
	return f
}

// readRuntimeFacts reads the real runtime sources of the tree under test.
func readRuntimeFacts(t *testing.T) (runtimeFacts, string, string, string) {
	t.Helper()
	dir := filepath.Join(testenv.GOROOT(t), "src", "runtime")
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("the gate anchors its model of morestack's return on %s, which cannot be read: %v", name, err)
		}
		return string(b)
	}
	asm, stackGo, procGo := read("asm_ppc.s"), read("stack.go"), read("proc.go")
	return deriveRuntimeFacts(asm, stackGo, procGo), asm, stackGo, procGo
}

// modelFunction walks a preprocessed Prog list and reports every real
// instruction up to and including the frame-allocating one (the first with a
// nonzero Spadj; its own PC is still frameless), with the state an interrupt
// there would observe. alloc is that instruction's index, or -1 for a function
// with no frame. The unsafe-point value and SP delta are computed the way
// linkpcln builds the tables (PCDATA, FUNCDATA and NOP are zero-width), and LR
// and R5 by a forward dataflow to a fixpoint over the branches, so the guard's
// back edge to its own top is followed rather than assumed.
func modelFunction(text *obj.Prog, facts runtimeFacts) (insns []prologueInsn, alloc int) {
	type raw struct {
		p         *obj.Prog
		pc        int64
		unsafeVal int64
		spBefore  int64
	}
	var all []raw
	index := map[*obj.Prog]int{} // a zero-width prog maps to the next real instruction
	var pending []*obj.Prog
	pc, acc := int64(0), int64(0)
	val := int64(abi.UnsafePointSafe) // linkpcln seeds the table with Safe
	for p := text.Link; p != nil; p = p.Link {
		if isUnsafePointPCDATA(p) {
			val = p.To.Offset
			continue
		}
		if p.As == obj.APCDATA || p.As == obj.AFUNCDATA || p.As == obj.ANOP {
			pending = append(pending, p)
			continue
		}
		for _, q := range pending {
			index[q] = len(all)
		}
		pending = pending[:0]
		index[p] = len(all)
		all = append(all, raw{p: p, pc: pc, unsafeVal: val, spBefore: acc})
		pc += insnSize
		acc += int64(p.Spadj) // the delta applies from the next instruction on
	}
	alloc = -1
	for i, r := range all {
		if r.p.Spadj != 0 {
			alloc = i
			break
		}
	}
	if alloc < 0 {
		return nil, -1
	}

	type state struct {
		lr      map[int64]bool
		r5      bool
		reached bool
	}
	st := make([]state, alloc+1)
	st[0] = state{lr: map[int64]bool{externalLinkReg: true}, r5: false, reached: true}
	work := []int{0}
	for len(work) > 0 {
		i := work[len(work)-1]
		work = work[:len(work)-1]
		r := all[i]
		out := map[int64]bool{}
		for v := range st[i].lr {
			out[v] = true
		}
		outR5 := st[i].r5
		switch {
		case isMorestackCall(r.p):
			// Returns through gogo: LR is the entry LR iff R5 held the slot value
			// when morestack saved it AND the runtime still does what we claim.
			out = map[int64]bool{garbageLinkReg: true}
			if facts.morestackResumesWithEntryLR() && st[i].r5 {
				out = map[int64]bool{externalLinkReg: true}
			}
		case setsLinkRegister(r.p):
			out = map[int64]bool{r.pc + insnSize: true}
		}
		if r.p.As == ALWZ && r.p.To.Type == obj.TYPE_REG && r.p.To.Reg == REG_R5 &&
			r.p.From.Type == obj.TYPE_MEM && r.p.From.Reg == REGSP && r.p.From.Offset == 8 && r.spBefore == 0 {
			outR5 = true
		} else if writesRegister(r.p, REG_R5) {
			outR5 = false
		}
		var succ []int
		target := func() {
			if r.p.To.Type == obj.TYPE_BRANCH && r.p.To.Target() != nil {
				if j, ok := index[r.p.To.Target()]; ok {
					succ = append(succ, j)
				}
			}
		}
		switch r.p.As {
		case AB:
			target()
		case ABC:
			target()
			succ = append(succ, i+1)
		case ABLR, obj.ARET:
		default:
			succ = append(succ, i+1)
		}
		for _, j := range succ {
			if j > alloc {
				continue
			}
			changed := !st[j].reached
			if !st[j].reached {
				st[j] = state{lr: map[int64]bool{}, r5: outR5, reached: true}
			}
			for v := range out {
				if !st[j].lr[v] {
					st[j].lr[v] = true
					changed = true
				}
			}
			if st[j].r5 && !outR5 {
				st[j].r5 = false
				changed = true
			}
			if changed {
				work = append(work, j)
			}
		}
	}
	for i := 0; i <= alloc; i++ {
		in := prologueInsn{p: all[i].p, pc: all[i].pc, unsafeVal: all[i].unsafeVal, spBefore: all[i].spBefore,
			r5Slot: st[i].r5, reached: st[i].reached}
		for v := range st[i].lr {
			in.lr = append(in.lr, v)
		}
		slices.Sort(in.lr)
		insns = append(insns, in)
	}
	return insns, alloc
}

// isMorestackCall reports a branch-and-link to the stack-growth entry points.
func isMorestackCall(p *obj.Prog) bool {
	return p.As == ABL && p.To.Sym != nil &&
		(p.To.Sym.Name == "runtime.morestack" || p.To.Sym.Name == "runtime.morestack_noctxt")
}

// writesRegister reports whether p stores its result in register r. Compares
// carry a register in To but do not write it; stores and branches carry memory or
// a branch target there.
func writesRegister(p *obj.Prog, r int16) bool {
	return p.To.Type == obj.TYPE_REG && p.To.Reg == r && !strings.HasPrefix(insnName(p), "cmp")
}

// setsLinkRegister reports whether executing p leaves LR holding the following
// PC: the forms that branch with the link bit set, which are exactly what
// preprocess counts as "making calls" when it decides whether a function is a
// leaf, plus the DUFF intrinsics, which are calls too. Plain AB/BLR do not touch
// LR.
func setsLinkRegister(p *obj.Prog) bool {
	switch p.As {
	case ABL, obj.ACALL, ABCTRL, ABLRL, obj.ADUFFCOPY, obj.ADUFFZERO:
		return true
	}
	return false
}

// gateFindings applies every check to a preprocessed function and returns each
// rejection with the check that made it.
//
//   - sigprof-state: at EVERY prologue PC, LR must be exactly the caller's. The
//     unsafe-point table is not consulted, because SIGPROF does not consult it.
//   - async-state: the same requirement, at the PCs isAsyncSafePoint would accept
//     (everything not marked Unsafe; Restart1/Restart2 are accepted too).
//   - no-bl-stackcheck: no branch-and-link to runtime.stackcheck anywhere in the
//     function. This is the shape gate: it fails if the call form returns, however
//     its callee returns and however its PCs are marked.
//   - model-coverage: every prologue instruction was reached by the dataflow. An
//     unreached instruction would be one the state checks silently never looked at.
func gateFindings(text *obj.Prog, facts runtimeFacts) []finding {
	var fs []finding
	insns, _ := modelFunction(text, facts)
	for _, in := range insns {
		insn := insnName(in.p)
		if !in.reached {
			fs = append(fs, finding{checkModelCoverage, in.pc, insn, "instruction not reached by the dataflow"})
			continue
		}
		if len(in.lr) == 1 && in.lr[0] == externalLinkReg {
			continue
		}
		detail := fmt.Sprintf("PCSP delta %d (no frame) but LR may be %s, not the caller's return address", in.spBefore, lrSetString(in.lr, in.pc))
		if slices.Contains(in.lr, in.pc) {
			detail += `; pc == lr && sp == fp, so an unwind from here throws "traceback stuck" (traceback.go:488-492)`
		}
		fs = append(fs, finding{checkSigprofState, in.pc, insn, detail})
		if in.unsafeVal != abi.UnsafePointUnsafe {
			fs = append(fs, finding{checkAsyncState, in.pc, insn, fmt.Sprintf("preemptible (unsafe-point %d); %s", in.unsafeVal, detail)})
		}
	}
	pcs := map[*obj.Prog]int64{}
	for _, in := range insns {
		pcs[in.p] = in.pc
	}
	for p := text.Link; p != nil; p = p.Link {
		if (p.As == ABL || p.As == obj.ACALL) && p.To.Sym != nil && p.To.Sym.Name == "runtime.stackcheck" {
			pc, ok := pcs[p]
			if !ok {
				pc = -1
			}
			fs = append(fs, finding{checkNoStackcheckBL, pc, insnName(p), "branch-and-link to runtime.stackcheck; the inline guard replaced it"})
		}
	}
	return fs
}

func lrSetString(set []int64, pc int64) string {
	var parts []string
	for _, v := range set {
		switch {
		case v == externalLinkReg:
			parts = append(parts, "the caller's")
		case v == garbageLinkReg:
			parts = append(parts, "an unrestored value")
		case v == pc:
			parts = append(parts, fmt.Sprintf("this very pc (%d)", v))
		default:
			parts = append(parts, fmt.Sprintf("pc %d of this function", v))
		}
	}
	return strings.Join(parts, " or ")
}

// insnName is the assembler mnemonic, falling back to obj's name for the
// pseudo-ops the encoder has no form for.
func insnName(p *obj.Prog) string {
	if m := Mnemonic(Inst(p.As)); m != "" {
		return m
	}
	return p.As.String()
}

// TestRuntimeResumeFacts anchors the model's one runtime-dependent assumption --
// that runtime.morestack returns to the prologue's back edge with LR restored to
// the entry LR -- on the real sources, and shows each link of that chain can
// break the anchor. A model whose assumption nobody checks would pass for
// whatever the runtime did.
func TestRuntimeResumeFacts(t *testing.T) {
	facts, asm, stackGo, procGo := readRuntimeFacts(t)
	if len(facts.failed) != 0 {
		t.Fatalf("the runtime no longer does what the prologue model assumes about morestack's return; anchors that failed: %v", facts.failed)
	}
	replaceIn := func(src, header, old, repl string) string {
		t.Helper()
		i := strings.Index(src, header)
		if i < 0 {
			t.Fatalf("no %q to mutate", header)
		}
		end := len(src)
		if j := strings.Index(src[i+len(header):], "\nTEXT "); j >= 0 {
			end = i + len(header) + j
		} else if j := strings.Index(src[i+len(header):], "\nfunc "); j >= 0 {
			end = i + len(header) + j
		}
		body := src[i:end]
		if strings.Count(body, old) != 1 {
			t.Fatalf("mutation %q matches %d times in %q, want exactly 1", old, strings.Count(body, old), header)
		}
		return src[:i] + strings.Replace(body, old, repl, 1) + src[end:]
	}
	for _, tc := range []struct {
		name string
		mut  func() (a, s, p string)
		want string
	}{
		{"gogo-never-reloads-LR", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·gogo(", "MTLR R31", "NOP"), stackGo, procGo
		}, "gogo-restores-gobuf_lr"},
		{"gogo-restores-the-wrong-register", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·gogo(", "MTLR R31", "MTLR R4"), stackGo, procGo
		}, "gogo-restores-gobuf_lr"},
		{"gogo-resumes-elsewhere", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·gogo(", "MTCTR R12", "MTCTR R4"), stackGo, procGo
		}, "gogo-resumes-at-gobuf_pc"},
		{"morestack-saves-the-wrong-register-as-gobuf_lr", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·morestack(", "STW R5, (g_sched+gobuf_lr)(R30)", "STW R8, (g_sched+gobuf_lr)(R30)"), stackGo, procGo
		}, "morestack-saves-R5-as-gobuf_lr"},
		{"morestack-clobbers-R5-before-saving-it", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·morestack(", "MFLR R8\n", "MFLR R8\n\tADDI R5, R0, $0\n"), stackGo, procGo
		}, "morestack-keeps-R5-until-saved"},
		{"morestack-saves-the-wrong-pc", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·morestack(", "STW R8, (g_sched+gobuf_pc)(R30)", "STW R7, (g_sched+gobuf_pc)(R30)"), stackGo, procGo
		}, "morestack-saves-LR-as-gobuf_pc"},
		{"noctxt-entry-clobbers-LR", func() (string, string, string) {
			return replaceIn(asm, "TEXT runtime·morestack_noctxt(", "ADDI R11, R0, $0", "MTLR R11"), stackGo, procGo
		}, "morestack_noctxt-leaves-LR-alone"},
		{"newstack-has-a-non-gogo-exit", func() (string, string, string) {
			return asm, replaceIn(stackGo, "func newstack(", "\tgogo(&gp.sched)\n}", "\t\n}"), procGo
		}, "newstack-resumes-via-gogo"},
		{"execute-does-not-use-gogo", func() (string, string, string) {
			return asm, stackGo, replaceIn(procGo, "func execute(", "gogo(&gp.sched)", "")
		}, "execute-resumes-via-gogo"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, s, p := tc.mut()
			got := deriveRuntimeFacts(a, s, p)
			if !slices.Contains(got.failed, tc.want) {
				t.Fatalf("mutation %q: anchors failed = %v, want %q among them", tc.name, got.failed, tc.want)
			}
			if got.morestackResumesWithEntryLR() {
				t.Fatalf("mutation %q still leaves the model trusting morestack's return", tc.name)
			}
			t.Logf("rejected by anchor %v", got.failed)
		})
	}
}

// growableFixture preprocesses the smallest growable non-leaf function.
func growableFixture(t *testing.T, locals int64, needCtxt bool) (*obj.Link, *obj.LSym, *obj.Prog) {
	t.Helper()
	ctxt := obj.Linknew(&Linkppc)
	ctxt.DiagFunc = func(format string, args ...interface{}) { t.Fatalf(format, args...) }
	s := ctxt.Lookup("growable")
	s.Set(obj.AttrNeedCtxt, needCtxt)
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: locals, Val: int32(4)}}
	text.Link = &obj.Prog{As: obj.ARET}
	s.NewFuncInfo().Text = text
	preprocess(ctxt, s, func() *obj.Prog { return new(obj.Prog) })
	return ctxt, s, text
}

// TestPrologueLRIsCallersAtEveryFramelessPC is the gate itself, on the real
// emitter's output: every check in gateFindings must come back empty.
func TestPrologueLRIsCallersAtEveryFramelessPC(t *testing.T) {
	facts, _, _, _ := readRuntimeFacts(t)
	if len(facts.failed) != 0 {
		t.Fatalf("runtime anchors failed: %v", facts.failed)
	}
	for _, tc := range []struct {
		name     string
		locals   int64
		needCtxt bool
	}{
		{"d-form-frame", 4096, false},
		{"indexed-frame", 65512, false},
		{"largest-d-form", 32752 - 16, false},
		{"small-frame", 4, false},
		{"needs-context", 64, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctxt, s, text := growableFixture(t, tc.locals, tc.needCtxt)

			if fs := gateFindings(text, facts); len(fs) > 0 {
				t.Fatalf("%d rejection(s) by %v; an interrupt at these PCs leaves LR not the caller's with no frame, and a stack walk from there throws \"traceback stuck\":\n%s",
					len(fs), checksOf(fs), describeFindings(fs))
			}

			// Non-vacuity. A gate that looks at nothing passes anything.
			insns, alloc := modelFunction(text, facts)
			if alloc < 0 {
				t.Fatalf("no frame-allocating instruction found; the model is not looking at the prologue\n%s", dumpProgs(text))
			}
			morestackAt := -1
			for i, in := range insns {
				if !in.reached {
					t.Fatalf("instruction %d (%s) not reached by the dataflow", i, insnName(in.p))
				}
				if in.spBefore != 0 {
					t.Fatalf("instruction %d (%s) has spBefore=%d inside the frameless region", i, insnName(in.p), in.spBefore)
				}
				if isMorestackCall(in.p) {
					morestackAt = i
				}
			}
			if len(insns) < 10 {
				t.Fatalf("only %d frameless prologue instructions modelled; expected the whole guard (>= 10)", len(insns))
			}
			if morestackAt < 0 || morestackAt+1 >= len(insns) {
				t.Fatalf("no BL runtime.morestack inside the frameless region: the grow path (and the gogo assumption) was never exercised by this gate")
			}
			if !insns[morestackAt].r5Slot {
				t.Fatalf("R5 does not hold the linkage-slot value at the BL runtime.morestack; gobuf_lr would be garbage")
			}
			if after := insns[morestackAt+1]; len(after.lr) != 1 || after.lr[0] != externalLinkReg {
				t.Fatalf("the instruction after the morestack call has LR %v, want only the caller's", after.lr)
			}
			// SIGPROF reads every one of these PCs; the async consumer only the
			// unmarked ones. Both populations must be non-empty or one half of the
			// gate is vacuous.
			samplable, preemptible := len(insns), 0
			for _, in := range insns {
				if in.unsafeVal != abi.UnsafePointUnsafe {
					preemptible++
				}
			}
			if preemptible == 0 || preemptible == samplable {
				t.Fatalf("%d of %d frameless PCs are preemptible; want some but not all, or the async half and the SIGPROF half are indistinguishable", preemptible, samplable)
			}

			// The modelled PCs must be the emitted ones, or the reasoning is about
			// a different program than the one that ships.
			assemble(ctxt, s, nil)
			for i, in := range insns {
				if in.p.Pc != in.pc {
					t.Fatalf("instruction %d modelled at pc %d but assembled at pc %d -- the model no longer matches the emitted layout", i, in.pc, in.p.Pc)
				}
			}
			// lwz r4,8(r30) = 32<<26 | 4<<21 | 30<<16 | 8 = 0x809e0008: the
			// g->stackguard0 read that replaced the call, exactly once.
			const guardWord = uint32(0x809e0008)
			hits := 0
			for i := 0; i+4 <= len(s.P); i += 4 {
				if binary.BigEndian.Uint32(s.P[i:]) == guardWord {
					hits++
				}
			}
			if hits != 1 {
				t.Fatalf("assembled function contains %d copies of the inline guard load 0x809e0008, want exactly 1", hits)
			}
		})
	}
}

// preFixSplitPrologue hand-builds the prologue as this package emitted it before
// the guard was inlined: Restart1 over MFLR/STW, then the frame-size constant,
// BL runtime.stackcheck, CMPWI, BEQ, the grow path (slot reload, BL
// runtime.morestack_noctxt, back edge) and the allocation. With markedThroughAlloc
// the guard run additionally carries PCDATA Unsafe from after the R0 zeroing to
// just after the allocation, which is what the marking-only fix did.
func preFixSplitPrologue(ctxt *obj.Link, markedThroughAlloc bool) *obj.Prog {
	mkr := func(val int64) *obj.Prog {
		return &obj.Prog{As: obj.APCDATA,
			From: obj.Addr{Type: obj.TYPE_CONST, Offset: abi.PCDATA_UnsafePoint},
			To:   obj.Addr{Type: obj.TYPE_CONST, Offset: val}}
	}
	reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
	at := func(off int64) obj.Addr { return obj.Addr{Type: obj.TYPE_MEM, Reg: REGSP, Offset: off} }
	call := func(sym string) obj.Addr {
		return obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup(sym)}
	}
	text := &obj.Prog{As: obj.ATEXT, To: obj.Addr{Type: obj.TYPE_TEXTSIZE, Offset: 128, Val: int32(4)}}
	recheck := &obj.Prog{As: obj.ANOP}
	done := &obj.Prog{As: obj.ANOP}
	beq := &obj.Prog{As: ABC, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 12}, Reg: 2, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	back := &obj.Prog{As: AB, From: obj.Addr{Type: obj.TYPE_CONST}, To: obj.Addr{Type: obj.TYPE_BRANCH}}
	back.To.SetTarget(recheck)
	beq.To.SetTarget(done)
	seq := []*obj.Prog{
		mkr(abi.UnsafePointRestart1),
		{As: AMFLR, To: reg(REG_R0)},
		{As: ASTW, From: reg(REG_R0), To: at(8)},
		mkr(abi.UnsafePointSafe),
		{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST}, Reg: REG_R0, To: reg(REG_R0)},
	}
	if markedThroughAlloc {
		seq = append(seq, mkr(abi.UnsafePointUnsafe))
	}
	seq = append(seq,
		recheck,
		&obj.Prog{As: AADDI, From: obj.Addr{Type: obj.TYPE_CONST, Offset: 128}, Reg: REG_R0, To: reg(REG_R3)},
		&obj.Prog{As: ABL, To: call("runtime.stackcheck")},
		&obj.Prog{As: ACMPWI, From: reg(REG_R3), To: obj.Addr{Type: obj.TYPE_CONST}},
		beq,
		&obj.Prog{As: ALWZ, From: at(8), To: reg(REG_R5)},
		&obj.Prog{As: ABL, To: call("runtime.morestack_noctxt")},
		back,
		done,
		&obj.Prog{As: ASTWU, From: reg(REGSP), To: at(-128), Spadj: 128},
	)
	if markedThroughAlloc {
		seq = append(seq, mkr(abi.UnsafePointSafe))
	}
	for i := 0; i+1 < len(seq); i++ {
		seq[i].Link = seq[i+1]
	}
	text.Link = seq[0]
	return text
}

func unlinkProg(text *obj.Prog, want func(*obj.Prog) bool) bool {
	for prev := text; prev.Link != nil; prev = prev.Link {
		if want(prev.Link) {
			prev.Link = prev.Link.Link
			return true
		}
	}
	return false
}

// TestPrologueGateRejectsCounterfeits runs the gate against prologues that are
// wrong in a named way and requires that each is rejected by the check(s) named
// here -- not merely rejected. The two halves of the gate must be separable: the
// marking-only counterfeit passes the async check by construction and has to be
// caught by the SIGPROF one, while the pre-fix shape has to fail the async one as
// well.
func TestPrologueGateRejectsCounterfeits(t *testing.T) {
	real, asm, stackGo, procGo := readRuntimeFacts(t)
	if len(real.failed) != 0 {
		t.Fatalf("runtime anchors failed: %v", real.failed)
	}
	firstInsn := func(fs []finding, name string) (finding, bool) {
		for _, f := range fs {
			if f.insn == name {
				return f, true
			}
		}
		return finding{}, false
	}
	type counterfeit struct {
		name   string
		build  func(t *testing.T) (*obj.Prog, runtimeFacts)
		want   []string // exactly the checks that must reject it, sorted
		verify func(t *testing.T, fs []finding)
	}
	realFunc := func(t *testing.T) (*obj.Link, *obj.Prog) {
		ctxt, _, text := growableFixture(t, 4096, false)
		return ctxt, text
	}
	for _, tc := range []counterfeit{
		{
			// The shape the hardware crashed on (archive/zip, pc=0x11d30bc):
			// Restart1 over MFLR/STW and nothing else, BL runtime.stackcheck. Both
			// consumers reach the CMPWI after the BL, and it is preemptible.
			name: "async-and-sigprof: BL stackcheck, Restart1 window only (the pre-fix prologue)",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				return preFixSplitPrologue(obj.Linknew(&Linkppc), false), real
			},
			want: []string{checkAsyncState, checkNoStackcheckBL, checkSigprofState},
			verify: func(t *testing.T, fs []finding) {
				f, ok := firstInsn(findingsOf(fs, checkAsyncState), "cmpwi")
				if !ok || !strings.Contains(f.detail, "pc == lr") || !strings.Contains(f.detail, "PCSP delta 0") {
					t.Fatalf("async-state did not name the CMPWI after the BL as pc == lr with PCSP delta 0: %v", findingsOf(fs, checkAsyncState))
				}
			},
		},
		{
			// What a marking-only fix produces: the same BL stackcheck, but every
			// bad PC marked Unsafe, through the allocation. isAsyncSafePoint
			// refuses all of them, so an async-only gate (the first version of
			// this file, and 1ye4's stackcheckunsafe_test.go) is satisfied. SIGPROF
			// still walks into the CMPWI.
			name: "sigprof-only: BL stackcheck with the whole run marked Unsafe (marking-only fix)",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				return preFixSplitPrologue(obj.Linknew(&Linkppc), true), real
			},
			want: []string{checkNoStackcheckBL, checkSigprofState},
			verify: func(t *testing.T, fs []finding) {
				if a := findingsOf(fs, checkAsyncState); len(a) != 0 {
					t.Fatalf("the counterfeit was meant to satisfy the async half, but async-state rejects it: %v", a)
				}
				f, ok := firstInsn(findingsOf(fs, checkSigprofState), "cmpwi")
				if !ok || !strings.Contains(f.detail, "pc == lr") || !strings.Contains(f.detail, "PCSP delta 0") {
					t.Fatalf("sigprof-state did not name the CMPWI after the BL as pc == lr with PCSP delta 0: %v", findingsOf(fs, checkSigprofState))
				}
			},
		},
		{
			name: "BL stackcheck re-inserted into the real inline prologue",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				ctxt, text := realFunc(t)
				for p := text; p.Link != nil; p = p.Link {
					if isUnsafePointPCDATA(p) && p.To.Offset == abi.UnsafePointUnsafe {
						bl := &obj.Prog{As: ABL, To: obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("runtime.stackcheck")}, Link: p.Link}
						p.Link = bl
						return text, real
					}
				}
				t.Fatal("no Unsafe marker in the real prologue")
				return nil, real
			},
			want: []string{checkAsyncState, checkNoStackcheckBL, checkSigprofState},
			verify: func(t *testing.T, fs []finding) {
				if s, a := len(findingsOf(fs, checkSigprofState)), len(findingsOf(fs, checkAsyncState)); s <= a {
					t.Fatalf("sigprof-state (%d) must see PCs async-state (%d) is excused from by the Unsafe window", s, a)
				}
			},
		},
		{
			// A pure state defect: nothing about the marking or any symbol is wrong;
			// the grow path simply does not reload the linkage slot into R5, so
			// morestack records garbage as gobuf_lr and gogo resumes with it.
			name: "grow path without the LWZ R5,8(R1) reload",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				_, text := realFunc(t)
				if !unlinkProg(text, func(p *obj.Prog) bool { return p.As == ALWZ && p.To.Reg == REG_R5 }) {
					t.Fatal("no R5 reload in the real prologue")
				}
				return text, real
			},
			want: []string{checkAsyncState, checkSigprofState},
			verify: func(t *testing.T, fs []finding) {
				if _, ok := firstInsn(findingsOf(fs, checkSigprofState), "b"); !ok {
					t.Fatalf("sigprof-state did not reject the back edge after the morestack return: %v", findingsOf(fs, checkSigprofState))
				}
				if _, ok := firstInsn(findingsOf(fs, checkAsyncState), "b"); ok {
					t.Fatalf("the back edge is inside the Unsafe window, async-state should not see it: %v", findingsOf(fs, checkAsyncState))
				}
			},
		},
		{
			// The same prologue, correct as emitted, but the runtime stops restoring
			// LR in gogo. The model must notice because it reads the runtime.
			name: "runtime change: gogo no longer reloads LR",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				_, text := realFunc(t)
				i := strings.Index(asm, "TEXT runtime·gogo(")
				j := i + strings.Index(asm[i:], "MTLR R31")
				if i < 0 || j < i {
					t.Fatal("no MTLR R31 in gogo")
				}
				mutated := asm[:j] + "NOP" + asm[j+len("MTLR R31"):]
				facts := deriveRuntimeFacts(mutated, stackGo, procGo)
				if !slices.Contains(facts.failed, "gogo-restores-gobuf_lr") {
					t.Fatalf("anchor did not notice the mutation: %v", facts.failed)
				}
				return text, facts
			},
			want: []string{checkAsyncState, checkSigprofState},
			verify: func(t *testing.T, fs []finding) {
				if _, ok := firstInsn(findingsOf(fs, checkSigprofState), "b"); !ok {
					t.Fatalf("sigprof-state did not reject the back edge: %v", findingsOf(fs, checkSigprofState))
				}
			},
		},
		{
			// Control that must ACCEPT: strip every unsafe-point marker from the real
			// prologue. If the gate were really reading the marking, this would be
			// rejected or the unmarked arm of the real test would have been; it is
			// accepted because the inline form is correct by STATE, with no BL to
			// leave a stale LR. (The Unsafe window is still emitted, for issue
			// #35470 -- see TestGuardRegionIsUnsafeForDoubleGrowth -- but this gate
			// does not depend on it.)
			name: "control (accepted): real inline prologue with every marker removed",
			build: func(t *testing.T) (*obj.Prog, runtimeFacts) {
				_, text := realFunc(t)
				for unlinkProg(text, isUnsafePointPCDATA) {
				}
				return text, real
			},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text, facts := tc.build(t)
			fs := gateFindings(text, facts)
			got := checksOf(fs)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("rejected by %v, want exactly %v\n%s", got, tc.want, describeFindings(fs))
			}
			if tc.verify != nil {
				tc.verify(t, fs)
			}
			for _, f := range fs {
				t.Logf("%s", f)
			}
		})
	}
}

// TestGuardRegionIsUnsafeForDoubleGrowth pins the marking, for its own reason and
// not as part of the SIGPROF argument (it plays no part there): a preemption
// request taken inside the guard is cleared, yet the guard still goes on to call
// morestack, which grows the stack a second time (upstream issue #35470;
// obj/ppc64/obj9.go marks the same run). The window must open after the
// MFLR/STW restart window, cover the whole guard including the grow call and back
// edge, and close before the allocation, so the allocation PC itself and the body
// stay preemptible.
func TestGuardRegionIsUnsafeForDoubleGrowth(t *testing.T) {
	facts, _, _, _ := readRuntimeFacts(t)
	for _, locals := range []int64{24, 65512} {
		_, _, text := growableFixture(t, locals, false)
		insns, alloc := modelFunction(text, facts)
		if alloc < 4 {
			t.Fatalf("locals=%d: allocation at index %d; the prologue is not the shape this test reads", locals, alloc)
		}
		// MFLR, STW (Restart1), the R0 zeroing (Safe), then the Unsafe run.
		if insns[0].p.As != AMFLR || insns[1].p.As != ASTW {
			t.Fatalf("locals=%d: prologue starts %s, %s, want MFLR, STW", locals, insnName(insns[0].p), insnName(insns[1].p))
		}
		for i := 0; i < 2; i++ {
			if insns[i].unsafeVal != abi.UnsafePointRestart1 {
				t.Errorf("locals=%d: %s at pc %d has unsafe-point %d, want Restart1", locals, insnName(insns[i].p), insns[i].pc, insns[i].unsafeVal)
			}
		}
		if insns[2].unsafeVal != abi.UnsafePointSafe {
			t.Errorf("locals=%d: the R0 zeroing at pc %d has unsafe-point %d, want Safe", locals, insns[2].pc, insns[2].unsafeVal)
		}
		end := 3
		for end <= alloc && insns[end].unsafeVal == abi.UnsafePointUnsafe {
			end++
		}
		if end-3 < 8 {
			t.Errorf("locals=%d: only %d Unsafe instructions after the R0 zeroing, want the whole guard (>= 8)", locals, end-3)
		}
		for i := 3; i < end; i++ {
			if isMorestackCall(insns[i].p) {
				goto covered
			}
		}
		t.Errorf("locals=%d: the BL runtime.morestack is not inside the Unsafe run", locals)
	covered:
		for i := end; i <= alloc; i++ {
			if insns[i].unsafeVal != abi.UnsafePointSafe {
				t.Errorf("locals=%d: %s at pc %d after the guard has unsafe-point %d, want Safe restored before the allocation", locals, insnName(insns[i].p), insns[i].pc, insns[i].unsafeVal)
			}
		}
	}
}

// stackGuardLoad reports whether p is the inline prologue guard's load of
// g.stackguard0 -- `lwz r4, 2*ptrSize(r30)`. The tests in this package that pin
// the prologue shape anchor on it rather than on a call, because the call is the
// defect (bead goppc-h890): `BL runtime.stackcheck` left the very next PC with
// pc == lr and no frame, which is the unconditional throw("traceback stuck").
func stackGuardLoad(p *obj.Prog, ptrSize int64) bool {
	return p.As == ALWZ && p.From.Type == obj.TYPE_MEM && p.From.Reg == REGG &&
		p.From.Offset == 2*ptrSize && p.To.Type == obj.TYPE_REG && p.To.Reg == REG_R4
}

// enoughStackBranch reports the guard's "enough stack, skip the grow path"
// branch: a branch conditional on the unsigned-compare LT bit being clear
// (BO=4 tests the condition negated, BI=0 is bit 0 of CR0).
func enoughStackBranch(p *obj.Prog) bool {
	return p.As == ABC && p.From.Type == obj.TYPE_CONST && p.From.Offset == 4 &&
		p.Reg == 0 && p.To.Type == obj.TYPE_BRANCH && p.To.Target() != nil
}

// isZeroWidthProg reports whether p occupies no bytes, so that a branch whose
// target prog is one of these lands on the next real instruction. PCDATA and
// FUNCDATA pseudo-ops are emitted with no width by both the assembler and
// linkpcln (any pseudo-prog sharing its successor's PC is dropped), and the
// inserted prologue uses ANOP as a branch anchor for exactly that reason.
func isZeroWidthProg(p *obj.Prog) bool {
	return p.As == obj.ANOP || p.As == obj.APCDATA || p.As == obj.AFUNCDATA
}
