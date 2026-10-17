// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc

package atomic_test

import (
	"bytes"
	"fmt"
	"internal/cpu"
	"internal/runtime/atomic"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// Tests for the 57-stripe address-hashed lock behind every ppc32 64-bit atomic
// (goppc-nzqrh). A "colliding" pair of operands is one whose addresses are 57
// words apart, which hash to the same stripe; a "distinct" pair is adjacent
// words, which hash to adjacent stripes.

const (
	stripeHoldArg    = "stripeHoldDistinct"
	stripeHoldWall   = 30 * time.Second
	stripeHoldMarker = "stripe: distinct operand completed"
)

func init() {
	if len(os.Args) == 2 && os.Args[1] == stripeHoldArg {
		stripeHoldChild()
	}
}

// operandWords returns n 8-byte-aligned words from the heap.
func operandWords(t testing.TB, n int) []uint64 {
	w := make([]uint64, n)
	if uintptr(unsafe.Pointer(&w[0]))&7 != 0 {
		t.Fatalf("heap uint64 slice at %p is not 8-byte aligned", &w[0])
	}
	return w
}

func wantStripe(p *uint64) int { return int((uintptr(unsafe.Pointer(p)) >> 3) % 57) }

func TestStripeSelection(t *testing.T) {
	if atomic.Stripes != 57 {
		t.Fatalf("Stripes = %d, want 57", atomic.Stripes)
	}
	w := operandWords(t, 3*57)
	seen := map[int]bool{}
	for i := range w {
		got, want := atomic.StripeIndex(&w[i]), wantStripe(&w[i])
		if got != want {
			t.Fatalf("StripeIndex(%p) = %d, want (addr>>3)%%57 = %d", &w[i], got, want)
		}
		if i < 57 {
			seen[got] = true
		}
		if j := i + 57; j < len(w) && atomic.StripeIndex(&w[j]) != got {
			t.Fatalf("words %d and %d, 57 apart, hash to different stripes", i, j)
		}
		if j := i + 1; j < len(w) && atomic.StripeIndex(&w[j]) == got {
			t.Fatalf("adjacent words %d and %d hash to the same stripe", i, j)
		}
	}
	if len(seen) != 57 {
		t.Fatalf("57 consecutive words reach %d stripes, want 57", len(seen))
	}
}

func TestStripeLayout(t *testing.T) {
	for i := 1; i < atomic.Stripes; i++ {
		if d := atomic.StripeAddr(i) - atomic.StripeAddr(i-1); d != cpu.CacheLinePadSize {
			t.Fatalf("stripe %d is %d bytes after stripe %d, want %d", i, d, i-1, cpu.CacheLinePadSize)
		}
	}
	if cpu.CacheLinePadSize < 128 {
		t.Fatalf("CacheLinePadSize = %d, want at least the 970's 128", cpu.CacheLinePadSize)
	}
}

// TestStripeAPISequential checks the result of every 64-bit operation, old and
// new values, across the 32-bit word boundary.
func TestStripeAPISequential(t *testing.T) {
	w := operandWords(t, 1)
	p := &w[0]
	const lo = 0x00000000ffffffff
	const hi = 0xfffffffe00000001
	check := func(name string, got, want uint64) {
		t.Helper()
		if got != want {
			t.Fatalf("%s = %#x, want %#x", name, got, want)
		}
	}
	atomic.Store64(p, lo)
	check("Load64", atomic.Load64(p), lo)
	check("LoadAcq64", atomic.LoadAcq64(p), lo)
	check("Xadd64 carry", atomic.Xadd64(p, 1), 0x100000000)
	check("Load64 after Xadd64", *p, 0x100000000)
	check("Xadd64 negative", atomic.Xadd64(p, -1), lo)
	check("Xchg64 old", atomic.Xchg64(p, hi), lo)
	check("Xchg64 new", *p, hi)
	if atomic.Cas64(p, lo, 7) {
		t.Fatal("Cas64 with wrong old succeeded")
	}
	check("Cas64 failed leaves value", *p, hi)
	if !atomic.Cas64(p, hi, lo) {
		t.Fatal("Cas64 with right old failed")
	}
	check("Cas64 new", *p, lo)
	check("And64 old", atomic.And64(p, 0xffff0000ffff0000), lo)
	check("And64 new", *p, lo&0xffff0000ffff0000)
	check("Or64 old", atomic.Or64(p, 0x8000000000000001), lo&0xffff0000ffff0000)
	check("Or64 new", *p, lo&0xffff0000ffff0000|0x8000000000000001)
	atomic.StoreRel64(p, hi)
	check("StoreRel64", atomic.Load64(p), hi)
	i := (*int64)(unsafe.Pointer(p))
	atomic.Storeint64(i, -5)
	check("Loadint64", uint64(atomic.Loadint64(i)), uint64(0xfffffffffffffffb))
	check("Xaddint64", uint64(atomic.Xaddint64(i, 6)), 1)
	check("Xchgint64", uint64(atomic.Xchgint64(i, 9)), 1)
	if !atomic.Casint64(i, 9, 10) || atomic.Casint64(i, 9, 11) {
		t.Fatal("Casint64 result wrong")
	}
	check("Casint64 value", *p, 10)
	if atomic.AnyStripeHeld() {
		t.Fatal("a stripe is still held after the sequence")
	}
}

// opIndex is the operand goroutine g uses: with offset 0 there is only one.
func opIndex(offset, g int) int {
	if offset == 0 {
		return 0
	}
	return g & 1
}

// contended runs each of n goroutines iter times with GOMAXPROCS(n) and waits.
func contended(n, iter int, f func(g, j int)) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(n))
	var wg sync.WaitGroup
	for g := 0; g < n; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iter; j++ {
				f(g, j)
			}
		}()
	}
	wg.Wait()
}

