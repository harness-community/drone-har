// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

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
	Source      string `envconfig:"PLUGIN_SOURCE"`
	Name        string `envconfig:"PLUGIN_NAME"`
	Version     string `envconfig:"PLUGIN_VERSION"`
	Description string `envconfig:"PLUGIN_DESCRIPTION"`
	Filename    string `envconfig:"PLUGIN_FILENAME"`
	PkgURL      string `envconfig:"PLUGIN_PKG_URL"`

	// Package type for push operations
	PackageType string `envconfig:"PLUGIN_PACKAGE_TYPE"`

	// Pull/Download parameters
	Destination string `envconfig:"PLUGIN_DESTINATION"`

	// Additional parameters
	Retries     int    `envconfig:"PLUGIN_RETRIES"`
	EnableProxy string `envconfig:"PLUGIN_ENABLE_PROXY"`
}

// Exec executes the plugin.
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

	// Route to appropriate command handler
	switch command {
	case "push", "upload":
		return execPushCommand(args)
	case "pull", "download":
		return execPullCommand(args)
	case "get", "info":
		return execGetCommand(args)
	case "delete", "remove":
		return execDeleteCommand(args)
	default:
		return fmt.Errorf("unsupported command: %s. Supported commands: push, pull, get, delete", command)
	}
}

// execPushCommand handles artifact upload operations
func execPushCommand(args Args) error {
	logrus.Println("Executing push command")

	// Validate required parameters for push
	if args.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if args.Source == "" {
		return fmt.Errorf("source file path must be set")
	}
	if args.Name == "" {
		return fmt.Errorf("artifact name must be set")
	}
	if args.Token == "" {
		return fmt.Errorf("authentication token must be set")
	}
	if args.Account == "" {
		return fmt.Errorf("account ID must be set")
	}
	if args.PkgURL == "" {
		return fmt.Errorf("package URL must be set")
	}

	// Set default package type and version
	packageType := args.PackageType
	if packageType == "" {
		packageType = "generic"
		logrus.Printf("No package type specified, using default: %s", packageType)
	}

	version := args.Version
	if version == "" {
		version = "1.0.0"
		logrus.Printf("No version specified, using default: %s", version)
	}

	// Build Harness CLI command - only registry and source as positional args
	cmdArgs := []string{getHarnessBin(), "ar", "push", packageType, args.Registry, args.Source}

	// Add required flags
	cmdArgs = append(cmdArgs, "--name", args.Name)
	cmdArgs = append(cmdArgs, "--version", version)
	cmdArgs = append(cmdArgs, "--token", args.Token)
	cmdArgs = append(cmdArgs, "--account", args.Account)
	cmdArgs = append(cmdArgs, "--pkg-url", args.PkgURL)

	// Add optional flags
	if args.Org != "" {
		cmdArgs = append(cmdArgs, "--org", args.Org)
	}
	if args.Project != "" {
		cmdArgs = append(cmdArgs, "--project", args.Project)
	}
	if args.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", args.ApiURL)
	}
	if args.Description != "" {
		cmdArgs = append(cmdArgs, "--description", args.Description)
	}
	if args.Filename != "" {
		cmdArgs = append(cmdArgs, "--filename", args.Filename)
	}

	// Add format flag for consistent output
	cmdArgs = append(cmdArgs, "--format", "json")

	return executeCommand(cmdArgs, fmt.Sprintf("push artifact '%s' to registry '%s'", args.Name, args.Registry))
}

// execPullCommand handles artifact download operations
func execPullCommand(args Args) error {
	logrus.Println("Executing pull command")

	// Validate required parameters for pull
	if args.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if args.Name == "" {
		return fmt.Errorf("package name must be set")
	}
	if args.Version == "" {
		return fmt.Errorf("package version must be set")
	}
	if args.Filename == "" {
		return fmt.Errorf("filename must be set")
	}
	if args.Destination == "" {
		return fmt.Errorf("destination path must be set")
	}
	if args.Token == "" {
		return fmt.Errorf("authentication token must be set")
	}
	if args.Account == "" {
		return fmt.Errorf("account ID must be set")
	}
	if args.PkgURL == "" {
		return fmt.Errorf("package URL must be set")
	}

	// Construct package path in the format expected by harness-cli: <package_name>/<version>/<filename>
	packagePath := fmt.Sprintf("%s/%s/%s", args.Name, args.Version, args.Filename)

	// Build Harness CLI command
	cmdArgs := []string{getHarnessBin(), "ar", "pull", "generic", args.Registry, packagePath, args.Destination}

	// Add required flags
	cmdArgs = append(cmdArgs, "--token", args.Token)
	cmdArgs = append(cmdArgs, "--account", args.Account)
	cmdArgs = append(cmdArgs, "--pkg-url", args.PkgURL)

	// Add optional flags
	if args.Org != "" {
		cmdArgs = append(cmdArgs, "--org", args.Org)
	}
	if args.Project != "" {
		cmdArgs = append(cmdArgs, "--project", args.Project)
	}
	if args.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", args.ApiURL)
	}

	// Add format flag for consistent output
	cmdArgs = append(cmdArgs, "--format", "json")

	return executeCommand(cmdArgs, fmt.Sprintf("pull artifact '%s' (version '%s', file '%s') from registry '%s' to '%s'", 
		args.Name, args.Version, args.Filename, args.Registry, args.Destination))
}

