#!/bin/bash
# gen-host.sh — host-side generation of the darwin/ppc32 golang.org/x/sys/unix
# tables that need no Mac (goppc-jic6x.3).
#
# usage: gen-host.sh XSYS_MODULE_DIR GOROOT_WORKTREE STAGING_DIR [ZTYPES_RAW ZERRORS_RAW]
#
#   XSYS_MODULE_DIR  hash-verified golang.org/x/sys@v0.39.0 module directory
#   GOROOT_WORKTREE  the Go worktree whose src/cmd/vendor carries the vendored
#                    x/sys copy and whose bin/go runs the generators
#   STAGING_DIR      must exist and be empty; everything is written there
#   ZTYPES_RAW       optional: cgo -godefs output from the target (step G)
#   ZERRORS_RAW      optional: mkerrors.sh output from the target (step G)
#
# Outputs in STAGING_DIR: zsyscall_darwin_ppc.go, zsyscall_darwin_ppc.s,
# zsysnum_darwin_ppc.go and, when the raws are given, ztypes_darwin_ppc.go and
# zerrors_darwin_ppc.go. Nothing is copied into the worktree: the implementer
# copies the files it wants.
set -euo pipefail
export LC_ALL=C
unset GOOS GOARCH

if [ $# -lt 3 ] || [ $# -gt 5 ]; then
	echo "usage: $0 XSYS_MODULE_DIR GOROOT_WORKTREE STAGING_DIR [ZTYPES_RAW ZERRORS_RAW]" >&2
	exit 2
fi
XSYS=$1
W=$2
STAGE=$3
ZTYPES_RAW=${4:-}
ZERRORS_RAW=${5:-}

die() { echo "gen-host.sh: $*" >&2; exit 2; }

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
FILTER=$SCRIPT_DIR/leopard-absent-sys.txt
GENSHA=$SCRIPT_DIR/xsys-generators.sha256
MKASM_PATCH=$SCRIPT_DIR/mkasm-ppc.patch
GO=$W/bin/go
GOFMT=$W/bin/gofmt
V=$W/src/cmd/vendor/golang.org/x/sys/unix
TAB=$'\t'

# ROOT is the project root: the nearest ancestor of this script that holds the
# frozen xnu syscall.h the syscall numbers are generated from.
ROOT=$SCRIPT_DIR
while [ ! -f "$ROOT/reference/xnu/bsd/sys/syscall.h" ]; do
	if [ "$ROOT" = "/" ]; then
		die "no reference/xnu/bsd/sys/syscall.h above $SCRIPT_DIR"
	fi
	ROOT=$(dirname "$ROOT")
done

for f in "$FILTER" "$GENSHA" "$MKASM_PATCH" "$GO" "$GOFMT"; do
	[ -e "$f" ] || die "missing $f"
done

# (i) the generator sources must be the pinned v0.39.0 ones.
(cd "$XSYS/unix" && sha256sum -c --quiet "$GENSHA") || die "generator hashes do not match xsys-generators.sha256"

# (ii) an empty staging directory, so nothing stale is picked up by mkasm.go.
[ -d "$STAGE" ] || die "staging dir $STAGE does not exist"
[ -z "$(ls -A "$STAGE")" ] || die "staging dir $STAGE is not empty"

# (iii) generators, with the one documented ppc pointer-size patch.
cp "$XSYS/unix/mksyscall.go" "$XSYS/unix/mkasm.go" "$XSYS/unix/mkpost.go" "$STAGE"/
chmod u+w "$STAGE/mksyscall.go" "$STAGE/mkasm.go" "$STAGE/mkpost.go"
(cd "$STAGE" && patch --quiet -p1 < "$MKASM_PATCH") || die "mkasm patch did not apply"
grep -q 'case "386", "arm", "ppc":' "$STAGE/mkasm.go" || die "mkasm.go was not patched"

# (iv) the generic and arch inputs that stay verbatim.
cp "$V/syscall_bsd.go" "$V/syscall_darwin_ppc.go" "$STAGE"/

# (v) syscall_darwin.go with the Leopard-absent //sys declarations removed.
filtered=0
while IFS=$'\t' read -r goname libcsym _replacement; do
	case $goname in '#'*|'') continue ;; esac
	hit=$(grep -cE "^//sys(nb)?[[:space:]]+$goname\(" "$V/syscall_darwin.go" || true)
	[ "$hit" = "1" ] || die "filter row $goname matches $hit //sys lines in syscall_darwin.go, want 1"
	filtered=$((filtered + 1))
done < "$FILTER"
[ "$filtered" = "23" ] || die "filter list has $filtered data rows, want 23"
names=$(awk -F'\t' '!/^#/ && NF >= 3 { printf "%s%s", sep, $1; sep = "|" }' "$FILTER")
before=$(grep -cE '^//sys(nb)?[[:space:]]' "$V/syscall_darwin.go" || true)
grep -v -E "^//sys(nb)?[[:space:]]+($names)\(" "$V/syscall_darwin.go" > "$STAGE/syscall_darwin.go"
after=$(grep -cE '^//sys(nb)?[[:space:]]' "$STAGE/syscall_darwin.go" || true)
[ $((before - after)) -eq 23 ] || die "removed $((before - after)) //sys lines, want 23"

