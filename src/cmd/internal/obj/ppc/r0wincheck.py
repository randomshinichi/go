#!/usr/bin/env python3
"""r0wincheck.py — goppc-lau1 gate: the ppc32 MFLR/STW prologue Restart1 window.

Reads `go tool compile -S` output for GOARCH=ppc (ppc32). That listing is emitted
after Preprocess+Assemble, so it carries the real instruction PCs *and* the PCDATA
pseudo-instructions -- exactly the two facts this gate needs.

THE BUG (bead goppc-lye2). A non-leaf ppc32 prologue is

        MFLR  R0            ; LR -> R0
        STW   R0, 8(SP)     ; R0 -> caller's linkage area
        ADDI  $0, R0, R0    ; R0 = 0

R0 holds the caller's return address across the STW. runtime.asyncPreempt's
generated ppc32 prologue zeroes R0, and genPPC32 saves only R3..R29, so an async
preemption landing on the STW stores 0 as the return address.

THE FIX (bead goppc-lau1). Mark MFLR..STW as PCDATA_UnsafePoint = Restart1, so
isAsyncSafePoint backs the resumption PC off to the start of the run
(runtime/preempt.go:485-491). That backing-off throws "bad restart PC" unless
startpc <= pc and pc-startpc <= 20, so the window must be <= 20 bytes wide.

ASSERTIONS. For EVERY function whose prologue is `MFLR Rt ; STW Rt,8(SP|R1)` with
Rt == R0:

  MARKED        the unsafe-point value in effect at the MFLR is Restart1/Restart2;
  RIGHT-START   that value's run starts AT the MFLR's own PC (so a preemption on
                either instruction resumes by re-executing the MFLR);
  BOTH          the same value and run still cover the STW (a window starting at
                the STW would leave the MFLR preemptible with a dead R0);
  CLOSED        the first real instruction AFTER the STW is no longer under that
                value and run -- the window did not leak;
  WIDTH         window width (end PC - startpc) <= 20 bytes.
                Width and closure are reported independently: an over-wide window
                that also leaks must fire BOTH assertions, or the leak would hide
                the condition the runtime turns into throw("bad restart PC").

A prologue linking through a register other than R0 is counted separately as
already-safe by construction, so an R0->R31 change elsewhere is not charged to
this gate.

Exit 0 only if matching prologues exist and there are zero violations. A census
that matched nothing FAILS: an instrument that examined nothing is not evidence.

Usage: r0wincheck.py [--verbose] [--label NAME] FILE [FILE...]
"""
import re
import sys

INSN = re.compile(r"^\s+0x([0-9a-f]+)\s+(\d+)\s+\(([^)]*)\)\t([A-Za-z0-9_.]+)\t?(.*)$")
FUNC = re.compile(r"^(\S+) (STEXT|TEXT)\b")

SAFE, UNSAFE = -1, -2
RESTART1, RESTART2 = -3, -4
UNSAFE_POINT_PCDATA_INDEX = 0  # abi.PCDATA_UnsafePoint
MAX_WINDOW_BYTES = 20          # runtime/preempt.go: pc-startpc > 20 => throw

# The prologue pair as the ppc32 backend prints it. mem(8) renders the base
# register as SP, but accept R1 in case the rendering changes.
STW_LINK_SLOT = re.compile(r"^R0,\s*8\((SP|R1)\)$")

KINDS = ["unmarked", "unsafe_inert", "wrong_start", "not_both", "not_closed",
         "too_wide", "open_at_func_end", "nonzero_link_reg"]


