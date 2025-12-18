// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// GenericHandler handles generic package operations
type GenericHandler struct {
	BaseHandler
}

// NewGenericHandler creates a new generic package handler
func NewGenericHandler() *GenericHandler {
	return &GenericHandler{
		BaseHandler: NewBaseHandler(Generic),
	}
}

// Validate checks if the configuration is valid for generic packages
func (h *GenericHandler) Validate(config Config) error {
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

// Push uploads generic artifacts to the registry
func (h *GenericHandler) Push(ctx context.Context, config Config) error {
	logrus.Println("Executing generic push command")

	// Validate configuration
	if err := h.Validate(config); err != nil {
		return err
	}

	// Set default version if not provided
	version := config.Version
	if version == "" {
		version = "v1"
		logrus.Printf("No version specified, using default: %s", version)
	}

	// Check if source is a file or directory
	fileInfo, err := os.Stat(config.Source)
	if err != nil {
		return fmt.Errorf("failed to access source path '%s': %w", config.Source, err)
	}

	logrus.Printf("Source path: %s", config.Source)
	logrus.Printf("Is directory: %v", fileInfo.IsDir())

	if fileInfo.IsDir() {
		// Reject directories - only single files are allowed
		return fmt.Errorf("directories are not supported, only single files can be pushed. Source '%s' is a directory", config.Source)
	}

	// Handle single file
	logrus.Printf("Detected file, calling pushSingleFile")
	return h.pushSingleFile(config, version, config.Source, config.Name, "")
}

// Pull downloads generic artifacts from the registry
func (h *GenericHandler) Pull(ctx context.Context, config Config) error {
	logrus.Println("Executing generic pull command")

	// Validate required parameters for pull
	if config.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if config.Name == "" {
		return fmt.Errorf("package name must be set")
	}
	if config.Version == "" {
		return fmt.Errorf("package version must be set")
	}
	if config.Filename == "" {
		return fmt.Errorf("filename must be set")
	}
	if config.Destination == "" {
		return fmt.Errorf("destination path must be set")
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

	// Construct package path in the format expected by harness-cli: <package_name>/<version>/<filename>
	packagePath := fmt.Sprintf("%s/%s/%s", config.Name, config.Version, config.Filename)

	// Build Harness CLI command
	cmdArgs := []string{getHarnessBin(), "artifact", "pull", string(Generic), config.Registry, packagePath, config.Destination}

	// Add required flags
	cmdArgs = append(cmdArgs, "--token", config.Token)
	cmdArgs = append(cmdArgs, "--account", config.Account)
	cmdArgs = append(cmdArgs, "--pkg-url", config.PkgURL)

	// Add optional flags
	if config.Org != "" {
		cmdArgs = append(cmdArgs, "--org", config.Org)
	}
	if config.Project != "" {
		cmdArgs = append(cmdArgs, "--project", config.Project)
	}
	if config.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", config.ApiURL)
	}

	// Add format flag for consistent output
	cmdArgs = append(cmdArgs, "--format", "json")

	return executeCommand(cmdArgs, fmt.Sprintf("pull artifact '%s' (version '%s', file '%s') from registry '%s' to '%s'",
		config.Name, config.Version, config.Filename, config.Registry, config.Destination))
}

// Get retrieves generic artifact information
func (h *GenericHandler) Get(ctx context.Context, config Config) error {
	logrus.Println("Executing generic get command")

	if config.Registry == "" {
		return fmt.Errorf("registry name must be set")
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

	// Use 'hc artifact get' command with name as positional arg and registry as flag
	cmd := []string{
		getHarnessBin(),
		"artifact", "get", config.Name,
	}

	// Add required flags
	cmd = append(cmd, "--registry", config.Registry)
	cmd = append(cmd, "--token", config.Token)
	cmd = append(cmd, "--account", config.Account)
	if config.Org != "" {
		cmd = append(cmd, "--org", config.Org)
	}
	if config.Project != "" {
		cmd = append(cmd, "--project", config.Project)
	}
	if config.ApiURL != "" {
		cmd = append(cmd, "--api-url", config.ApiURL)
	}

	cmd = append(cmd, "--format", "json")

	return executeCommand(cmd, fmt.Sprintf("get info for artifact '%s' in registry '%s'", config.Name, config.Registry))
}

// Delete removes generic artifacts from the registry
func (h *GenericHandler) Delete(ctx context.Context, config Config) error {
	logrus.Println("Executing generic delete command")

	// Validate required parameters for delete
	if config.Registry == "" {
		return fmt.Errorf("registry name must be set")
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

	// Build Harness CLI command - use 'hc artifact delete' with name as argument and registry as flag
	cmdArgs := []string{getHarnessBin(), "artifact", "delete", config.Name}

	// Add required flags
	cmdArgs = append(cmdArgs, "--registry", config.Registry)
	cmdArgs = append(cmdArgs, "--token", config.Token)
	cmdArgs = append(cmdArgs, "--account", config.Account)

	// Add optional flags
	if config.Org != "" {
		cmdArgs = append(cmdArgs, "--org", config.Org)
	}
	if config.Project != "" {
		cmdArgs = append(cmdArgs, "--project", config.Project)
	}
	if config.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", config.ApiURL)
	}

	// Add format flag for consistent output
	cmdArgs = append(cmdArgs, "--format", "json")

	return executeCommand(cmdArgs, fmt.Sprintf("delete artifact '%s' from registry '%s'", config.Name, config.Registry))
}

// pushDirectory handles pushing all files in a directory for generic packages
func (h *GenericHandler) pushDirectory(config Config, version string) error {
	logrus.Printf("Source is a directory, pushing all files from: %s", config.Source)

	// Walk through directory recursively to find all files
	var filesToPush []string
	err := filepath.Walk(config.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Printf("Warning: failed to access path '%s': %v", path, err)
			return nil // Continue walking despite errors
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories (like .venv, .git, etc.)
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		// Only collect regular files
		filesToPush = append(filesToPush, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory '%s': %w", config.Source, err)
	}

	if len(filesToPush) == 0 {
		return fmt.Errorf("no files found in directory '%s'", config.Source)
	}

	logrus.Printf("Found %d files to push", len(filesToPush))

	// Track success/failure statistics
	var successCount, failureCount, skippedCount int

	// Push each file
	for i, file := range filesToPush {
		filename := filepath.Base(file)
		logrus.Printf("[%d/%d] Pushing file: %s", i+1, len(filesToPush), filename)

		// Get relative path from the source directory
		relPath, err := filepath.Rel(config.Source, file)
		if err != nil {
			// Fallback to just filename if relative path fails
			relPath = filename
		}

		// For multiple files, use relative path as part of artifact name
		artifactName := config.Name

		// Validate artifact name
		if artifactName == "" || strings.HasSuffix(artifactName, "_") || strings.HasSuffix(artifactName, "-") {
			logrus.Printf("⚠ Warning: Invalid artifact name '%s' for file '%s', skipping", artifactName, filename)
			skippedCount++
			continue
		}

		err = h.pushSingleFile(config, version, file, artifactName, relPath)
		if err != nil {
			// Log the error but continue with other files instead of failing the entire process
			logrus.Errorf("✗ Failed to push file '%s': %v", filename, err)
			logrus.Printf("Continuing with remaining files...")
			failureCount++
			continue
		}

		logrus.Printf("✓ Successfully pushed file: %s as artifact: %s", filename, artifactName)
		successCount++
	}

	// Print final statistics
	logrus.Printf("=== DIRECTORY UPLOAD SUMMARY ===")
	logrus.Printf("Total files processed: %d", len(filesToPush))
	logrus.Printf("✓ Successfully uploaded: %d", successCount)
	if failureCount > 0 {
		logrus.Printf("✗ Failed uploads: %d", failureCount)
	}
	if skippedCount > 0 {
		logrus.Printf("⚠ Skipped files: %d", skippedCount)
	}

	if failureCount > 0 || skippedCount > 0 {
		logrus.Printf("Directory upload completed with %d failures and %d skipped files", failureCount, skippedCount)
	} else {
		logrus.Printf("✓ All files uploaded successfully!")
	}

	return nil
}

// pushSingleFile handles pushing a single file for generic packages
func (h *GenericHandler) pushSingleFile(config Config, version, filePath, customName, relativePath string) error {
	// Use custom name if provided, otherwise use the original artifact name
	artifactName := config.Name
	if customName != "" {
		artifactName = customName
	}

	// Build command using shared helper
	cmdArgs, err := buildPushCommand(Generic, config, version, filePath, artifactName, true)
	if err != nil {
		return err
	}

	// Add path parameter for generic packages - use relative path if provided, otherwise use filename
	if relativePath != "" {
		cmdArgs = append(cmdArgs, "--path", relativePath)
	}

	return executeCommand(cmdArgs, fmt.Sprintf("push artifact '%s' to registry '%s'", artifactName, config.Registry))
}

// Helper functions

// AuthConfig represents the Harness authentication configuration
type AuthConfig struct {
	BaseURL   string `json:"base_url"`
	Token     string `json:"token"`
	AccountID string `json:"account_id"`
}

// createAuthFile creates ~/.harness/auth.json with authentication details
func createAuthFile(config Config) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .harness directory if it doesn't exist
	harnessDir := filepath.Join(homeDir, ".harness")
	if err := os.MkdirAll(harnessDir, 0755); err != nil {
		return fmt.Errorf("failed to create .harness directory: %w", err)
	}

	// Prepare auth configuration
	authConfig := AuthConfig{
		BaseURL:   config.PkgURL,
		Token:     fmt.Sprintf("CIManager %s", config.Token),
		AccountID: config.Account,
	}

	// Write auth.json file
	authFile := filepath.Join(harnessDir, "auth.json")
	authData, err := json.MarshalIndent(authConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}

	if err := os.WriteFile(authFile, authData, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	logrus.Printf("Created auth file: %s", authFile)
	return nil
}

// buildPushCommand builds a common push command for any package type
func buildPushCommand(packageType PackageType, config Config, version, filePath, artifactName string, includeFileAndVersion bool) ([]string, error) {
	// Create auth file first
	if err := createAuthFile(config); err != nil {
		return nil, fmt.Errorf("failed to create auth file: %w", err)
	}

	cmdArgs := []string{getHarnessBin(), "artifact"}

	// Add context flags (no token needed as it's in auth.json)
	if config.Org != "" {
		cmdArgs = append(cmdArgs, "--org", config.Org)
	}
	if config.Project != "" {
		cmdArgs = append(cmdArgs, "--project", config.Project)
	}

	// Add the rest of the command: push, package-type, registry, filepath
	// All package types need the file path to know what to push
	cmdArgs = append(cmdArgs, "push", strings.ToLower(string(packageType)), config.Registry, filePath)

	// Add other required flags
	if includeFileAndVersion {
		// Only generic packages need explicit name and version
		cmdArgs = append(cmdArgs, "--name", artifactName)
		cmdArgs = append(cmdArgs, "--version", version)
	}
	cmdArgs = append(cmdArgs, "--pkg-url", config.PkgURL)

	// Add remaining optional flags
	if config.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", config.ApiURL)
	}
	if config.Filename != "" {
		cmdArgs = append(cmdArgs, "--filename", config.Filename)
	}

	return cmdArgs, nil
}

func getHarnessBin() string {
	if runtime.GOOS == "windows" {
		if _, err := os.Stat("C:/bin/hc.exe"); err == nil {
			return "C:/bin/hc.exe"
		}
	}
	return "hc"
}

// executeCommand executes a Harness CLI command
func executeCommand(cmdArgs []string, operation string) error {
	cmdStr := strings.Join(cmdArgs, " ")
	logrus.Printf("Executing command: %s", cmdStr)

	// Execute command directly without shell to avoid argument parsing issues
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to %s: %w", operation, err)
	}

	logrus.Printf("Successfully completed: %s\n", operation)
	return nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	// Only show trace in debug mode to reduce noise
	if logrus.GetLevel() >= logrus.DebugLevel {
		fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
	}
}
