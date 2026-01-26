# Release Notes - v0.0.1

**Release Date**: January 26, 2026

## 🎉 Initial Release

We're excited to announce the first release of the official ArubaCloud Terraform Provider! This provider enables infrastructure as code for the complete ArubaCloud platform.

## ✨ What's Included

### Supported Resources (25)

**Compute & Networking**
- `arubacloud_project` - Project management
- `arubacloud_cloudserver` - Virtual machines
- `arubacloud_keypair` - SSH key pairs
- `arubacloud_elasticip` - Elastic IP addresses
- `arubacloud_vpc` - Virtual Private Clouds
- `arubacloud_subnet` - VPC subnets
- `arubacloud_securitygroup` - Security groups
- `arubacloud_securityrule` - Security group rules
- `arubacloud_vpcpeering` - VPC peering connections
- `arubacloud_vpcpeeringroute` - VPC peering routes
- `arubacloud_vpntunnel` - VPN tunnels
- `arubacloud_vpnroute` - VPN routes

**Storage**
- `arubacloud_blockstorage` - Block storage volumes
- `arubacloud_snapshot` - Volume snapshots
- `arubacloud_backup` - Volume backups
- `arubacloud_restore` - Volume restores

**Containers & Kubernetes**
- `arubacloud_kaas` - Kubernetes as a Service
- `arubacloud_containerregistry` - Container Registry

**Database**
- `arubacloud_dbaas` - DBaaS instances (MySQL/PostgreSQL)
- `arubacloud_database` - Database within DBaaS
- `arubacloud_dbaasuser` - DBaaS users
- `arubacloud_databasegrant` - Database grants
- `arubacloud_databasebackup` - Database backups

**Automation & Security**
- `arubacloud_schedulejob` - Scheduled jobs
- `arubacloud_kms` - Key Management Service

### Data Sources (25)

All resources have corresponding data sources for importing and referencing existing infrastructure.

## 🔧 Key Features

- **Complete API Coverage**: Support for Compute, Storage, Networking, Kubernetes, DBaaS, and Security services
- **Smart State Management**: Automatic waiting for resources to reach active state after creation
- **Import Support**: Import existing infrastructure into Terraform state
- **Comprehensive Examples**: Working examples for all resources in the `examples/` directory
- **Full Documentation**: Auto-generated documentation with schemas and examples
- **Official SDK**: Built on the official ArubaCloud Go SDK

## 📚 Documentation

- **Provider Configuration**: See [docs/index.md](docs/index.md)
- **Resources**: See [docs/resources/](docs/resources/)
- **Data Sources**: See [docs/data-sources/](docs/data-sources/)
- **Examples**: See [examples/](examples/)

## 🚀 Getting Started

### Installation

Add to your Terraform configuration:

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

### Quick Example

```hcl
# Create a project
resource "arubacloud_project" "example" {
  name        = "my-project"
  description = "Managed by Terraform"
}

# Create a Cloud Server
resource "arubacloud_cloudserver" "web" {
  name           = "web-server"
  project_id     = arubacloud_project.example.id
  flavor         = "small"
  image          = "ubuntu-22.04"
  admin_password = "SecureP@ssw0rd!"
}
```

## ⚠️ Known Limitations

- **Key and KMIP Resources**: Temporarily disabled pending SDK updates for proper field mapping
- **Test Coverage**: While all resources are functionally tested with real infrastructure, automated acceptance tests require infrastructure setup and are ongoing

## 🐛 Reporting Issues

Please report any issues or feature requests on our [GitHub Issues](https://github.com/arubacloud/terraform-provider-arubacloud/issues) page.

## 📝 Changelog

See [CHANGELOG.md](CHANGELOG.md) for detailed changes.

## 🙏 Acknowledgments

Thank you to all contributors and early testers who helped make this release possible!

---

**Note**: This is an initial release (v0.0.1). While all resources have been tested with real infrastructure, we welcome feedback and bug reports as the provider matures. Future releases will include enhanced test coverage and additional features based on community feedback.
