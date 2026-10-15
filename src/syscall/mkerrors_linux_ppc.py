#!/usr/bin/env python3
"""Overlay PowerPC UAPI constants onto host-width mkerrors output.

The vendored x/sys table was generated for linux/ppc from a PowerPC include
root (see its generated-file header). mkerrors.sh -m32 alone selects host
width, not target architecture; retain its non-architecture constants but use
this PowerPC table for constants it defines.
"""

import re
import sys


def constants(text):
    pattern = r"^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(0x[0-9a-fA-F]+|[0-9]+)\s*$"
    return {name: int(value, 0) for name, value in re.findall(pattern, text, re.M)}


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: mkerrors_linux_ppc.py vendor-zerrors-linux-ppc.go")
    generated = sys.stdin.read()
    with open(sys.argv[1], encoding="utf-8") as source:
        target = source.read()
    host_values = constants(generated)
    ppc_values = constants(target)
    names = host_values.keys() & ppc_values.keys()
    output = generated
    for name in sorted(names):
        if host_values[name] == ppc_values[name]:
            continue
        line = re.compile(
            r"^(\s*" + re.escape(name) + r"\s*=\s*)(?:0x[0-9a-fA-F]+|[0-9]+)(\s*)$",
            re.M,
        )
        output, count = line.subn(lambda match: match.group(1) + hex(ppc_values[name]) + match.group(2), output)
        if count != 1:
            raise SystemExit(f"expected one generated numeric constant for {name}, found {count}")
    sys.stdout.write(output)


if __name__ == "__main__":
    main()
