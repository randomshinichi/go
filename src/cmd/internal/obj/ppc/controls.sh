#!/bin/bash
# Negative controls for the goppc-lau1 gate. Each variant deliberately breaks the
# marker in a scratch copy of obj/ppc/link.go, rebuilds cmd/compile into a FRESH
# GOCACHE (rebuilding cmd/compile does NOT invalidate package builds on this
# toolchain, so a shared cache would reuse the fixed package), compiles one
# package, and runs the checker. The link.go under test is always restored.
set -u
W=/data/source/projects/go-g5-port/go/worktrees/ppc32-r31nf
L=$W/src/cmd/internal/obj/ppc/link.go
PRISTINE=/data/tmp-r31nf/pre-edit/link.go.FIXED-pristine
# GOOS/GOARCH are set ONLY on the package build. Exporting them around
# `go install cmd/compile` cross-compiles the tool for darwin/ppc and leaves
# pkg/tool/linux_amd64/compile untouched, which would make every variant measure
# the same (already-fixed) binary -- an invalid control.
export GOROOT=$W GO111MODULE=off TMPDIR=/data/tmp-r31nf/tmp CGO_ENABLED=0
unset GOOS GOARCH GOFLAGS
FIXED_SHA=88c023d2f12e75a88620d8736f5be5adb83d125d5517fc2208fd9ab9daef8c48

run_variant() {
  local name="$1"
  cp -f "$PRISTINE" "$L"
  python3 /data/tmp-r31nf/mutate_$name.py || { echo "$name: MUTATION FAILED"; return 1; }
  echo "--- $name: diff applied to link.go ---"
  diff "$PRISTINE" "$L" | head -20
  rm -rf /data/tmp-r31nf/gocache-$name /data/tmp-r31nf/asm-$name
  mkdir -p /data/tmp-r31nf/gocache-$name /data/tmp-r31nf/asm-$name
  GOCACHE=/data/tmp-r31nf/gocache-$name timeout 1200 env GOROOT=$W GO111MODULE=off \
      GOOS=linux GOARCH=amd64 $W/bin/go install cmd/compile || {
    echo "$name: BUILD FAILED"; cp -f "$PRISTINE" "$L"; return 1; }
  SHA=$(sha256sum $W/pkg/tool/linux_amd64/compile | cut -d' ' -f1)
  echo "$name compile sha256: $SHA"
  if [ "$SHA" = "$FIXED_SHA" ]; then
    echo "$name: INVALID CONTROL -- host compile binary is byte-identical to the fixed one"
    cp -f "$PRISTINE" "$L"; return 1
  fi
  GOCACHE=/data/tmp-r31nf/gocache-$name GOOS=darwin GOARCH=ppc \
      $W/bin/go build -gcflags='sort=-S' \
      sort > /data/tmp-r31nf/asm-$name/sort.S.txt 2>&1
  echo "--- $name: checker verdict ---"
  python3 /data/tmp-r31nf/r0wincheck.py --label "$name" /data/tmp-r31nf/asm-$name/sort.S.txt
  echo "$name checker rc=$?"
  cp -f "$PRISTINE" "$L"
}

for v in "$@"; do run_variant "$v"; done

# Always leave the tree and the toolchain in the fixed state.
cd $W || exit 1
cp -f "$PRISTINE" "$L"   # NOT git checkout: it silently reverted an uncommitted refinement
git status --porcelain | sed 's/^/UNEXPECTED-DIRTY: /
'
GOOS=linux GOARCH=amd64 GOCACHE=/data/tmp-r31nf/gocache-final timeout 1200 $W/bin/go install cmd/compile
echo "final compile sha256: $(sha256sum $W/pkg/tool/linux_amd64/compile | cut -d' ' -f1)"
echo "expect $FIXED_SHA"
