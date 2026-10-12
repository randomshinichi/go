//go:build darwin && ppc && arc4randomtest

package runtime

import (
	"internal/abi"
	"unsafe"
)

// Arc4randomBufTrampolineTestCall invokes the raw libc trampoline with a caller-
// controlled argument block. It is included only in discriminator test builds.
func Arc4randomBufTrampolineTestCall(arg unsafe.Pointer) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(arc4random_buf_trampoline)), arg)
}
