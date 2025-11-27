---
page_title: "arubacloud_vpcpeering Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud VPC Peering.
---

# arubacloud_vpcpeering

Manages an ArubaCloud VPC Peering.

## Example Usage

```terraform
resource "arubacloud_vpcpeering" "example" {
  name       = "example-vpc-peering"
  location   = "example-location"
  tags       = ["tag1", "tag2"]
  peer_vpc   = "peer-vpc-id"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_vpcpeering.example <peering_id>
```
