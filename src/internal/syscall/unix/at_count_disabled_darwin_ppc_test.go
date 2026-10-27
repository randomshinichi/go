// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && !goppc_atcount_test

package unix_test

import "testing"

// Native counting exists only in the dedicated test variant.
func atBegin349Observation(t *testing.T) func() { return func() {} }
