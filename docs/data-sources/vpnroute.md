---
page_title: "arubacloud_vpn_route"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPN route.
---

# arubacloud_vpn_route

Reads an existing ArubaCloud VPN route.

```terraform
data "arubacloud_vpnroute" "basic" {
  id             = "vpn-route-id"
  project_id     = "your-project-id"
  vpn_tunnel_id  = "your-vpn-tunnel-id"
}

output "vpnroute_name" {
  value = data.arubacloud_vpnroute.basic.name
}
output "vpnroute_destination" {
  value = data.arubacloud_vpnroute.basic.destination
}
output "vpnroute_gateway" {
  value = data.arubacloud_vpnroute.basic.gateway
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
