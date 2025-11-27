---
page_title: "arubacloud_securityrule Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud Security Rule.
---

# arubacloud_securityrule

Manages an ArubaCloud Security Rule.

## Example Usage

```terraform
resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  vpc_id            = "example-vpc-id"
  security_group_id = "example-security-group-id"
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_securityrule.example <securityrule_id>
```
