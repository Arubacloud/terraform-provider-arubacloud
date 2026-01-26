# Terraform Provider for ArubaCloud

[![GitHub release](https://img.shields.io/github/tag/arubacloud/terraform-provider-arubacloud.svg?label=release)](https://github.com/arubacloud/terraform-provider-arubacloud/releases/latest)
[![Tests](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml)
[![Release](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml)
[![Coverage](https://img.shields.io/badge/coverage-9.9%25-green)](TEST_RESULTS.md)

Manage your ArubaCloud infrastructure with Terraform! This provider enables you to create and manage Cloud Servers, VPCs, Kubernetes clusters, DBaaS instances, and more using infrastructure as code.

## Features

- **Compute**: Cloud Servers, SSH keypairs, Elastic IPs
- **Storage**: Block Storage, Snapshots, Backups
- **Networking**: VPCs, Subnets, Security Groups, VPN, VPC Peering
- **Kubernetes**: KaaS (Kubernetes as a Service)
- **Database**: DBaaS with MySQL/PostgreSQL support
- **Container**: Container Registry
- **Security**: Key Management Service (KMS)

> **Note**: This is the initial v0.0.1 release. While all resources are functional and tested with real infrastructure, comprehensive automated testing is ongoing. Please report any issues on GitHub.

## Documentation

- [Terraform Registry Documentation](https://registry.terraform.io/providers/arubacloud/arubacloud/latest/docs) (after publication)
- [Examples](examples/) - Complete working examples for all resources
- [Generated Docs](docs/) - Detailed documentation for all resources and data sources

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.24.x (only needed to build the provider from source)

## Installation

### Via Terraform Registry (Recommended)

Once published, the provider will be automatically downloaded when you run `terraform init`:

```hcl
terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = "~> 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
}
```

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/arubacloud/terraform-provider-arubacloud/releases)
2. Extract and place in your Terraform plugins directory:
   - Linux: `~/.terraform.d/plugins/registry.terraform.io/arubacloud/arubacloud/0.0.1/linux_amd64/`
   - macOS: `~/.terraform.d/plugins/registry.terraform.io/arubacloud/arubacloud/0.0.1/darwin_amd64/`
   - Windows: `%APPDATA%\terraform.d\plugins\registry.terraform.io\arubacloud\arubacloud\0.0.1\windows_amd64\`

## Quick Start

## Quick Start

### 1. Configure the provider

To use the provider, configure it with your ArubaCloud credentials:

```hcl
terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = "~> 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
```

You can also use environment variables:

- **ARUBACLOUD_API_KEY** - Your ArubaCloud API key
- **ARUBACLOUD_API_SECRET** - Your ArubaCloud API secret

```bash
export ARUBACLOUD_API_KEY="your-api-key"
export ARUBACLOUD_API_SECRET="your-api-secret"
terraform plan
```

### 2. Create your first resource

Example: Create a project and a Cloud Server

```hcl
# Create a project
resource "arubacloud_project" "example" {
  name        = "my-terraform-project"
  description = "Project managed by Terraform"
}

# Create a VPC
resource "arubacloud_vpc" "example" {
  name       = "my-vpc"
  project_id = arubacloud_project.example.id
  cidr       = "10.0.0.0/16"
}

# Create a subnet
resource "arubacloud_subnet" "example" {
  name       = "my-subnet"
  vpc_id     = arubacloud_vpc.example.id
  cidr       = "10.0.1.0/24"
}

# Create a Cloud Server
resource "arubacloud_cloudserver" "example" {
  name                = "my-server"
  project_id          = arubacloud_project.example.id
  flavor              = "small"
  image               = "ubuntu-22.04"
  vpc_uri             = arubacloud_vpc.example.uri
  subnet_uri          = arubacloud_subnet.example.uri
  admin_password      = "SecureP@ssw0rd!"
  enable_monitoring   = true
}
```

### 3. Apply the configuration

```bash
terraform init
terraform plan
terraform apply
```

For more examples, see the [examples/](examples/) directory.


## Build the provider

Clone repository to your workspace:

```bash
git clone https://github.com/arubacloud/terraform-provider-arubacloud.git
cd terraform-provider-arubacloud
```

Build the provider:
```bash
make build
```

Or run all checks (format, lint, test, build, generate):
```bash
make
```

## Develop the provider

- Install Go >= 1.24
- Run `make build` to build the provider binary
- Run `make test` to run unit tests
- Run `make testacc` to run acceptance tests (may create real resources)
- Run `make ci-test` to run all CI checks locally (build, lint, generate, test)

### Testing

The provider includes comprehensive unit and acceptance tests.

**Run all CI checks locally** (build, lint, generate, test):
```bash
make ci-test
```
This runs the same checks as the CI pipeline:
- Downloads dependencies
- Builds the provider
- Runs linter (auto-installs `golangci-lint` if needed)
- Generates documentation
- Runs `go mod tidy`
- Checks for uncommitted changes
- Runs unit tests
- Generates coverage report

**Run unit tests** (fast, no external dependencies):
```bash
make test
```

**Run linter** (auto-installs `golangci-lint` if needed):
```bash
make lint
```

**Run acceptance tests** (requires `TF_ACC=1`, may create real resources):
```bash
make testacc
```

**Run a specific test**:
```bash
make testacc-run TEST=TestAccBackupResource
```

**Generate coverage report**:
```bash
make testcov
# Opens coverage.html in your browser
```

**Format code**:
```bash
make fmt
```

**Generate documentation**:
```bash
make generate
```

For more details, see the [Testing Guide](docs/TESTING.md).

### Local Development

To use a locally built provider, set up a Terraform CLI config file:
> export TF_CLI_CONFIG_FILE="terraform.tfrc"

In the config file, override the lookup path for the provider:

```bash
provider_installation {
  dev_overrides {
    "arubacloud/arubacloud" = "$PWD"
  }
  direct {}
}
```

Build the provider:
> make build

Run Terraform:
> terraform plan

You should see a warning about provider development overrides.