def main() -> int:
    args = []
    skip = False
    for a in sys.argv[1:]:
        if skip:
            skip = False
            continue
        if a == "--label":
            skip = True
            continue
        if a.startswith("--"):
            continue
        args.append(a)
    verbose = "--verbose" in sys.argv
    label = ""
    if "--label" in sys.argv:
        label = sys.argv[sys.argv.index("--label") + 1]

    total = 0                      # prologues matching MFLR R0 ; STW R0,8(SP)
    covered = 0                     # clean: MARKED+RIGHT-START+BOTH+CLOSED+WIDTH
    nonzero_link_reg = 0            # MFLR Rt;STW Rt,8(SP) with Rt != R0 -> already safe
    max_width = 0
    max_disp = 0
    counts = dict.fromkeys(KINDS, 0)
    examples = []

    def note(kind, fn, *a):
        counts[kind] += 1
        if len(examples) < 25:
            examples.append((kind, label, fn) + tuple(str(x) for x in a))

    def close_open(w, cur_pc, cur_val, cur_val_start, after_end):
        """Track the restart run past the STW and evaluate it when it truly ends.

        Width and closure are measured INDEPENDENTLY: a window that both leaks and
        exceeds 20 bytes must report both violations, otherwise the leak hides the
        over-wide range -- and the over-wide range is the one the runtime turns
        into throw("bad restart PC"). cur_val/cur_val_start are the unsafe-point
        state in effect AT this instruction. (Testing the window's *stored* value
        instead would be vacuously true -- it is RESTART1 by construction.)"""
        nonlocal covered, max_width
        if w is None:
            return None
        if cur_val in (RESTART1, RESTART2) and cur_val_start == w["mflr_pc"]:
            # Still inside the same run: keep measuring.
            w["run_len"] += 1
            if not after_end:
                return w
            # Run reaches the end of the function: measure, then report.
            width = (w["stw_pc"] + w["run_len"] * 4) - w["mflr_pc"]
            if width > max_width:
                max_width = width
            note("open_at_func_end", w["fn"], "mflr=%s" % hex(w["mflr_pc"]),
                 "width=%d" % width)
            if width > MAX_WINDOW_BYTES:
                note("too_wide", w["fn"], "mflr=%s" % hex(w["mflr_pc"]), "width=%d" % width)
            return None
        width = cur_pc - w["mflr_pc"]
        if width > max_width:
            max_width = width
        leaked = w["run_len"] > 2
        if leaked:
            note("not_closed", w["fn"], "mflr=%s" % hex(w["mflr_pc"]),
                 "insn_in_run=%d" % w["run_len"], "width=%d" % width)
        if width > MAX_WINDOW_BYTES:
            note("too_wide", w["fn"], "mflr=%s" % hex(w["mflr_pc"]), "width=%d" % width)
        if not leaked and width <= MAX_WINDOW_BYTES:
            covered += 1
        return None

    for path in args:
        fn = "?"
        val, val_start = SAFE, None      # unsafe-point value in effect, and its run start
        prev = None                      # (pc, mnem, ops, val, val_start) previous real insn
        pending_mflr = None              # pc of an MFLR R0 awaiting a matching STW
        w = None                         # open window awaiting closure

        with open(path, errors="replace") as fh:
            for ln in fh:
                ln = ln.rstrip("\n")
                m = FUNC.match(ln)
                if m:
                    w = close_open(w, 0, SAFE, None, after_end=True)
                    fn = m.group(1)
                    val, val_start = SAFE, None
                    prev, pending_mflr, w = None, None, None
                    continue
                m = INSN.match(ln)
                if not m:
                    continue
                pc = int(m.group(1), 16)
                mn = m.group(4)
                ops = m.group(5).strip()

                if mn == "PCDATA":
                    parts = [p.strip() for p in ops.split(",")]
                    if len(parts) == 2 and parts[0] == "$%d" % UNSAFE_POINT_PCDATA_INDEX:
                        v = int(parts[1].lstrip("$"))
                        if v != val:
                            val, val_start = v, pc
                    continue

                # ---------- a real instruction ----------
                w = close_open(w, pc, val, val_start, after_end=False)

                if mn == "STW" and prev is not None and prev[1] == "MFLR":
                    if STW_LINK_SLOT.match(ops) and prev[2] == "R0":
                        total += 1
                        mflr_pc = prev[0]
                        v0, s0 = prev[3], prev[4]   # at the MFLR
                        v1, s1 = val, val_start     # at the STW
                        # The listing's PCs are function-relative, so the entry
                        # MFLR is at PC 0: compare unconditionally (a truthiness
                        # test on mflr_pc would silently drop every entry window).
                        if (pc - mflr_pc) > max_disp:
                            max_disp = pc - mflr_pc
                        if v0 == SAFE:
                            note("unmarked", fn, "mflr=%s" % hex(mflr_pc))
                        elif v0 == UNSAFE:
                            note("unsafe_inert", fn, "mflr=%s" % hex(mflr_pc))
                        elif v0 not in (RESTART1, RESTART2):
                            note("unmarked", fn, "mflr=%s" % hex(mflr_pc), "val=%d" % v0)
                        elif s0 != mflr_pc:
                            note("wrong_start", fn, "mflr=%s" % hex(mflr_pc),
                                 "run_start=%s" % hex(s0 or 0))
                        elif v1 != v0 or s1 != mflr_pc:
                            note("not_both", fn, "mflr=%s" % hex(mflr_pc),
                                 "stw_val=%d" % v1, "stw_start=%s" % hex(s1 or 0))
                        else:
                            w = dict(fn=fn, mflr_pc=mflr_pc, stw_pc=pc, val=v0,
                                     val_start=s0, run_len=2)
                    elif prev[2] != "R0":
                        # e.g. the rejected R0->R31 variant: safe by construction.
                        nonzero_link_reg += 1
                elif mn == "MFLR":
                    pending_mflr = pc if ops == "R0" else None

                prev = (pc, mn, ops, val, val_start)

            w = close_open(w, 0, SAFE, None, after_end=True)

    print("label                 %s" % (label or "-"))
    print("files                 %d" % len(args))
    print("prologues_MFLR_R0_STW %d" % total)
    print("COVERED_ok            %d" % covered)
    print("nonzero_link_reg      %d" % nonzero_link_reg)
    for k in KINDS:
        if counts[k]:
            print("VIOL_%-18s %d" % (k, counts[k]))
    print("max_window_width      %d bytes (limit %d)" % (max_width, MAX_WINDOW_BYTES))
    print("max_pc_minus_startpc  %d bytes (limit %d)" % (max_disp, MAX_WINDOW_BYTES))
    if verbose:
        for e in examples:
            print("  VIOLATION %s" % (e,))

    bad = sum(counts[k] for k in KINDS if k != "nonzero_link_reg")
    if total == 0:
        print("RESULT: FAIL (no matching prologues examined -- instrument saw nothing)")
        return 1
    if bad:
        print("RESULT: FAIL (%d violations, %d/%d covered)" % (bad, covered, total))
        return 1
    print("RESULT: PASS (%d/%d covered, max width %d <= %d)" %
          (covered, total, max_width, MAX_WINDOW_BYTES))
    return 0


if __name__ == "__main__":
    sys.exit(main())
