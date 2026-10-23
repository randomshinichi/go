// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && goppc_atcompat_disabled

package unix

// atCompatForceDisabled is true only in a build made with the
// goppc_atcompat_disabled tag. It exists so the disabled-backend dispatch can be
// exercised on a kernel that has the anchored backend (internal/syscall/unix/
// at_disabled_darwin_ppc_test.go asserts that every real-descriptor operation
// refuses without mutating anything, and that the path-based fallbacks still
// work). It is a compile-time constant, so no test can race it, and it is the
// same mechanism as the goppc_atcompound_mutant tag: no shipped binary has it.
const atCompatForceDisabled = true
