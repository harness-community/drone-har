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

// ComposerHandler handles Composer package operations
type ComposerHandler struct {
	BaseHandler
}

// NewComposerHandler creates a new Composer package handler
func NewComposerHandler() *ComposerHandler {
	return &ComposerHandler{
		BaseHandler: NewBaseHandler(Composer),
	}
}

// Validate checks if the configuration is valid for Composer packages
func (h *ComposerHandler) Validate(config Config) error {
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

// Push uploads Composer packages to the registry
func (h *ComposerHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Composer push command")

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

// pushSingleFile handles pushing a single file for Composer packages
func (h *ComposerHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Composer)
	cmdArgs, err := buildPushCommand(Composer, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Composer artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Composer packages from the registry
func (h *ComposerHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Composer pull logic
	return fmt.Errorf("Composer pull is not yet implemented")
}

// Get retrieves Composer package information
func (h *ComposerHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Composer get logic
	return fmt.Errorf("Composer get is not yet implemented")
}

// Delete removes Composer packages from the registry
func (h *ComposerHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Composer delete logic
	return fmt.Errorf("Composer delete is not yet implemented")
}
