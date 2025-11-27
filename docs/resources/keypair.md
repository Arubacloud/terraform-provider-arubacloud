
---
page_title: "arubacloud_keypair Resource - ArubaCloud"
subcategory: "Compute"
description: |-
  Manages an ArubaCloud KeyPair.
---

# arubacloud_keypair

Manages an ArubaCloud KeyPair.

```terraform
resource "arubacloud_keypair" "basic" {
  name            = "example-keypair"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_keypair.example <keypair_id>
```