// stripeCases are the operand arrangements every contention test runs under.
// Each returns the operands the goroutines share: for "same" every goroutine
// uses word 0; for "collide" goroutine g uses word 0 or word 57 (same stripe,
// different operand); for "distinct" word 0 or word 1.
var stripeCases = []struct {
	name   string
	offset int
}{
	{"same", 0},
	{"colliding", 57},
	{"distinct", 1},
}

func TestStripeXaddContended(t *testing.T) {
	const n, iter = 4, 2000
	for _, c := range stripeCases {
		t.Run(c.name, func(t *testing.T) {
			w := operandWords(t, 58)
			ops := [2]*uint64{&w[0], &w[c.offset]}
			// Each operand's returned new values must be exactly 1..total, each
			// once: no lost update and no duplicated or torn result.
			var mu sync.Mutex
			seen := [2]map[uint64]int{{}, {}}
			contended(n, iter, func(g, j int) {
				k := opIndex(c.offset, g)
				v := atomic.Xadd64(ops[k], 1)
				mu.Lock()
				seen[k][v]++
				mu.Unlock()
			})
			var total [2]uint64
			if c.offset == 0 {
				total[0] = n * iter
			} else {
				total[0], total[1] = n/2*iter, n/2*iter
			}
			for k := 0; k < 2; k++ {
				if k == 1 && c.offset == 0 {
					continue
				}
				if got := atomic.Load64(ops[k]); got != total[k] {
					t.Fatalf("operand %d final %d, want %d", k, got, total[k])
				}
				for v := uint64(1); v <= total[k]; v++ {
					if seen[k][v] != 1 {
						t.Fatalf("operand %d: Xadd64 returned %d %d times, want once", k, v, seen[k][v])
					}
				}
			}
		})
	}
}

func TestStripeCasContended(t *testing.T) {
	const n, iter = 4, 1000
	for _, c := range stripeCases {
		t.Run(c.name, func(t *testing.T) {
			w := operandWords(t, 58)
			ops := [2]*uint64{&w[0], &w[c.offset]}
			contended(n, iter, func(g, j int) {
				p := ops[g&1]
				for {
					old := atomic.Load64(p)
					if atomic.Cas64(p, old, old+0x100000001) {
						return
					}
				}
			})
			if c.offset == 0 {
				if got, want := *ops[0], uint64(n*iter)*0x100000001; got != want {
					t.Fatalf("final %#x, want %#x", got, want)
				}
			} else {
				want := uint64(n/2*iter) * 0x100000001
				if *ops[0] != want || *ops[1] != want {
					t.Fatalf("finals %#x %#x, want %#x each", *ops[0], *ops[1], want)
				}
			}
		})
	}
}

