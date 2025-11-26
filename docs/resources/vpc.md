---
page_title: "arubacloud_vpc Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud VPC Network.
---

# arubacloud_vpc (Resource)

Manages an ArubaCloud VPC Network.

## Example Usage

```terraform
resource "arubacloud_vpc" "basic" {
  name = "basic-vpc"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_vpc.example <vpc_id>
```
