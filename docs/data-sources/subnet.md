---
page_title: "arubacloud_subnet"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud subnet.
---

# arubacloud_subnet

Reads an existing ArubaCloud subnet.

```terraform
data "arubacloud_subnet" "basic" {
  id         = "subnet-id"
  project_id = "your-project-id"
  vpc_id     = "your-vpc-id"
}

output "subnet_name" {
  value = data.arubacloud_subnet.basic.name
}
output "subnet_address" {
  value = data.arubacloud_subnet.basic.address
}
output "subnet_dhcp_enabled" {
  value = data.arubacloud_subnet.basic.dhcp_enabled
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) Subnet identifier
- `project_id` (String) ID of the project this subnet belongs to
- `vpc_id` (String) ID of the VPC this subnet belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `address` (String) Address of the network in CIDR notation
- `dhcp_enabled` (Boolean) Whether DHCP is enabled
- `dhcp_range_count` (Number) Number of available IP addresses in DHCP range
- `dhcp_range_start` (String) Starting IP address for DHCP range
- `dhcp_routes` (List of Object) DHCP routes configuration (address, gateway)
- `dns` (List of String) List of DNS IP addresses
- `location` (String) Subnet location
- `name` (String) Subnet name
- `tags` (List of String) List of tags for the subnet
- `type` (String) Subnet type (Basic or Advanced)
- `uri` (String) Subnet URI
