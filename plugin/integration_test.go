// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

//go:build integration
// +build integration

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Integration tests require real credentials
// Run with: go test -tags=integration -v ./plugin/...
//
// Required environment variables:
// - HARNESS_TOKEN
// - HARNESS_ACCOUNT
// - HARNESS_ORG (optional, defaults to "default")
// - HARNESS_PROJECT (optional, defaults to "jatintest")
// - HARNESS_PKG_URL (optional, defaults to "https://pkg.harness.io")

func getTestCredentials(t *testing.T) (token, account, org, project, pkgURL string) {
	token = os.Getenv("HARNESS_TOKEN")
	account = os.Getenv("HARNESS_ACCOUNT")
	org = os.Getenv("HARNESS_ORG")
	project = os.Getenv("HARNESS_PROJECT")
	pkgURL = os.Getenv("HARNESS_PKG_URL")

	if token == "" {
		t.Skip("HARNESS_TOKEN not set, skipping integration test")
	}
	if account == "" {
		t.Skip("HARNESS_ACCOUNT not set, skipping integration test")
	}

	// Set defaults
	if org == "" {
		org = "default"
	}
	if project == "" {
		project = "jatintest"
	}
	if pkgURL == "" {
		pkgURL = "https://pkg.harness.io"
	}

	return
}

func TestIntegration_PushArtifact(t *testing.T) {
	token, account, org, project, pkgURL := getTestCredentials(t)

	// Create a test file
	testFile := filepath.Join(t.TempDir(), "README.txt")
	testContent := "This is a test artifact for integration testing.\nCreated by drone-har plugin tests.\n"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	args := Args{
		Command:  "push",
		Registry: "testt",
		Source:   testFile,
		Name:     "sample-readme",
		Version:  "1.0.0",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
		PkgURL:   pkgURL,
	}

	t.Logf("Pushing artifact to registry 'testt'...")
	err = Exec(context.Background(), args)
	if err != nil {
		t.Fatalf("Failed to push artifact: %v", err)
	}

	t.Log("✅ Successfully pushed artifact")
}

func TestIntegration_GetArtifact(t *testing.T) {
	token, account, org, project, _ := getTestCredentials(t)

	args := Args{
		Command:  "get",
		Registry: "testt",
		Name:     "sample-readme",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
	}

	t.Logf("Getting artifact info from registry 'testt'...")
	err := Exec(context.Background(), args)
	if err != nil {
		t.Fatalf("Failed to get artifact info: %v", err)
	}

	t.Log("✅ Successfully retrieved artifact info")
}

func TestIntegration_PullArtifact(t *testing.T) {
	token, account, org, project, pkgURL := getTestCredentials(t)

	// Create a temporary download directory
	downloadDir := filepath.Join(t.TempDir(), "downloads")
	err := os.MkdirAll(downloadDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create download directory: %v", err)
	}

	args := Args{
		Command:     "pull",
		Registry:    "testt",
		Name:        "sample-readme",
		Version:     "1.0.0",
		Filename:    "README.txt",
		Destination: downloadDir,
		Token:       token,
		Account:     account,
		Org:         org,
		Project:     project,
		PkgURL:      pkgURL,
	}

	t.Logf("Pulling artifact from registry 'testt' to '%s'...", downloadDir)
	err = Exec(context.Background(), args)
	if err != nil {
		t.Fatalf("Failed to pull artifact: %v", err)
	}

	// Verify the file was downloaded
	downloadedFile := filepath.Join(downloadDir, "README.txt")
	if _, err := os.Stat(downloadedFile); os.IsNotExist(err) {
		t.Fatalf("Downloaded file not found at %s", downloadedFile)
	}

	t.Logf("✅ Successfully pulled artifact to %s", downloadedFile)
}

func TestIntegration_DeleteArtifact(t *testing.T) {
	token, account, org, project, _ := getTestCredentials(t)

	args := Args{
		Command:  "delete",
		Registry: "testt",
		Name:     "sample-readme",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
	}

	t.Logf("Deleting artifact from registry 'testt'...")
	err := Exec(context.Background(), args)
	if err != nil {
		t.Fatalf("Failed to delete artifact: %v", err)
	}

	t.Log("✅ Successfully deleted artifact")
}

// TestIntegration_FullWorkflow runs all 4 operations in sequence
func TestIntegration_FullWorkflow(t *testing.T) {
	token, account, org, project, pkgURL := getTestCredentials(t)

	// 1. PUSH
	t.Log("\n=== Step 1: PUSH Artifact ===")
	testFile := filepath.Join(t.TempDir(), "README.txt")
	testContent := "This is a test artifact for integration testing.\nCreated by drone-har plugin tests.\n"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pushArgs := Args{
		Command:  "push",
		Registry: "testt",
		Source:   testFile,
		Name:     "sample-readme",
		Version:  "1.0.0",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
		PkgURL:   pkgURL,
	}

	err = Exec(context.Background(), pushArgs)
	if err != nil {
		t.Fatalf("Failed to push artifact: %v", err)
	}
	t.Log("✅ Push completed successfully")

	// 2. GET
	t.Log("\n=== Step 2: GET Artifact Info ===")
	getArgs := Args{
		Command:  "get",
		Registry: "testt",
		Name:     "sample-readme",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
	}

	err = Exec(context.Background(), getArgs)
	if err != nil {
		t.Fatalf("Failed to get artifact info: %v", err)
	}
	t.Log("✅ Get completed successfully")

	// 3. PULL
	t.Log("\n=== Step 3: PULL Artifact ===")
	downloadDir := filepath.Join(t.TempDir(), "downloads")
	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create download directory: %v", err)
	}

	pullArgs := Args{
		Command:     "pull",
		Registry:    "testt",
		Name:        "sample-readme",
		Version:     "1.0.0",
		Filename:    "README.txt",
		Destination: downloadDir,
		Token:       token,
		Account:     account,
		Org:         org,
		Project:     project,
		PkgURL:      pkgURL,
	}

	err = Exec(context.Background(), pullArgs)
	if err != nil {
		t.Fatalf("Failed to pull artifact: %v", err)
	}

	// Verify the file was downloaded
	downloadedFile := filepath.Join(downloadDir, "README.txt")
	if _, err := os.Stat(downloadedFile); os.IsNotExist(err) {
		t.Fatalf("Downloaded file not found at %s", downloadedFile)
	}
	t.Log("✅ Pull completed successfully")

	// 4. DELETE
	t.Log("\n=== Step 4: DELETE Artifact ===")
	deleteArgs := Args{
		Command:  "delete",
		Registry: "testt",
		Name:     "sample-readme",
		Token:    token,
		Account:  account,
		Org:      org,
		Project:  project,
	}

	err = Exec(context.Background(), deleteArgs)
	if err != nil {
		t.Fatalf("Failed to delete artifact: %v", err)
	}
	t.Log("✅ Delete completed successfully")

	t.Log("\n🎉 Full workflow completed successfully!")
	t.Log("   ✅ Push")
	t.Log("   ✅ Get")
	t.Log("   ✅ Pull")
	t.Log("   ✅ Delete")
}

