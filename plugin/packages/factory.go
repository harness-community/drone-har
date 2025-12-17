// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"fmt"
	"strings"
)

// HandlerFactory creates package handlers based on package type
type HandlerFactory struct {
	handlers map[PackageType]PackageHandler
}

// NewHandlerFactory creates a new handler factory with all registered handlers
func NewHandlerFactory() *HandlerFactory {
	factory := &HandlerFactory{
		handlers: make(map[PackageType]PackageHandler),
	}
	
	// Register all available handlers
	factory.registerHandler(NewGenericHandler())
	factory.registerHandler(NewNPMHandler())
	factory.registerHandler(NewPythonHandler())
	factory.registerHandler(NewGoHandler())
	factory.registerHandler(NewRPMHandler())
	factory.registerHandler(NewNuGetHandler())
	factory.registerHandler(NewMavenHandler())
	
	return factory
}

// registerHandler registers a package handler with the factory
func (f *HandlerFactory) registerHandler(handler PackageHandler) {
	f.handlers[handler.GetPackageType()] = handler
}

// GetHandler returns the appropriate handler for the given package type
func (f *HandlerFactory) GetHandler(packageType string) (PackageHandler, error) {
	// Normalize package type to uppercase
	normalizedType := PackageType(packageType)
	
	// Default to generic if empty
	if normalizedType == "" {
		normalizedType = Generic
	}
	
	handler, exists := f.handlers[normalizedType]
	if !exists {
		return nil, fmt.Errorf("unsupported package type: %s. Supported types: %s", 
			packageType, f.GetSupportedTypes())
	}
	
	return handler, nil
}

// GetSupportedTypes returns a comma-separated list of supported package types
func (f *HandlerFactory) GetSupportedTypes() string {
	var types []string
	for packageType := range f.handlers {
		types = append(types, string(packageType))
	}
	return strings.Join(types, ", ")
}

// IsSupported checks if a package type is supported
func (f *HandlerFactory) IsSupported(packageType string) bool {
	normalizedType := PackageType(strings.ToUpper(packageType))
	_, exists := f.handlers[normalizedType]
	return exists
}

// GetImplementedTypes returns only the package types that are fully implemented
func (f *HandlerFactory) GetImplementedTypes() []PackageType {
	// All package types now have push functionality implemented
	return []PackageType{Generic, NPM, RPM, Python, Go, NuGet, Maven}
}

// GetPlannedTypes returns the package types that are planned but not yet implemented
func (f *HandlerFactory) GetPlannedTypes() []PackageType {
	// All basic functionality is now implemented
	return []PackageType{}
}
