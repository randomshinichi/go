// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package runtime_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"internal/testenv"
)

// This file is the oracle behind the fatal-path witnesses in
// badsystemstack_test.go (goppc-0v7x). The witnesses re-exec the test binary into
// a child that is going to crash on purpose, and the question the parent has to
// answer is not "did the child die" but "did the child finish its diagnostic and
// terminate by itself". Three observations look alike from outside and must not
// be confused:
//
//   - a runtime wedge: the fatal report faulted again and the fault was never
//     delivered (on darwin the kernel just retries the faulting instruction,
//     forever, in system time). The child is alive and consuming CPU.
//   - a starved child: a loaded one-CPU target never gave the child the CPU it
//     needs. The child is alive and has consumed almost no CPU.
//   - a counterfeit: the child prints the first line of the report and then
//     exits some other way. It is dead, and it did not finish the report.
//
// Wall-clock expiry alone cannot tell the first two apart, so the parent reads
// back the CPU the child actually consumed (user plus system, from the wait
// status). Only a child that has spent its whole CPU budget without exiting is a
// wedge. A child that was still alive when the wall budget ran out with its CPU
// budget unspent says nothing about the runtime, and the test says so and skips.
//
// The child also sets RLIMIT_CPU on itself, as a backstop that ends a wedge
// early. That backstop is NOT relied on: Linux enforces it (the host run of the
// spinner counterfeit dies of it after the hard limit), but xnu-1228 on the iBook
// was observed not to - a child that had consumed 52 CPU-seconds against a 10 s
// limit was still running when the parent's wall timer killed it. So on Leopard
// the wall budget is what ends a wedge, and the CPU figure is what names it.

// Checked by the witnesses in badsystemstack_test.go and by the oracle self-test.
const (
	badssFatalLine    = "fatal: systemstack called from unexpected goroutine"
	badssReturnedLine = "badss: returned without taking the fatal branch"
	badssControlLine  = "badss: control returned from systemstack"
)

var badssFaultLine = func() string {
	if runtime.GOOS == "linux" {
		return "SIGSEGV: segmentation violation"
	}
	return "SIGBUS: bus error"
}()

const (
	// badssReadyLine is the first thing every child writes, after it has
	// installed its CPU limit and before it does anything that can fault. The
	// parent starts its wall timer when it sees this line, not at process
	// creation, so the exec and loader time of a loaded target is not charged to
	// the fault path.
	badssReadyLine = "badss: ready"

	// badssCPUEnv carries the child's CPU budget in whole seconds.
	badssCPUEnv = "BADSS_CHILD_CPU_SECS"

	// badssPrologueFailedExit is the exit status of a child that could not
	// install its CPU limit. It is outside every status the fatal ladder uses.
	badssPrologueFailedExit = 96
)

// The counterfeit children below are ONLY used by TestBadssOracleCounterfeits, to
// prove the oracle can reject what it claims to reject. None of them touches the
// runtime fault path.
const (
	badssCounterfeitPrefixExit3Arg = "badssCounterfeitPrefixExit3"
	badssCounterfeitFullExit3Arg   = "badssCounterfeitFullTextExit3"
	badssCounterfeitCompleteArg    = "badssCounterfeitComplete"
	badssCounterfeitReturnedArg    = "badssCounterfeitReturned"
	badssCounterfeitSleeperArg     = "badssCounterfeitSleeper"
	badssCounterfeitSpinnerArg     = "badssCounterfeitSpinner"
	badssCounterfeitNeverReadyArg  = "badssCounterfeitNeverReady"
)

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case badssCounterfeitPrefixExit3Arg:
		// The motivating counterfeit: the first line of the fatal report, then a
		// plain exit 3. The rest of the report is lost.
		badssChildReady()
		os.Stderr.WriteString(badssFatalLine)
		os.Exit(3)
	case badssCounterfeitFullExit3Arg:
		// Every marker of the nested report, in order, and still the wrong exit.
		badssChildReady()
		os.Stderr.WriteString(badssNestedReport)
		os.Exit(3)
	case badssCounterfeitCompleteArg:
		// Positive control for the oracle: exactly the shape of a good witness.
		badssChildReady()
		os.Stderr.WriteString(badssNestedReport)
		os.Exit(4)
	case badssCounterfeitReturnedArg:
		badssChildReady()
		os.Stderr.WriteString(badssReturnedLine + "\n")
		os.Exit(97)
	case badssCounterfeitSleeperArg:
		// Alive, ready, and using no CPU at all: a perfectly starved child.
		badssChildReady()
		time.Sleep(time.Hour)
	case badssCounterfeitSpinnerArg:
		// Alive, ready, and burning CPU without ever finishing: a wedge.
		badssChildReady()
		for {
		}
	case badssCounterfeitNeverReadyArg:
		// Never reaches the ready marker.
		time.Sleep(time.Hour)
	}
}

