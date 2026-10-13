#!/bin/sh
# Build the stackcheck_lr witness for darwin/ppc.
#   build.sh fixed   OUT   -- the tree's asm_ppc.s (stackcheck returns via CTR, LR reloaded)
#   build.sh mutant  OUT   -- the NAMED MUTANT "stackcheck-plain-RET": the pre-change
#                             asm_ppc.s (both returns are a bare RET), overlaid.
# Run from anywhere with GOROOT pointing at the worktree and its bin/go first on PATH.
set -eu
variant=$1; out=$2
here=$(cd "$(dirname "$0")" && pwd)
root=$(go env GOROOT)
ov=${TMPDIR:-/tmp}/stackcheck_lr_overlay_$variant.json
asm=$root/src/runtime/asm_ppc.s
[ "$variant" = mutant ] && asm_src=$here/overlay/asm_ppc.s.mutant || asm_src=$asm
cat > "$ov" <<JSON
{"Replace": {
 "$root/src/runtime/proc.go": "$here/overlay/proc.go",
 "$root/src/runtime/stack.go": "$here/overlay/stack.go",
 "$root/src/runtime/stackcheck_lr_1ye4c.go": "$here/overlay/stackcheck_lr_1ye4c.go",
 "$root/src/runtime/asm_ppc.s": "$asm_src"
}}
JSON
cd "$here"
GOOS=darwin GOARCH=ppc CGO_ENABLED=0 GO111MODULE=off go build -overlay "$ov" -o "$out" .
