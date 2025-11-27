---
page_title: "arubacloud_securitygroup Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud SecurityGroup.
---

# arubacloud_securitygroup

Manages an ArubaCloud SecurityGroup.

```terraform
resource "arubacloud_security_group" "basic" {
  name = "basic-security-group"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_security_group.example <securitygroup_id>
```
