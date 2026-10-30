## Limitations: where we work around old OS X bugs
Apparently even OS kernels change their APIs and capabilities all the time, and downstream software may assume
certain capabilities that weren't available back then (usually because the CPU doesn't support some primitive)

So we split development into two branches:
`-strict` branch ports Golang to ppc/ppc64 darwin/linux with strict semantics
`-darwin` branch: based on `-strict`, adds compatibility shims for Leopard. May relax semantics within reason.

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
darwin/ppc64 we set the options the kernel does have, **ignore `ENOPROTOOPT` for the two it does not**, and let the
system defaults stand: the connection is made and works, on Leopard's schedule rather than the requested one. Any
other errno still propagates, so a real failure is not swallowed. The other Darwin ports, where the options exist,
keep the strict behaviour.

**4. Leopard's kernel rejects `MADV_FREE` — so a page is released with `msync(MS_KILLPAGES)`.**

`madvise(MADV_FREE)` is how modern Go hands memory back: the pages stay mapped but the kernel may reclaim them. On
xnu-1228 that call is `EINVAL`, so darwin/ppc uses the kernel's other answer, `msync(MS_KILLPAGES)`, which moves
eligible pages to the inactive queue and leaves the mapping valid. Two things are worth knowing about it. The kernel
skips objects it is not the sole owner of (`ref_count != 1`, `needs_copy`, a shadow) and skips wired, busy or
gobbled pages even when the call succeeds — so a successful `msync` means "these pages may now be reclaimed", not
"the RSS dropped". A measured 1 GiB anonymous mapping did move active-to-inactive under no pressure at all. The
other is that 10.6's `MADV_FREE` is a wrapper around this same operation.

**5. Leopard has no `*at` syscall family — so an anchored operation is rebuilt from two kernel interfaces.**

There is no `openat`, `renameat`, `linkat`, `unlinkat`, `fstatat`, `mkdirat`, `symlinkat`, `readlinkat` or
`fchmodat` in the kernel or in libSystem, which is why `os.Root` has nothing to call. What xnu-1228 does have is
`fcntl(F_OPENFROM)` (56) and `fcntl(F_UNLINKFROM)` (57) — anchored open and anchored unlink — plus syscall 349,
`__pthread_fchdir`, a working directory that belongs to one thread. So one anchored operation runs under a
thread-private cwd inside a single `entersyscall`/`exitsyscall` scope, and most of `os.Root` becomes expressible
without ever reconstructing a pathname. The dispatch and reconstruction follow MacPorts legacy-support's `atfuncs`,
credited in the file.

Two places it cannot be made equivalent, and both are refused rather than approximated. Flagless `linkat` must link
the source symlink itself, but Leopard's `link(2)` resolves its first name with `FOLLOW` and there is no `llink(2)`,
so `AT_SYMLINK_NOFOLLOW` answers `ENOTSUP`. And on a kernel where the anchored backend is not enabled (an older
9.x, or Tiger), `RemoveAll` falls back to the pre-Go-1.12 path-based walk, which has the symlink race that version
fixed; `Root.RemoveAll` takes no such fallback, and returns `ENOTSUP`, because being confined is the whole point.

**6. Leopard has no `O_CLOEXEC` — so close-on-exec is a second call.**

The flag arrived in 10.7, so `syscall.O_CLOEXEC` is 0 on darwin/ppc and every open is followed by an explicit
`fcntl(F_SETFD, FD_CLOEXEC)`. That leaves a window in which a fork can inherit the descriptor, which an
open-with-flag would not — accepted rather than dropping close-on-exec entirely.