// badssWitnessExpect is what the taken-branch witness must look like when it is
// healthy: systemstack_bad prints its message, abort() faults, the report walks a
// stack it cannot (the unregistered copy has no saved sp), that nested fault is
// delivered because the fault signals were unblocked, and the m.dying ladder in
// startpanic_m (panic.go) prints "panic during panic" on the second entry and
// "stack trace unavailable" on the third, then exit(4). Every marker is
// terminal evidence that the ladder ran to its end.
var badssWitnessExpect = badssExpect{
	exitCode:  4,
	markers:   []string{badssFatalLine, badssFaultLine, "panic during panic", badssFaultLine, "stack trace unavailable"},
	forbidden: []string{badssReturnedLine},
}

// badssCounterfeitFaultLine describes the actual report in the positive-control
// child independently of the oracle's expected string. Keeping these separate
// lets TestBadssOracleCounterfeits detect a wrong platform expectation.
var badssCounterfeitFaultLine = func() string {
	if runtime.GOOS == "linux" {
		return "SIGSEGV: segmentation violation"
	}
	return "SIGBUS: bus error"
}()

// badssNestedReport is the shape of the report the nested-fault witness produces
// when the fatal ladder runs to the end: the original message, the original
// fault, the nested fault, and the ladder's last line, in that order.
var badssNestedReport = badssFatalLine +
	badssCounterfeitFaultLine + "\nPC=0x1 m=0 sigcode=1 addr=0x0\n\ngoroutine 1 [running]:\n" +
	"panic during panic\n" +
	badssCounterfeitFaultLine + "\nPC=0x2 m=0 sigcode=1 addr=0x28\n\n" +
	"stack trace unavailable\n"

// badssChildReady is the child's prologue: set a CPU limit on this process, then
// announce readiness. The soft limit is the budget; the hard limit is two seconds
// later, because a Go process ignores the SIGXCPU sent at the soft limit and the
// kernel's SIGKILL at the hard limit is what ends a spinner where the kernel
// enforces the limit at all (see the file comment: xnu-1228 does not).
func badssChildReady() {
	secs, err := strconv.Atoi(os.Getenv(badssCPUEnv))
	if err != nil || secs <= 0 {
		os.Stderr.WriteString("badss: " + badssCPUEnv + " missing or invalid\n")
		os.Exit(badssPrologueFailedExit)
	}
	lim := syscall.Rlimit{Cur: uint64(secs), Max: uint64(secs) + 2}
	if err := syscall.Setrlimit(syscall.RLIMIT_CPU, &lim); err != nil {
		os.Stderr.WriteString("badss: setrlimit(RLIMIT_CPU): " + err.Error() + "\n")
		os.Exit(badssPrologueFailedExit)
	}
	os.Stderr.WriteString(badssReadyLine + "\n")
}

// badssBudget is what a child may consume before the parent stops waiting.
type badssBudget struct {
	// cpu is the child's own CPU limit: user plus system seconds it may spend.
	cpu time.Duration
	// startup is the wall time allowed from process creation to the ready line.
	startup time.Duration
	// wall is the wall time allowed from the ready line to termination.
	wall time.Duration
}

