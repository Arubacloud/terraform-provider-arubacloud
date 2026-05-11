# Terraform Provider for ArubaCloud

[![GitHub release](https://img.shields.io/github/tag/arubacloud/terraform-provider-arubacloud.svg?label=release)](https://github.com/arubacloud/terraform-provider-arubacloud/releases/latest)
[![Tests](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml)
[![Acceptance Tests](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/acceptance.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/acceptance.yml)
[![Release](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/Arubacloud/terraform-provider-arubacloud/graph/badge.svg)](https://codecov.io/gh/Arubacloud/terraform-provider-arubacloud)

Manage your [ArubaCloud](https://arubacloud.com/) infrastructure with Terraform — a European cloud platform offering virtual machines, managed Kubernetes, managed databases, private networking, block storage, and security services.

## Resources

| Category | Resources & Data Sources |
|---|---|
| **Compute** | `arubacloud_cloudserver`, `arubacloud_keypair`, `arubacloud_elasticip` |
| **Network** | `arubacloud_vpc`, `arubacloud_subnet`, `arubacloud_securitygroup`, `arubacloud_securityrule`, `arubacloud_vpcpeering`, `arubacloud_vpcpeeringroute`, `arubacloud_vpntunnel`, `arubacloud_vpnroute` |
| **Storage** | `arubacloud_blockstorage`, `arubacloud_snapshot`, `arubacloud_backup`, `arubacloud_restore` |
| **Container** | `arubacloud_kaas`, `arubacloud_containerregistry` |
| **Database** | `arubacloud_dbaas`, `arubacloud_database`, `arubacloud_databasegrant`, `arubacloud_databasebackup`, `arubacloud_dbaasuser` |
| **Management** | `arubacloud_project`, `arubacloud_schedulejob` |
| **Security** | `arubacloud_kms` |

## Documentation

- [Terraform Registry](https://registry.terraform.io/providers/Arubacloud/arubacloud/latest/docs) — full provider reference
- [examples/](examples/) — complete working examples for all resources
- [docs/](docs/) — generated reference for all resources and data sources

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.24 (only needed to build from source)

## Installation

Add the provider to your Terraform configuration and run `terraform init`:

```hcl
terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.1.3"
    }
  }
}

provider "arubacloud" {
  api_key    = var.arubacloud_api_key
  api_secret = var.arubacloud_api_secret
}
```

Credentials can also be supplied via environment variables:

```bash
export ARUBACLOUD_API_KEY="your-api-key"
export ARUBACLOUD_API_SECRET="your-api-secret"
terraform plan
```

## Quick Start

A full end-to-end example (project → keypair → VPC → subnet → security group → boot disk → CloudServer) is in [`examples/provider/quick-start.tf`](examples/provider/quick-start.tf).

```bash
terraform init
terraform plan
terraform apply
```

See [`examples/test/`](examples/test/) for more complete scenario-based examples.

## Debugging & Logging

The provider can emit full HTTP request/response traces to help diagnose API errors. Two independent filters must both be open before any trace output appears:

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
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

The `Authorization` header is auto-redacted as `Bearer [REDACTED]`. Other body content is not redacted — **do not commit trace files to version control**.

For more details see the [Logging & Troubleshooting](docs/index.md#logging--troubleshooting) section.

## Build the Provider

```bash
git clone https://github.com/arubacloud/terraform-provider-arubacloud.git
cd terraform-provider-arubacloud
make build        # build provider binary
make              # fmt → lint → test → build → generate
```

## Develop the Provider

- Install Go >= 1.24
- Run `make test` to run unit tests
- Run `make testacc` to run acceptance tests (requires `TF_ACC=1`, creates real resources)
- Run `make lint` to run the linter (auto-installs `golangci-lint` if needed)
- Run `make generate` to regenerate docs after schema changes
- Run `make testcov` to generate an HTML coverage report

### Local Development Override

To use a locally built binary instead of the registry version:

```bash
# 1. Build
make build

# 2. Create a dev override config (terraform.tfrc)
cat > terraform.tfrc <<'EOF'
provider_installation {
  dev_overrides {
    "arubacloud/arubacloud" = "$PWD"
  }
  direct {}
}
EOF

# 3. Point Terraform at it
export TF_CLI_CONFIG_FILE="$PWD/terraform.tfrc"
terraform plan   # will warn about dev overrides — that's expected
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow: formatting, linting, doc regeneration, and the PR checklist.
