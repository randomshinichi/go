# Darwin ppc32 kevent: libc sentinel versus runtime errno

## Boundary and fix

Darwin libc returns `-1` on failure and stores the actual error in the calling
thread's errno slot. The runtime's `kevent` callers instead require `-errno`.
Without that translation, an interrupted wait (EINTR = 4) appears to the runtime
as error 1, and `netpoll` aborts rather than retrying.

The fix in `src/runtime/sys_darwin_ppc.s` mirrors the ppc64 trampoline: compare
R3 with -1, branch past error handling on success, call `libc_error`, load the
32-bit errno, negate it, return. `CMPWI`/`BC 4,2`/`LWZ` are this ppc32 assembler's
spellings of compare/BNE/word load; the disassembly confirms those operations.
The existing `NOSPLIT,$8-0` declaration and argument loads are unchanged.
`kqueue_trampoline` is unchanged: its ppc64 counterpart also just calls libc and
returns. No consumer changes or error suppression were made.

## Executed on hardware: own discriminating control (2026-10-12)

Machine: iBook G4, PowerBook6,5, 7450, one CPU, Darwin 9.8.0 / Leopard 10.5.8.
G5 unavailable: **no G5 execution or ABI claim**. No emulation was used.

The same diagnostic overlay and localhost TCP probe were used before and after
against base `d542fe44c0`. The probe sets GOMAXPROCS=2 and uses a spinning second
P to send SIGURG specifically to the M inside a blocking kevent. A process-wide
signal is not equivalent: another thread can receive it without interrupting the
poller. The overlay publishes that M only around nonzero-delay polls, logs the
return and `*getg().m.errnoAddr`, and counts accepted EINTRs. It is diagnostic
only and is **not** in production runtime code. CPU profiling is never started.

Before, `probe-before2`, wrapper FAILED rc=2:

```
O6CJ directed poller SIGURG sent; no CPU profiling
O6CJ netpoll kq/return/errno: 4 -1 4
runtime: kevent on fd 4 failed with 1
fatal error: runtime: netpoll failed
```

After, `probe-after`, wrapper OK rc=0:

```
O6CJ directed poller SIGURG sent; no CPU profiling
O6CJ netpoll kq/return/errno: 4 -4 4
O6CJ EINTR accepted; retry/recalculate
O6CJ ordinary TCP read n=0 err=read tcp 127.0.0.1:57172->127.0.0.1:57173: i/o timeout
O6CJ PASS: EINTR observed; TCP read resumed
```

The logged return is the trampoline return: before it is also libc's raw -1;
after it is the translated -4. The after disassembly shows the -1 test immediately
following the libc call. For this finite-delay wait, netpoll returns to recalculate
the deadline rather than its indefinite-wait `goto retry`; the scheduler polls
again, the intended deadline expires normally, and the probe then verifies a byte
can still be written/read on the **same connection**. This is not a timeout-bound
pass: the external 600-second wall watchdog did not fire.

An initial sleep-between-attempts probe did not observe a blocking M and exited
with `O6CJ no EINTR: control inconclusive`. It was discarded as a reproduction;
its log is retained. Keeping a second P active, rather than repeatedly waking the
poller with timers, produced the red control above before any production edit.

The committed `TestKeventErrnoDarwinPPC` independently tests the ABI boundary with
an invalid descriptor. On hardware, an assembly overlay restoring the original
trampoline yields `kevent(-1) = -1, want -9 (-EBADF)` (FAIL rc=1); the fixed binary
passes. The export is test-only and ppc32/Darwin-selected by filename.

## Per-consumer audit (inspected, plus build selection)

`GOOS=darwin GOARCH=ppc go list -f '{{join .GoFiles "\n"}}' runtime`
selects `netpoll_darwin_pipe.go`, `netpoll_kqueue.go`, and `os_darwin.go`, **not**
`netpoll_kqueue_pipe.go` or `netpoll_kqueue_event.go`. The recorded file list is
`runtime-selected.txt` in the evidence archive.

