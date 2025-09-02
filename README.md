# Deployer v0.1 - Repsoft Limited

A professional, interactive command-line deployment tool that automates Docker container deployments to remote servers with colorful interface and real-time progress tracking.

## Table of Contents

- [Overview](#overview)
- [What Does Deployer Do?](#what-does-deployer-do)
- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Adding New Services](#adding-new-services)
- [Command Line Options](#command-line-options)
- [Troubleshooting](#troubleshooting)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

**Deployer** streamlines the deployment process for containerized applications with an intuitive, interactive workflow that provides real-time feedback through a colorful command-line interface.

### What Does Deployer Do?

Deployer automates the complete deployment pipeline:

1. **Build** Docker images from your local code (optional)
2. **Tag & Push** images to your private registry  
3. **Connect** to remote servers via SSH
4. **Pull** images on the target server
5. **Stop/Remove** old containers safely
6. **Deploy** new containers with your specifications
7. **Verify** deployment by checking container mounts and status

## Features

- **Progress Tracking** - Visual progress bar with real-time percentage (0-100%) 
- **Interactive Mode** - User-friendly service selection with numbered options
- **Docker Integration** - Complete Docker workflow automation
- **SSH Support** - Password and SSH key-based authentication
- **Multi-Service Management** - Deploy multiple applications from single configuration
- **Container Verification** - Mount inspection and status validation
- **Dry Run Mode** - Test deployments without making actual changes
- **Smart Building** - Skip builds for registry-only deployments

## Quick Start

### Prerequisites

- Go 1.19+ installed
- Docker installed locally
- SSH access to target server
- Private Docker registry access

### Installation

1. **Clone and build:**
```bash
git clone <repository-url>
cd deployer
go build -o deployer.exe ./cmd/deployer
```

2. **Configure services** - Create or edit `config.json`:

```json
{
  "registry": {
    "host": "prod-registry.cefero.com",
    "username": "registryadmin", 
    "password": "your-registry-password"
  },
  "ssh": {
    "host": "10.10.10.41",
    "username": "prod-apps",
    "port": 22,
    "password": "your-ssh-password"
  },
  "services": {
    "microsrv": {
      "service_name": "microsrv",
      "image_name": "microsrv",
      "build_path": "./microsrv",
      "container_name": "microsrv",
      "docker_run_args": "--volumes-from docportal-vols --env-file /etc/cefero/microsrv.env --restart unless-stopped",
      "health_timeout": 90
    }
  }
}
```

3. **Run the deployer:**
```bash
# Interactive mode (recommended)
./deployer.exe

# Command line mode
./deployer.exe -service microsrv -version 0.86
```

## Usage

### Interactive Mode (Recommended)

Run the deployer and follow the colored prompts:

```bash
./deployer.exe
```

**Sample Interactive Session:**
```
Deployer v0.1 - Repsoft Limited
===============================

Available Services:
  [1] cefero.docportal.api
  [2] microsrv
  [3] doctrllm

Select service (enter number): 2
Selected: microsrv
Enter version (e.g., 1.0.0): 0.86
Current build path: ./microsrv
Override build path? (Enter for default, or specify new path): 
Dry run mode? (y/n): n

Starting deployment of microsrv:0.86
Build path: ./microsrv
===============================

[████████████████████████████████████████████░░░░░] 85%
[INFO] [9/10] Running new container
[SUCCESS] [9/10] COMPLETED: Running new container
```

### Command Line Mode

Direct deployment with command-line arguments:

```bash
# Deploy specific service
./deployer.exe -service microsrv -version 0.86

# Dry run to preview
./deployer.exe -service doctrllm -version 0.80 -dry-run

# Custom build path
./deployer.exe -service myapp -version 1.0 -build-path "C:\Custom Path"

# List available services
./deployer.exe -list
```

## Configuration

The `config.json` file contains all deployment settings.

### Configuration Structure

| Section | Field | Description | Required |
|---------|-------|-------------|-----------|
| **Registry** | `host` | Docker registry hostname | Yes |
| | `username` | Registry username | Yes |
| | `password` | Registry password | Yes |
| **SSH** | `host` | Target server IP/hostname | Yes |
| | `username` | SSH username | Yes |
| | `port` | SSH port (default: 22) | Yes |
| | `password` | SSH password | No* |
| | `key_file` | Path to SSH private key | No* |
| **Services** | `service_name` | Unique service identifier | Yes |
| | `image_name` | Docker image name | Yes |
| | `build_path` | Build context path (empty = skip build) | No |
| | `container_name` | Container name on target server | Yes |
| | `docker_run_args` | Docker run arguments | No |
| | `health_timeout` | Health check timeout (unused) | No |

*Either `password` or `key_file` must be provided for SSH authentication.

## Adding New Services

To deploy a new service, add it to the `services` section in `config.json`:

### For Applications That Need Building:

```json
{
  "my-api": {
    "service_name": "my-api",
    "image_name": "my-api",
    "build_path": "./api-source",
    "container_name": "my-api-prod",
    "docker_run_args": "-p 8080:80 --restart unless-stopped",
    "health_timeout": 30
  }
}
```

### For Pre-built Images (Registry Only):

```json
{
  "doctrllm": {
    "service_name": "doctrllm",
    "image_name": "doctrllm",
    "build_path": "",
    "container_name": "doctrllm",
    "docker_run_args": "-p 8000:8000",
    "health_timeout": 30
  }
}
```

### Important Guidelines:

- **Build Path**: Set to directory containing Dockerfile, or empty "" to skip building
- **Paths with Spaces**: Use quotes: `"C:\\Path With Spaces"`
- **Docker Args**: Combine multiple: `"-p 8080:80 --restart unless-stopped --env-file /etc/app/.env"`
- **Port Mapping**: Use `-p host:container` format
- **Environment**: Use `--env-file /path/to/.env` or `-e VAR=value`

## Visual Features

Deployer provides professional colored output:

- **Green**: Success messages, completed steps, progress bar fill
- **Blue**: Info messages, in-progress status
- **Yellow**: Warnings, custom paths, build path overrides  
- **Red**: Errors, failures
- **Cyan**: Section headers, service names
- **Bold**: Emphasis, final success messages

### Progress Bar

Real-time visual progress tracking:
```
[██████████████████████████████████████░░░░░░░░░░] 78%
[INFO] [8/10] Running new container
```

## Command Line Options

| Flag | Description | Example |
|------|-------------|---------|
| `-service` | Service name to deploy | `-service microsrv` |
| `-version` | Version tag for image | `-version 1.2.3` |
| `-build-path` | Override build path | `-build-path ./custom/path` |
| `-dry-run` | Preview without executing | `-dry-run` |
| `-config` | Configuration file path | `-config prod-config.json` |
| `-list` | List available services | `-list` |

### Usage Examples:
```bash
# Interactive mode
./deployer.exe

# Deploy specific service  
./deployer.exe -service doctrllm -version 0.80

# Test deployment
./deployer.exe -service microsrv -version 0.86 -dry-run

# Custom build path
./deployer.exe -service myapp -version 1.0 -build-path "C:\Custom Path"

# List services
./deployer.exe -list
```

## Deployment Pipeline

Deployer executes these steps with visual progress tracking:

1. **Building Docker image** (if build_path specified)
2. **Tagging image for registry**
3. **Logging into registry**
4. **Pushing image to registry**
5. **Connecting to remote server**
6. **Pulling image on remote**
7. **Stopping existing container**
8. **Removing existing container**
9. **Running new container**
10. **Verifying container mounts**

Each step shows colored status messages and updates the progress bar.

## Examples

### Example 1: .NET API Service
```json
{
  "api-service": {
    "service_name": "api-service",
    "image_name": "my-dotnet-api",
    "build_path": "./src/MyApi",
    "container_name": "api-prod",
    "docker_run_args": "-p 7067:80 --env-file /etc/api/.env --restart unless-stopped",
    "health_timeout": 45
  }
}
```

### Example 2: Python Microservice (Pre-built)
```json
{
  "python-ml": {
    "service_name": "python-ml",
    "image_name": "ml-service",
    "build_path": "",
    "container_name": "ml-service",
    "docker_run_args": "-p 8000:8000 -v /data:/app/data --restart unless-stopped",
    "health_timeout": 60
  }
}
```

### Example 3: Node.js Application
```json
{
  "node-app": {
    "service_name": "node-app",
    "image_name": "my-node-app",
    "build_path": "./client",
    "container_name": "node-app-prod",
    "docker_run_args": "-p 3000:3000 --link postgres-db --env-file /etc/node/.env",
    "health_timeout": 30
  }
}
```

## Troubleshooting

### Common Issues

**SSH Connection Failed:**
- Verify SSH credentials in `config.json`
- Test manual connection: `ssh username@hostname`
- Check firewall/network access to target server
- Ensure SSH service is running on target server

**Docker Build Failed:**
- Verify Dockerfile exists in specified `build_path`
- Check Docker daemon is running locally
- Ensure build context contains all required files
- Review build logs for specific errors

**Registry Authentication:**
- Verify registry credentials are correct
- Test manual login: `docker login your-registry.com`
- Check network connectivity to registry
- Ensure registry URL format is correct

**Path Issues (Windows):**
- Use quotes for paths with spaces: `"C:\\Program Files\\My App"`
- Use double backslashes in JSON: `"C:\\\\path\\\\to\\\\app"`
- Avoid special characters in paths

**Container Mount Issues:**
- Verify volume paths exist on target server
- Check permissions on mount directories
- Review `docker_run_args` for proper syntax

### Debug Mode

Use dry run to preview all commands:
```bash
./deployer.exe -service myapp -version 1.0 -dry-run
```

### Log Analysis

Review colored output:
- **Red errors** indicate failure points
- **Yellow warnings** suggest configuration issues
- **Green success** messages confirm completion

## Best Practices

### Service Selection

**Choose the right deployment option:**

- **Local Build + Deploy**: Services you develop locally
  - Set `build_path` to your source directory
  - Use for .NET APIs, Node.js apps, custom services

- **Registry Only**: Pre-built services, third-party images
  - Set `build_path` to `""` (empty)
  - Use for Python ML services, databases, tools

### Configuration Management

- Keep separate configs for different environments
- Use descriptive service names
- Document custom `docker_run_args`
- Test with dry-run before production deployment

### Security

- Store credentials securely
- Use SSH keys when possible
- Restrict SSH user permissions
- Regularly rotate passwords

---

## License

Copyright 2024 Repsoft Limited. All rights reserved.

**Important:** Always test deployments in staging before production. Use dry-run mode to verify configurations.
