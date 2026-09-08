// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

//go:build !(linux && (amd64 || arm64)) && !(darwin && (amd64 || arm64)) && !(windows && (amd64 || arm64))

package packages

// embeddedHC is empty on platforms we do not publish releases for, so the
// plugin still builds there and falls back to downloading harness-cli.
var embeddedHC []byte
