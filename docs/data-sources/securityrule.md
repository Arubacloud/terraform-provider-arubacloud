---
page_title: "arubacloud_securityrule"
subcategory: "Network"
description: |-
  Reads an existing ArubaCloud Security Rule.
---

# arubacloud_securityrule

Reads an existing ArubaCloud Security Rule.

```terraform
data "arubacloud_security_rule" "example" {
  id = "your-securityrule-id"
}

output "securityrule_name" {
  value = data.arubacloud_security_rule.example.name
}
output "securityrule_location" {
  value = data.arubacloud_security_rule.example.location
}
output "securityrule_project_id" {
  value = data.arubacloud_security_rule.example.project_id
}
output "securityrule_vpc_id" {
  value = data.arubacloud_security_rule.example.vpc_id
}
output "securityrule_security_group_id" {
  value = data.arubacloud_security_rule.example.security_group_id
}
output "securityrule_direction" {
  value = data.arubacloud_security_rule.example.direction
}
output "securityrule_protocol" {
  value = data.arubacloud_security_rule.example.protocol
}
output "securityrule_port" {
  value = data.arubacloud_security_rule.example.port
}
output "securityrule_target_kind" {
  value = data.arubacloud_security_rule.example.target_kind
}
output "securityrule_target_value" {
  value = data.arubacloud_security_rule.example.target_value
}
```

## Argument Reference

<!-- tfplugindocs injects arguments -->

## Attribute Reference

<!-- tfplugindocs injects attributes -->

