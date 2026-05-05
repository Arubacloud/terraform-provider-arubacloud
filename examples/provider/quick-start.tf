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