| Consumer | Boundary | ppc32 conversion | Result |
| --- | --- | --- | --- |
| `netpoll_kqueue.go`, `netpoll`, comparisons with `-_EINTR`/`-_ETIMEDOUT` | `kevent` -> `kevent_trampoline` | Missing before; now libc_error + word load + NEG | Fixed sole gap found |
| `netpoll_darwin_pipe.go`, `wakeNetpoll`, comparisons with `-_EAGAIN`/`-_EINTR` | `write` -> `write1` -> `write_trampoline` | Already compares -1, calls libc_error, loads errno and NEG (`sys_darwin_ppc.s:262-275`) | No change |
| `netpoll_kqueue_pipe.go`, `wakeNetpoll`, same comparisons | Same write abstraction in the copied pipe algorithm | Not selected for Darwin ppc32; the selected Darwin copy uses the already-converting write trampoline | No Darwin ppc32 gap; no claim about other OS ABIs |
| `os_darwin.go`, `sigNoteSleep`, comparison with `-_EINTR` | `read` -> `read_trampoline` | Already compares -1, calls libc_error, loads errno and NEG (`sys_darwin_ppc.s:211-224`) | No change |

The ppc64 read/write trampolines also perform this conversion. No second genuine
gap was found in these consumers. Most libc calls are **not** negative-errno APIs;
this does not justify adding conversions to the remaining trampolines.

## Targeted regression results (no full-package pass claims)

Each leg used `env GOMAXPROCS=2`, `-test.v -test.timeout=0`, anchored `-test.run`
patterns, and the serial bounded wrapper with CPU=1800 / wall=2000 seconds.
No wall-bounded leg and no `[no tests to run]` was counted as a pass.

| Package | Test names | Observed result |
| --- | --- | --- |
| runtime/pprof | TestBlockProfile (debug=1, proto) | PASS |
| runtime/pprof | TestMutexProfile | **Fatal in proto**: `traceback did not unwind completely`, during copystack/shrinkstack; combined leg rc=2. Not a kevent fatal; not fixed here. No overall TestMutexProfile pass claimed. |
| runtime/pprof | TestCPUProfile | PASS, 288 samples, 14 in expected cpuHog1; this placement is not an idle-main-thread/profiler-routing claim |
| runtime/pprof | TestGoroutineCounts, TestProcRegionFilename (text, shortBuffer) | PASS in separate leg after the earlier fatal prevented these from running |
| syscall | TestDirent, TestMmap, TestPositionalIODarwinPPC, TestGeneratedWordArgsDarwinPPC, TestGetfsstat | PASS |
| net | TestDialLocal, TestTCPListener, TestReadWriteDeadlineRace, TestConcurrentSetDeadline | PASS |
| os (pipe split) | TestPipeEOF, TestPipeIOCloseRace, TestPipeCloseRace | PASS |
| os (I/O split) | TestReadAt, TestWriteAt, TestStat | PASS |
| runtime (poll split) | TestKeventErrnoDarwinPPC, TestNetpollBreak | PASS |
| runtime (channel split) | TestChan, TestChanSendInterface | PASS |

The pprof fatal is retained for independent review, not suppressed or retried.
The CPU-profiling design gap tracked separately in goppc-b3n7 was not changed.

## Built and inspected evidence

Host tools were built with GOOS/GOARCH unset by `preflight.py` (full `install cmd`,
including compile/asm/link). The worktree go driver was copied from the integrated
worktree, then run with GOROOT pointing at the dedicated `ppc32-o6cj` worktree.
All target builds used `-a` **and** a distinct fresh `/data/o6cj/cache-*` GOCACHE:
GOOS=darwin, GOARCH=ppc, CGO_ENABLED=0, GO111MODULE=off, GOTOOLCHAIN=local,
TMPDIR=/data/o6cj/tmp. Both preflights were cold SAFE-TO-MEASURE with compiler
fingerprint `8af25a9767dcf428`; unchanged compiler fingerprint is expected for a
runtime assembly fix. After-assembly preflight reinstalled the host tools.

