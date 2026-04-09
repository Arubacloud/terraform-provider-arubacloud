---
page_title: "arubacloud_vpn_tunnel"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPN tunnel.
---

# arubacloud_vpntunnel

Reads an existing ArubaCloud VPN tunnel.

```terraform
data "arubacloud_vpntunnel" "basic" {
  id         = "vpn-tunnel-id"
  project_id = "your-project-id"
}

output "vpntunnel_name" {
  value = data.arubacloud_vpntunnel.basic.name
}
output "vpntunnel_status" {
  value = data.arubacloud_vpntunnel.basic.status
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) VPN Tunnel identifier
- `project_id` (String) ID of the project this VPN Tunnel belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `name` (String) VPN Tunnel name
- `remote_peer` (String) Remote peer IP address
- `status` (String) VPN Tunnel status
