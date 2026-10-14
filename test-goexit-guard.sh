#!/bin/bash
set -euo pipefail

if (($# != 0)); then
	printf '%s\n' 'goexit guard acceptance rejects arguments; do not pass filters, cache options, or test selectors' >&2
	exit 2
fi

root=$(cd -- "$(dirname -- "$0")" && pwd)
project=$(cd -- "$root/../../.." && pwd)
go="$root/bin/go"
builder="$project/probes/ppc-mac/build-toolchain.py"

# This command requires a provisioned worktree-local host Go toolchain. It does
# not fall back to a system Go: the compiler, linker, and runtime sources must
# all come from this worktree. On a clean checkout, provision it with the
# build-toolchain command printed by fail_prerequisite.
fail_prerequisite() {
	printf 'goexit guard acceptance cannot execute: %s\n' "$1" >&2
	printf 'required clean-checkout prerequisite: python3 %s --worktree %s --variant goexit-guard --no-prune\n' "$builder" "$root" >&2
	exit 1
}

if [[ ! -x "$go" ]]; then
	fail_prerequisite "missing worktree-local Go executable $go"
fi
for source in "$root/src/runtime/asm_ppc.s" "$root/src/runtime/asm_ppc64x.s"; do
	if [[ ! -r "$source" ]]; then
		fail_prerequisite "missing required runtime source $source"
	fi
done
for tool in compile link; do
	if ! env GOROOT="$root" GOOS= GOARCH= "$go" tool "$tool" -V >/dev/null 2>&1; then
		fail_prerequisite "missing usable host tool '$tool' in the worktree Go toolchain at $root"
	fi
done

output=$(mktemp "${TMPDIR:-/tmp}/goexit-guard-acceptance.XXXXXX")
trap 'rm -f "$output"' EXIT HUP INT TERM

if ! (cd "$root" && env GOROOT="$root" GOFLAGS= GOOS= GOARCH= "$go" test cmd/internal/obj/ppc -count=1 -v) >"$output" 2>&1; then
	cat "$output"
	printf '%s\n' 'goexit guard acceptance failed: required test command returned nonzero' >&2
	exit 1
fi
cat "$output"

if grep -Fq '(cached)' "$output"; then
	printf '%s\n' 'goexit guard acceptance failed: cached results are not fresh execution' >&2
	exit 1
fi
if grep -Fq 'warning: no tests to run' "$output"; then
	printf '%s\n' 'goexit guard acceptance failed: no tests ran' >&2
	exit 1
fi
for event in \
	'^--- PASS: TestRuntimeGoexitIsFrameless \([0-9]+([.][0-9]+)?s\)$' \
	'^[[:space:]]+--- PASS: TestRuntimeGoexitIsFrameless/ppc \([0-9]+([.][0-9]+)?s\)$' \
	'^[[:space:]]+--- PASS: TestRuntimeGoexitIsFrameless/ppc64 \([0-9]+([.][0-9]+)?s\)$'; do
	if ! grep -Eq -- "$event" "$output"; then
		printf 'goexit guard acceptance failed: missing anchored fresh event %s\n' "$event" >&2
		exit 1
	fi
done
if ! tail -n 1 "$output" | grep -Eq '^ok[[:space:]]+cmd/internal/obj/ppc[[:space:]]+[0-9]+([.][0-9]+)?s$'; then
	printf '%s\n' 'goexit guard acceptance failed: missing terminal successful cmd/internal/obj/ppc package summary' >&2
	exit 1
fi
printf '%s\n' 'ACCEPTED: fresh ppc and ppc64 goexit guard checks both passed'
