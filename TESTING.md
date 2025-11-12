# Testing Guide for drone-har

This guide explains how to test the drone-har plugin to ensure it works correctly before deployment.

## 🧪 Testing Levels

### 1. Basic Validation Testing (Safe - No API calls)
```bash
# Run the basic test script
./test-plugin.sh
```

This will:
- ✅ Build the plugin
- ✅ Run unit tests
- ✅ Test parameter validation
- ✅ Test environment variable parsing
- ⚠️ Show expected failure when trying to call Harness CLI (normal)

### 2. Real Integration Testing (Makes actual API calls)
```bash
# Only run this if you have valid Harness credentials
./test-with-real-credentials.sh
```

This will:
- ✅ Check if Harness CLI is installed
- ✅ Verify authentication
- ✅ Upload a real test artifact
- ✅ Confirm successful upload

## 📋 Pre-Testing Checklist

Before running tests, ensure you have:

- [ ] Go 1.22.7+ installed
- [ ] Plugin built successfully (`go build -o drone-har .`)
- [ ] Unit tests passing (`go test ./...`)

For real testing, additionally ensure:
- [ ] Harness CLI installed (`hc` command available)
- [ ] Valid Harness account with HAR access
- [ ] API token with artifact upload permissions
- [ ] Registry created in Harness UI

## 🔧 Manual Testing Steps

If you prefer manual testing:

### Step 1: Build and Test Locally
```bash
# Build the plugin
go build -o drone-har .

# Run unit tests
go test ./... -v

# Test with fake credentials (will fail at CLI call)
export PLUGIN_REGISTRY="test-registry"
export PLUGIN_SOURCE="test-sample.txt"
export PLUGIN_NAME="test-artifact"
export PLUGIN_VERSION="1.0.0"
export PLUGIN_TOKEN="fake-token"
export PLUGIN_ACCOUNT="fake-account"

./drone-har
```

### Step 2: Test with Docker
```bash
# Build Docker image
docker build -t drone-har-test .

# Test with environment variables
docker run --rm \
  -e PLUGIN_REGISTRY=your-registry \
  -e PLUGIN_SOURCE=/test-sample.txt \
  -e PLUGIN_NAME=test-artifact \
  -e PLUGIN_VERSION=1.0.0 \
  -e PLUGIN_TOKEN=your-token \
  -e PLUGIN_ACCOUNT=your-account \
  -v $(pwd)/test-sample.txt:/test-sample.txt \
  drone-har-test
```

## ✅ Success Indicators

### Basic Testing Success:
- Plugin builds without errors
- Unit tests pass
- Parameter validation works
- Environment variables are parsed correctly
- Command generation is correct (visible in logs)

### Integration Testing Success:
- Harness CLI executes successfully
- Artifact uploads to HAR
- No authentication errors
- Artifact visible in Harness UI

## ❌ Common Issues & Solutions

### Issue: "hc: command not found"
**Solution**: Install Harness CLI
```bash
curl -L https://github.com/harness/harness-cli/releases/latest/download/hc-linux-amd64 -o /usr/local/bin/hc
chmod +x /usr/local/bin/hc
```

### Issue: "authentication token must be set"
**Solution**: Ensure you have a valid Harness API token with HAR permissions

### Issue: "registry name must be set"
**Solution**: Verify the registry exists in your Harness account

### Issue: "failed to upload package"
**Solution**: Check:
- Registry permissions
- Token permissions
- Network connectivity
- File exists and is readable

## 🚀 Ready for Deployment?

Your plugin is ready for Git push and team usage when:

- ✅ `./test-plugin.sh` passes completely
- ✅ `./test-with-real-credentials.sh` uploads successfully
- ✅ Docker image builds and runs
- ✅ Documentation is complete
- ✅ Examples work as expected

## 📞 Getting Help

If tests fail:
1. Check the error messages in the logs
2. Verify your Harness credentials and permissions
3. Ensure the registry exists and is accessible
4. Test Harness CLI manually: `hc artifact push generic --help`