# (vi) the libc wrappers. mksyscall.go is input-compatible but the //sys list is
# not upstream's any more, so the generated header says so.
(cd "$STAGE" && GOOS_TARGET=darwin GOARCH_TARGET=ppc "$GO" run mksyscall.go -b32 -tags darwin,ppc \
	syscall_bsd.go syscall_darwin.go syscall_darwin_ppc.go) | "$GOFMT" > "$STAGE/zsyscall_darwin_ppc.go"
{
	head -2 "$STAGE/zsyscall_darwin_ppc.go"
	echo '// Port-local: syscall_darwin.go was filtered first; see src/syscall/testdata/darwin_ppc32/xsys/README.'
	tail -n +3 "$STAGE/zsyscall_darwin_ppc.go"
} > "$STAGE/zsyscall_darwin_ppc.go.tmp"
mv "$STAGE/zsyscall_darwin_ppc.go.tmp" "$STAGE/zsyscall_darwin_ppc.go"

# (vii) the trampolines. mkasm.go leaves a darwin/ppc link test that would need
# a Mac to run, and the vendored package has no test files.
(cd "$STAGE" && "$GO" run mkasm.go darwin ppc) || die "mkasm.go failed"
rm -f "$STAGE/darwin_ppc_test.go"

# (viii) the syscall numbers, from the frozen xnu header.
mkdir -p "$STAGE/MacOSX10.5.sdk/usr/include/sys"
cp "$ROOT/reference/xnu/bsd/sys/syscall.h" "$STAGE/MacOSX10.5.sdk/usr/include/sys/syscall.h"
got=$(md5sum "$STAGE/MacOSX10.5.sdk/usr/include/sys/syscall.h" | cut -d' ' -f1)
[ "$got" = "33dfdddbb1866a4e524711b01d59aebe" ] || die "syscall.h md5 $got, want 33dfdddbb1866a4e524711b01d59aebe"
(cd "$STAGE" && perl "$W/src/syscall/mksysnum_darwin.pl" MacOSX10.5.sdk/usr/include/sys/syscall.h) > "$STAGE/zsysnum_darwin_ppc.go.raw"
# std's mksysnum names syscall 202 SYS___SYSCTL. x/sys names it SYS_SYSCTL and
# provides const SYS___SYSCTL = SYS_SYSCTL in syscall_darwin.go (that is how
# zsysnum_darwin_amd64.go and zsysnum_darwin_arm64.go spell it). Rename that
# one name; the line count and every value stay unchanged.
subs=$(grep -cE "^${TAB}SYS___SYSCTL = 202$" "$STAGE/zsysnum_darwin_ppc.go.raw" || true)
[ "$subs" = "1" ] || die "found $subs 'SYS___SYSCTL = 202' lines, want 1"
sed -e "s/^${TAB}SYS___SYSCTL = 202$/${TAB}SYS_SYSCTL = 202/" -e 's/^package syscall$/package unix/' \
	"$STAGE/zsysnum_darwin_ppc.go.raw" | "$GOFMT" > "$STAGE/zsysnum_darwin_ppc.go"
rm -f "$STAGE/zsysnum_darwin_ppc.go.raw"
if grep -qE "^${TAB}SYS___SYSCTL +=" "$STAGE/zsysnum_darwin_ppc.go"; then
	die "zsysnum_darwin_ppc.go still declares SYS___SYSCTL"
fi
{
	head -2 "$STAGE/zsysnum_darwin_ppc.go"
	echo '// Port-local: std'"'"'s mksysnum names syscall 202 SYS___SYSCTL; x/sys names it'
	echo '// SYS_SYSCTL and provides const SYS___SYSCTL = SYS_SYSCTL in syscall_darwin.go'
	echo '// (as for amd64/arm64). Value 202 is unchanged.'
	tail -n +3 "$STAGE/zsysnum_darwin_ppc.go"
} > "$STAGE/zsysnum_darwin_ppc.go.tmp"
mv "$STAGE/zsysnum_darwin_ppc.go.tmp" "$STAGE/zsysnum_darwin_ppc.go"
n=$(grep -cE "^${TAB}SYS_[A-Z0-9_]+ += [0-9]+$" "$STAGE/zsysnum_darwin_ppc.go" || true)
[ "$n" = "338" ] || die "zsysnum_darwin_ppc.go has $n SYS_ lines, want 338"

# (ix) the target-generated types and errors, when the iBook outputs are given.
if [ -n "$ZTYPES_RAW" ] || [ -n "$ZERRORS_RAW" ]; then
	[ -n "$ZTYPES_RAW" ] && [ -n "$ZERRORS_RAW" ] || die "give both ZTYPES_RAW and ZERRORS_RAW, or neither"
	(cd "$STAGE" && GOOS_TARGET=darwin GOARCH_TARGET=ppc "$GO" run mkpost.go < "$ZTYPES_RAW") > "$STAGE/ztypes_darwin_ppc.go"
	"$GOFMT" < "$ZERRORS_RAW" > "$STAGE/zerrors_darwin_ppc.go"
fi

# (x) the receipt.
for f in zsyscall_darwin_ppc.go zsyscall_darwin_ppc.s zsysnum_darwin_ppc.go ztypes_darwin_ppc.go zerrors_darwin_ppc.go; do
	if [ -f "$STAGE/$f" ]; then
		(cd "$STAGE" && sha256sum "$f")
	fi
done
if [ -n "$("$GOFMT" -l "$STAGE"/zsyscall_darwin_ppc.go "$STAGE"/zsysnum_darwin_ppc.go)" ]; then
	die "generated Go files are not gofmt-clean"
fi
