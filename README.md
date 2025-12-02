# Terraform Provider for ArubaCloud

[![GitHub release](https://img.shields.io/github/tag/arubacloud/terraform-provider-arubacloud.svg?label=release)](https://github.com/arubacloud/terraform-provider-arubacloud/releases/latest)
[![Tests](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/test.yml)
[![Release](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml/badge.svg)](https://github.com/arubacloud/terraform-provider-arubacloud/actions/workflows/release.yml)

> **⚠️ Development Status**: This provider is currently under active development and is **not production-ready yet**. APIs and resource schemas may change. Use at your own risk in production environments.

- Documentation: See generated docs in the `docs/` folder or [Terraform Registry](https://registry.terraform.io/providers/arubacloud/arubacloud/latest/docs) (when published)

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.24.x (to build the provider plugin)

## Using the provider

To use the provider, configure it with your ArubaCloud credentials:

```hcl
provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
```
You can also use environment variables:

- **ARUBACLOUD_API_KEY**
- **ARUBACLOUD_API_SECRET**

## Example Usage

```hcl
resource "arubacloud_keypair" "example" {
  name     = "example-keypair"
  location = "ITBG-Bergamo"
  tags     = ["keypair", "test"]
  value    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
}

resource "arubacloud_cloudserver" "example" {
  name            = "example-cloudserver"
  location        = "ITBG-Bergamo"
  project_id      = arubacloud_project.example.id
  zone            = "ITBG-1"
  vpc_id          = arubacloud_vpc.example.id
  flavor_name     = "c2.medium"
  elastic_ip_id   = arubacloud_elasticip.example.id
  boot_volume     = arubacloud_blockstorage.example.id
  key_pair_id     = arubacloud_keypair.example.id
  subnets         = [arubacloud_subnet.example.id, arubacloud_subnet2.example.id]
  securitygroups  = [arubacloud_securitygroup.example.id, arubacloud_securitygroup2.example.id]
}
```

## Build the provider

Clone repository to your workspace:

```bash
git clone https://github.com/arubacloud/terraform-provider-arubacloud.git
cd terraform-provider-arubacloud
```

Build the provider
```bash
make build
```

## Develop the provider

- Install Go >= 1.24
- Run `make build` to build the provider binary
- Run `make test` to run unit tests
- Run `make testacc` to run acceptance tests (may create real resources)

### Testing

The provider includes comprehensive unit and acceptance tests.

**Run unit tests** (fast, no external dependencies):
```bash
make test
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