// badssDefaultBudget is generous on purpose. A healthy witness needs about 0.2 s
// of wall on an idle iBook G4; the wedge needs the whole CPU budget before it can
// be called one, and the wall budget is long enough to deliver that budget at
// roughly a tenth of a CPU. Past that the target is too loaded to say anything.
var badssDefaultBudget = badssBudget{
	cpu:     10 * time.Second,
	startup: 60 * time.Second,
	wall:    120 * time.Second,
}

// badssBudgetScaleEnv, when set to a positive number, multiplies the startup and
// wall budgets of the witnesses. It exists to drill the starvation verdict on a
// real wedging runtime: a scale of 0.05 gives the wedge about six seconds of wall,
// which on a one-CPU target is less CPU than the budget, which is exactly what a
// CPU-starved child looks like, and the witness must then report
// INFRASTRUCTURE-INCONCLUSIVE and not the regression. Unset in every normal run.
const badssBudgetScaleEnv = "BADSS_BUDGET_SCALE"

// badssWitnessBudget is the budget the witnesses run their children with.
func badssWitnessBudget(t *testing.T) badssBudget {
	t.Helper()
	b := badssDefaultBudget
	if v := os.Getenv(badssBudgetScaleEnv); v != "" {
		scale, err := strconv.ParseFloat(v, 64)
		if err != nil || scale <= 0 {
			t.Fatalf("%s=%q is not a positive number", badssBudgetScaleEnv, v)
		}
		b.startup = time.Duration(float64(b.startup) * scale)
		b.wall = time.Duration(float64(b.wall) * scale)
		t.Logf("DRILL: %s=%s: startup budget %v, wall budget %v", badssBudgetScaleEnv, v, b.startup, b.wall)
	}
	return b
}

// badssSink collects a child's stderr and notices the ready line.
type badssSink struct {
	mu    sync.Mutex
	buf   bytes.Buffer
	seen  bool
	ready chan struct{}
}

func newBadssSink() *badssSink { return &badssSink{ready: make(chan struct{})} }

func (s *badssSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf.Write(p)
	if !s.seen && bytes.Contains(s.buf.Bytes(), []byte(badssReadyLine)) {
		s.seen = true
		close(s.ready)
	}
	return len(p), nil
}

func (s *badssSink) isReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen
}

func (s *badssSink) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// badssRun is everything observed about one child.
type badssRun struct {
	arg    string
	budget badssBudget
	out    string
	state  *os.ProcessState
	// ready reports that the child printed badssReadyLine.
	ready bool
	// startupExpired is true if the startup budget ran out before the ready line.
	startupExpired bool
	// killedByUs is true only if the parent's Kill succeeded AND the child's wait
	// status says it died of that SIGKILL. A child that exited by itself in the
	// instant the deadline fired is not killed, whatever the timer says.
	killedByUs bool
	// cpu is the user plus system time the child consumed.
	cpu time.Duration
	// startupElapsed and runElapsed are the wall time to the ready line and from
	// it to termination.
	startupElapsed, runElapsed time.Duration
}

// runBadssChild re-execs this test binary in one of the child modes.
func runBadssChild(t *testing.T, arg string, budget badssBudget) badssRun {
	t.Helper()
	return runBadssCommand(t, arg, budget, testenv.Executable(t), arg)
}

