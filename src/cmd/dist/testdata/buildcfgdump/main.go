// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Buildcfgdump prints what the real internal/buildcfg derives from the
// GO$GOARCH environment, for TestBuildcfgAgreement in cmd/dist to compare
// with dist's own derivation. It prints the GO$GOARCH setting that applies to
// $GOARCH, then the architecture sub-version tags (ToolTags minus the
// goexperiment tags), one per line.
package main

import (
	"fmt"
	"internal/buildcfg"
	"strings"
)

func main() {
	name, value := buildcfg.GOGOARCH()
	fmt.Printf("setting %s=%s\n", name, value)
	for _, tag := range buildcfg.ToolTags {
		if strings.HasPrefix(tag, "goexperiment.") {
			continue
		}
		fmt.Printf("tag %s\n", tag)
	}
}
