// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package packages

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
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

	logrus.Printf("Source path: %s", config.Source)

	cmdArgs, err := buildPushCommand(Generic, config, version, config.Source, config.Name, true)
	if err != nil {
		return err
	}
	return executeCommand(config, cmdArgs, fmt.Sprintf("push artifact '%s' to registry '%s'", config.Name, config.Registry))
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

	// harness-har refuses a destination that does not exist yet.
	if err := os.MkdirAll(config.Destination, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory '%s': %w", config.Destination, err)
	}

	harnessBin, err := getHarnessBin()
	if err != nil {
		return err
	}
	cmdArgs := []string{
		harnessBin, "pull", "artifact", config.Registry + "/" + config.Name, config.Destination,
		"--version", config.Version,
		"--filename", config.Filename,
		"--package-type", "generic",
	}

	return executeCommand(config, cmdArgs, fmt.Sprintf("pull artifact '%s' (version '%s', file '%s') from registry '%s' to '%s'",
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

	harnessBin, err := getHarnessBin()
	if err != nil {
		return err
	}
	cmdArgs := []string{harnessBin, "get", "artifact", config.Registry + "/" + config.Name, "--format", "json"}

	return executeCommand(config, cmdArgs, fmt.Sprintf("get info for artifact '%s' in registry '%s'", config.Name, config.Registry))
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

	harnessBin, err := getHarnessBin()
	if err != nil {
		return err
	}
	// --force skips the confirmation prompt, which a pipeline step cannot answer.
	cmdArgs := []string{harnessBin, "delete", "artifact", config.Registry + "/" + config.Name, "--force"}

	return executeCommand(config, cmdArgs, fmt.Sprintf("delete artifact '%s' from registry '%s'", config.Name, config.Registry))
}

// Helper functions

// cliAuthEnv returns the variables that authenticate harness-har. A pat., sat.
// or st. API key uses the CLI's env mode; any other token is the pipeline's
// build token (a JWT), which only the CLI's pipeline mode sends as
// "Authorization: CIManager <token>". Plugin settings win over the pipeline's
// own HARNESS_*_ID variables, which only fill gaps.
func cliAuthEnv(config Config) (map[string]string, error) {
	org := firstNonEmpty(config.Org, os.Getenv("HARNESS_ORG_ID"))
	project := firstNonEmpty(config.Project, os.Getenv("HARNESS_PROJECT_ID"))
	if org == "" || project == "" {
		return nil, fmt.Errorf("organization and project must be set (PLUGIN_ORG and PLUGIN_PROJECT)")
	}

	token := config.Token
	switch {
	case strings.HasPrefix(token, "pat."), strings.HasPrefix(token, "sat."), strings.HasPrefix(token, "st."):
		// HARNESS_PROFILE=env also stops the CLI from switching to pipeline
		// mode just because HARNESS_PIPELINE_ID is set.
		return map[string]string{
			"HARNESS_PROFILE":      "env",
			"HARNESS_API_KEY":      token,
			"HARNESS_ACCOUNT":      config.Account,
			"HARNESS_ORG":          org,
			"HARNESS_PROJECT":      project,
			"HARNESS_API_URL":      config.ApiURL,
			"HARNESS_REGISTRY_URL": config.PkgURL,
		}, nil
	default:
		return map[string]string{
			"HARNESS_PROFILE":     "pipeline",
			"HARNESS_HAR_TOKEN":   token,
			"HARNESS_HAR_API_URL": config.ApiURL,
			"HARNESS_HAR_URL":     config.PkgURL,
			"HARNESS_ACCOUNT_ID":  firstNonEmpty(config.Account, os.Getenv("HARNESS_ACCOUNT_ID")),
			"HARNESS_ORG_ID":      org,
			"HARNESS_PROJECT_ID":  project,
		}, nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// commandEnv is the plugin's own environment with vars replacing any existing
// entries of the same name.
func commandEnv(vars map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(vars))
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if _, ok := vars[key]; ok {
			continue
		}
		env = append(env, kv)
	}
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+vars[key])
	}
	return env
}

// buildPushCommand builds a common push command for any package type
func buildPushCommand(packageType PackageType, config Config, version, filePath, artifactName string, includeFileAndVersion bool) ([]string, error) {
	harnessBin, err := getHarnessBin()
	if err != nil {
		return nil, err
	}

	// Only generic packages take the artifact name and version from the
	// plugin; every other type reads them from the package itself.
	target := config.Registry
	if includeFileAndVersion {
		target += "/" + artifactName
	}
	cmdArgs := []string{harnessBin, "push", "artifact:" + strings.ToLower(string(packageType)), target, filePath}
	if includeFileAndVersion {
		cmdArgs = append(cmdArgs, "--version", version)
	}
	if packageType == Maven {
		cmdArgs = append(cmdArgs, "--pom-file", config.PomFile)
	}

	if config.Filename != "" {
		logrus.Warnf("PLUGIN_FILENAME is ignored for push: the file is stored under its own name")
	}

	return cmdArgs, nil
}

// executeCommand runs a harness-har command authenticated for config.
func executeCommand(config Config, cmdArgs []string, operation string) error {
	vars, err := cliAuthEnv(config)
	if err != nil {
		return err
	}
	// The plugin pins its CLI version, so the CLI's own update notice is noise.
	vars["HARNESS_NO_UPDATE_CHECK"] = "1"

	cmdStr := strings.Join(cmdArgs, " ")
	logrus.Printf("Executing command: %s", cmdStr)

	// Execute command directly without shell to avoid argument parsing issues
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = commandEnv(vars)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)

	if err := cmd.Run(); err != nil {
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
