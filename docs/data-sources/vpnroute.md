---
page_title: "arubacloud_vpn_route"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPN route.
---

# arubacloud_vpnroute

Reads an existing ArubaCloud VPN route.

```terraform
data "arubacloud_vpnroute" "basic" {
  id            = "vpn-route-id"
  project_id    = "your-project-id"
  vpn_tunnel_id = "your-vpn-tunnel-id"
}

output "vpnroute_destination" {
  value = data.arubacloud_vpnroute.basic.destination
}
output "vpnroute_gateway" {
  value = data.arubacloud_vpnroute.basic.gateway
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) VPN Route identifier
- `project_id` (String) ID of the project this VPN Route belongs to
- `vpn_tunnel_id` (String) ID of the VPN Tunnel this route belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `destination` (String) Cloud subnet destination (CIDR)
- `gateway` (String) On-premises subnet gateway (CIDR)
- `name` (String) VPN Route name