// TestStripeXchgContended passes a token: every Xchg64 must return a value some
// goroutine stored, and every stored value is returned exactly once, so the old
// values form a permutation.
func TestStripeXchgContended(t *testing.T) {
	const n, iter = 4, 1000
	for _, c := range stripeCases {
		t.Run(c.name, func(t *testing.T) {
			w := operandWords(t, 58)
			ops := [2]*uint64{&w[0], &w[c.offset]}
			var mu sync.Mutex
			got := [2]map[uint64]int{{}, {}}
			contended(n, iter, func(g, j int) {
				k := opIndex(c.offset, g)
				val := uint64(g*iter+j+1)<<32 | uint64(g*iter+j+1) // both words equal
				old := atomic.Xchg64(ops[k], val)
				if old>>32 != old&0xffffffff {
					t.Errorf("Xchg64 returned torn value %#x", old)
				}
				mu.Lock()
				got[k][old]++
				mu.Unlock()
			})
			for k := 0; k < 2; k++ {
				if k == 1 && c.offset == 0 {
					continue
				}
				for v, cnt := range got[k] {
					if cnt != 1 {
						t.Fatalf("operand %d: old value %#x returned %d times", k, v, cnt)
					}
				}
				// n*iter exchanges returned n*iter distinct olds; add the final
				// stored value and the initial zero is accounted for by being one of them.
				if want := n * iter; c.offset == 0 && len(got[0]) != want {
					t.Fatalf("distinct olds %d, want %d", len(got[0]), want)
				}
			}
		})
	}
}

// TestStripeBitOpsContended sets and clears per-goroutine bits spanning both
// words and checks the returned old values.
func TestStripeBitOpsContended(t *testing.T) {
	const n, iter = 4, 500
	for _, c := range stripeCases {
		t.Run(c.name, func(t *testing.T) {
			w := operandWords(t, 58)
			ops := [2]*uint64{&w[0], &w[c.offset]}
			contended(n, iter, func(g, j int) {
				p := ops[g&1]
				bit := uint64(1) << uint(g) // low word
				hibit := bit << 32          // high word
				if old := atomic.Or64(p, bit|hibit); old&(bit|hibit) != 0 {
					t.Errorf("g%d: Or64 old %#x already had own bits", g, old)
				}
				if v := atomic.Load64(p); v&(bit|hibit) != bit|hibit {
					t.Errorf("g%d: own bits missing after Or64: %#x", g, v)
				}
				if old := atomic.And64(p, ^(bit | hibit)); old&(bit|hibit) != bit|hibit {
					t.Errorf("g%d: And64 old %#x lost own bits", g, old)
				}
			})
			for k := 0; k < 2; k++ {
				if *ops[k] != 0 {
					t.Fatalf("operand %d final %#x, want 0", k, *ops[k])
				}
			}
		})
	}
}

// TestStripeNoTearing alternates two values whose halves differ; any observed
// mix of halves means a load or store ran outside the stripe.
func TestStripeNoTearing(t *testing.T) {
	const a, b = 0x00000000ffffffff, 0xffffffff00000000
	for _, c := range stripeCases {
		t.Run(c.name, func(t *testing.T) {
			w := operandWords(t, 58)
			p := &w[c.offset]
			atomic.Store64(p, a)
			contended(4, 2000, func(g, j int) {
				switch g {
				case 0:
					atomic.Store64(p, a)
				case 1:
					atomic.Store64(p, b)
				case 2:
					if old := atomic.Xchg64(p, a); old != a && old != b {
						t.Errorf("torn Xchg64 old %#x", old)
					}
				default:
					if v := atomic.Load64(p); v != a && v != b {
						t.Errorf("torn load %#x", v)
					}
				}
			})
		})
	}
}

