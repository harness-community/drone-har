// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package main

import (
	// embed is imported for its side effect: it enables the //go:embed
	// directive below.
	_ "embed"
)

// cliVersionFile holds the contents of the repo-root HC_VERSION file, which is
// the single source of truth for the pinned harness-har version. The container
// images read the same file at build time, so an image and the plugin binary
// built from a given commit always agree on which CLI they expect.
//
//go:embed HC_VERSION
var cliVersionFile string
