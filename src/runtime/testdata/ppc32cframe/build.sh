#!/bin/sh
# Build the ppc32cframe probe: cross-compile the Go program here, compile the C
# helpers on the target with the Leopard system gcc.
#   ./build.sh <GOROOT worktree> <outdir> [ssh target]
# then run it with:  ssh ibook 'sh -s -- --cpu 120 --wall 300 --prefix frames -- \
#   /tmp/frames/ppc32cframe /tmp/frames/libframes.dylib' < probes/ppc-mac/bounded-run.sh
set -eu
root=$1 out=$2 target=${3:-ibook}
here=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$out"
(cd "$here" && GOROOT="$root" GOOS=darwin GOARCH=ppc CGO_ENABLED=0 GOTOOLCHAIN=local \
	"$root/bin/go" build -ldflags=-checklinkname=0 -o "$out/ppc32cframe" .)
ssh "$target" 'mkdir -p /tmp/frames'
scp "$here/helpers.c" "$target:/tmp/frames/helpers.c"
ssh "$target" 'cd /tmp/frames && gcc -arch ppc -O0 -dynamiclib -o libframes.dylib helpers.c'
scp "$out/ppc32cframe" "$target:/tmp/frames/ppc32cframe"
