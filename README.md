# The Go Programming Language ported to PowerPC Mac
A fork of Go 1.26.8 `c293dd49` that adds a big endian `darwin/ppc64` port, tested on a PowerMac G5 Quad running 10.5.8 Leopard.

- bootstraps itself!
- up to date SSL/TLS support!
- 234/247 stdlib test legs pass on the G5 - the rest are platform limits and test-fixture gaps, not defects in the port
- profile guided optimization works!??
- G4, G3 support planned
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

## What works

- The full self-hosting chain: `cmd/dist` → `toolchain1` → `go_bootstrap` → `toolchain2` → `toolchain3` (PGO'd) → packages and commands.
- `runtime`, `sync`, `os`, `net`, `encoding/json`, `os/exec`, `crypto/sha256`, `crypto/tls`, and the rest of the 230 passing legs.
- Profile-guided optimisation, using the target's own `preprofile`.
- The `runtime/pprof` **CPU sampler** (verified sampling on hardware), `net` (sockets, DNS via file-based resolvers), `/proc`-free Leopard syscalls (`getfsstat64`, `readdir_r$INODE64`, `lutimes`, `F_GETPATH`).

## Where we work around Mac OS X bugs

This port runs on a kernel old enough to have bugs the rest of the Go world never had to care about. When one of
them changes observable behaviour, we work around it in the runtime so that a Go program behaves as documented,
and we write down what and why.

**1. Leopard delivers profiling signals to the wrong thread — so we sample in user space.**

`ITIMER_PROF` raises `SIGPROF` on the *process*, and the kernel chooses which thread receives it. On Linux, and on
macOS from about 2012 onward, that is the thread that used the CPU — exactly what a CPU profiler needs. On Mac OS X
10.5 the kernel (xnu-1228) instead wakes the first thread that is not blocking the signal. This is upstream Go issue
6047, and the reason upstream *skipped* the Darwin CPU-profile tests from 2013 until 2021. Measured on the G5 with a
C probe — four threads, two of them spinning:

| | main (idle) | A (idle) | B (spin) | C (spin) |
|---|---|---|---|---|
| Leopard ppc64 | **493** | 0 | 0 | 0 |
| Leopard ppc64, main+A block SIGPROF | 0 | 0 | **419** | 0 |
| Linux x86_64, same probe | 0 | 0 | 306 | 184 |

A profiler that trusts the kernel therefore samples whichever thread sits first — usually an idle one, giving either
"0 samples" or a stack full of `pthread_cond_wait`. `TestCPUProfile` passing alone and failing in the suite was
purely a matter of which goroutine earlier tests had left on that thread.

So on darwin/ppc64 the runtime does not ask the kernel. A sampler thread wakes twice per period, asks Mach how much
CPU each Go thread has used (`thread_info(THREAD_BASIC_INFO)`), and sends `SIGPROF` with `pthread_kill` — which *is*
thread-directed, and which Leopard does honour. It is a user-space version of Linux's per-thread CPU timers, and it
is confined to darwin/ppc64: darwin/amd64 and darwin/arm64 keep the original two-line path, where the kernel behaves.

**2. Leopard's `nanosleep` takes a lock it may already hold — so our `osyield` is not `usleep`.**

Go's Darwin `osyield` is `usleep(1)`. That is fine except from inside a signal handler: Leopard's `nanosleep` calls
`_pthread_testcancel`, which takes the *calling thread's own* libc spinlock. If `SIGPROF` interrupts a thread inside
that region, and the handler then spins waiting for a profile lock, its `usleep` waits forever on a lock that the
interrupted code on the same thread still holds — and if a GC was stopping the world, the program is gone. A C probe
with no Go in it hangs 3/3 that way, and never with `sched_yield`. On darwin/ppc64 `osyield` now calls `sched_yield`
(as Linux does); the other Darwin ports keep `usleep(1)`.

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
