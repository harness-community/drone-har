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

// CargoHandler handles Cargo package operations
type CargoHandler struct {
	BaseHandler
}

// NewCargoHandler creates a new Cargo package handler
func NewCargoHandler() *CargoHandler {
	return &CargoHandler{
		BaseHandler: NewBaseHandler(Cargo),
	}
}

// Validate checks if the configuration is valid for Cargo packages
func (h *CargoHandler) Validate(config Config) error {
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

// Push uploads Cargo packages to the registry
func (h *CargoHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Cargo push command")

	// Validate configuration
	if err := h.Validate(config); err != nil {
		return err
	}

	// Check if source path exists
	_, err := os.Stat(config.Source)
	if err != nil {
		return fmt.Errorf("failed to access source path '%s': %w", config.Source, err)
	}

	logrus.Printf("Source path: %s", config.Source)

	return h.pushSingleFile(config, config.Source, config.Name)
}

// pushSingleFile handles pushing a single file for Cargo packages
func (h *CargoHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Cargo)
	cmdArgs, err := buildPushCommand(Cargo, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Cargo artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Cargo packages from the registry
func (h *CargoHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Cargo pull logic
	return fmt.Errorf("Cargo pull is not yet implemented")
}

// Get retrieves Cargo package information
func (h *CargoHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Cargo get logic
	return fmt.Errorf("Cargo get is not yet implemented")
}

// Delete removes Cargo packages from the registry
func (h *CargoHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Cargo delete logic
	return fmt.Errorf("Cargo delete is not yet implemented")
}
