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

// MavenHandler handles Maven package operations
type MavenHandler struct {
	BaseHandler
}

// NewMavenHandler creates a new Maven package handler
func NewMavenHandler() *MavenHandler {
	return &MavenHandler{
		BaseHandler: NewBaseHandler(Maven),
	}
}

// Validate checks if the configuration is valid for Maven packages
func (h *MavenHandler) Validate(config Config) error {
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

// Push uploads Maven packages to the registry
func (h *MavenHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing Maven push command")

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

// pushSingleFile handles pushing a single file for Maven packages
func (h *MavenHandler) pushSingleFile(config Config, filePath, artifactName string) error {
	// Build command using shared helper (no file path and version in command for Maven)
	cmdArgs, err := buildPushCommand(Maven, config, "", filePath, artifactName, false)
	if err != nil {
		return err
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push Maven artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Pull downloads Maven packages from the registry
func (h *MavenHandler) Pull(ctx context.Context, config Config) error {
	// TODO: Implement Maven pull logic
	return fmt.Errorf("Maven pull is not yet implemented")
}

// Get retrieves Maven package information
func (h *MavenHandler) Get(ctx context.Context, config Config) error {
	// TODO: Implement Maven get logic
	return fmt.Errorf("Maven get is not yet implemented")
}

// Delete removes Maven packages from the registry
func (h *MavenHandler) Delete(ctx context.Context, config Config) error {
	// TODO: Implement Maven delete logic
	return fmt.Errorf("Maven delete is not yet implemented")
}
