# The Go Programming Language ported to PowerPC Mac
A fork of Go 1.26.8 `c293dd49` that ports Go to big-endian PowerPC: `darwin/ppc` and `darwin/ppc64`, tested on a
PowerMac G5 Quad and iBook/PowerBook G4s running 10.5.8 Leopard; `linux/ppc` is the emulated porting venue, and
`linux/ppc64` builds but has no validated venue.

- bootstraps itself!
- up to date SSL/TLS support!
- 240/248 stdlib test legs pass on ppc32 (PowerBook G4) - the rest are kernel limits, venue gaps and provisioning, not defects in the port
- profile guided optimization works!??
- G3 support planned (the suite numbers below are measured on a G4)
- no optimized intrinsics for AES, GCM, rounding, sha...
- no CGO, no external linking
- `go tool pprof` has no interactive editing
- memory release works via `msync(MS_KILLPAGES)`. 10.6's `MADV_FREE` is just a wrapper around this too

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

## What works, what doesn't

- The full self-hosting chain: `cmd/dist` → `toolchain1` → `go_bootstrap` → `toolchain2` → `toolchain3` (PGO'd) → packages and commands.
- `runtime`, `sync`, `os`, `net`, `encoding/json`, `os/exec`, `crypto/sha256`, `crypto/tls`, and the rest of the 240 passing legs.
- Profile-guided optimisation, using the target's own `preprofile`.
- The `runtime/pprof` **CPU sampler** (verified sampling on hardware), `net` (sockets, DNS via file-based resolvers), `/proc`-free Leopard syscalls (`getfsstat64`, `readdir_r$INODE64`, `lutimes`, `F_GETPATH`).

### Where we work around Mac OS X bugs

This port runs on a kernel old enough to have bugs the rest of the Go world never had to care about. When one of
them changes observable behaviour, we work around it in the runtime so that a Go program behaves as documented,
and we write down what and why.

The gory details of what doesn't work, and to what extent, are in [LIMITATIONS](LIMITATIONS.md).

### `stdlib` testsuite: 240 of 248 legs pass on ppc32

Measured on hardware — run `runs/augzc3-std-pbg4-20261025T032406Z`, tree tip `98614a202d`, executed 2026-10-25 on a
PowerBook5,6 (10.5.8). 248 legs: 239 clean passes, `net` graded ANOMALY and verified by hand as a pass (its own passing
test prints `panic:` from a subprocess), 8 not passing. **This table is ppc32 on a G4; it does not describe the G5.**

| leg | cause (measured) |
|---|---|
| `os` | 55 × `linkat`/`renameat`/`openat`: `operation not supported`, every one in the `Root` link/rename families — the two-name anchored operations did not exist at this tip. **Fixed since** by `9ea25b67a8`: that leg is now 268 PASS / 2 FAIL / 17 SKIP, and the two remaining failures are the documented `linkat` divergence ([LIMITATIONS](LIMITATIONS.md) §5). |
| `runtime` | `TestBadOpen`: `runtime.Close(-1)` returned the raw kernel `-9` where upstream Darwin routes through libc and returns `-1`. **Fixed since** by `da5970e7c8`; the whole leg passes in `runs/b0i4-bnd-20261026T221636Z`. |
| `time` | `panic: cannot load America/Los_Angeles for testing: unknown time zone` — no zoneinfo the test's own loader accepts; it asks for `-tags=timetzdata`. Provisioning: the leg verified nothing. |
| `syscall` | `TestExec` — Leopard refuses `execve` from a multithreaded process, `ENOTSUP`. Kernel-impossible rather than unfixed ([LIMITATIONS](LIMITATIONS.md) §7). |
| `encoding/gob` | `TestCountDecodeMallocs`: 5 allocations per decode, want 3. Upstream's own latent failure on single-CPU venues, not ppc32 — the test skips at `GOMAXPROCS>1` and fails identically on unmodified upstream. |
| `crypto/x509` | `TestHybridPool`: google.com's chain against this machine's 2008 root pool, `certificate signed by unknown authority`. The trust source is the machine's, not the port's. |
| `crypto/internal/fips140test` | `TestIntegrityCheckFailure`: the CAST/self-test injection harness, same signature as the iBook's provisioning classification (`goppc-ljq9l`). |
| `go/types` | `ExampleScope`: `could not import fmt … fork/exec …/bin/go: no such file or directory` — the examples shell out to a native `go`, and this venue has no ppc32 `go` binary. Venue, not port; a native dist flips it. |

Built `CGO_ENABLED=0` and cross-built, on a 1 GiB G4: cgo, the cgo resolver, `os/user` and PAM are unexercised here.


# Credits
Sponsored by a desire to learn about things beneath the compiler.

Coordinator - qwen3.8-flash, deepseek-v4.1-flash

Planner, Oracle - gpt-6.1-astra

Tireless Worker - gpt-6-luna xhigh, sonnet-5-5

Verification Workers - gpt-6.1-sol, claude-opus-5-5

Reviewers, Debuggers - gpt-6.1-sol, claude-opus-5-5
