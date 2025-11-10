# droneHarPlugin

A Drone plugin for managing generic artifacts in Harness Artifact Registry (HAR) using the Harness CLI. Supports multiple operations including upload, download, info retrieval, and deletion.

## Overview

This plugin provides a comprehensive way to manage artifacts in Harness Artifact Registry as part of your Drone CI/CD pipeline. It uses the Harness CLI (`hc`) internally to handle authentication and supports multiple operations:

- **Push**: Upload artifacts to the registry
- **Pull**: Download artifacts from the registry
- **Get**: Retrieve artifact information
- **Delete**: Remove artifacts from the registry

## Usage

### Basic Upload Example

```yaml
kind: pipeline
type: docker
name: default

steps:
- name: upload-artifact
  image: harness/droneHarPlugin
  settings:
    command: push  # Optional - defaults to push
    registry: my-registry
    source: ./dist/my-app.zip
    name: my-application
    version: 1.0.0
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: my-org
    project: my-project
    pkg_url: https://pkg.qa.harness.io
```

### Multi-Command Pipeline Example

```yaml
kind: pipeline
type: docker
name: artifact-management

steps:
# Upload artifact
- name: upload-artifact
  image: harness/droneHarPlugin
  settings:
    command: push
    registry: my-registry
    source: ./build/artifact.tar.gz
    name: my-service
    version: ${DRONE_BUILD_NUMBER}
    description: "Build artifact for ${DRONE_COMMIT_SHA}"
    filename: "my-service-${DRONE_BUILD_NUMBER}.tar.gz"
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    org: my-org
    project: my-project
    pkg_url: https://pkg.qa.harness.io
    enable_proxy: true

# Get artifact info
- name: get-artifact-info
  image: harness/droneHarPlugin
  settings:
    command: get
    registry: my-registry
    name: my-service
    version: ${DRONE_BUILD_NUMBER}
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    pkg_url: https://pkg.qa.harness.io

# Download artifact (in a different pipeline/stage)
- name: download-artifact
  image: harness/droneHarPlugin
  settings:
    command: pull
    registry: my-registry
    name: my-service
    version: 1.0.0
    filename: my-service.tar.gz
    destination: ./downloaded/
    token:
      from_secret: harness_token
    account:
      from_secret: harness_account
    pkg_url: https://pkg.qa.harness.io
```

## Settings

### Common Required Settings

| Setting | Description | Example |
|---------|-------------|---------|
| `registry` | Name of the Harness Artifact Registry | `my-registry` |
| `token` | Harness authentication token | `pat.abc123...` |
| `account` | Harness account ID | `abc123def456` |
| `pkg_url` | Base URL for the Packages | `https://pkg.qa.harness.io` |

### Command-Specific Required Settings

#### Push Command
| Setting | Description | Example |
|---------|-------------|---------|
| `source` | Path to the artifact file to upload | `./dist/app.zip` |
| `name` | Name for the artifact in the registry | `my-application` |

#### Pull Command
| Setting | Description | Example |
|---------|-------------|---------|
| `name` | Name of the artifact to download | `my-application` |
| `version` | Version of the artifact to download | `1.0.0` |
| `filename` | Filename of the artifact to download | `app.zip` |
| `destination` | Local destination path for download | `./downloads/` |

#### Get Command
| Setting | Description | Example |
|---------|-------------|---------|
| `name` | Name of the artifact to get info for | `my-application` |

#### Delete Command
| Setting | Description | Example |
|---------|-------------|---------|
| `name` | Name of the artifact to delete | `my-application` |

### Optional Settings