// runBadssCommand runs name with args under the same two-phase timer as the child
// modes. what labels the run in reports. The command must either be a child mode
// or, like gdb, run one and pass its stderr through.
func runBadssCommand(t *testing.T, what string, budget badssBudget, name string, args ...string) badssRun {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd = testenv.CleanCmdEnv(cmd)
	cmd.Env = append(cmd.Env, badssCPUEnv+"="+strconv.Itoa(int(budget.cpu/time.Second)))
	// The fatal path and every marker here write to fd 2. A debugger prints its own
	// messages on fd 1 and the child's output interleaves with them, so both go to
	// one sink (exec shares one copier when the writers are identical).
	sink := newBadssSink()
	cmd.Stderr = sink
	cmd.Stdout = sink
	created := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting %s: %v", what, err)
	}
	done := make(chan struct{})
	go func() { cmd.Wait(); close(done) }()

	r := badssRun{arg: what, budget: budget}
	killSent := false
	kill := func() {
		killSent = cmd.Process.Kill() == nil
		<-done
	}

	// Phase 1: process creation to the ready line.
	startupTimer := time.NewTimer(budget.startup)
	select {
	case <-sink.ready:
		r.ready = true
	case <-done:
	case <-startupTimer.C:
		// The ready line and the timer can both be due; the line wins.
		if sink.isReady() {
			r.ready = true
		} else {
			r.startupExpired = true
			kill()
		}
	}
	startupTimer.Stop()
	if !r.ready && !r.startupExpired {
		// The child ended before the select noticed its ready line. Wait has
		// returned, so the sink is complete.
		r.ready = sink.isReady()
	}
	readyAt := time.Now()
	r.startupElapsed = readyAt.Sub(created)

	// Phase 2: the ready line to termination. Nothing before the ready line is
	// charged to this timer.
	if r.ready {
		wallTimer := time.NewTimer(budget.wall)
		select {
		case <-done:
		case <-wallTimer.C:
			kill()
		}
		wallTimer.Stop()
		r.runElapsed = time.Since(readyAt)
	}
	<-done

	r.out = sink.String()
	r.state = cmd.ProcessState
	r.cpu = r.state.UserTime() + r.state.SystemTime()
	if ws, ok := r.state.Sys().(syscall.WaitStatus); ok {
		r.killedByUs = killSent && ws.Signaled() && ws.Signal() == syscall.SIGKILL
	}
	return r
}

// badssExpect is what a child that terminated by itself must have looked like.
type badssExpect struct {
	// exitCode is the exit status the child must have ended with; negative means
	// the status is not graded.
	exitCode int
	// markers must all appear in out, in this order.
	markers []string
	// forbidden must not appear in out.
	forbidden []string
	// once must each appear exactly once: the original fault is attributed one
	// time, not replaced or repeated by a later one.
	once []string
}

const (
	badssPass         = "PASS"
	badssFail         = "FAIL"
	badssInconclusive = "INFRASTRUCTURE-INCONCLUSIVE"
)

// badssVerdict is the oracle's answer. oracles names every check that rejected the
// child, so a report can say which one caught which counterfeit.
type badssVerdict struct {
	status  string
	oracles []string
	detail  string
}

func (v badssVerdict) has(oracle string) bool {
	for _, o := range v.oracles {
		if o == oracle {
			return true
		}
	}
	return false
}

func (v badssVerdict) String() string {
	if v.status == badssPass {
		return badssPass
	}
	return fmt.Sprintf("%s [oracle: %s]: %s", v.status, strings.Join(v.oracles, ", "), v.detail)
}

// orderedMarkers reports the first marker that is missing, or out of order.
func orderedMarkers(out string, markers []string) (missing string, ok bool) {
	pos := 0
	for _, m := range markers {
		i := strings.Index(out[pos:], m)
		if i < 0 {
			return m, false
		}
		pos += i + len(m)
	}
	return "", true
}

