// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"context"
)

// PackageType represents the supported package types
type PackageType string

const (
	Generic PackageType = "GENERIC"
	NPM     PackageType = "NPM"
	Dart    PackageType = "DART"
	Composer PackageType = "COMPOSER"
	RPM     PackageType = "RPM"
	Python  PackageType = "PYTHON"
	Go      PackageType = "GO"
	Cargo   PackageType = "CARGO"
	NuGet   PackageType = "NUGET"
	Maven   PackageType = "MAVEN"
	Conda   PackageType = "CONDA"
)

// Config holds the common configuration for all package handlers
type Config struct {
	// Authentication
	Token   string
	Account string
	Org     string
	Project string
	ApiURL  string
	PkgURL  string

	// Registry and artifact details
	Registry    string
	Name        string
	Version     string
	Description string
	Filename    string

	// Operation details
	Source      string
	Destination string
	Retries     int
}

// PackageHandler defines the interface that all package type handlers must implement
type PackageHandler interface {
	// Push uploads artifacts to the registry
	Push(ctx context.Context, config Config) error
	
	// Pull downloads artifacts from the registry
	Pull(ctx context.Context, config Config) error
	
	// Get retrieves artifact information
	Get(ctx context.Context, config Config) error
	
	// Delete removes artifacts from the registry
	Delete(ctx context.Context, config Config) error
	
	// Validate checks if the configuration is valid for this package type
	Validate(config Config) error
	
	// GetPackageType returns the package type this handler supports
	GetPackageType() PackageType
}

// BaseHandler provides common functionality for all package handlers
type BaseHandler struct {
	packageType PackageType
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(packageType PackageType) BaseHandler {
	return BaseHandler{
		packageType: packageType,
	}
}

// GetPackageType returns the package type
func (h BaseHandler) GetPackageType() PackageType {
	return h.packageType
}
