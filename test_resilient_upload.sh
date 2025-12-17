#!/bin/bash
# Make sure the plugin is built first
go build -o drone-har

# Test script for single file upload functionality
# This demonstrates the single file upload feature

echo "=== SINGLE FILE UPLOAD TEST ==="
echo "Testing generic package handler with:"
echo "1. Single file upload only"
echo "2. Directory uploads are rejected"
echo ""

# Add current directory to PATH so hc can be found
export PATH="$PATH:$(pwd)"

# Create test file for single file upload
TEST_FILE="/tmp/test-artifact.txt"
rm -f $TEST_FILE

# Create a single test file
echo "This is a test artifact for single file upload" > $TEST_FILE
echo "Created by drone-har test script" >> $TEST_FILE
echo "Timestamp: $(date)" >> $TEST_FILE

echo "Created test file: $TEST_FILE"
echo "File contents:"
cat $TEST_FILE
echo ""

# Set plugin environment variables
export PLUGIN_COMMAND=push
export PLUGIN_REGISTRY=generic-local-1
export PLUGIN_PACKAGE_TYPE=GENERIC
export PLUGIN_SOURCE=$TEST_FILE
export PLUGIN_NAME=resilient_test
export PLUGIN_VERSION=1.0.0
export PLUGIN_DESCRIPTION="Testing single file upload"
export PLUGIN_TOKEN=eyJhbGciOiJIUzI1NiJ9
export PLUGIN_ACCOUNT=iWnhltqOT7GFt7R-F_zP7Q
export PLUGIN_PKG_URL=https://pkg.qa.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=sourabh_test

echo "Starting single file upload test..."
echo "Expected behavior:"
echo "- Should upload the single file successfully"
echo "- Should reject directories if provided"
echo "- Should fail the entire step if upload fails"
echo ""

# Run the plugin
./drone-har

echo ""
echo "=== TEST COMPLETED ==="
echo "Check the output above to see:"
echo "✓ Single file upload result"
echo "✗ Any upload failures (should fail entire step)"

# Cleanup
rm -f $TEST_FILE
