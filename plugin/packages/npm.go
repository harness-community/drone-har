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

// NPMHandler handles NPM package operations
type NPMHandler struct {
	BaseHandler
}

// NewNPMHandler creates a new NPM package handler
func NewNPMHandler() *NPMHandler {
	return &NPMHandler{
		BaseHandler: NewBaseHandler(NPM),
	}
}

// Validate checks if the configuration is valid for NPM packages
func (h *NPMHandler) Validate(config Config) error {
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

// Push uploads NPM packages to the registry
func (h *NPMHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing NPM push command")

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

// pushSingleFile handles pushing a single file for NPM packages
func (h *NPMHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for NPM)
	cmdArgs, err := buildPushCommand(NPM, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push NPM artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads NPM packages from the registry
func (h *NPMHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement NPM pull logic
	return fmt.Errorf("NPM pull is not yet implemented")
}

// Get retrieves NPM package information
func (h *NPMHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement NPM get logic
	return fmt.Errorf("NPM get is not yet implemented")
}

// Delete removes NPM packages from the registry
func (h *NPMHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement NPM delete logic
	return fmt.Errorf("NPM delete is not yet implemented")
}
