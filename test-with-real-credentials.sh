#!/bin/bash

# Test script for droneHarPlugin with REAL Harness credentials
# ⚠️  ONLY run this if you have valid Harness CLI and credentials

set -e

echo "🔐 Testing droneHarPlugin with REAL credentials..."
echo "=================================================="
echo "⚠️  WARNING: This will make actual API calls to Harness!"
echo ""

# Check if Harness CLI is available
if ! command -v hc &> /dev/null; then
    echo "❌ Harness CLI (hc) not found!"
    echo "   Please install Harness CLI first:"
    echo "   curl -L https://github.com/harness/harness-cli/releases/latest/download/hc-linux-amd64 -o /usr/local/bin/hc"
    echo "   chmod +x /usr/local/bin/hc"
    exit 1
fi

echo "✅ Harness CLI found: $(which hc)"

# Check if user is authenticated
echo "🔍 Checking Harness CLI authentication..."
if hc auth status &> /dev/null; then
    echo "✅ Harness CLI is authenticated"
else
    echo "❌ Harness CLI not authenticated!"
    echo "   Please run: hc auth login"
    echo "   Or set up authentication with token"
    exit 1
fi

# Prompt for required parameters
echo ""
echo "📝 Please provide the following information:"
read -p "Registry name: " REGISTRY_NAME
read -p "Account ID: " ACCOUNT_ID
read -p "Organization ID (optional): " ORG_ID
read -p "Project ID (optional): " PROJECT_ID
read -s -p "API Token: " API_TOKEN
echo ""

# Validate inputs
if [[ -z "$REGISTRY_NAME" || -z "$ACCOUNT_ID" || -z "$API_TOKEN" ]]; then
    echo "❌ Registry name, Account ID, and API Token are required!"
    exit 1
fi

# Build the plugin
echo "📦 Building plugin..."
go build -o droneHarPlugin .

# Set environment variables
export PLUGIN_REGISTRY="$REGISTRY_NAME"
export PLUGIN_SOURCE="test-sample.txt"
export PLUGIN_NAME="droneHarPlugin-test"
export PLUGIN_VERSION="1.0.0-test-$(date +%s)"
export PLUGIN_TOKEN="$API_TOKEN"
export PLUGIN_ACCOUNT="$ACCOUNT_ID"
export PLUGIN_PKG_URL="https://app.harness.io"
export PLUGIN_DESCRIPTION="Test upload from droneHarPlugin"
export PLUGIN_LOG_LEVEL="debug"

if [[ -n "$ORG_ID" ]]; then
    export PLUGIN_ORG="$ORG_ID"
fi

if [[ -n "$PROJECT_ID" ]]; then
    export PLUGIN_PROJECT="$PROJECT_ID"
fi

echo ""
echo "🚀 Testing plugin with real credentials..."
echo "Registry: $PLUGIN_REGISTRY"
echo "Artifact: $PLUGIN_NAME"
echo "Version: $PLUGIN_VERSION"
echo "Source: $PLUGIN_SOURCE"

# Run the plugin
if ./droneHarPlugin; then
    echo ""
    echo "🎉 SUCCESS! Plugin worked correctly!"
    echo "✅ Artifact uploaded to Harness Artifact Registry"
    echo ""
    echo "You can now:"
    echo "1. Check your Harness UI to see the uploaded artifact"
    echo "2. Push this plugin to Git"
    echo "3. Build and publish the Docker image"
else
    echo ""
    echo "❌ Plugin execution failed!"
    echo "Check the error messages above for details"
    exit 1
fi

# Clean up environment variables
unset PLUGIN_REGISTRY PLUGIN_SOURCE PLUGIN_NAME PLUGIN_VERSION 
unset PLUGIN_TOKEN PLUGIN_ACCOUNT PLUGIN_PKG_URL PLUGIN_DESCRIPTION PLUGIN_LOG_LEVEL
unset PLUGIN_ORG PLUGIN_PROJECT
