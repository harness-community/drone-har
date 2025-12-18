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

// CondaHandler handles Conda package operations
type CondaHandler struct{}

// NewCondaHandler creates a new Conda handler
func NewCondaHandler() *CondaHandler {
	return &CondaHandler{}
}

// GetPackageType returns the package type for Conda packages
func (h *CondaHandler) GetPackageType() PackageType {
	return Conda
}

// Validate checks if the configuration is valid for Conda packages
func (h *CondaHandler) Validate(config Config) error {
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

// Push uploads Conda packages to the registry
func (h *CondaHandler) Push(ctx context.Context, config Config) error {
	// Validate configuration
	if err := h.Validate(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if source exists
	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", config.Source)
	}

	fileInfo, err := os.Stat(config.Source)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("directories are not supported, only single files can be pushed. Source '%s' is a directory", config.Source)
	}

	logrus.Printf("Source path: %s", config.Source)
	logrus.Printf("Detected file, calling pushSingleFile")

	return h.pushSingleFile(config, config.Source, config.Name)
}

// pushSingleFile handles pushing a single file for Conda packages
func (h *CondaHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Conda)
	cmdArgs, err := buildPushCommand(Conda, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Conda artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Conda packages from the registry
func (h *CondaHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Conda pull logic
	return fmt.Errorf("Conda pull is not yet implemented")
}

// Get retrieves information about Conda packages
func (h *CondaHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Conda get logic
	return fmt.Errorf("Conda get is not yet implemented")
}

// Delete removes Conda packages from the registry
func (h *CondaHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Conda delete logic
	return fmt.Errorf("Conda delete is not yet implemented")
}
