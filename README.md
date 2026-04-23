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

### 2. How to start?

Have a look to [examples](examples/test/)


### 3. Apply the configuration

```bash
terraform init
terraform plan
terraform apply
```

## Debugging & Logging

The provider can emit full HTTP request/response traces to help diagnose API errors. Two filters must both be open before any trace output appears:

| Filter | How to set |
|---|---|
| Provider `log_level` | HCL: `log_level = "DEBUG"` or env: `ARUBACLOUD_LOG_LEVEL=DEBUG` |
| Terraform log pipeline | Env: `TF_LOG=DEBUG` (or narrower: `TF_LOG_PROVIDER_ARUBACLOUD_SDK=DEBUG`) |

```hcl
provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
  log_level  = "DEBUG"
}
```

```bash
# Both env vars and terraform must be on the same command line (or exported first).
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

Each outbound HTTP request (method, URL, headers, body) and response (status, headers, body) is logged. The `Authorization` header is auto-redacted as `Bearer [REDACTED]`. Other body content is not redacted — do not commit trace files to version control.

For subsystem-scoped output (SDK HTTP only, no other provider noise) and more detail, see the [Logging & Troubleshooting](docs/index.md#logging--troubleshooting) section in the provider docs.

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow: formatting, linting, doc regeneration, and the PR checklist.

**Quick summary:**
- `make fmt` — format (enforced by lint)
- `make lint` — lint (auto-installs golangci-lint v2 if missing)
- `make test` — unit tests
- `make generate` — regenerate docs after schema/template changes (requires Terraform in PATH; use WSL on Windows)
- `make ci-test` — full CI pipeline locally

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