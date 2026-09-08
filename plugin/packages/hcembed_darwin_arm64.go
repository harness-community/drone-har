// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

//go:build darwin && arm64

package packages

import (
	// embed is imported for its side effect: it enables the //go:embed
	// directive below.
	_ "embed"
)

// embeddedHC is the pinned harness-cli release archive for this platform,
// fetched into hcbin by scripts/fetch-hc.sh at build time.
//
//go:embed hcbin/hc_darwin_arm64.tar.gz
var embeddedHC []byte
