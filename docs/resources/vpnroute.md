---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpnroute"
sidebar_current: "docs-resource-vpnroute"
description: |-
  VPN Route defines routing for VPN tunnels.
---

# arubacloud_vpnroute

VPN Routes allow you to control network traffic for VPN tunnels.

## Usage example

```hcl
resource "arubacloud_vpnroute" "example" {
  name          = "example-vpn-route"
  location      = "ITBG-Bergamo"
  tags          = ["route", "vpn"]
  project_id    = arubacloud_project.example.id
  vpn_tunnel_id = arubacloud_vpntunnel.example.id
  properties = {
    cloud_subnet   = "10.0.0.0/24"
    on_prem_subnet = "192.168.1.0/24"
  }
}
```

## Argument reference

* `id` - (Computed)[string] The ID of the VPN Route.
* `name` - (Required)[string] The name of the VPN Route.
* `properties` - (Required)[map] VPN route properties.
* ...other arguments...

## Import

To import a VPN Route, define an empty resource in your plan:

```
resource "arubacloud_vpnroute" "example" {
}
```

Import using the VPN Route ID:

```
terraform import arubacloud_vpnroute.example <vpnroute_id>
```
