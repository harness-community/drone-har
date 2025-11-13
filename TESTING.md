# Testing Documentation

## Overview

This document describes the testing strategy for the drone-har plugin. The plugin includes both unit tests and integration tests to ensure reliability and correctness.

## Test Results Summary

```
Total Tests: 26
  - Unit Tests: 21 ✅
  - Integration Tests: 5 ✅
Status: ALL PASSING
Coverage: 53.3%
```

## Test Types

### Unit Tests
Unit tests validate input parameters, error handling, and command generation without making actual API calls.

### Integration Tests
Integration tests execute real operations against Harness Artifact Registry to verify end-to-end functionality.

## Test Coverage

### ✅ Push Command Tests
- Missing registry validation
- Missing source validation
- Missing name validation
- Missing token validation
- Missing account validation
- Missing pkg_url validation
- Default command behavior (defaults to push)
- Real artifact upload (integration)

### ✅ Pull Command Tests
- Missing package name validation
- Missing version validation
- Missing filename validation
- Missing destination validation
- Real artifact download (integration)

### ✅ Get Command Tests
- Missing registry validation
- Missing name validation
- Real artifact info retrieval (integration)

### ✅ Delete Command Tests
- Missing registry validation
- Missing name validation
- Real artifact deletion (integration)

### ✅ Utility Tests
- parseBoolOrDefault function
- copyEnvVariableIfExists function
- getHarnessBin function
- Unsupported command handling

### ✅ Integration Tests
- Full workflow test (push → get → pull → delete)

## Generated CLI Commands

### 1. Push Command

**Plugin generates:**
```bash
hc artifact push generic <registry> <source> \
  --name <name> \
  --version <version> \
  --token <token> \
  --account <account> \
  --pkg-url <pkg-url> \
  --org <org> \
  --project <project> \
  --format json
```

**Matches your working command:**
```bash
hc artifact push generic testt sample-artifacts/README.txt \
  --name sample-readme \
  --version 1.0.0 \
  --pkg-url https://pkg.harness.io
```

### 2. Get Command

**Plugin generates:**
```bash
hc artifact get <name> \
  --registry <registry> \
  --token <token> \
  --account <account> \
  --org <org> \
  --project <project> \
  --format json
```

**Matches your working command:**
```bash
hc artifact get sample-readme \
  --registry testt \
  --org default \
  --project jatintest \
  --format json
```

### 3. Pull Command

**Plugin generates:**
```bash
hc artifact pull generic <registry> <name>/<version>/<filename> <destination> \
  --token <token> \
  --account <account> \
  --pkg-url <pkg-url> \
  --org <org> \
  --project <project> \
  --format json
```

**Matches your working command:**
```bash
hc artifact pull generic testt sample-readme/1.0.0/README.txt ./downloads \
  --pkg-url https://pkg.harness.io \
  --org default \
  --project jatintest
```

### 4. Delete Command

**Plugin generates:**
```bash
hc artifact delete <name> \
  --registry <registry> \
  --token <token> \
  --account <account> \
  --org <org> \
  --project <project> \
  --format json
```

**Matches your working command:**
```bash
hc artifact delete sample-readme \
  --registry testt \
  --org default \
  --project jatintest
```

## Example Pipeline Configurations

### Push Example
```yaml
- name: upload-artifact
  image: harness/drone-har
  settings:
    command: push
    registry: testt
    source: sample-artifacts/README.txt
    name: sample-readme
    version: 1.0.0
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: default
    project: jatintest
    pkg_url: https://pkg.harness.io
```

### Get Example
```yaml
- name: get-artifact-info
  image: harness/drone-har
  settings:
    command: get
    registry: testt
    name: sample-readme
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: default
    project: jatintest
```

### Pull Example
```yaml
- name: download-artifact
  image: harness/drone-har
  settings:
    command: pull
    registry: testt
    name: sample-readme
    version: 1.0.0
    filename: README.txt
    destination: ./downloads
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: default
    project: jatintest
    pkg_url: https://pkg.harness.io
```

### Delete Example
```yaml
- name: delete-artifact
  image: harness/drone-har
  settings:
    command: delete
    registry: testt
    name: sample-readme
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: default
    project: jatintest
```

## Running Tests

### Unit Tests Only

```bash
# Using make
make test

# Or directly with go
go test -v ./plugin/...

# With coverage
make test-coverage
# or
go test -cover ./plugin/...
```

### Integration Tests Only

Integration tests require valid Harness credentials:

```bash
# Set credentials
export HARNESS_TOKEN="your-harness-token"
export HARNESS_ACCOUNT="your-account-id"
export HARNESS_ORG="default"              # optional
export HARNESS_PROJECT="your-project"     # optional
export HARNESS_PKG_URL="https://pkg.harness.io"  # optional

# Run integration tests
make test-integration
# or
go test -tags=integration -v ./plugin/...
```

### Using the Helper Script

```bash
# Make the script executable
chmod +x run_integration_tests.sh

# Run all integration tests
./run_integration_tests.sh

# Run specific test
./run_integration_tests.sh full    # Full workflow test
./run_integration_tests.sh push    # Push test only
./run_integration_tests.sh get     # Get test only
./run_integration_tests.sh pull    # Pull test only
./run_integration_tests.sh delete  # Delete test only
```

### Run All Tests

```bash
make test-all
```

### Run Specific Test

```bash
go test -v -run TestExec_PushCommand ./plugin/...
go test -tags=integration -v -run TestIntegration_FullWorkflow ./plugin/...
```

## Key Changes Made

1. ✅ Changed from `hc ar` to `hc artifact` for all commands
2. ✅ Updated push command structure
3. ✅ Updated pull command structure
4. ✅ Updated get command structure
5. ✅ Updated delete command structure
6. ✅ All commands now match the actual harness-cli command format
7. ✅ Added comprehensive validation tests for all 4 operations

## Validation

All parameter validations are working correctly:

- **Push**: Requires registry, source, name, token, account, pkg_url
- **Pull**: Requires registry, name, version, filename, destination, token, account, pkg_url
- **Get**: Requires registry, name, token, account
- **Delete**: Requires registry, name, token, account

Optional parameters (org, project, api_url, description, etc.) are properly handled.
