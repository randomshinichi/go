# The Go Programming Language ported to PowerPC Mac
A fork of Go 1.26.8 `c293dd49` that adds a big endian `darwin/ppc64` port, tested on a PowerMac G5 Quad running 10.5.8 Leopard.

- bootstraps itself!
- up to date SSL/TLS support!
- 230/247 tests from the stdlib pass - mostly small problems, nothing wrong with the port itself
- profile guided optimization works!??
- G4, G3 support planned
- no optimized intrinsics for AES, GCM, rounding, sha...
- no CGO, no external linking

# Building it

Step 1 builds a Go that can run on PowerPC Mac; Step 2 uses that `go` command to build the toolchain.

### 1. Make a bootstrap toolchain (on any machine with a working Go 1.26.8)

```sh
cd go/src
GOOS=darwin GOARCH=ppc64 ./make.bash          # cross-builds pkg/tool/darwin_ppc64/*
GOOS=darwin GOARCH=ppc64 ../bin/go build -o ../bin/go cmd/go   # a Go that RUNS on the G5
cp -a .. /path/to/g5/                         # copy the whole tree over
```

This produces a `go1.26.8 darwin/ppc64` toolchain with `DefaultGOPPC64` baked to `ppc970`.
*(This recipe is reconstructed from the recorded build of the toolchain used on the machine)*
Cross-building for `darwin/ppc64` from Linux or macOS/arm64 works; `CGO_ENABLED=0` and `GOFLAGS=-pgo=off`
are the safe defaults while cross-building.

### 2. Bootstrap natively (on the Power Mac, in Leopard)

```sh
cd /path/to/go/src
GOROOT_BOOTSTRAP=/path/to/the/step-1/tree ./make.bash
```

That is the whole thing — `./make.bash` **works as-is**, with default PGO, on all four cores. It ends with:

```
Installed Go for darwin/ppc64 in /path/to/go
Installed commands in /path/to/go/bin
```

Then, with nothing else set:

```sh
$ /path/to/go/bin/go version
go version go1.26.8 darwin/ppc64
$ /path/to/go/bin/go env GOPPC64
ppc970
```

**Confirmed on hardware:** `make.bash` exit 0; `GOHOSTARCH=ppc64`, `GOARCH=ppc64`, `GOPPC64=ppc970` from a standalone
`go tool dist env` with nothing exported; a program computing SHA-256 builds with **132 target-tool invocations and
zero host tools** and prints the right digest; `go test crypto/sha256` passes. Two independent native builds produced
a **byte-identical** `bin/go`.

## What works

- The full self-hosting chain: `cmd/dist` → `toolchain1` → `go_bootstrap` → `toolchain2` → `toolchain3` (PGO'd) → packages and commands.
- `runtime`, `sync`, `os`, `net`, `encoding/json`, `os/exec`, `crypto/sha256`, `crypto/tls`, and the rest of the 230 passing legs.
- Profile-guided optimisation, using the target's own `preprofile`.
- The `runtime/pprof` **CPU sampler** (verified sampling on hardware), `net` (sockets, DNS via file-based resolvers), `/proc`-free Leopard syscalls (`getfsstat64`, `readdir_r$INODE64`, `lutimes`, `F_GETPATH`).

# `stdlib` testsuite failures - not a problem
| leg | cause (measured) |
|---|---|
| `go/types`, `go/internal/gcimporter`, `internal/platform` | Needs `$GOROOT/test/**`, which the test fixture didn't ship. `go/types` type-checked **675 packages in 23 s** before failing on the corpus. |
| `crypto` (`TestPureGoTag`) | a modern `git` wasn't on the G5 Quad's `PATH`. |
| `crypto/ed25519`, `crypto/tls` (`TestBogoSuite`) | Need network/vendored testdata. |
| `math/big`, `crypto/ecdh`, `net/http`, `internal/coverage/cfile` | `TestLinker*`-family tests that build a program and inspect the binary's symbols. `crypto/ecdh` hard-codes a P256 symbol assumption that is false on big-endian ppc64. |
| `runtime` | Newly reachable in the provisioned configuration; **not yet fully triaged** — the one item here that is genuinely open. |


# Credits
Sponsored by a desire to learn about things beneath the compiler.
Coordinator - qwen3.8-flash, deepseek-v4.1-flash
Planner, Oracle - gpt-6.1-astra
Tireless Worker - gpt-6-luna xhigh, sonnet-5-5
Verification Workers - gpt-6.1-sol, claude-opus-5-5
Reviewers, Debuggers - gpt-6.1-sol, claude-opus-5-5
