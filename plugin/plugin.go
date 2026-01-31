// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/harness/drone-har/plugin/packages"
	"github.com/sirupsen/logrus"
)

const (
	harnessHTTPProxy  = "HARNESS_HTTP_PROXY"
	harnessHTTPSProxy = "HARNESS_HTTPS_PROXY"
	harnessNoProxy    = "HARNESS_NO_PROXY"
	httpProxy         = "HTTP_PROXY"
	httpsProxy        = "HTTPS_PROXY"
	noProxy           = "NO_PROXY"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Harness CLI authentication
	Token   string `envconfig:"PLUGIN_TOKEN"`
	Account string `envconfig:"PLUGIN_ACCOUNT"`
	Org     string `envconfig:"PLUGIN_ORG"`
	Project string `envconfig:"PLUGIN_PROJECT"`
	ApiURL  string `envconfig:"PLUGIN_API_URL"`

	// HAR commands
	Command string `envconfig:"PLUGIN_COMMAND"`

	// Artifact upload/download parameters
	Registry    string `envconfig:"PLUGIN_REGISTRY"`
	Source      string `envconfig:"PLUGIN_SOURCE"` // File path or directory path - if directory, all files will be pushed recursively
	Name        string `envconfig:"PLUGIN_NAME"`   // Base artifact name - for directories, files get unique names with relative paths
	Version     string `envconfig:"PLUGIN_VERSION"`
	Description string `envconfig:"PLUGIN_DESCRIPTION"`
	Filename    string `envconfig:"PLUGIN_FILENAME"`
	PkgURL      string `envconfig:"PLUGIN_PKG_URL"`
	PomFile     string `envconfig:"PLUGIN_POM_FILE"`

	// Package type for push operations
	PackageType string `envconfig:"PLUGIN_PACKAGE_TYPE"`

	// Pull/Download parameters
	Destination string `envconfig:"PLUGIN_DESTINATION"`

	// Additional parameters
	Retries     int    `envconfig:"PLUGIN_RETRIES"`
	EnableProxy string `envconfig:"PLUGIN_ENABLE_PROXY"`
}

// Exec executes the plugin using the new modular architecture.
func Exec(ctx context.Context, args Args) error {
	logrus.Println("Starting Harness Artifact Registry operation")

	// Check if command is specified, default to push for backward compatibility
	command := args.Command
	if command == "" {
		command = "push"
		logrus.Printf("No command specified, defaulting to: %s", command)
	}

	// Enable proxy if configured
	enableProxy := parseBoolOrDefault(false, args.EnableProxy)
	if enableProxy {
		logrus.Printf("Setting proxy config for operation")
		setSecureConnectProxies()
	}

	// Create handler factory
	factory := packages.NewHandlerFactory()

	// Get package type, default to generic
	packageType := args.PackageType
	if packageType == "" {
		packageType = "generic"
		logrus.Printf("No package type specified, using default: %s", packageType)
	}

	// Get the appropriate handler for the package type
	handler, err := factory.GetHandler(packageType)
	if err != nil {
		return fmt.Errorf("failed to get package handler: %w", err)
	}

	logrus.Printf("Using %s package handler", handler.GetPackageType())

	// Convert args to config
	config := argsToConfig(args)

	// Route to appropriate command handler
	switch command {
	case "push", "upload":
		return handler.Push(ctx, config)
	case "pull", "download":
		return handler.Pull(ctx, config)
	case "get", "info":
		return handler.Get(ctx, config)
	case "delete", "remove":
		return handler.Delete(ctx, config)
	default:
		return fmt.Errorf("unsupported command: %s. Supported commands: push, pull, get, delete", command)
	}
}

// argsToConfig converts Args to packages.Config
func argsToConfig(args Args) packages.Config {
	registry := strings.TrimSpace(args.Registry)
	if idx := strings.LastIndex(registry, "."); idx != -1 {
		registry = registry[idx+1:]
	}

	return packages.Config{
		// Authentication
		Token:   args.Token,
		Account: args.Account,
		Org:     args.Org,
		Project: args.Project,
		ApiURL:  args.ApiURL,
		PkgURL:  args.PkgURL,

		// Registry and artifact details
		Registry:    registry,
		Name:        args.Name,
		Version:     args.Version,
		Description: args.Description,
		Filename:    args.Filename,
		PomFile:     args.PomFile,

		// Operation details
		Source:      args.Source,
		Destination: args.Destination,
		Retries:     args.Retries,
	}
}

func parseBoolOrDefault(defaultValue bool, s string) bool {
	if s == "" {
		return defaultValue
	}
	return strings.ToLower(s) == "true" || s == "1"
}

func setSecureConnectProxies() {
	copyEnvVariableIfExists(harnessHTTPProxy, httpProxy)
	copyEnvVariableIfExists(harnessHTTPSProxy, httpsProxy)
	copyEnvVariableIfExists(harnessNoProxy, noProxy)
}

func copyEnvVariableIfExists(src string, dest string) {
	srcValue := os.Getenv(src)
	if srcValue == "" {
		return
	}
	err := os.Setenv(dest, srcValue)
	if err != nil {
		logrus.Printf("Failed to copy env variable from %s to %s with error %v", src, dest, err)
	}
}
