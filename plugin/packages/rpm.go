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

// RPMHandler handles RPM package operations
type RPMHandler struct {
	BaseHandler
}

// NewRPMHandler creates a new RPM package handler
func NewRPMHandler() *RPMHandler {
	return &RPMHandler{
		BaseHandler: NewBaseHandler(RPM),
	}
}

// Validate checks if the configuration is valid for RPM packages
func (h *RPMHandler) Validate(config Config) error {
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

// Push uploads RPM packages to the registry
func (h *RPMHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing RPM push command")

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

// pushSingleFile handles pushing a single file for RPM packages
func (h *RPMHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for RPM)
	cmdArgs, err := buildPushCommand(RPM, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push RPM artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads RPM packages from the registry
func (h *RPMHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement RPM pull logic
	return fmt.Errorf("RPM pull is not yet implemented")
}

// Get retrieves RPM package information
func (h *RPMHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement RPM get logic
	return fmt.Errorf("RPM get is not yet implemented")
}

// Delete removes RPM packages from the registry
func (h *RPMHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement RPM delete logic
	return fmt.Errorf("RPM delete is not yet implemented")
}
