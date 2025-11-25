---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpntunnel"
sidebar_current: "docs-resource-vpntunnel"
description: |-
  VPN Tunnel provides secure connectivity between cloud and on-premises networks.
---

# arubacloud_vpntunnel

VPN Tunnels allow secure communication between your ArubaCloud VPC and external networks.

## Usage example

```hcl
resource "arubacloud_vpntunnel" "example" {
  name       = "example-vpn-tunnel"
  location   = "ITBG-Bergamo"
  tags       = ["vpn", "prod"]
  project_id = arubacloud_project.example.id
  properties = {
    vpn_type            = "Site-To-Site"
    vpn_client_protocol = "ikev2"
    ip_configurations = {
      vpc = {
        id = arubacloud_vpc.example.id
      }
      subnet = {
        id = arubacloud_subnet.example.id
      }
      public_ip = {
        id = arubacloud_elasticip.example.id
      }
    }
    vpn_client_settings = {
      ike = {
        lifetime     = 3600
        encryption   = "AES256"
        hash         = "SHA256"
        dh_group     = "group14"
        dpd_action   = "restart"
        dpd_interval = 30
        dpd_timeout  = 120
      }
      esp = {
        lifetime   = 3600
        encryption = "AES256"
        hash       = "SHA256"
        pfs        = "group14"
      }
      psk = {
        cloud_site   = "cloud-site-1"
        on_prem_site = "on-prem-site-1"
        secret       = "supersecretkey"
      }
    }
    peer_client_public_ip = "203.0.113.1"
  }
}
```

## Argument reference

* `id` - (Computed)[string] The ID of the VPN Tunnel.
* `name` - (Required)[string] The name of the VPN Tunnel.
* `properties` - (Required)[map] VPN tunnel properties.
* ...other arguments...

## Import

To import a VPN Tunnel, define an empty resource in your plan:

```
resource "arubacloud_vpntunnel" "example" {
}
```

Import using the VPN Tunnel ID:

```
terraform import arubacloud_vpntunnel.example <vpntunnel_id>
```