The actual effective GOROOT here was `/data/source/projects/go-g5-port/go/worktrees/ppc32-o6cj`.
Overlay keys used that path. `strings` found `O6CJ netpoll kq/return/errno:` in both
probe binaries, and hardware printed it; no silent overlay no-op was trusted.

`go tool objdump` from this worktree disassembled the two discriminating probes
and all five fixed test binaries. `isa_floor_census.py` reported **zero illegal
G4 7450 instructions and zero undecoded words** in their function/code census:

| Binary | Instruction words |
| --- | ---: |
| probe-before2 | 404500 |
| probe-after | 404505 |
| runtime-pprof.test | 715958 |
| syscall.test | 604471 |
| net.test | 1050587 |
| os.test | 928234 |
| runtime.test | 1977879 |

Only `_go:buildid` was excluded from objdump's function blocks: it is a data
string, not executable code. The original unfiltered disassemblies and censuses
are retained too; interpreting its ASCII bytes as instructions produced false
violations. The real executed binaries above are positive examples; an explicit
synthetic NOP control yields zero illegal instructions, and a deliberately broken
64-bit `tdi` encoding (`0x0a20ff00`) yields one illegal G4 instruction. The census
also runs its existing injected instruction controls. No newer-CPU emulator was
used to infer legality.

## Reproducibility and retained bytes

Sealed evidence archive SHA-256:
`47da7a1d66ab4105933ad0fddcf4511533e16a88141f2de39508a34ae560ddd7`.
It contains sources/overlays, all built binaries, compressed unfiltered and
code-only disassemblies, census outputs, preflight status JSONs, wrapper logs,
per-test logs, regression commands (`regressions.json` / `run-regressions.py`),
and file identities (`sha256.json`). Host location: `/data/o6cj/evidence.tar.gz`;
a copy accompanies the lane's authoritative RESULT artifact. SHA-256 identities:

* probe-before2: `dfdf0722ae80c3da9719c81e682d95900b0346836bc8e4911bddeebdb12d440b`
* probe-after: `644803cbae6f8d38690534916b211fcde251168fadcfcaad0b1f70a77e84398d`

From the project root, reproduce host builds with the recorded paths (or regenerate
overlay keys if relocating):

```
PREFLIGHT_TARGET_HOST=ibook python3 probes/ppc-mac/preflight.py \
  --worktree go/worktrees/ppc32-o6cj --variant o6cj-after --no-prune
ROOT=$(realpath go/worktrees/ppc32-o6cj)
OUT=$(realpath out/ppc-mac/o6cj)
GOROOT=$ROOT GOCACHE=/data/o6cj/cache-after TMPDIR=/data/o6cj/tmp \
  GOOS=darwin GOARCH=ppc CGO_ENABLED=0 GO111MODULE=off GOTOOLCHAIN=local \
  "$ROOT/bin/go" build -a -overlay "$OUT/overlay.json" \
  -o "$OUT/probe-after" "$OUT/probe.go"
# Test binaries: same environment, fresh cache per package, no overlay:
# "$ROOT/bin/go" test -a -c -o "$OUT/<name>.test" <package>
scp "$OUT/probe-after" ibook:/tmp/o6cj/
ssh ibook 'sh -s -- --cpu 300 --wall 600 --prefix o6cj-after \
  --out /tmp/o6cj/after.log -- /tmp/o6cj/probe-after' \
  < probes/ppc-mac/bounded-run.sh
```

The diagnostic overlay source and complete probe are archived, not installed in
the runtime. No target run was unwrapped. Workload exit codes were obtained
directly, never from a pipeline. `/tmp/o6cj` was removed after log retrieval;
process inspection found no o6cj/bounded-run survivors. The owner's `top` was not
touched. Wrapper limits used: probes 300/600, tests 1800/2000; no gdb leg needed.

Review gate remains required; the separate pprof unwind failure means this is
not a clean full-suite result, despite the discriminating fix controls passing.
Its before/after provenance was not investigated here; no causal attribution is
claimed from the different fatal message alone.
