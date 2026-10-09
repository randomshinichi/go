#!/bin/bash
# mkerrors.sh cross-check wrapper used on pmg5.
# Go 1.26.8 cgo passed -fno-lto, which Apple's GCC 4.0.1 does not support.
args=()
for arg in "$@"; do
	[ "$arg" = "-fno-lto" ] && continue
	args+=("$arg")
done
exec /usr/bin/gcc -arch ppc "${args[@]}"
