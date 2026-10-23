// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin && ppc && !goppc_atcompat_disabled

package unix

// atCompatForceDisabled is false in every production build of this package. Its
// true value lives in at_compat_disabled_darwin_ppc.go, which is compiled only
// under the goppc_atcompat_disabled tag; see atCompatEnabled.
const atCompatForceDisabled = false
