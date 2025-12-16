#!/bin/bash
# Make sure the plugin is built first
go build -o drone-har

# Test script for resilient directory upload functionality
# This demonstrates the new features: path parameter and failure resilience

echo "=== RESILIENT DIRECTORY UPLOAD TEST ==="
echo "Testing enhanced generic package handler with:"
echo "1. Path parameter in commands"
echo "2. Failure resilience (continues on individual file failures)"
echo ""

# Add current directory to PATH so hc can be found
export PATH="$PATH:$(pwd)"

# Create test directory with various file types
TEST_DIR="/tmp/resilient-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR/subdir

# Create valid files
echo "Valid content 1" > $TEST_DIR/valid1.txt
echo "Valid content 2" > $TEST_DIR/valid2.md
echo "Subdirectory content" > $TEST_DIR/subdir/nested.json

# Create files that might cause issues (for testing resilience)
echo "File with special chars" > "$TEST_DIR/file-with-dashes.txt"
echo "Another valid file" > $TEST_DIR/another_valid.log

# Create a file with problematic name (to test skipping)
touch "$TEST_DIR/.hidden-file"

echo "Created test directory structure:"
find $TEST_DIR -type f -exec echo "  {}" \;
echo ""

# Set plugin environment variables
export PLUGIN_COMMAND=push
export PLUGIN_REGISTRY=testt
export PLUGIN_PACKAGE_TYPE=generic
export PLUGIN_SOURCE=$TEST_DIR
export PLUGIN_NAME=resilient_test
export PLUGIN_VERSION=1.0.0
export PLUGIN_DESCRIPTION="Testing resilient upload with path parameter"
export PLUGIN_TOKEN=pat.rfmnLA8cRVGqwC8S-Quo6A.691422fdbc4ac02fc793739c.T2OD9qky61ms6PWCFxcQ
export PLUGIN_ACCOUNT=rfmnLA8cRVGqwC8S-Quo6A
export PLUGIN_PKG_URL=https://pkg.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=jatintest

echo "Starting resilient directory upload test..."
echo "Expected behavior:"
echo "- Should process all valid files"
echo "- Should skip hidden files automatically"
echo "- Should continue even if some files fail"
echo "- Should show comprehensive summary at the end"
echo ""

# Run the plugin
./drone-har

echo ""
echo "=== TEST COMPLETED ==="
echo "Check the summary above to see:"
echo "✓ Successfully uploaded files"
echo "✗ Any failed uploads (should continue processing)"
echo "⚠ Any skipped files (hidden/invalid names)"

# Cleanup
rm -rf $TEST_DIR
