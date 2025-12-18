#!/bin/bash
# Make sure the plugin is built first
go build -o drone-har

# Test script for single file upload functionality
# This demonstrates the single file upload feature

echo "=== PACKAGE UPLOAD TEST ==="
echo "Testing multiple package handlers:"
echo "1. Generic package (single file upload)"
echo "2. NPM package (.tgz tarball file)"
echo "3. RPM package (.rpm file)"
echo "4. Conda package (.conda/.tar.bz2 file)"
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
export PLUGIN_NAME=resilient_test-2
export PLUGIN_VERSION=1.0.0
export PLUGIN_DESCRIPTION="Testing single file upload"
export PLUGIN_TOKEN=eyJhbGciOiJIUzI1NiJ9.
export PLUGIN_ACCOUNT=iWnhltqOT7GFt7R-F_zP7Q
export PLUGIN_PKG_URL=https://pkg.qa.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=sourabh_test

echo "=== GENERIC PACKAGE TEST ==="
echo "Starting generic package upload test..."
echo "Expected behavior:"
echo "- Should upload the single file successfully"
echo "- Should reject directories if provided"
echo "- Should fail the entire step if upload fails"
echo ""

# Run the plugin for generic package
./drone-har

echo ""
echo "=== NPM PACKAGE TEST ==="

# Create NPM test package
NPM_TEST_DIR="/tmp/npm-test-package"
NPM_TAR_FILE="/tmp/test-npm-package-1.0.0.tgz"
rm -rf $NPM_TEST_DIR
rm -f $NPM_TAR_FILE
mkdir -p $NPM_TEST_DIR

# Create package.json
cat > $NPM_TEST_DIR/package.json << EOF
{
  "name": "test-npm-package",
  "version": "1.0.0",
  "description": "Test NPM package for drone-har",
  "main": "index.js",
  "author": "Test Author",
  "license": "MIT"
}
EOF

# Create index.js
cat > $NPM_TEST_DIR/index.js << EOF
console.log("Hello from test NPM package!");
module.exports = {
  greeting: "Hello World"
};
EOF

# Create NPM tarball (.tgz file)
cd /tmp
tar -czf test-npm-package-1.0.0.tgz -C npm-test-package .
cd - > /dev/null

echo "Created NPM test package directory at: $NPM_TEST_DIR"
echo "Created NPM tarball at: $NPM_TAR_FILE"
echo "Package.json contents:"
cat $NPM_TEST_DIR/package.json
echo ""

# Set NPM plugin environment variables
export PLUGIN_COMMAND=push
export PLUGIN_REGISTRY=npm-local-1
export PLUGIN_PACKAGE_TYPE=NPM
export PLUGIN_SOURCE=$NPM_TAR_FILE
# Note: No PLUGIN_NAME or PLUGIN_VERSION for NPM - derived from package.json
export PLUGIN_DESCRIPTION="Testing NPM package upload"
export PLUGIN_TOKEN=eyJhbGciOiJIUzI1NiJ9.eyJhY2NvdW50SWQiOiJpV25obHRxT1Q3R0Z0N1ItRl96UDdRIiwicm9sZSI6IiIsImlzcyI6Ikhhcm5lc3MgSW5jIiwibmFtZSI6ImRpeDhnTjdnUTdXTW55OTlyRXA0LWciLCJhbGxvd2VkUmVzb3VyY2VzIjpbImh0dHBzOi8vcGtnLnFhLmhhcm5lc3MuaW8iXSwiZXhwIjoxNzY2MDM0NTc3LCJ0eXBlIjoiVVNFUiIsImlhdCI6MTc2NTk0ODE3NywiZW1haWwiOiJzb3VyYWJoLmF3YXNodGlAaGFybmVzcy5pbyIsInVuaXF1ZUlkIjoiZGl4OGdON2dRN1dNbnk5OXJFcDQtZyIsInVzZXJuYW1lIjoiU291cmFiaCBhd2FzaHRpIn0.I320K0riKFgWpNx6XtpPGODLb2i1_lN3J0AtNmFlgc8
export PLUGIN_ACCOUNT=iWnhltqOT7GFt7R-F_zP7Q
export PLUGIN_PKG_URL=https://pkg.qa.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=sourabh_test

echo "Starting NPM package upload test..."
echo "Expected behavior:"
echo "- Should read package name and version from tarball metadata"
echo "- Should upload the NPM .tgz file successfully"
echo "- Should accept .tgz tarball files (single file)"
echo ""

# Run the plugin for NPM package
./drone-har

echo ""
echo "=== RPM PACKAGE TEST ==="

# Create RPM test package (mock RPM file)
RPM_TEST_FILE="/tmp/test-rpm-package-1.0.0-1.x86_64.rpm"
rm -f $RPM_TEST_FILE

# Create a mock RPM file (in real scenario, this would be built with rpmbuild)
# For testing purposes, we'll create a simple file with RPM-like name
cat > $RPM_TEST_FILE << 'EOF'
This is a mock RPM package file for testing purposes.
In a real scenario, this would be a binary RPM file created with rpmbuild.

Package: test-rpm-package
Version: 1.0.0
Release: 1
Architecture: x86_64
Summary: Test RPM package for drone-har
Description: This is a test RPM package used to validate the RPM handler functionality.
EOF

echo "Created RPM test package at: $RPM_TEST_FILE"
echo "RPM file size: $(ls -lh $RPM_TEST_FILE | awk '{print $5}')"
echo ""

