// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

func KeventInvalidDescriptor() int32 {
	return kevent(-1, nil, 0, nil, 0, nil)
}
