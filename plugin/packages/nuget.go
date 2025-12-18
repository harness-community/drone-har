// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// NuGetHandler handles NuGet package operations
type NuGetHandler struct {
	BaseHandler
}

// NewNuGetHandler creates a new NuGet package handler
func NewNuGetHandler() *NuGetHandler {
	return &NuGetHandler{
		BaseHandler: NewBaseHandler(NuGet),
	}
}

// Validate checks if the configuration is valid for NuGet packages
func (h *NuGetHandler) Validate(config Config) error {
	if config.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if config.Source == "" {
		return fmt.Errorf("source file path must be set")
	}
	if config.Token == "" {
		return fmt.Errorf("authentication token must be set")
	}
	if config.Account == "" {
		return fmt.Errorf("account ID must be set")
	}
	if config.PkgURL == "" {
		return fmt.Errorf("package URL must be set")
	}
	return nil
}

// Push uploads NuGet packages to the registry
func (h *NuGetHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing NuGet push command")

	// Validate configuration
	if err := h.Validate(config); err != nil {
		return err
	}

	// Check if source is a file (only single files allowed)
	fileInfo, err := os.Stat(config.Source)
	if err != nil {
		return fmt.Errorf("failed to access source path '%s': %w", config.Source, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("directories are not supported, only single files can be pushed. Source '%s' is a directory", config.Source)
	}

	logrus.Printf("Source path: %s", config.Source)
	logrus.Printf("Detected file, calling pushSingleFile")

	return h.pushSingleFile(config, config.Source, config.Name)
}

// pushSingleFile handles pushing a single file for NuGet packages
func (h *NuGetHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for NuGet)
	cmdArgs, err := buildPushCommand(NuGet, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push NuGet artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads NuGet packages from the registry
func (h *NuGetHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement NuGet pull logic
	return fmt.Errorf("NuGet pull is not yet implemented")
}

// Get retrieves NuGet package information
func (h *NuGetHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement NuGet get logic
	return fmt.Errorf("NuGet get is not yet implemented")
}

// Delete removes NuGet packages from the registry
func (h *NuGetHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement NuGet delete logic
	return fmt.Errorf("NuGet delete is not yet implemented")
}
