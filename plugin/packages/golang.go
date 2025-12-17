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

// GoHandler handles Go package operations
type GoHandler struct {
	BaseHandler
}

// NewGoHandler creates a new Go package handler
func NewGoHandler() *GoHandler {
	return &GoHandler{
		BaseHandler: NewBaseHandler(Go),
	}
}

// Validate checks if the configuration is valid for Go packages
func (h *GoHandler) Validate(config Config) error {
	if config.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if config.Source == "" {
		return fmt.Errorf("source file path must be set")
	}
	if config.Name == "" {
		return fmt.Errorf("artifact name must be set")
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

// Push uploads Go packages to the registry
func (h *GoHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Go push command")

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

// pushSingleFile handles pushing a single file for Go packages
func (h *GoHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Go)
	cmdArgs := buildPushCommand(Go, config, "", filePath, artifactName, false)

	return executeCommand(cmdArgs, fmt.Sprintf("push Go artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Go packages from the registry
func (h *GoHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Go pull logic
	return fmt.Errorf("Go pull is not yet implemented")
}

// Get retrieves Go package information
func (h *GoHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Go get logic
	return fmt.Errorf("Go get is not yet implemented")
}

// Delete removes Go packages from the registry
func (h *GoHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Go delete logic
	return fmt.Errorf("Go delete is not yet implemented")
}
