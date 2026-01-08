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

// DartHandler handles Dart package operations
type DartHandler struct {
	BaseHandler
}

// NewDartHandler creates a new Dart package handler
func NewDartHandler() *DartHandler {
	return &DartHandler{
		BaseHandler: NewBaseHandler(Dart),
	}
}

// Validate checks if the configuration is valid for Dart packages
func (h *DartHandler) Validate(config Config) error {
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

// Push uploads Dart packages to the registry
func (h *DartHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Dart push command")

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

// pushSingleFile handles pushing a single file for Dart packages
func (h *DartHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Dart)
	cmdArgs, err := buildPushCommand(Dart, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Dart artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Dart packages from the registry
func (h *DartHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Dart pull logic
	return fmt.Errorf("Dart pull is not yet implemented")
}

// Get retrieves Dart package information
func (h *DartHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Dart get logic
	return fmt.Errorf("Dart get is not yet implemented")
}

// Delete removes Dart packages from the registry
func (h *DartHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Dart delete logic
	return fmt.Errorf("Dart delete is not yet implemented")
}
