# Quick Start (OpenTofu) — end-to-end ArubaCloud CloudServer with networking
# Replace placeholder values with your own before applying.
# Run: tofu init && tofu plan && tofu apply

terraform {
  required_providers {
    arubacloud = {
      source  = "registry.opentofu.org/Arubacloud/arubacloud"
      version = ">= 0.3.0"
    }
  }
}

provider "arubacloud" {
  client_id     = var.client_id
  client_secret = var.client_secret
}

variable "client_id" {
  description = "ArubaCloud OAuth2 client ID"
  type        = string
  sensitive   = true
}

variable "client_secret" {
  description = "ArubaCloud OAuth2 client secret"
  type        = string
  sensitive   = true
}

resource "arubacloud_project" "quickstart" {
  name        = "quickstart-project"
  description = "Quick-start project managed by OpenTofu"
  tags        = ["quickstart", "opentofu"]
}

resource "arubacloud_keypair" "quickstart" {
  name       = "quickstart-keypair"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.quickstart.id
  value      = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC3... user@host"
  tags       = ["quickstart"]
}

resource "arubacloud_vpc" "quickstart" {
  name       = "quickstart-vpc"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.quickstart.id
  tags       = ["quickstart"]
}

resource "arubacloud_subnet" "quickstart" {
  name       = "quickstart-subnet"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.quickstart.id
  vpc_id     = arubacloud_vpc.quickstart.id
  type       = "Advanced"
  network = {
    address = "10.0.0.0/24"
    dhcp = {
      enabled = true
      range = {
        start = "10.0.0.10"
        count = 200
      }
      routes = [
        {
          address = "0.0.0.0/0"
          gateway = "10.0.0.1"
        }
      ]
      dns = ["8.8.8.8", "8.8.4.4"]
    }
  }
  tags = ["quickstart"]
}

resource "arubacloud_securitygroup" "quickstart" {
  name       = "quickstart-sg"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.quickstart.id
  vpc_id     = arubacloud_vpc.quickstart.id
  tags       = ["quickstart"]
}

resource "arubacloud_securityrule" "ssh_inbound" {
  name              = "quickstart-allow-ssh"
  location          = "ITBG-Bergamo"
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
      value = "0.0.0.0/0"
    }
  }
}

resource "arubacloud_blockstorage" "quickstart_boot" {
  name           = "quickstart-boot-disk"
  project_id     = arubacloud_project.quickstart.id
  location       = "ITBG-Bergamo"
  zone           = "ITBG-1"
  size_gb        = 50
  billing_period = "Hour"
  type           = "Performance"
  bootable       = true
  image          = "LU22-001"
  tags           = ["quickstart", "boot"]
}

resource "arubacloud_cloudserver" "quickstart" {
  name       = "quickstart-server"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.quickstart.id
  zone       = "ITBG-1"
  tags       = ["quickstart", "compute"]

  network = {
    vpc_uri_ref            = arubacloud_vpc.quickstart.uri
    subnet_uri_refs        = [arubacloud_subnet.quickstart.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.quickstart.uri]
  }

  settings = {
    flavor_name      = "CSO4A8"
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