// Recovered faults must happen before a stripe is taken and must not poison
// later operations.
func TestStripeFaultsDoNotPoison(t *testing.T) {
	var nilp *uint64
	x := make([]uint32, 4)
	mis := (*uint64)(unsafe.Pointer(uintptr(unsafe.Pointer(&x[0])) | 4))
	ops := map[string]func(p *uint64){
		"Load64":    func(p *uint64) { atomic.Load64(p) },
		"LoadAcq64": func(p *uint64) { atomic.LoadAcq64(p) },
		"Store64":   func(p *uint64) { atomic.Store64(p, 1) },
		"StoreRel":  func(p *uint64) { atomic.StoreRel64(p, 1) },
		"Xchg64":    func(p *uint64) { atomic.Xchg64(p, 1) },
		"Xadd64":    func(p *uint64) { atomic.Xadd64(p, 1) },
		"Cas64":     func(p *uint64) { atomic.Cas64(p, 0, 1) },
		"And64":     func(p *uint64) { atomic.And64(p, 1) },
		"Or64":      func(p *uint64) { atomic.Or64(p, 1) },
	}
	w := operandWords(t, 57)
	for name, op := range ops {
		for what, p := range map[string]*uint64{"nil": nilp, "misaligned": mis} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("%s(%s) did not panic", name, what)
					}
				}()
				op(p)
			}()
			if atomic.AnyStripeHeld() {
				t.Fatalf("%s(%s) left a stripe held", name, what)
			}
			// Every stripe must still be usable.
			for i := range w {
				atomic.Xadd64(&w[i], 1)
			}
		}
	}
	for i := range w {
		if w[i] == 0 {
			t.Fatalf("stripe %d never worked after the faults", i)
		}
	}
}

// TestStripeHeldIsObservable is the positive control for the hold hooks: a held
// stripe is visible in its lock word and to AnyStripeHeld, and no other.
func TestStripeHeldIsObservable(t *testing.T) {
	w := operandWords(t, 2)
	if atomic.AnyStripeHeld() {
		t.Fatal("a stripe is held before the test")
	}
	atomic.StripeAcquire(&w[0])
	held := atomic.StripeIndex(&w[0])
	for i := 0; i < atomic.Stripes; i++ {
		if (atomic.StripeWord(i) != 0) != (i == held) {
			t.Errorf("stripe %d word %d, held stripe is %d", i, atomic.StripeWord(i), held)
		}
	}
	if !atomic.AnyStripeHeld() {
		t.Error("AnyStripeHeld false while a stripe is held")
	}
	atomic.StripeRelease(&w[0])
	if atomic.AnyStripeHeld() {
		t.Error("AnyStripeHeld true after release")
	}
	atomic.StripeAcquireAll()
	for i := 0; i < atomic.Stripes; i++ {
		if atomic.StripeWord(i) != 1 {
			t.Errorf("AcquireAll left stripe %d free", i)
		}
	}
	atomic.StripeReleaseAll()
	if atomic.AnyStripeHeld() {
		t.Error("AnyStripeHeld true after ReleaseAll")
	}
}

// TestStripeEveryOpExcludesSameStripe: while the stripe of an operand is held,
// every 64-bit operation on that operand, and on a different operand that
// collides with it, must wait; once the stripe is released it must complete.
// A mutant that gives one operation a different stripe from the others
// completes immediately and fails, deterministically. The waiting side cannot
// fail by mistake: a correct operation cannot complete while the stripe is held.
func TestStripeEveryOpExcludesSameStripe(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))
	ops := []struct {
		name string
		f    func(p *uint64)
	}{
		{"Load64", func(p *uint64) { atomic.Load64(p) }},
		{"LoadAcq64", func(p *uint64) { atomic.LoadAcq64(p) }},
		{"Store64", func(p *uint64) { atomic.Store64(p, 1) }},
		{"StoreRel64", func(p *uint64) { atomic.StoreRel64(p, 1) }},
		{"Xchg64", func(p *uint64) { atomic.Xchg64(p, 1) }},
		{"Xadd64", func(p *uint64) { atomic.Xadd64(p, 1) }},
		{"Cas64", func(p *uint64) { atomic.Cas64(p, 0, 1) }},
		{"And64", func(p *uint64) { atomic.And64(p, 1) }},
		{"Or64", func(p *uint64) { atomic.Or64(p, 1) }},
		{"Loadint64", func(p *uint64) { atomic.Loadint64((*int64)(unsafe.Pointer(p))) }},
		{"Storeint64", func(p *uint64) { atomic.Storeint64((*int64)(unsafe.Pointer(p)), 1) }},
		{"Xaddint64", func(p *uint64) { atomic.Xaddint64((*int64)(unsafe.Pointer(p)), 1) }},
		{"Xchgint64", func(p *uint64) { atomic.Xchgint64((*int64)(unsafe.Pointer(p)), 1) }},
		{"Casint64", func(p *uint64) { atomic.Casint64((*int64)(unsafe.Pointer(p)), 0, 1) }},
	}
	w := operandWords(t, 58)
	held := &w[0]
	for _, op := range ops {
		for _, target := range []struct {
			name string
			p    *uint64
		}{{"same operand", &w[0]}, {"colliding operand", &w[57]}} {
			var started, finished uint32
			atomic.StripeAcquire(held)
			go func() {
				atomic.Store(&started, 1)
				op.f(target.p)
				atomic.Store(&finished, 1)
			}()
			for i := 0; atomic.Load(&started) == 0; i++ {
				if i > 5000 {
					atomic.StripeRelease(held)
					t.Fatalf("%s on %s: goroutine never started", op.name, target.name)
				}
				time.Sleep(time.Millisecond)
			}
			time.Sleep(100 * time.Millisecond)
			early := atomic.Load(&finished) != 0
			atomic.StripeRelease(held)
			if early {
				t.Errorf("%s on %s completed while its stripe was held", op.name, target.name)
			}
			for i := 0; atomic.Load(&finished) == 0; i++ {
				if i > 5000 {
					t.Fatalf("%s on %s did not complete after the stripe was released", op.name, target.name)
				}
				time.Sleep(time.Millisecond)
			}
		}
	}
}

