package runtime

import "internal/runtime/atomic"

// Test-only sigprof hook for the goppc-1ye4 stackcheck witness; applied with
// go build -overlay, never part of the tree.

const sig1ye4cMax = 1 << 17

var (
	sig1ye4cN     uint32
	sig1ye4cEq    uint32
	sig1ye4cPC    [sig1ye4cMax]uint32
	sig1ye4cLR    [sig1ye4cMax]uint32
	newstack1ye4c uint32

	Sigprof1ye4cSkipEq bool
)

// sigprofHook1ye4c runs on the signal stack. It reports true when the sample
// must not be passed on to the traceback.
func sigprofHook1ye4c(pc, lr uintptr) bool {
	n := atomic.Xadd(&sig1ye4cN, 1)
	if n <= sig1ye4cMax {
		sig1ye4cPC[n-1] = uint32(pc)
		sig1ye4cLR[n-1] = uint32(lr)
	}
	if pc == lr {
		atomic.Xadd(&sig1ye4cEq, 1)
		return Sigprof1ye4cSkipEq
	}
	return false
}

func Sigprof1ye4cCounts() (n, eq uint32) {
	return atomic.Load(&sig1ye4cN), atomic.Load(&sig1ye4cEq)
}

func Sigprof1ye4cSamples() (pcs, lrs []uint32) {
	n := atomic.Load(&sig1ye4cN)
	if n > sig1ye4cMax {
		n = sig1ye4cMax
	}
	return sig1ye4cPC[:n], sig1ye4cLR[:n]
}

func Newstack1ye4cCount() uint32 { return atomic.Load(&newstack1ye4c) }
