// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// CPU profiling on darwin/ppc64 (Mac OS X 10.5 Leopard).
//
// Elsewhere on Unix the runtime arms the process-wide ITIMER_PROF timer and
// the kernel sends each SIGPROF to the thread that was using the CPU when the
// period expired, so the signal handler samples the right stack. The Leopard
// kernel (xnu-1228) does not. It raises the tick as a process-directed
// signal, and a process-directed signal goes to the first thread, in creation
// order, that does not block it. Measured on the G5, every tick reaches m0,
// whatever m0 is doing. When m0 is parked the profile describes an idle
// thread, or is empty if m0 last ran Go code before profiling started
// (sigprof ignores an M whose profilehz is 0). Later kernels send the tick to
// the running thread (psignal_try_thread in bsd_ast); see go.dev/issue/6047.
//
// So darwin/ppc64 does not use ITIMER_PROF. A sampler thread wakes twice per
// profiling period, reads each M's thread CPU time with the Mach call
// thread_info, and sends SIGPROF with pthread_kill to every M that has used
// a whole period of CPU since its last sample. That emulates the per-thread
// CPU timers Linux uses (timer_create on CLOCK_THREAD_CPUTIME_ID). The
// SIGPROF handler and sigprof are unchanged.

// mOSProf is the sampler's state for one M. Only the sampler thread uses it,
// with sched.lock held.
type mOSProf struct {
	port   uint32 // Mach port of the M's thread; 0 until the sampler looks it up
	epoch  uint32 // profSampler.epoch when cpu was last read
	cpu    int64  // thread CPU time in µs at the previous look
	credit int64  // CPU time in µs used and not yet sampled
}

// profSampler is protected by sched.lock.
var profSampler struct {
	hz      int32  // sampling rate; 0 when profiling is off
	epoch   uint32 // incremented each time profiling starts
	started bool   // the sampler thread exists
	parked  bool   // the sampler thread sleeps on wake
	wake    note
}

// setProcessCPUProfiler is called with prof.signalLock held.
func setProcessCPUProfiler(hz int32) {
	if hz != 0 {
		// setProcessCPUProfilerTimer installs the SIGPROF handler.
		// This port does not use the timer it also arms.
		setProcessCPUProfilerTimer(hz)
		setitimer(_ITIMER_PROF, &itimerval{}, nil)
	}
	start := false
	lock(&sched.lock)
	profSampler.hz = hz
	if hz != 0 {
		profSampler.epoch++
		start = !profSampler.started
		profSampler.started = true
		if profSampler.parked {
			profSampler.parked = false
			notewakeup(&profSampler.wake)
		}
	}
	unlock(&sched.lock)
	if start {
		newm(profSamplerLoop, nil, -1)
	}
	if hz == 0 {
		// The sampler rechecks hz under sched.lock before sending,
		// so it sends nothing more. Restore the previous handler.
		setProcessCPUProfilerTimer(0)
	}
}

func setThreadCPUProfiler(hz int32) {
	setThreadCPUProfilerHz(hz)
}

// profSamplerLoop is the body of the sampler thread. It runs on an M
// without a P, so it must not have write barriers.
//
//go:nowritebarrierrec
func profSamplerLoop() {
	lock(&sched.lock)
	sched.nmsys++
	checkdead()
	for {
		if profSampler.hz == 0 {
			noteclear(&profSampler.wake)
			profSampler.parked = true
			unlock(&sched.lock)
			notesleep(&profSampler.wake)
			lock(&sched.lock)
			continue
		}
		period := 1000000 / int64(profSampler.hz)
		unlock(&sched.lock)
		// Look twice per period so that one SIGPROF per look always
		// keeps up with a thread that runs continuously, even when
		// this thread wakes late.
		usleep(uint32(period / 2))
		lock(&sched.lock)
		if profSampler.hz != 0 {
			profSamplerTick(period)
		}
	}
}

// profSamplerTick sends SIGPROF to each M whose thread has used another
// period µs of CPU time. It is called with sched.lock held, which keeps the
// thread of every M on allm alive: mexit clears procid and then takes
// sched.lock to unlink the M before its thread exits.
//
//go:nowritebarrierrec
func profSamplerTick(period int64) {
	self := getg().m
	for mp := allm; mp != nil; mp = mp.alllink {
		if mp == self || mp.procid == 0 {
			continue
		}
		pr := &mp.prof
		if pr.port == 0 {
			pr.port = pthread_mach_thread_np(pthread(mp.procid))
		}
		cpu := threadCPUTime(pr.port)
		if cpu < 0 {
			continue
		}
		if pr.epoch != profSampler.epoch || mp.profilehz == 0 {
			// Profiling has just started, or this M is not
			// profiled (sigprof would drop its samples): CPU used
			// until now is not sampled.
			pr.epoch = profSampler.epoch
			pr.cpu = cpu
			pr.credit = 0
			continue
		}
		pr.credit += cpu - pr.cpu
		pr.cpu = cpu
		if pr.credit >= period {
			pr.credit -= period
			signalM(mp, _SIGPROF)
		}
	}
}

// threadBasicInfo is Mach's struct thread_basic_info (mach/thread_info.h):
// ten 32-bit words, the same on ppc and ppc64 (measured on the G5).
type threadBasicInfo struct {
	userSec, userUsec     int32
	systemSec, systemUsec int32
	cpuUsage              int32
	policy                int32
	runState              int32
	flags                 int32
	suspendCount          int32
	sleepTime             int32
}

const _THREAD_BASIC_INFO = 3

// threadCPUTime returns the user plus system CPU time of a thread in µs,
// or -1 if the kernel does not answer for that thread.
//
//go:nowritebarrierrec
func threadCPUTime(port uint32) int64 {
	var info threadBasicInfo
	count := uint32(unsafe.Sizeof(info) / 4)
	if thread_info(port, _THREAD_BASIC_INFO, &info, &count) != 0 {
		return -1
	}
	return (int64(info.userSec)+int64(info.systemSec))*1000000 + int64(info.userUsec) + int64(info.systemUsec)
}