// judgeBadss classifies one child against what it was supposed to do.
func judgeBadss(r badssRun, e badssExpect) badssVerdict {
	for _, marker := range e.markers {
		if marker == "" {
			return badssVerdict{badssFail, []string{"empty-marker"},
				"expected report contains an empty marker, which would match any output"}
		}
	}
	spent := r.cpu >= r.budget.cpu*3/4
	facts := fmt.Sprintf("child %s: cpu %v of budget %v, startup %v of %v, run %v of %v",
		r.arg, r.cpu.Round(time.Millisecond), r.budget.cpu, r.startupElapsed.Round(time.Millisecond),
		r.budget.startup, r.runElapsed.Round(time.Millisecond), r.budget.wall)
	ws, _ := r.state.Sys().(syscall.WaitStatus)

	switch {
	case !r.ready && r.startupExpired:
		return badssVerdict{badssInconclusive, []string{"startup-budget"},
			facts + ": no ready line before the startup budget ran out; nothing was learned about the fault path"}
	case !r.ready:
		return badssVerdict{badssFail, []string{"never-ready"},
			facts + ": the child ended without ever printing the ready line: " + r.state.String()}
	case r.killedByUs && spent:
		return badssVerdict{badssFail, []string{"wedge"},
			facts + ": the wall budget ran out with the whole CPU budget spent and the child still alive: " +
				"the fatal report did not terminate (goppc-0v7x)"}
	case r.killedByUs:
		return badssVerdict{badssInconclusive, []string{"wall-budget-cpu-unspent"},
			facts + ": the wall budget ran out with the CPU budget unspent; the child was starved of CPU, " +
				"this is not evidence of a runtime wedge"}
	case ws.Signaled() && (ws.Signal() == syscall.SIGKILL || ws.Signal() == syscall.SIGXCPU) && spent:
		return badssVerdict{badssFail, []string{"wedge"},
			facts + ": the kernel ended the child with " + ws.Signal().String() +
				" after the child spent its whole CPU budget: the fatal report did not terminate (goppc-0v7x)"}
	case ws.Signaled():
		return badssVerdict{badssFail, []string{"unexpected-signal"},
			facts + ": the child was ended by " + ws.Signal().String() + " instead of exiting by itself"}
	}

	// The child exited by itself. Run every check so the report names all of them.
	var oracles []string
	var why []string
	reject := func(oracle, format string, args ...any) {
		oracles = append(oracles, oracle)
		why = append(why, oracle+": "+fmt.Sprintf(format, args...))
	}
	for _, f := range e.forbidden {
		if strings.Contains(r.out, f) {
			reject("forbidden-text", "found %q", f)
		}
	}
	if r.state.Success() && e.exitCode > 0 {
		reject("exit0", "the child exited 0")
	}
	if e.exitCode < 0 {
		// Exit status not graded: the command is a debugger, whose own status
		// says nothing about the program it ran. Its markers still are.
	} else if code := r.state.ExitCode(); code == badssPrologueFailedExit {
		reject("prologue", "the child could not install its CPU limit")
	} else if code != e.exitCode {
		reject("exit-code", "exit %d, want %d", code, e.exitCode)
	}
	if missing, ok := orderedMarkers(r.out, e.markers); !ok {
		reject("terminal-evidence", "marker %q missing or out of order (the report did not run to its end)", missing)
	}
	for _, o := range e.once {
		if n := strings.Count(r.out, o); n != 1 {
			reject("attribution", "%q appears %d times, want exactly once", o, n)
		}
	}
	if len(oracles) > 0 {
		return badssVerdict{badssFail, oracles, facts + ": " + strings.Join(why, "; ")}
	}
	return badssVerdict{status: badssPass, detail: facts}
}

// badssLoad is the host's load average line, for the report. Best effort.
func badssLoad() string {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return "uptime unavailable: " + err.Error()
	}
	return strings.TrimSpace(string(out))
}

// badssApply reports a verdict the way every witness does: a pass is logged, a
// rejection is an error naming the oracle, and an exhausted infrastructure budget
// is a skip - never a pass and never the regression.
func badssLinuxWitnessVerdict(r badssRun) badssVerdict {
	var oracles []string
	var why []string
	if r.killedByUs {
		oracles = append(oracles, "harness-killed")
		why = append(why, "child was killed by the harness")
	}
	ws, ok := r.state.Sys().(syscall.WaitStatus)
	if !ok || !ws.Signaled() || ws.Signal() != syscall.SIGSEGV {
		oracles = append(oracles, "termination")
		why = append(why, fmt.Sprintf("child status %v, want termination by SIGSEGV", r.state))
	}
	for _, marker := range []string{badssFatalLine, badssFaultLine, "runtime.abort()"} {
		if !strings.Contains(r.out, marker) {
			oracles = append(oracles, "report-markers")
			why = append(why, fmt.Sprintf("output missing %q", marker))
		}
	}
	if len(oracles) != 0 {
		return badssVerdict{badssFail, oracles, strings.Join(why, "; ")}
	}
	return badssVerdict{status: badssPass, detail: "child reached the fatal report and terminated by SIGSEGV"}
}