// stripeHoldChild holds the stripe of word 0 and then performs 64-bit atomics on
// word 1, which hashes to a different stripe. With one global lock that
// operation spins forever on this very thread, so the parent bounds the child
// with a wall-clock kill.
func stripeHoldChild() {
	runtime.LockOSThread()
	w := make([]uint64, 2)
	a, b := &w[0], &w[1]
	if atomic.StripeIndex(a) == atomic.StripeIndex(b) {
		fmt.Println("stripe: operands share a stripe")
		os.Exit(3)
	}
	atomic.StripeAcquire(a)
	atomic.Xadd64(b, 5)
	atomic.Store64(b, atomic.Load64(b)+1)
	atomic.Cas64(b, 6, 7)
	atomic.Xchg64(b, atomic.Or64(b, 8)+atomic.And64(b, 0xff))
	stillA := atomic.StripeWord(atomic.StripeIndex(a))
	bFree := atomic.StripeWord(atomic.StripeIndex(b))
	atomic.StripeRelease(a)
	if stillA != 1 || bFree != 0 {
		fmt.Printf("stripe: held stripe word %d (want 1), other stripe word %d (want 0)\n", stillA, bFree)
		os.Exit(3)
	}
	fmt.Println(stripeHoldMarker)
	os.Exit(0)
}

// TestStripeDistinctOperandNotBlocked: holding the stripe of one operand must
// not block an operand at another stripe. A single global lock fails this by
// hanging, which is why it runs in a child with a wall-clock kill.
func TestStripeDistinctOperandNotBlocked(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe, stripeHoldArg)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var killed bool
	var mu sync.Mutex
	timer := time.AfterFunc(stripeHoldWall, func() { mu.Lock(); killed = true; mu.Unlock(); cmd.Process.Kill() })
	werr := cmd.Wait()
	timer.Stop()
	mu.Lock()
	defer mu.Unlock()
	if killed {
		t.Fatalf("child HUNG: killed after %v wall: an operation at a different stripe waited for a held stripe\n%s", stripeHoldWall, buf.String())
	}
	if werr != nil || !strings.Contains(buf.String(), stripeHoldMarker) {
		t.Fatalf("child failed (%v):\n%s", werr, buf.String())
	}
}

// TestStripeNoEscape: the pointer operand of every 64-bit operation must not
// escape. If it did, the stack array below would be moved to the heap and the
// run would allocate.
func TestStripeNoEscape(t *testing.T) {
	n := testing.AllocsPerRun(100, func() {
		var buf [3]uint32
		p := (*uint64)(unsafe.Add(unsafe.Pointer(&buf[0]), (8-uintptr(unsafe.Pointer(&buf[0]))&7)&7))
		atomic.Store64(p, 1)
		atomic.Xadd64(p, 1)
		atomic.Xchg64(p, 3)
		atomic.Cas64(p, 3, 4)
		atomic.Or64(p, 8)
		atomic.And64(p, 12)
		if atomic.Load64(p) != 12 {
			panic("bad value")
		}
	})
	if n != 0 {
		t.Fatalf("64-bit atomics on a stack operand allocate %v times: the operand escapes", n)
	}
}
