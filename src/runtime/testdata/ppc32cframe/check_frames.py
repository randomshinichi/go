#!/usr/bin/env python3
"""Static census of the darwin/ppc C-call frame contract in a built binary.

usage: check_frames.py <binary> [path-to-go [path-to-sys_darwin_ppc.s]]

Every function from sys_darwin_ppc.s that makes a C call (a call to a
_macho_callstub.libc_* import, or an indirect CALL through CTR) must

  1. allocate a frame of at least 56 bytes: 24 bytes of linkage plus the
     32-byte home area in which a C callee stores r3-r10, and
  2. never use the home area, SP+24..SP+55, as its own scratch (a load, a
     store, or the address of a slot there).

Reference: reference/xnu/EXTERNAL_HEADERS/architecture/ppc/asm_help.h (ARG(n),
BUILD_FRAME). Exit 0 when every such function passes, 1 on any violation, 2 when
no C-calling function was found at all (the instrument saw nothing).

The 8-byte saved-LR slot sits in the CALLER's linkage area, entry SP+8, so a
frame of exactly N bytes has it at SP+N+8: that is why a 32-byte frame puts it
inside a callee's home area (SP+40 is home word five).
"""
import re
import subprocess
import sys

LINKAGE, HOME = 24, 32
funcs = {}
cur = None
out = subprocess.run([sys.argv[2] if len(sys.argv) > 2 else 'go', 'tool', 'objdump', sys.argv[1]],
                     capture_output=True, text=True, check=True).stdout
for line in out.splitlines():
    m = re.match(r'TEXT (\S+)\(SB\) .*sys_darwin_ppc\.s$', line)
    if m:
        cur = funcs.setdefault(m.group(1), [])
    elif line.startswith('TEXT '):
        cur = None
    elif cur is not None and line.strip():
        f = [x for x in line.split('\t') if x.strip()]
        if len(f) >= 4:
            cur.append(f[3].strip())

bad = 0
checked = 0
for name, insns in sorted(funcs.items()):
    ops = insns
    calls_c = any(re.match(r'(CALL|BL) (_macho_callstub\.libc_|\(CTR\))', t) for t in ops)
    if not calls_c:
        continue
    checked += 1
    frame = 0
    for t in ops:
        m = re.match(r'MOVWU R1,-(\d+)\(R1\)', t)
        if m:
            frame = int(m.group(1))
            break
    problems = []
    if frame < LINKAGE + HOME:
        problems.append('frame %d < %d' % (frame, LINKAGE + HOME))
    if frame and LINKAGE <= frame + 8 < LINKAGE + HOME:
        problems.append('saved LR at SP+%d lies inside the callee home area' % (frame + 8))
    for t in ops:
        # The prologue/epilogue's own accesses (reload of the saved LR at
        # frame+8, the frame pop) are not scratch; the line above reports them.
        if re.match(r'MOVWZ %d\(R1\),R31$' % (frame + 8), t) or re.match(r'ADD R1,\$%d,R1$' % frame, t):
            continue
        for m in re.finditer(r'(-?\d+)\(R1\)', t):
            if LINKAGE <= int(m.group(1)) < LINKAGE + HOME:
                problems.append('home-area scratch: ' + t)
        m = re.match(r'ADD R1,\$(\d+),R\d+', t)
        if m and LINKAGE <= int(m.group(1)) < LINKAGE + HOME:
            problems.append('home-area address: ' + t)
    print('%s %-52s frame=%-3d %s' % ('FAIL' if problems else 'ok  ', name, frame, '; '.join(problems)))
    bad += bool(problems)
print('checked %d C-calling functions, %d violation(s)' % (checked, bad))
if len(sys.argv) > 3:
    # The linker drops unreferenced functions; say which source TEXT symbols
    # this binary did not contain, so a clean result is not read as covering them.
    src = re.findall(r'^TEXT runtime·(\w+)\(SB\)', open(sys.argv[3]).read(), re.M)
    absent = [n for n in src if 'runtime.' + n not in funcs]
    print('not in this binary (dead-stripped): ' + (', '.join(absent) or 'none'))
    # For those, fall back to the declared locals in the source: a frame of
    # 24 + locals rounded up to 16 holds the home area iff locals >= 32.
    blocks = re.split(r'^(?=TEXT )', open(sys.argv[3]).read(), flags=re.M)
    for b in blocks:
        m = re.match(r'TEXT runtime·(\w+)\(SB\), [^,]*, \$(\d+)-', b)
        if m and m.group(1) in absent and re.search(r'\bBL\s+libc_|BCTRL', b):
            small = int(m.group(2)) < HOME
            print('%s %-52s declared locals=$%s (source only, not binary)' % ('FAIL' if small else 'ok  ', 'runtime.' + m.group(1), m.group(2)))
            bad += small
    if bad:
        sys.exit(1)
sys.exit(2 if checked == 0 else (1 if bad else 0))
