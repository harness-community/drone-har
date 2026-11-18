// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

//go:build linux && arm64

package packages

import (
	// embed is imported for its side effect: it enables the //go:embed
	// directive below.
	_ "embed"
)

// embeddedCLI is the pinned harness-har release archive for this platform,
// fetched into hcbin by scripts/fetch-hc.sh at build time.
//
//go:embed hcbin/harness-har_linux_arm64.tar.gz
var embeddedCLI []byte