# Set RPM plugin environment variables
export PLUGIN_COMMAND=push
export PLUGIN_REGISTRY=rpm-local-1
export PLUGIN_PACKAGE_TYPE=RPM
export PLUGIN_SOURCE=$RPM_TEST_FILE
# Note: No PLUGIN_NAME or PLUGIN_VERSION for RPM - derived from RPM metadata
export PLUGIN_DESCRIPTION="Testing RPM package upload"
export PLUGIN_TOKEN=eyJhbGciOiJIUzI1NiJ9.eyJhY2NvdW50SWQiOiJpV25obHRxT1Q3R0Z0N1ItRl96UDdRIiwicm9sZSI6IiIsImlzcyI6Ikhhcm5lc3MgSW5jIiwibmFtZSI6ImRpeDhnTjdnUTdXTW55OTlyRXA0LWciLCJhbGxvd2VkUmVzb3VyY2VzIjpbImh0dHBzOi8vcGtnLnFhLmhhcm5lc3MuaW8iXSwiZXhwIjoxNzY2MDM0NTc3LCJ0eXBlIjoiVVNFUiIsImlhdCI6MTc2NTk0ODE3NywiZW1haWwiOiJzb3VyYWJoLmF3YXNodGlAaGFybmVzcy5pbyIsInVuaXF1ZUlkIjoiZGl4OGdON2dRN1dNbnk5OXJFcDQtZyIsInVzZXJuYW1lIjoiU291cmFiaCBhd2FzaHRpIn0.I320K0riKFgWpNx6XtpPGODLb2i1_lN3J0AtNmFlgc8
export PLUGIN_ACCOUNT=iWnhltqOT7GFt7R-F_zP7Q
export PLUGIN_PKG_URL=https://pkg.qa.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=sourabh_test

echo "Starting RPM package upload test..."
echo "Expected behavior:"
echo "- Should read package name and version from RPM metadata"
echo "- Should upload the RPM .rpm file successfully"
echo "- Should accept .rpm files (single file)"
echo ""

# Run the plugin for RPM package
./drone-har

echo ""
echo "=== CONDA PACKAGE TEST ==="

# Create Conda test package (mock Conda file)
CONDA_TEST_FILE="/tmp/test-conda-package-1.0.0-py39_0.conda"
rm -f $CONDA_TEST_FILE

# Create a mock Conda package file (in real scenario, this would be built with conda-build)
# For testing purposes, we'll create a simple file with Conda-like name
cat > $CONDA_TEST_FILE << 'EOF'
This is a mock Conda package file for testing purposes.
In a real scenario, this would be a binary Conda package file created with conda-build.

Package: test-conda-package
Version: 1.0.0
Build: py39_0
Platform: linux-64
Summary: Test Conda package for drone-har
Description: This is a test Conda package used to validate the Conda handler functionality.
Dependencies:
  - python >=3.9
  - numpy
EOF

echo "Created Conda test package at: $CONDA_TEST_FILE"
echo "Conda file size: $(ls -lh $CONDA_TEST_FILE | awk '{print $5}')"
echo ""

# Set Conda plugin environment variables
export PLUGIN_COMMAND=push
export PLUGIN_REGISTRY=conda-local-1
export PLUGIN_PACKAGE_TYPE=CONDA
export PLUGIN_SOURCE=$CONDA_TEST_FILE
# Note: No PLUGIN_NAME or PLUGIN_VERSION for Conda - derived from Conda metadata
export PLUGIN_DESCRIPTION="Testing Conda package upload"
export PLUGIN_TOKEN=eyJhbGciOiJIUzI1NiJ9.eyJhY2NvdW50SWQiOiJpV25obHRxT1Q3R0Z0N1ItRl96UDdRIiwicm9sZSI6IiIsImlzcyI6Ikhhcm5lc3MgSW5jIiwibmFtZSI6ImRpeDhnTjdnUTdXTW55OTlyRXA0LWciLCJhbGxvd2VkUmVzb3VyY2VzIjpbImh0dHBzOi8vcGtnLnFhLmhhcm5lc3MuaW8iXSwiZXhwIjoxNzY2MDM0NTc3LCJ0eXBlIjoiVVNFUiIsImlhdCI6MTc2NTk0ODE3NywiZW1haWwiOiJzb3VyYWJoLmF3YXNodGlAaGFybmVzcy5pbyIsInVuaXF1ZUlkIjoiZGl4OGdON2dRN1dNbnk5OXJFcDQtZyIsInVzZXJuYW1lIjoiU291cmFiaCBhd2FzaHRpIn0.I320K0riKFgWpNx6XtpPGODLb2i1_lN3J0AtNmFlgc8
export PLUGIN_ACCOUNT=iWnhltqOT7GFt7R-F_zP7Q
export PLUGIN_PKG_URL=https://pkg.qa.harness.io
export PLUGIN_ORG=default
export PLUGIN_PROJECT=sourabh_test

echo "Starting Conda package upload test..."
echo "Expected behavior:"
echo "- Should read package name and version from Conda metadata"
echo "- Should upload the Conda .conda file successfully"
echo "- Should accept .conda/.tar.bz2 files (single file)"
echo ""

# Run the plugin for Conda package
./drone-har

echo ""
echo "=== ALL TESTS COMPLETED ==="
echo "Check the output above to see:"
echo "✓ Generic package upload result"
echo "✓ NPM package upload result"
echo "✓ RPM package upload result"
echo "✓ Conda package upload result"
echo "✗ Any upload failures (should fail entire step)"

# Cleanup
rm -f $TEST_FILE
rm -rf $NPM_TEST_DIR
rm -f $NPM_TAR_FILE
rm -f $RPM_TEST_FILE
rm -f $CONDA_TEST_FILE
