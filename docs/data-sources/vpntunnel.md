---
page_title: "arubacloud_vpn_tunnel"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud VPN tunnel.
---

# arubacloud_vpn_tunnel

Reads an existing ArubaCloud VPN tunnel.

```terraform
data "arubacloud_vpntunnel" "basic" {
  id         = "vpn-tunnel-id"
  project_id = "your-project-id"
}

output "vpntunnel_name" {
  value = data.arubacloud_vpntunnel.basic.name
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->
