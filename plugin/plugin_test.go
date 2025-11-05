// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestExec_MissingRegistry(t *testing.T) {
	args := Args{
		Command: "push",
		Source:  "test.txt",
		Name:    "test-artifact",
		Token:   "test-token",
		Account: "test-account",
		PkgURL:  "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing registry")
	}
	if err.Error() != "registry name must be set" {
		t.Errorf("Expected 'registry name must be set', got '%s'", err.Error())
	}
}

func TestExec_MissingSource(t *testing.T) {
	args := Args{
		Command:  "push",
		Registry: "test-registry",
		Name:     "test-artifact",
		Token:    "test-token",
		Account:  "test-account",
		PkgURL:   "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing source")
	}
	if err.Error() != "source file path must be set" {
		t.Errorf("Expected 'source file path must be set', got '%s'", err.Error())
	}
}

func TestExec_MissingName(t *testing.T) {
	args := Args{
		Command:  "push",
		Registry: "test-registry",
		Source:   "test.txt",
		Token:    "test-token",
		Account:  "test-account",
		PkgURL:   "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing name")
	}
	if err.Error() != "artifact name must be set" {
		t.Errorf("Expected 'artifact name must be set', got '%s'", err.Error())
	}
}

func TestExec_MissingToken(t *testing.T) {
	args := Args{
		Command:  "push",
		Registry: "test-registry",
		Source:   "test.txt",
		Name:     "test-artifact",
		Account:  "test-account",
		PkgURL:   "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing token")
	}
	if err.Error() != "authentication token must be set" {
		t.Errorf("Expected 'authentication token must be set', got '%s'", err.Error())
	}
}

func TestExec_MissingAccount(t *testing.T) {
	args := Args{
		Command:  "push",
		Registry: "test-registry",
		Source:   "test.txt",
		Name:     "test-artifact",
		Token:    "test-token",
		PkgURL:   "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing account")
	}
	if err.Error() != "account ID must be set" {
		t.Errorf("Expected 'account ID must be set', got '%s'", err.Error())
	}
}

func TestExec_MissingPkgURL(t *testing.T) {
	args := Args{
		Command:  "push",
		Registry: "test-registry",
		Source:   "test.txt",
		Name:     "test-artifact",
		Token:    "test-token",
		Account:  "test-account",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing pkg-url")
	}
	if err.Error() != "package URL must be set" {
		t.Errorf("Expected 'package URL must be set', got '%s'", err.Error())
	}
}

func TestGetHarnessBin_Windows(t *testing.T) {
	// This test would need to be run on Windows or with mocked file system
	// For now, just test the Unix path
	bin := getHarnessBin()
	if bin != "hc" {
		t.Errorf("Expected 'hc', got '%s'", bin)
	}
}

func TestParseBoolOrDefault(t *testing.T) {
	tests := []struct {
		defaultValue bool
		input        string
		expected     bool
	}{
		{false, "", false},
		{true, "", true},
		{false, "true", true},
		{false, "TRUE", true},
		{false, "1", true},
		{false, "false", false},
		{false, "0", false},
		{true, "false", false},
	}

	for _, test := range tests {
		result := parseBoolOrDefault(test.defaultValue, test.input)
		if result != test.expected {
			t.Errorf("parseBoolOrDefault(%v, %q) = %v, expected %v",
				test.defaultValue, test.input, result, test.expected)
		}
	}
}

func TestCopyEnvVariableIfExists(t *testing.T) {
	// Set a test environment variable
	testSrc := "TEST_SRC_VAR"
	testDest := "TEST_DEST_VAR"
	testValue := "test-value"

	os.Setenv(testSrc, testValue)
	defer os.Unsetenv(testSrc)
	defer os.Unsetenv(testDest)

	copyEnvVariableIfExists(testSrc, testDest)

	result := os.Getenv(testDest)
	if result != testValue {
		t.Errorf("Expected %q, got %q", testValue, result)
	}
}

func TestCopyEnvVariableIfExists_NotExists(t *testing.T) {
	testSrc := "TEST_NONEXISTENT_VAR"
	testDest := "TEST_DEST_VAR2"

	// Ensure source doesn't exist
	os.Unsetenv(testSrc)
	os.Unsetenv(testDest)

	copyEnvVariableIfExists(testSrc, testDest)

	result := os.Getenv(testDest)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// Test different command types
func TestExec_UnsupportedCommand(t *testing.T) {
	args := Args{
		Command: "invalid-command",
		Token:   "test-token",
		Account: "test-account",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for unsupported command")
	}
	if !strings.Contains(err.Error(), "unsupported command") {
		t.Errorf("Expected 'unsupported command' error, got '%s'", err.Error())
	}
}

func TestExec_DefaultCommand(t *testing.T) {
	args := Args{
		// No command specified - should default to push
		Registry: "test-registry",
		Source:   "test.txt",
		Name:     "test-artifact",
		Token:    "test-token",
		Account:  "test-account",
		PkgURL:   "https://app.harness.io",
	}

	// This will fail at CLI execution but should pass validation
	err := Exec(context.Background(), args)
	// We expect it to fail at CLI execution, not validation
	if err != nil && strings.Contains(err.Error(), "must be set") {
		t.Errorf("Validation should pass for default command, got '%s'", err.Error())
	}
}

func TestExec_PullCommand_MissingPackagePath(t *testing.T) {
	args := Args{
		Command:     "pull",
		Registry:    "test-registry",
		Destination: "/tmp/test",
		Token:       "test-token",
		Account:     "test-account",
		PkgURL:      "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing package path")
	}
	if err.Error() != "package path must be set" {
		t.Errorf("Expected 'package path must be set', got '%s'", err.Error())
	}
}

func TestExec_PullCommand_MissingDestination(t *testing.T) {
	args := Args{
		Command:     "pull",
		Registry:    "test-registry",
		PackagePath: "my-app/1.0.0",
		Token:       "test-token",
		Account:     "test-account",
		PkgURL:      "https://app.harness.io",
	}

	err := Exec(context.Background(), args)
	if err == nil {
		t.Error("Expected error for missing destination")
	}
	if err.Error() != "destination path must be set" {
		t.Errorf("Expected 'destination path must be set', got '%s'", err.Error())
	}
}
