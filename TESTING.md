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

The plugin runs the embedded `harness-har` binary. Credentials are never passed
as flags: the plugin sets them in the child process environment.

- Pipeline build token (JWT): `HARNESS_PROFILE=pipeline` with `HARNESS_HAR_TOKEN`,
  `HARNESS_HAR_API_URL`, `HARNESS_HAR_URL` and `HARNESS_ACCOUNT_ID`/`ORG_ID`/`PROJECT_ID`.
  The CLI sends `Authorization: CIManager <token>`.
- PAT, SAT or scoped token (`st.`): `HARNESS_PROFILE=env` with `HARNESS_API_KEY`, `HARNESS_ACCOUNT`,
  `HARNESS_ORG`, `HARNESS_PROJECT`, `HARNESS_API_URL` and `HARNESS_REGISTRY_URL`.
  The CLI sends `x-api-key: <token>`.

### 1. Push Command

```bash
harness-har push artifact:generic <registry>/<name> <source> --version <version>
```

### 2. Get Command

```bash
harness-har get artifact <registry>/<name> --format json
```

### 3. Pull Command

```bash
harness-har pull artifact <registry>/<name> <destination> \
  --version <version> --filename <filename> --package-type generic
```

### 4. Delete Command

```bash
harness-har delete artifact <registry>/<name> --force
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

1. ✅ Moved from the deprecated `hc` to the embedded `harness-har` binary
2. ✅ Credentials are passed through the environment instead of `~/.harness/auth.json`
3. ✅ Pipeline build tokens use the CLI's pipeline auth; PATs, SATs and scoped tokens use env auth
4. ✅ Added comprehensive validation tests for all 4 operations

## Validation

All parameter validations are working correctly:

- **Push**: Requires registry, source, name, token, account, pkg_url
- **Pull**: Requires registry, name, version, filename, destination, token, account, pkg_url
- **Get**: Requires registry, name, token, account
- **Delete**: Requires registry, name, token, account

Optional parameters (org, project, api_url, description, etc.) are properly handled.
