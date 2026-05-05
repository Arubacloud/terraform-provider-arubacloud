---
page_title: "Provider: ArubaCloud"
description: |-
  The ArubaCloud Terraform provider manages compute, network, storage, database, container, and security resources on the ArubaCloud platform.
---

# ArubaCloud Provider

The ArubaCloud Terraform provider lets you manage infrastructure on [ArubaCloud](https://arubacloud.com) — a European cloud platform offering virtual machines, managed Kubernetes, managed databases, private networking, block storage, and security services.

Use the navigation on the left to browse all available resources and data sources.

## Example Usage

```terraform
terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = "YOUR_API_KEY"
  api_secret = "YOUR_API_SECRET"
}
```

## Quick Start

The following end-to-end example provisions a CloudServer with all required dependencies (project, SSH key pair, VPC, subnet, security group, and bootable volume):

```terraform
# Quick Start — end-to-end ArubaCloud CloudServer with networking
# Replace placeholder values with your own before applying.
# See https://api.arubacloud.com/docs/metadata for valid locations, flavors, and images.

terraform {
  required_providers {
    arubacloud = {
      source  = "arubacloud/arubacloud"
      version = ">= 0.0.1"
    }
  }
}

provider "arubacloud" {
  api_key    = var.api_key
  api_secret = var.api_secret
}

variable "api_key" {
  description = "ArubaCloud API key"
  type        = string
  sensitive   = true
}

variable "api_secret" {
  description = "ArubaCloud API secret"
  type        = string
  sensitive   = true
}

# 1. Project — the root organisational unit that owns all resources below.
resource "arubacloud_project" "quickstart" {
  name        = "quickstart-project"
  description = "Quick-start project managed by Terraform"
  tags        = ["quickstart", "terraform"]
}

# 2. SSH key pair — used for passwordless access to the CloudServer.
resource "arubacloud_keypair" "quickstart" {
  name       = "quickstart-keypair"
  location   = "it-mil1" # Milan region; any valid region string works here
  project_id = arubacloud_project.quickstart.id
  # Replace with your own public key.
  value = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC3... user@host"
  tags  = ["quickstart"]
}

# 3. VPC — isolated private network for all resources in this project.
resource "arubacloud_vpc" "quickstart" {
  name       = "quickstart-vpc"
  location   = "it-mil1"
  project_id = arubacloud_project.quickstart.id
  tags       = ["quickstart"]
}

# 4. Subnet (Advanced type) — carves an address space out of the VPC.
#    The Advanced type enables DHCP, which is required for CloudServer attachment.
resource "arubacloud_subnet" "quickstart" {
  name       = "quickstart-subnet"
  location   = "it-mil1"
  project_id = arubacloud_project.quickstart.id
  vpc_id     = arubacloud_vpc.quickstart.id
  type       = "Advanced" # Must be "Advanced" to support DHCP
  network = {
    address = "10.0.0.0/24" # CIDR for this subnet; adjust to avoid conflicts
    dhcp = {
      enabled = true # Required for Advanced type
      range = {
        start = "10.0.0.10"
        count = 200
      }
      routes = [
        {
          address = "0.0.0.0/0"   # Default route
          gateway = "10.0.0.1"    # First usable address in the CIDR
        }
      ]
      dns = ["8.8.8.8", "8.8.4.4"]
    }
  }
  tags = ["quickstart"]
}

# 5. Security group — a named container for firewall rules attached to the VPC.
resource "arubacloud_securitygroup" "quickstart" {
  name       = "quickstart-sg"
  location   = "it-mil1"
  project_id = arubacloud_project.quickstart.id
  vpc_id     = arubacloud_vpc.quickstart.id
  tags       = ["quickstart"]
}

# 6. Security rule — allows inbound SSH (TCP/22) from any source.
#    Add additional arubacloud_securityrule resources for other ports (e.g. 80, 443).
resource "arubacloud_securityrule" "ssh_inbound" {
  name              = "quickstart-allow-ssh"
  location          = "it-mil1"
  project_id        = arubacloud_project.quickstart.id
  vpc_id            = arubacloud_vpc.quickstart.id
  security_group_id = arubacloud_securitygroup.quickstart.id
  tags              = ["quickstart"]
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "22"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0" # Restrict to your IP range in production
    }
  }
}

# 7. Block storage (bootable) — the boot disk for the CloudServer.
#    Set bootable = true and supply an image ID. Find image IDs at
#    https://api.arubacloud.com/docs/metadata/#cloud-server-bootvolume
resource "arubacloud_blockstorage" "quickstart_boot" {
  name           = "quickstart-boot-disk"
  project_id     = arubacloud_project.quickstart.id
  location       = "it-mil1"
  zone           = "it-mil1-1" # Zone must match the CloudServer zone below
  size_gb        = 50
  billing_period = "Hour"
  type           = "Performance" # Performance type recommended for boot volumes
  bootable       = true
  image          = "LU22-001" # Ubuntu 22.04; see metadata API for full list
  tags           = ["quickstart", "boot"]
}

# 8. CloudServer — the virtual machine wired to all resources above.
#    flavor_name selects the vCPU/RAM profile; see
#    https://api.arubacloud.com/docs/metadata/#cloudserver-flavors for the full list.
resource "arubacloud_cloudserver" "quickstart" {
  name       = "quickstart-server"
  location   = "it-mil1"
  project_id = arubacloud_project.quickstart.id
  zone       = "it-mil1-1" # Must match the boot-volume zone above
  tags       = ["quickstart", "compute"]

  network = {
    vpc_uri_ref            = arubacloud_vpc.quickstart.uri
    subnet_uri_refs        = [arubacloud_subnet.quickstart.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.quickstart.uri]
  }

  settings = {
    flavor_name      = "CSO4A8" # 4 vCPU, 8 GB RAM; change to suit your workload
    key_pair_uri_ref = arubacloud_keypair.quickstart.uri
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.quickstart_boot.uri
  }
}

output "cloudserver_id" {
  description = "The ID of the provisioned CloudServer"
  value       = arubacloud_cloudserver.quickstart.id
}
```

## Resources by Category

| Category | Resources | Data Sources |
|---|---|---|
| **Compute** | [`arubacloud_cloudserver`](resources/cloudserver), [`arubacloud_keypair`](resources/keypair), [`arubacloud_elasticip`](resources/elasticip) | [`arubacloud_cloudserver`](data-sources/cloudserver), [`arubacloud_keypair`](data-sources/keypair), [`arubacloud_elasticip`](data-sources/elasticip) |
| **Network** | [`arubacloud_vpc`](resources/vpc), [`arubacloud_subnet`](resources/subnet), [`arubacloud_securitygroup`](resources/securitygroup), [`arubacloud_securityrule`](resources/securityrule), [`arubacloud_vpcpeering`](resources/vpcpeering), [`arubacloud_vpcpeeringroute`](resources/vpcpeeringroute), [`arubacloud_vpntunnel`](resources/vpntunnel), [`arubacloud_vpnroute`](resources/vpnroute) | [`arubacloud_vpc`](data-sources/vpc), [`arubacloud_subnet`](data-sources/subnet), [`arubacloud_securitygroup`](data-sources/securitygroup), [`arubacloud_securityrule`](data-sources/securityrule), [`arubacloud_vpcpeering`](data-sources/vpcpeering), [`arubacloud_vpcpeeringroute`](data-sources/vpcpeeringroute), [`arubacloud_vpntunnel`](data-sources/vpntunnel), [`arubacloud_vpnroute`](data-sources/vpnroute) |
| **Storage** | [`arubacloud_blockstorage`](resources/blockstorage), [`arubacloud_snapshot`](resources/snapshot), [`arubacloud_backup`](resources/backup), [`arubacloud_restore`](resources/restore) | [`arubacloud_blockstorage`](data-sources/blockstorage), [`arubacloud_snapshot`](data-sources/snapshot), [`arubacloud_backup`](data-sources/backup), [`arubacloud_restore`](data-sources/restore) |
| **Container** | [`arubacloud_kaas`](resources/kaas), [`arubacloud_containerregistry`](resources/containerregistry) | [`arubacloud_kaas`](data-sources/kaas), [`arubacloud_containerregistry`](data-sources/containerregistry) |
| **Database** | [`arubacloud_dbaas`](resources/dbaas), [`arubacloud_database`](resources/database), [`arubacloud_databasegrant`](resources/databasegrant), [`arubacloud_databasebackup`](resources/databasebackup), [`arubacloud_dbaasuser`](resources/dbaasuser) | [`arubacloud_dbaas`](data-sources/dbaas), [`arubacloud_database`](data-sources/database), [`arubacloud_databasegrant`](data-sources/databasegrant), [`arubacloud_databasebackup`](data-sources/databasebackup), [`arubacloud_dbaasuser`](data-sources/dbaasuser) |
| **Management** | [`arubacloud_project`](resources/project), [`arubacloud_schedulejob`](resources/schedulejob) | [`arubacloud_project`](data-sources/project), [`arubacloud_schedulejob`](data-sources/schedulejob) |
| **Security** | [`arubacloud_kms`](resources/kms) | [`arubacloud_kms`](data-sources/kms) |

## Argument Reference

The following arguments are supported:

- `api_key` - (Required, string) ArubaCloud API key. Can also be specified with the `ARUBACLOUD_API_KEY` environment variable.
- `api_secret` - (Required, string) ArubaCloud API secret. Can also be specified with the `ARUBACLOUD_API_SECRET` environment variable.
- `resource_timeout` - (Optional, string) Timeout for waiting for resources to become active after creation (e.g. `"5m"`, `"10m"`). Default: `"10m"`.
- `base_url` - (Optional, string) Override the ArubaCloud API base URL. Advanced use only.
- `token_issuer_url` - (Optional, string) Override the ArubaCloud token issuer URL. Advanced use only.
- `log_level` - (Optional, string) SDK log level for HTTP request/response tracing. Accepted values (case-insensitive): `OFF`, `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE`. Default: `OFF`. Can also be set via the `ARUBACLOUD_LOG_LEVEL` environment variable; the HCL attribute takes precedence.

## Logging & Troubleshooting

The provider exposes two independent log filters:

| Filter | Controls |
|---|---|
| `log_level` / `ARUBACLOUD_LOG_LEVEL` | What the SDK HTTP client forwards to Terraform logs |
| `TF_LOG` / `TF_LOG_PROVIDER` | What Terraform actually writes to stderr |

A message is visible only when **both** filters permit it. SDK messages are tagged with the `arubacloud-sdk` subsystem and can be targeted specifically with `TF_LOG_PROVIDER_ARUBACLOUD_SDK`.

### Enable full HTTP tracing

```hcl
provider "arubacloud" {
  api_key    = var.api_key
  api_secret = var.api_secret
  log_level  = "DEBUG"
}
```

```sh
TF_LOG=DEBUG terraform apply
# or target only SDK output:
TF_LOG_PROVIDER_ARUBACLOUD_SDK=DEBUG terraform apply
```

To route trace output to a file (keeps terminal output readable):

```sh
TF_LOG=DEBUG TF_LOG_PATH=./trace.log terraform plan
```

Filter the captured file to SDK HTTP lines only:

```sh
grep '@module=arubacloud.arubacloud-sdk' trace.log
```

This emits each outbound HTTP request (method, URL, headers, body) and response (status, headers, body) to stderr or the log file. The `Authorization` header is automatically redacted as `Bearer [REDACTED]`.

> **Warning**: request/response bodies may contain sensitive data. Do not commit debug logs to version control.

### Common pitfalls

- **`log_level = "DEBUG"` set but no output appears** — `log_level` is filter #1 (SDK → Terraform). Filter #2 (Terraform → stderr/file) is controlled by `TF_LOG`. Both must be set. Add `TF_LOG=DEBUG` (or `TF_LOG_PROVIDER_ARUBACLOUD_SDK=DEBUG`) to the same command.
- **Env vars set on one line, `terraform` run on the next** — In bash, the `VAR=value command` prefix form only applies to the command on that line. Setting them on a blank line is a no-op. Use a single-line invocation (`TF_LOG=DEBUG terraform plan`) or `export TF_LOG=DEBUG` first.
- **Want SDK silent but still see provider events** — Set `log_level = "OFF"` (the default) or omit it. Provider-level `tflog` calls (resource lifecycle events, wait/retry status) are unaffected by `log_level` and continue to respect `TF_LOG` alone.

### Suppress SDK output while keeping provider tracing

Set `log_level = "OFF"` (the default) or omit it entirely. The provider-level `tflog` calls (resource lifecycle events, wait/retry status) are unaffected by `log_level` and continue to respect `TF_LOG` alone.
