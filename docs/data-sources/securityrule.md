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

output "securityrule_name" {
  value = data.arubacloud_securityrule.example.name
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
output "securityrule_target_kind" {
  value = data.arubacloud_securityrule.example.target_kind
}
output "securityrule_target_value" {
  value = data.arubacloud_securityrule.example.target_value
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->

