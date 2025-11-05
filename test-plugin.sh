#!/bin/bash

# Test script for droneHarPlugin
# This script helps validate the plugin works correctly before deployment

set -e

echo "🧪 Testing droneHarPlugin..."
echo "=================================="

# Step 1: Build the plugin
echo "📦 Building plugin..."
go build -o droneHarPlugin .
echo "✅ Plugin built successfully"

# Step 2: Run unit tests
echo "🔬 Running unit tests..."
go test ./... -v
echo "✅ Unit tests passed"

# Step 3: Test plugin validation (should fail with missing params)
echo "🔍 Testing parameter validation..."
echo "Testing missing registry..."
if ./droneHarPlugin 2>&1 | grep -q "registry name must be set"; then
    echo "✅ Registry validation works"
else
    echo "❌ Registry validation failed"
    exit 1
fi

# Step 4: Test with environment variables (dry run)
echo "🌍 Testing with environment variables..."
export PLUGIN_REGISTRY="test-registry"
export PLUGIN_SOURCE="test-sample.txt"
export PLUGIN_NAME="test-artifact"
export PLUGIN_VERSION="1.0.0-test"
export PLUGIN_TOKEN="fake-token-for-testing"
export PLUGIN_ACCOUNT="fake-account-for-testing"
export PLUGIN_PKG_URL="https://app.harness.io"
export PLUGIN_LOG_LEVEL="debug"

echo "Environment variables set:"
echo "  PLUGIN_REGISTRY=$PLUGIN_REGISTRY"
echo "  PLUGIN_SOURCE=$PLUGIN_SOURCE"
echo "  PLUGIN_NAME=$PLUGIN_NAME"
echo "  PLUGIN_VERSION=$PLUGIN_VERSION"
echo "  PLUGIN_TOKEN=***hidden***"
echo "  PLUGIN_ACCOUNT=***hidden***"
echo "  PLUGIN_PKG_URL=$PLUGIN_PKG_URL"

# This will fail because we don't have real Harness CLI or credentials
# But it will show us the command being generated
echo "🚀 Testing plugin execution (will fail at CLI call, but shows command generation)..."
if ./droneHarPlugin 2>&1; then
    echo "✅ Plugin executed successfully"
else
    echo "⚠️  Plugin failed as expected (no real Harness CLI/credentials)"
    echo "   This is normal for local testing"
fi

# Clean up
unset PLUGIN_REGISTRY PLUGIN_SOURCE PLUGIN_NAME PLUGIN_VERSION PLUGIN_TOKEN PLUGIN_ACCOUNT PLUGIN_PKG_URL PLUGIN_LOG_LEVEL

echo ""
echo "🎉 Plugin testing completed!"
echo "=================================="
echo "✅ Build: SUCCESS"
echo "✅ Unit Tests: SUCCESS" 
echo "✅ Parameter Validation: SUCCESS"
echo "⚠️  CLI Execution: Expected failure (no real credentials)"
echo ""
echo "📋 Next Steps:"
echo "1. Test with real Harness credentials (see test-with-real-credentials.sh)"
echo "2. Build Docker image for deployment"
echo "3. Push to Git repository"