| Setting | Description | Default | Example | Commands |
|---------|-------------|---------|---------|----------|
| `command` | Operation to perform | `push` | `pull`, `get`, `delete` | All |
| `version` | Version for the artifact | `1.0.0` | `${DRONE_BUILD_NUMBER}` | push, get, delete |
| `description` | Description of the artifact | _(empty)_ | `Build artifact` | push |
| `filename` | Custom filename for the uploaded artifact | _(basename of source)_ | `app-v1.0.0.zip` | push |
| `package_type` | Type of package | `generic` | `generic` | push |
| `target` | Target path for operations | _(empty)_ | `./target/` | pull |
| `org` | Harness organization ID | _(empty)_ | `my-org` | All |
| `project` | Harness project ID | _(empty)_ | `my-project` | All |
| `api_url` | Base URL for the Harness API | _(empty)_ | `https://app.harness.io` | All |
| `enable_proxy` | Enable proxy configuration | `false` | `true` | All |
| `retries` | Number of operation retries | `0` | `3` | All |
| `log_level` | Plugin log level | _(empty)_ | `debug` | All |

## Authentication

The plugin requires a Harness Personal Access Token (PAT) for authentication. You can create one in your Harness account settings.

**Security Note**: Always store your token as a Drone secret, never hardcode it in your pipeline configuration.

```yaml
# Store as a secret
settings:
  token:
    from_secret: harness_token
```

## Environment Variables

The plugin uses the following environment variables (automatically set by Drone):

### Common Variables
- `PLUGIN_COMMAND` - Command to execute (push, pull, get, delete)
- `PLUGIN_REGISTRY` - Registry name
- `PLUGIN_TOKEN` - Authentication token
- `PLUGIN_ACCOUNT` - Account ID
- `PLUGIN_PKG_URL` - Base URL for packages
- `PLUGIN_ORG` - Organization ID
- `PLUGIN_PROJECT` - Project ID
- `PLUGIN_API_URL` - API base URL
- `PLUGIN_ENABLE_PROXY` - Enable proxy
- `PLUGIN_RETRIES` - Number of retries
- `PLUGIN_LOG_LEVEL` - Log level

### Push Command Variables
- `PLUGIN_SOURCE` - Source file path
- `PLUGIN_NAME` - Artifact name
- `PLUGIN_VERSION` - Artifact version
- `PLUGIN_DESCRIPTION` - Artifact description
- `PLUGIN_FILENAME` - Custom filename
- `PLUGIN_PACKAGE_TYPE` - Package type

### Pull Command Variables
- `PLUGIN_NAME` - Artifact name
- `PLUGIN_VERSION` - Artifact version
- `PLUGIN_FILENAME` - Artifact filename
- `PLUGIN_DESTINATION` - Destination path

### Get/Delete Command Variables
- `PLUGIN_NAME` - Artifact name
- `PLUGIN_VERSION` - Artifact version

## Proxy Support

The plugin supports proxy configuration through environment variables:

- `HARNESS_HTTP_PROXY` - HTTP proxy URL
- `HARNESS_HTTPS_PROXY` - HTTPS proxy URL
- `HARNESS_NO_PROXY` - Comma-separated list of hosts to bypass proxy

Set `enable_proxy: true` in your pipeline settings to activate proxy support.

## Building the Plugin

```bash
# Build the binary
go build -o droneHarPlugin

# Build Docker image
docker build -t harness/droneHarPlugin .
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Commands

### Push (Upload)
Uploads an artifact file to the registry.

**Required**: `registry`, `source`, `name`, `token`, `account`, `pkg_url`

### Pull (Download)
Downloads an artifact from the registry.

**Required**: `registry`, `name`, `version`, `filename`, `destination`, `token`, `account`, `pkg_url`

### Get (Info)
Retrieves information about an artifact.

**Required**: `registry`, `name`, `token`, `account`, `pkg_url`

### Delete (Remove)
Deletes an artifact from the registry.

**Required**: `registry`, `name`, `token`, `account`, `pkg_url`

## Requirements

- Harness CLI (`hc`) must be available in the container
- Valid Harness authentication token
- Access to the target Harness Artifact Registry

## License

This project is licensed under the Blue Oak Model License.