func badssApply(t *testing.T, what string, r badssRun, v badssVerdict) {
	t.Helper()
	t.Logf("%s: %s\n  load: %s\n  child output:\n%s", what, v, badssLoad(), r.out)
	switch v.status {
	case badssFail:
		t.Errorf("%s: oracle %s rejected: %s", what, strings.Join(v.oracles, ", "), v.detail)
	case badssInconclusive:
		t.Skipf("%s: %s: %s", what, badssInconclusive, v.detail)
	}
}

// TestBadssOracleCounterfeits exercises the oracle against children that are
// broken on purpose, because an oracle that has only ever seen a good child proves
// nothing. Each case names the oracle that must reject it, and the positive
// control must pass.
func TestBadssOracleCounterfeits(t *testing.T) {
	testenv.MustHaveExec(t)
	quick := badssBudget{cpu: 2 * time.Second, startup: 60 * time.Second, wall: 20 * time.Second}
	starved := badssBudget{cpu: 10 * time.Second, startup: 60 * time.Second, wall: 3 * time.Second}
	neverReady := badssBudget{cpu: 10 * time.Second, startup: 3 * time.Second, wall: 60 * time.Second}

	cases := []struct {
		name   string
		arg    string
		budget badssBudget
		status string
		// oracles that must all have rejected the child.
		oracles []string
		// hostile means the host being busy can legitimately turn the expected
		// verdict into an inconclusive one; the case is then skipped, not failed.
		hostile bool
	}{
		{"positive-control-complete-report-exit-4", badssCounterfeitCompleteArg, badssDefaultBudget, badssPass, nil, false},
		{"fatal-prefix-then-exit-3", badssCounterfeitPrefixExit3Arg, badssDefaultBudget, badssFail, []string{"exit-code", "terminal-evidence"}, false},
		{"full-report-text-then-exit-3", badssCounterfeitFullExit3Arg, badssDefaultBudget, badssFail, []string{"exit-code"}, false},
		{"returned-path", badssCounterfeitReturnedArg, badssDefaultBudget, badssFail, []string{"forbidden-text", "exit-code"}, false},
		{"alive-burning-cpu-never-exits-is-a-wedge", badssCounterfeitSpinnerArg, quick, badssFail, []string{"wedge"}, true},
		{"alive-using-no-cpu-is-starved-not-wedged", badssCounterfeitSleeperArg, starved, badssInconclusive, []string{"wall-budget-cpu-unspent"}, false},
		{"never-ready-is-startup-starved", badssCounterfeitNeverReadyArg, neverReady, badssInconclusive, []string{"startup-budget"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := runBadssChild(t, c.arg, c.budget)
			v := judgeBadss(r, badssWitnessExpect)
			t.Logf("verdict: %s (host load: %s)", v, badssLoad())
			if c.hostile && v.status == badssInconclusive {
				t.Skipf("host too busy to spend a %v CPU budget inside %v: %s", c.budget.cpu, c.budget.wall, v)
			}
			if v.status != c.status {
				t.Fatalf("verdict %s, want status %s\n%s", v, c.status, r.out)
			}
			if c.arg == badssCounterfeitPrefixExit3Arg && runtime.GOOS == "linux" {
				witness := badssLinuxWitnessVerdict(r)
				if witness.status != badssFail {
					t.Errorf("Linux witness accepted fatal-prefix counterfeit: %s", witness)
				}
				t.Logf("Linux witness rejects fatal-prefix counterfeit: %s", witness)
			}
			for _, o := range c.oracles {
				if !v.has(o) {
					t.Errorf("oracle %q did not reject this child: %s", o, v)
				}
			}
		})
	}
	t.Run("empty-expected-marker", func(t *testing.T) {
		r := runBadssChild(t, badssCounterfeitCompleteArg, badssDefaultBudget)
		expect := badssWitnessExpect
		expect.markers = []string{""}
		v := judgeBadss(r, expect)
		if v.status != badssFail || !v.has("empty-marker") {
			t.Fatalf("empty expected marker was not rejected: %s\n%s", v, r.out)
		}
	})
}
