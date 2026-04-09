---
page_title: "arubacloud_securityrule"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud Security Rule.
---

# arubacloud_securityrule

Reads an existing ArubaCloud Security Rule.

```terraform
data "arubacloud_securityrule" "example" {
  id                = "your-securityrule-id"
  project_id        = "your-project-id"
  vpc_id            = "your-vpc-id"
  security_group_id = "your-securitygroup-id"
}

output "securityrule_direction" {
  value = data.arubacloud_securityrule.example.direction
}
output "securityrule_protocol" {
  value = data.arubacloud_securityrule.example.protocol
}
output "securityrule_port" {
  value = data.arubacloud_securityrule.example.port
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) Security Rule identifier
- `project_id` (String) ID of the project this Security Rule belongs to
- `vpc_id` (String) ID of the VPC this Security Rule belongs to
- `security_group_id` (String) ID of the Security Group this rule belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `direction` (String) Rule direction (Inbound or Outbound)
- `location` (String) Security Rule location
- `name` (String) Security Rule name
- `port` (String) Port or port range
- `protocol` (String) Protocol (TCP, UDP, ICMP, ANY)
- `tags` (List of String) List of tags for the Security Rule
- `target_kind` (String) Target kind (Any, IP, SecurityGroup)
- `target_value` (String) Target value (IP address or Security Group ID)
- `uri` (String) Security Rule URI
