// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package runtime_test

import (
	"bytes"
	"fmt"
	"internal/profile"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// goppc-452y, goppc-nzqrh. ppc32 has no native 64-bit atomic: every 64-bit
// operation in internal/runtime/atomic takes a spinlock chosen by the operand's
// address from a table of 57 stripes (it was one global lock, lock64, until
// goppc-nzqrh), and the stripes are not reentrant. The SIGPROF handler writes
// its sample into a profBuf, whose indices are 64-bit atomics, so a SIGPROF
// that lands on a thread holding the stripe of one of those indices waits for
// the lock forever: the owner is the code the handler interrupted and cannot run
// again until the handler returns.
//
// Both children below run in a subprocess because the failure is a hang that no
// in-process timeout can recover. The parent bounds each with a wall-clock kill
// and reports the child's user and system CPU: a spin in the signal handler is
// USER time (about the wall time, system near zero). That is how it is told
// apart from goppc-0v7x, a kernel re-fault loop that burns SYSTEM time.
// The named mutant for these tests is "M-noguard": the tree with the
// `if atomic64AnyStripeHeld() { cpuprof.lostAtomic++; return }` statement removed
// from sigprof in proc.go and nothing else changed. TestSIGPROFWhileLock64Held
// hangs on it every time (3 of 3 on a G4, goppc-452y, under the global lock); the
// profile test hangs on it when a signal lands on a thread holding a stripe the
// handler needs.
const (
	lock64HoldArg    = "lock64HoldAcrossSIGPROF"
	lock64ProfileArg = "lock64ProfileUnderLoad"
	lock64ControlArg = "lock64ProfileWithoutAtomic64"
	lock64OtherArg   = "lock64HoldOtherStripeAcrossSIGPROF"
	lock64RHeldArg   = "lock64HoldReadIndexStripeAcrossSIGPROF"
	lock64DoneMarker = "lock64: survived"
	lock64ChildWall  = 60 * time.Second
)

func init() {
	if len(os.Args) != 2 {
		return
	}
	switch os.Args[1] {
	case lock64HoldArg:
		lock64HoldChild()
	case lock64RHeldArg:
		lock64RHeldChild()
	case lock64OtherArg:
		lock64HoldOtherChild()
	case lock64ProfileArg:
		lock64ProfileChild(true)
	case lock64ControlArg:
		lock64ProfileChild(false)
	}
}

// lock64HoldChild holds the stripe of the profile buffer's write index, which
// the handler must take, while SIGPROF is delivered to this thread.
func lock64HoldChild() {
	runtime.LockOSThread()
	runtime.SetCPUProfileRate(100)
	before := runtime.ProfLostAtomic()
	held := runtime.HoldAtomic64StripeAcrossSIGPROF(true)
	after := runtime.ProfLostAtomic()
	runtime.SetCPUProfileRate(0)
	fmt.Printf("%s stripe=%d profbuf_uses_stripe=%v lostAtomic_before=%d lostAtomic_after=%d\n",
		lock64DoneMarker, held.Stripe, held.ProfBufUses, before, after)
	if after <= before {
		fmt.Println("lock64: the signal was not counted as lost")
		os.Exit(3)
	}
	os.Exit(0)
}

// lock64RHeldChild holds the stripe of the profile buffer's READ index r while
// SIGPROF is delivered to this thread. The handler's cpuprof.add needs r's stripe
// as well as w's (goppc-l3r5t). It is the discriminator for a sigprof guard that
// checks fewer stripes than cpuprof.add takes: such a guard sees w's stripe free
// and the handler then waits for r's stripe, held by the thread it interrupted.
//
// Preconditions, printed BEFORE the signal so that a hang still carries them:
// r's and w's stripes differ. The delivery witness is the lost-sample count: the
// handler ran and dropped the sample, so the count must rise across the hook. A
// signal that never arrived would exit 3 here, not hang; a hang is therefore
// the guard's omission and nothing else.
func lock64RHeldChild() {
	runtime.LockOSThread()
	runtime.SetCPUProfileRate(100)
	plan := runtime.PlanRHeld()
	fmt.Printf("lock64: r-held plan r_stripe=%d w_stripe=%d distinct=%v\n", plan.R, plan.W, plan.Distinct)
	if !plan.Distinct {
		fmt.Println("lock64: r and w share a stripe; this is not an r-held discriminator")
		os.Exit(3)
	}
	before := runtime.ProfLostAtomic()
	plan.HoldAcrossSIGPROF()
	after := runtime.ProfLostAtomic()
	runtime.SetCPUProfileRate(0)
	fmt.Printf("%s r-held lostAtomic_before=%d lostAtomic_after=%d\n", lock64DoneMarker, before, after)
	if after <= before {
		fmt.Println("lock64: the signal was not counted as lost")
		os.Exit(3)
	}
	os.Exit(0)
}

// lock64HoldOtherChild holds a stripe that no profile-buffer word hashes to.
// The handler does not need it, but sigprof's guard asks about any stripe, so
// the sample is dropped and counted: a conservative cross-stripe drop. The child
// reports what happened and fails only on a hang or a crash; it does not
// require the drop, so a future guard that is exact per stripe is not forbidden.
func lock64HoldOtherChild() {
	runtime.LockOSThread()
	runtime.SetCPUProfileRate(100)
	before := runtime.ProfLostAtomic()
	held := runtime.HoldAtomic64StripeAcrossSIGPROF(false)
	after := runtime.ProfLostAtomic()
	runtime.SetCPUProfileRate(0)
	fmt.Printf("%s stripe=%d profbuf_uses_stripe=%v lostAtomic_before=%d lostAtomic_after=%d\n",
		lock64DoneMarker, held.Stripe, held.ProfBufUses, before, after)
	if held.ProfBufUses {
		fmt.Println("lock64: the hook chose a stripe the profile buffer uses")
		os.Exit(3)
	}
	os.Exit(0)
}

var lock64Sink uint64

// lock64ProfileChild profiles a program that does 64-bit atomics on several
// goroutines amid ordinary compute, and reports what the profile holds. With
// atomics false it is the control: the same program without the 64-bit
// operations, which must lose no sample.
func lock64ProfileChild(atomics bool) {
	var buf bytes.Buffer
	if err := pprof.StartCPUProfile(&buf); err != nil {
		fmt.Println("lock64: StartCPUProfile:", err)
		os.Exit(2)
	}
	var wg sync.WaitGroup
	var counter uint64
	stop := time.Now().Add(4 * time.Second)
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x := uint64(1)
			for time.Now().Before(stop) {
				for i := 0; i < 2000; i++ {
					x = x*6364136223846793005 + 1442695040888963407
				}
				if atomics {
					atomic.AddUint64(&counter, x&1)
					lock64Sink += atomic.LoadUint64(&counter)
				}
			}
		}()
	}
	wg.Wait()
	pprof.StopCPUProfile()
	p, err := profile.Parse(&buf)
	if err != nil {
		fmt.Println("lock64: profile.Parse:", err)
		os.Exit(2)
	}
	var total, lost int64
	flat := map[string]int64{}
	for _, s := range p.Sample {
		n := s.Value[0]
		total += n
		if len(s.Location) > 0 && len(s.Location[0].Line) > 0 {
			flat[s.Location[0].Line[0].Function.Name] += n
		}
		for _, loc := range s.Location {
			for _, line := range loc.Line {
				switch {
				case strings.Contains(line.Function.Name, "_LostSIGPROFDuringAtomic64"):
					lost += n
				}
			}
		}
	}
	fmt.Printf("%s samples=%d lost_during_atomic64=%d\n", lock64DoneMarker, total, lost)
	for name, n := range flat {
		fmt.Printf("lock64: flat %5d %s\n", n, name)
	}
	switch {
	case total-lost <= 0:
		fmt.Println("lock64: the profile holds no sample besides the lost ones")
		os.Exit(3)
	case !atomics && lost != 0:
		fmt.Println("lock64: a program without 64-bit atomics lost samples")
		os.Exit(3)
	case atomics && lost == 0:
		fmt.Println("lock64: no sample was lost while lock64 was held (workload did not exercise the guard)")
		os.Exit(3)
	}
	os.Exit(0)
}

