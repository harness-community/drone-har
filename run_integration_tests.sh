#!/bin/bash

# Integration Test Runner for drone-har plugin
# This script runs integration tests against a real Harness Artifact Registry

set -e

echo "🔧 Harness Artifact Registry - Integration Test Runner"
echo "======================================================="
echo ""

# Check if credentials are provided via environment variables
if [ -z "$HARNESS_TOKEN" ]; then
    echo "⚠️  HARNESS_TOKEN environment variable is not set"
    echo ""
    echo "Please set your Harness credentials:"
    echo "  export HARNESS_TOKEN='your-token'"
    echo "  export HARNESS_ACCOUNT='your-account-id'"
    echo "  export HARNESS_ORG='default'           # optional"
    echo "  export HARNESS_PROJECT='jatintest'     # optional"
    echo "  export HARNESS_PKG_URL='https://pkg.harness.io'  # optional"
    echo ""
    exit 1
fi

if [ -z "$HARNESS_ACCOUNT" ]; then
    echo "⚠️  HARNESS_ACCOUNT environment variable is not set"
    exit 1
fi

# Set defaults
export HARNESS_ORG=${HARNESS_ORG:-"default"}
export HARNESS_PROJECT=${HARNESS_PROJECT:-"jatintest"}
export HARNESS_PKG_URL=${HARNESS_PKG_URL:-"https://pkg.harness.io"}

echo "✅ Credentials configured:"
echo "   Account: $HARNESS_ACCOUNT"
echo "   Org: $HARNESS_ORG"
echo "   Project: $HARNESS_PROJECT"
echo "   Pkg URL: $HARNESS_PKG_URL"
echo ""

# Run integration tests
echo "🧪 Running integration tests..."
echo ""

if [ "$1" == "full" ]; then
    echo "Running full workflow test..."
    go test -tags=integration -v -run TestIntegration_FullWorkflow ./plugin/...
elif [ "$1" == "push" ]; then
    echo "Running push test..."
    go test -tags=integration -v -run TestIntegration_PushArtifact ./plugin/...
elif [ "$1" == "get" ]; then
    echo "Running get test..."
    go test -tags=integration -v -run TestIntegration_GetArtifact ./plugin/...
elif [ "$1" == "pull" ]; then
    echo "Running pull test..."
    go test -tags=integration -v -run TestIntegration_PullArtifact ./plugin/...
elif [ "$1" == "delete" ]; then
    echo "Running delete test..."
    go test -tags=integration -v -run TestIntegration_DeleteArtifact ./plugin/...
else
    echo "Running all integration tests..."
    go test -tags=integration -v ./plugin/...
fi

echo ""
echo "✅ Integration tests completed!"

