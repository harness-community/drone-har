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

// PythonHandler handles Python package operations
type PythonHandler struct {
	BaseHandler
}

// NewPythonHandler creates a new Python package handler
func NewPythonHandler() *PythonHandler {
	return &PythonHandler{
		BaseHandler: NewBaseHandler(Python),
	}
}

// Validate checks if the configuration is valid for Python packages
func (h *PythonHandler) Validate(config Config) error {
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

// Push uploads Python packages to the registry
func (h *PythonHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Python push command")

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

// pushSingleFile handles pushing a single file for Python packages
func (h *PythonHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Python)
	cmdArgs, err := buildPushCommand(Python, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Python artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Python packages from the registry
func (h *PythonHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Python pull logic
	return fmt.Errorf("Python pull is not yet implemented")
}

// Get retrieves Python package information
func (h *PythonHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Python get logic
	return fmt.Errorf("Python get is not yet implemented")
}

// Delete removes Python packages from the registry
func (h *PythonHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Python delete logic
	return fmt.Errorf("Python delete is not yet implemented")
}
