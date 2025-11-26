---
page_title: "arubacloud_subnet Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud Subnet.
---

# arubacloud_subnet (Resource)

Manages an ArubaCloud Subnet.

## Example Usage

```terraform
resource "arubacloud_subnet" "basic" {
  name = "basic-subnet"
  vpc_id = "vpc-id"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_subnet.example <subnet_id>
```