// runLock64Child runs this test binary as the named child and kills it after
// lock64ChildWall. It returns the combined output and, on a kill, the child's CPU split.
func runLock64Child(t *testing.T, arg string) (out string, killed bool, user, sys time.Duration, err error) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe, arg)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", arg, err)
	}
	var wasKilled atomic.Bool
	timer := time.AfterFunc(lock64ChildWall, func() { wasKilled.Store(true); cmd.Process.Kill() })
	werr := cmd.Wait()
	timer.Stop()
	killed = wasKilled.Load()
	if st := cmd.ProcessState; st != nil {
		user, sys = st.UserTime(), st.SystemTime()
	}
	return buf.String(), killed, user, sys, werr
}

// TestSIGPROFWhileLock64Held delivers SIGPROF to a thread that holds lock64 and
// requires the handler to return. Before the fix the child spins forever.
func TestSIGPROFWhileLock64Held(t *testing.T) {
	out, killed, user, sys, err := runLock64Child(t, lock64HoldArg)
	t.Logf("child output:\n%s", out)
	if killed {
		t.Fatalf("child HUNG: killed after %v wall; user=%v sys=%v. User time about the wall time and "+
			"system time near zero is a spin in the SIGPROF handler on lock64 (goppc-452y); "+
			"system time about the wall time would be a kernel re-fault loop (goppc-0v7x)", lock64ChildWall, user, sys)
	}
	if err != nil || !strings.Contains(out, lock64DoneMarker) {
		t.Fatalf("child failed (%v) or did not report survival:\n%s", err, out)
	}
}

