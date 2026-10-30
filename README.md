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

## What works

- The full self-hosting chain: `cmd/dist` → `toolchain1` → `go_bootstrap` → `toolchain2` → `toolchain3` (PGO'd) → packages and commands.
- `runtime`, `sync`, `os`, `net`, `encoding/json`, `os/exec`, `crypto/sha256`, `crypto/tls`, and the rest of the 240 passing legs.
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

So on darwin/ppc and darwin/ppc64 the runtime does not ask the kernel. A sampler thread wakes twice per period, asks Mach how much
CPU each Go thread has used (`thread_info(THREAD_BASIC_INFO)`), and sends `SIGPROF` with `pthread_kill` — which *is*
thread-directed, and which Leopard does honour. It is a user-space version of Linux's per-thread CPU timers, and it
is confined to the two PowerPC ports: darwin/amd64 and darwin/arm64 keep the original two-line path, where the kernel behaves.

**2. Leopard's `nanosleep` takes a lock it may already hold — so our `osyield` is not `usleep`.**

Go's Darwin `osyield` is `usleep(1)`. That is fine except from inside a signal handler: Leopard's `nanosleep` calls
`_pthread_testcancel`, which takes the *calling thread's own* libc spinlock. If `SIGPROF` interrupts a thread inside
that region, and the handler then spins waiting for a profile lock, its `usleep` waits forever on a lock that the
interrupted code on the same thread still holds — and if a GC was stopping the world, the program is gone. A C probe
with no Go in it hangs 3/3 that way, and never with `sched_yield`. On darwin/ppc64 `osyield` now calls libc's `sched_yield`
(as Linux does), and darwin/ppc issues the `swtch_pri` Mach trap directly; darwin/amd64 and darwin/arm64 keep
`usleep(1)`.

**3. Leopard has no per-socket TCP keepalive tuning — so we set what we can and connect anyway.**

TCP keepalive probes an idle connection to find out whether the far end is still there. It has three knobs: how long
the connection must be quiet before probing starts (the idle time), how long between probes (`TCP_KEEPINTVL`), and
how many missed probes before declaring it dead (`TCP_KEEPCNT`). Go's `net.Dialer.KeepAliveConfig` exposes all three
per connection, and on Darwin Go reaches for the socket options `0x101` and `0x102`.

Leopard's kernel has only the first. From the machine's own header:

```
/usr/include/netinet/tcp.h:203:  #define TCP_KEEPALIVE  0x10
```

`0x101` and `0x102` appear nowhere — they arrived in Snow Leopard. So `setsockopt` returns `ENOPROTOOPT`, which
means "no such option here", and Go's dial path treated that as fatal: **the connection was never made at all.** That
single interaction cost the three `TestTCPConnKeepAliveConfig*` tests out of 283 passing in `net`.

The feature is not missing, only the per-connection control. Leopard keeps connections alive on a system-wide
schedule — `net.inet.tcp.keepidle` (2 h), `net.inet.tcp.keepintvl` (75 s), and no `keepcnt` sysctl at all. So on
darwin/ppc and darwin/ppc64 we set the options the kernel does have, **ignore `ENOPROTOOPT` for the two it does not**, and let the
system defaults stand: the connection is made and works, on Leopard's schedule rather than the requested one. Any
other errno still propagates, so a real failure is not swallowed. The other Darwin ports, where the options exist,
keep the strict behaviour.

# `stdlib` testsuite: 240 of 248 legs pass on ppc32

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