// execGetCommand handles artifact info retrieval
func execGetCommand(args Args) error {
	logrus.Println("Executing get command")

	if args.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if args.Name == "" {
		return fmt.Errorf("artifact name must be set")
	}
	if args.Token == "" {
		return fmt.Errorf("authentication token must be set")
	}
	if args.Account == "" {
		return fmt.Errorf("account ID must be set")
	}

	// Use 'hc ar get artifact' command with registry and name flags
	cmd := []string{
		getHarnessBin(),
		"ar", "get", "artifact", args.Name,
	}

	// Add required flags
	cmd = append(cmd, "--registry", args.Registry)
	cmd = append(cmd, "--token", args.Token)
	cmd = append(cmd, "--account", args.Account)
	if args.Org != "" {
		cmd = append(cmd, "--org", args.Org)
	}
	if args.Project != "" {
		cmd = append(cmd, "--project", args.Project)
	}
	if args.ApiURL != "" {
		cmd = append(cmd, "--api-url", args.ApiURL)
	}

	cmd = append(cmd, "--format", "json")

	logrus.Infof("Executing command: %s", strings.Join(cmd, " "))
	return executeCommand(cmd, fmt.Sprintf("get info for artifact '%s' in registry '%s'", args.Name, args.Registry))
}

// execDeleteCommand handles artifact deletion
func execDeleteCommand(args Args) error {
	logrus.Println("Executing delete command")

	// Validate required parameters for delete
	if args.Registry == "" {
		return fmt.Errorf("registry name must be set")
	}
	if args.Name == "" {
		return fmt.Errorf("artifact name must be set")
	}
	if args.Token == "" {
		return fmt.Errorf("authentication token must be set")
	}
	if args.Account == "" {
		return fmt.Errorf("account ID must be set")
	}

	// Build Harness CLI command - use 'hc ar delete artifact' with name as argument and registry as flag
	cmdArgs := []string{getHarnessBin(), "ar", "delete", "artifact", args.Name}

	// Add required flags
	cmdArgs = append(cmdArgs, "--registry", args.Registry)
	cmdArgs = append(cmdArgs, "--token", args.Token)
	cmdArgs = append(cmdArgs, "--account", args.Account)

	// Add optional flags
	if args.Org != "" {
		cmdArgs = append(cmdArgs, "--org", args.Org)
	}
	if args.Project != "" {
		cmdArgs = append(cmdArgs, "--project", args.Project)
	}
	if args.ApiURL != "" {
		cmdArgs = append(cmdArgs, "--api-url", args.ApiURL)
	}

	// Add format flag for consistent output
	cmdArgs = append(cmdArgs, "--format", "json")

	return executeCommand(cmdArgs, fmt.Sprintf("delete artifact '%s' from registry '%s'", args.Name, args.Registry))
}

// executeCommand executes a Harness CLI command
func executeCommand(cmdArgs []string, operation string) error {
	cmdStr := strings.Join(cmdArgs[:], " ")
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

	logrus.Printf("Successfully completed: %s", operation)
	return nil
}

func getShell() (string, string) {
	if runtime.GOOS == "windows" {
		// First check for PowerShell Core (pwsh.exe) which is used in PowerShell Nanoserver
		if _, err := os.Stat("C:/Program Files/PowerShell/pwsh.exe"); err == nil {
			return "pwsh", "-Command"
		}

		// Fall back to traditional PowerShell
		return "powershell", "-Command"
	}

	return "sh", "-c"
}

func getHarnessBin() string {
	if runtime.GOOS == "windows" {
		if _, err := os.Stat("C:/bin/hc.exe"); err == nil {
			return "C:/bin/hc.exe"
		}
	}
	return "hc"
}

func parseBoolOrDefault(defaultValue bool, s string) bool {
	if s == "" {
		return defaultValue
	}
	return strings.ToLower(s) == "true" || s == "1"
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
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