// TestSIGPROFWhileReadIndexStripeHeld delivers SIGPROF to a thread that holds the
// stripe of profBuf.r, distinct from the stripe of profBuf.w. The existing
// required-stripe control holds w's stripe and the other-stripe control holds
// neither, so a guard that checks only w's stripe passes both; this one hangs on
// it, and is killed by the unchanged child wall. A hang with the precondition
// line present is the intended failure.
func TestSIGPROFWhileReadIndexStripeHeld(t *testing.T) {
	out, killed, user, sys, err := runLock64Child(t, lock64RHeldArg)
	t.Logf("child output:\n%s", out)
	if !strings.Contains(out, "lock64: r-held plan r_stripe=") || !strings.Contains(out, "distinct=true") {
		t.Fatalf("child did not establish the distinct r/w stripe precondition:\n%s", out)
	}
	if killed {
		t.Fatalf("child HUNG: killed after %v wall with r's stripe held and the precondition established; user=%v sys=%v. "+
			"User time about the wall time is a spin in the SIGPROF handler on a stripe the guard did not cover", lock64ChildWall, user, sys)
	}
	if err != nil || !strings.Contains(out, lock64DoneMarker) {
		t.Fatalf("child failed (%v) or did not report survival:\n%s", err, out)
	}
}

// TestSIGPROFWhileOtherStripeHeld delivers SIGPROF to a thread that holds a
// stripe the handler does not need. It must return. The log shows whether the
// sample was dropped (it is, under the any-stripe guard).
func TestSIGPROFWhileOtherStripeHeld(t *testing.T) {
	out, killed, user, sys, err := runLock64Child(t, lock64OtherArg)
	t.Logf("child output:\n%s", out)
	if killed {
		t.Fatalf("child HUNG: killed after %v wall; user=%v sys=%v", lock64ChildWall, user, sys)
	}
	if err != nil || !strings.Contains(out, lock64DoneMarker) {
		t.Fatalf("child failed (%v) or did not report survival:\n%s", err, out)
	}
}

// TestCPUProfileSamplesWithAtomic64 requires a profile of a workload that uses
// 64-bit atomics to complete AND to contain samples besides the lost ones:
// dropping the samples that find lock64 held must not drop the profile. The
// control child runs the same workload without 64-bit atomics and must lose none.
//
// The test does not assert that samples land in the workload's own functions. On
// darwin/ppc the kernel delivers the process-wide profiling signal to a thread
// parked in pthread_cond_wait, so both children attribute their samples there;
// that holds with or without 64-bit atomics and is not what this test is about.
func TestCPUProfileSamplesWithAtomic64(t *testing.T) {
	for _, arg := range []string{lock64ProfileArg, lock64ControlArg} {
		t.Run(arg, func(t *testing.T) {
			out, killed, user, sys, err := runLock64Child(t, arg)
			t.Logf("child output:\n%s", out)
			if killed {
				t.Fatalf("child HUNG: killed after %v wall; user=%v sys=%v (goppc-452y if user time dominates)",
					lock64ChildWall, user, sys)
			}
			if err != nil || !strings.Contains(out, lock64DoneMarker) {
				t.Fatalf("child failed (%v) or did not report a profile:\n%s", err, out)
			}
		})
	}
}
