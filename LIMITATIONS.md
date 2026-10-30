## Limitations: where we work around old OS X bugs

Every limitation below is a property of Leopard's kernel (xnu-1228) or libSystem, not of the CPU. Source citations
are in this tree at f06c93b449 (branch 1.26.8-darwin); kernel citations are into `reference/xnu` (xnu-1228).
Hardware numbers are labelled with the machine and width that produced them: **pmg5** is the G5 (ppc64 or ppc32),
**powerbook-g4** is the G4 (ppc32 only). Where a number comes from a bead note and its artifact is not in `runs/`,
we say so.

Branches: all of the code below is present on both `1.26.8-darwin` and `1.26.8-strict` today. The planned split
(bead goppc-13s2v, still open) would put only the ppc32 anchored `*at` shim (§5) in `-darwin`.

**1. Leopard delivers profiling signals to the wrong thread — so we sample in user space.**

`ITIMER_PROF` raises `SIGPROF` on the *process*. On xnu-1228 that is `psignal(p, SIGPROF)`
(`reference/xnu/bsd/kern/kern_sig.c:2675`), and `get_signalthread` (`kern_sig.c:1466`) hands it to the first thread
in the process's thread list that does not block it. A CPU profiler needs the thread that used the CPU. Upstream
Go disabled the macOS CPU-profile tests from August 2013 until February 2021 (`git show b445d6ea34`, "Fixes
#6047"; that commit blames a missing kernel patch, not Leopard in particular). A C probe — four threads, main and
A idle, B and C spinning, `setitimer(ITIMER_PROF, 10ms)`, signals counted per thread — measured on pmg5 (numbers
from bead goppc-proc-mil000.1.6.1; the probe `sigprof_who.c` is not on this host):

| | main (idle) | A (idle) | B (spin) | C (spin) |
|---|---|---|---|---|
| Leopard ppc64 | **493** | 0 | 0 | 0 |
| Leopard ppc64, main+A block SIGPROF | 0 | 0 | **419** | 0 |
| Leopard ppc32 | **425** | 0 | 0 | 0 |
| Leopard ppc32, main+A block SIGPROF | 0 | 0 | **404** | 0 |
| Linux (same probe) | 0 | 0 | 306 | 184 |

A profiler that trusts the kernel samples whichever thread sits first, usually an idle one: either "0 samples" or
a stack full of `pthread_cond_wait`.

So on darwin/ppc and darwin/ppc64 the runtime does not ask the kernel. A sampler thread wakes twice per period,
reads each Go thread's CPU time with `thread_info(THREAD_BASIC_INFO)`, and sends `SIGPROF` with `pthread_kill`,
which is thread-directed (`src/runtime/cpuprof_darwin_ppc64.go`, `cpuprof_darwin_ppc32.go`). darwin/amd64 and
darwin/arm64 keep the original path (`cpuprof_darwin.go`, `darwin && !ppc && !ppc64`). Results: ppc64, 220/220
`runtime/pprof` test passes over 20 full-suite runs on pmg5, 0 hangs (goppc-proc-mil000.1.6.1); ppc32, `runtime/pprof`
leg PASS in 210 s on powerbook-g4 (`runs/augzc3-std-pbg4-20261025T032406Z/ledger.tsv:5`; log under
`scratch/augzc3-std-pbg4-20261025T032406Z/collected/logs/runtime_pprof.log`;
`TestCPUProfileMultithreadMagnitude` is skipped), and `TestCPUProfile` / `TestCPUProfileIdleMainThread` 6/6 each on
pmg5 (goppc-xprg.13).

**2. Leopard's `nanosleep` takes a lock it may already hold — so `osyield` is not `usleep` on ppc or ppc64.**

Go's Darwin `osyield` is `usleep(1)`. That is unsafe inside a signal handler: per bead goppc-proc-mil000.1.6.1
(disassembly of the G5's libSystem), Leopard's `nanosleep` calls `_pthread_testcancel`, which takes the calling
thread's own libc spinlock. If `SIGPROF` interrupts a thread inside that region and the handler then spins on the
profile lock (`cpuprof.go:109,146`) with `usleep`, it waits forever on a lock its own thread holds. A C probe with
no Go in it hung 3 of 3 runs within 2 s with `usleep(1)` and did not hang in 2 of 2 runs of 60 s with `sched_yield`
(pmg5, same bead); before the fix the Go runtime hung 2 of 8 full runs with a GC waiting for the stuck thread.

| port | `osyield` |
|---|---|
| darwin/ppc64 | libc `sched_yield` (`sys_darwin_ppc64.go:51-57`) |
| darwin/ppc | the `swtch_pri` Mach trap (−59) issued directly (`osyield_darwin_ppc.go`, `osyield_darwin_ppc.s`; trap listed at `reference/xnu/osfmk/mach/syscall_sw.h:113`) |
| darwin/amd64, darwin/arm64 | `usleep(1)` (`osyield_darwin.go`, `darwin && !ppc && !ppc64`) |

**3. Leopard has no per-socket TCP keepalive interval or count — so we set what we can and connect anyway.**

TCP keepalive has three knobs: idle time before the first probe, the interval between probes, and the number of
missed probes. Go's `KeepAliveConfig` sets all three per connection; on Darwin it uses `TCP_KEEPALIVE` for idle and
socket options `0x101` (interval) and `0x102` (count) for the others. Leopard defines only the first
(`reference/xnu/bsd/netinet/tcp.h:203`: `TCP_KEEPALIVE 0x10`). `0x101` and `0x102` return `ENOPROTOOPT`, and Go's
dial path treated that as fatal, so the connection was never made: net on pmg5 (ppc64) was 283 PASS / 3 FAIL /
19 SKIP, the three failures being `TestTCPConnKeepAliveConfig`, `…Dialer`, `…Listener` (goppc-acq5).

Leopard keeps connections alive on system-wide values: idle 2 h and interval 75 s (`tcp_timer.h:130-131`), and a
fixed 8 probes with no sysctl (`tcp_timer.h:132`, `tcp_timer.c:154`); the sysctl values were measured on pmg5
(goppc-acq5). So on darwin/ppc **and** darwin/ppc64 we set idle per socket, **ignore `ENOPROTOOPT` for interval and
count only**, and leave the system values in force (`src/net/tcpsockopt_darwin_leopard.go`, `darwin && (ppc ||
ppc64)`). Any other errno still propagates; the idle path is unchanged. Other Darwin ports keep strict behaviour
(`tcpsockopt_darwin_other.go`). Results: ppc64, net 288 PASS / 0 FAIL / 19 SKIP on pmg5 (goppc-acq5); ppc32, the three
keepalive-config tests and `TestDarwinKeepAliveBestEffortRoundTrip` PASS on powerbook-g4
(`runs/augzc3-std-pbg4-20261025T032406Z/legs/net.log:1060-1114`; that leg is a manual PASS, grader artifact
disclosed in its REPORT.md). The merged tests check idle and the round trip but no longer assert interval or count
(goppc-us6l, open).

**4. Leopard's kernel rejects `MADV_FREE` — so a page is released with `msync(MS_KILLPAGES)`.**

xnu-1228's `madvise` accepts only RANDOM, SEQUENTIAL, NORMAL, WILLNEED and DONTNEED; anything else, `MADV_FREE`
included, is `EINVAL` (`reference/xnu/bsd/kern/kern_mman.c:846`). On darwin/ppc and darwin/ppc64 `sysUnusedOS` calls
`msync(MS_KILLPAGES)` instead (`src/runtime/mem_darwin_leopard.go:23-39`), which moves eligible pages to the
inactive queue and leaves the mapping valid (`kern_mman.c:643`, `osfmk/vm/vm_map.c:11159-11162`). The kernel skips
objects it does not solely own (`ref_count != 1`, `needs_copy`, a shadow) and wired or busy pages even when the call
succeeds, so success means "these pages may now be reclaimed", not "RSS dropped". If `msync` fails the runtime
throws. Measured on pmg5 with a C probe (not the Go runtime): a 1 GiB anonymous mapping moved 262,144 pages
active-to-inactive under `MS_KILLPAGES`, the control moved 0 (bead goppc-proc-tsk4462464506; the probe report is not
on this host). We have no receipt of the Go runtime itself taking this path: `TestMemoryLimit` is skipped in the
only full ppc32 runtime leg (`runs/augzc3-std-pbg4-20261025T032406Z/legs/runtime.log:2084`).

**5. Leopard has no `*at` syscall family — so ppc32 rebuilds anchored operations from private kernel interfaces, and ppc64 refuses them.**

xnu-1228 has no `openat`, `renameat`, `linkat`, `unlinkat` or `fstatat` (none in `reference/xnu/bsd/kern/syscalls.master`),
and the `*at` functions are absent from libSystem (`at_darwin_ppc64.go:14-18` cites the G5 symbol probe; artifact not
on this host). What the kernel does have is `fcntl(F_OPENFROM)` (56) and `fcntl(F_UNLINKFROM)` (57), both private
(`reference/xnu/bsd/sys/fcntl.h:235-236`), and syscall 349, `__pthread_fchdir`, a working directory belonging to one
thread (`syscalls.master:505`).

*darwin/ppc.* One anchored operation runs under a thread-private cwd inside a single `entersyscall`/`exitsyscall`
scope (`src/syscall/at_compound_darwin_ppc.go:60-92`, `src/internal/syscall/unix/at_darwin_ppc.go`; structure follows
MacPorts legacy-support `atfuncs`, credited in the file). The backend is enabled only when `kern.osrelease` parses as
9.8.x; elsewhere it refuses. On powerbook-g4 the full `os` package is 268 PASS / 2 FAIL / 17 SKIP
(`runs/at-v2-ppc32-20261026T194311Z/os-full.log`, REPORT.md); the same tests were 261 PASS / 8 FAIL before the
two-name operations existed (`runs/augzc3-std-pbg4-20261025T032406Z/legs/os.log`). The tree under `src/os`,
`src/syscall`, `src/internal/syscall/unix` is that run's subject plus four files. Where it is weaker than POSIX:
- Zero-flag `Linkat` follows a source symlink (`link(2)` resolves its first name with `FOLLOW`, `vfs_syscalls.c:2605-2608`,
  and there is no `llink`); `AT_SYMLINK_NOFOLLOW` returns `ENOTSUP`. The remaining two failing `os` tests,
  `TestRootLinkFrom` (symlink cases) and `TestRootConsistencyLink` (directory cases), fail for this reason and are not skipped.
- `Link`/`Rename` between two different directories rebuild the destination name with `F_GETPATH`. That is a name
  snapshot, not a held anchor: a concurrent rename of the destination directory can redirect the operation.
- On a kernel where the backend is off, `RemoveAll` falls back to the pre-Go-1.12 path-based walk, which has the
  symlink race that version fixed (`os/removeall_darwin_ppc.go`); `Root.RemoveAll` returns `ENOTSUP`.

*darwin/ppc64.* There is no anchored backend. Every `*at` call with a real `dirfd` returns `ENOTSUP`, so `os.Root`
and the descriptor-based halves of `os` fail rather than act on the wrong directory (`at_darwin_ppc64.go:14-26`), and
`RemoveAll` is always the path-based walk with the symlink race (`os/removeall_darwin_ppc64.go`). We have no ppc64
hardware receipt for this on this host; the tests that pin it are `os/root_darwin_ppc64_test.go`.

**6. Leopard has no `O_CLOEXEC` — so close-on-exec is a second call.**

The flag is not in xnu-1228's `fcntl.h`, and the G5's compiler rejects it (`docs/darwin-abi-reference.md:1612-1615`);
the source comments say it arrived in 10.7. `syscall.O_CLOEXEC` is 0 on darwin/ppc and darwin/ppc64
(`syscall_darwin_ppc.go:17`, `syscall_darwin_ppc64.go:18`). The `os` package's opens, directory opens and `Root`
opens follow with `fcntl(F_SETFD, FD_CLOEXEC)` (`os/file_unix.go:275,298`, `os/root_unix.go:51,126`,
`supportsCloseOnExec = false`). A fork between the open and the `fcntl` can inherit the descriptor, which an
open-with-flag would not allow — accepted rather than dropping close-on-exec. A caller that passes `O_CLOEXEC`
straight to `syscall.Open` gets the flag as 0 and no close-on-exec; it must call `syscall.CloseOnExec` itself.

**7. Leopard refuses `execve` from a multithreaded process — so `syscall.Exec` always returns `ENOTSUP`.**

xnu-1228 returns `ENOTSUP` for `execve`, and for `posix_spawn(POSIX_SPAWN_SETEXEC)`, from any task with more than one
thread, before it looks at the path (`reference/xnu/bsd/kern/kern_exec.c:1963-1972`). A Go process always has several
Mach threads, so `syscall.Exec` cannot succeed and returns the kernel's `ENOTSUP`; `TestExec` fails and is not
weakened. Measured on powerbook-g4 (`runs/diag-dsvm-pbg4-20261026T204019Z`): a plain Go process shows 3 threads in
`ps -M`; `Exec` of itself, `/bin/sh`, `/bin/echo` and a nonexistent path all return errno 45; a C `execve` from one
thread succeeds, from two threads returns 45. `fork` then `execve` in the child works, which is what `ForkExec`,
`StartProcess` and `os/exec` use. The same limit on ppc64 is recorded in `syscall_darwin_ppc64.go:274-288` from G5
probes whose artifacts are not on this host.
