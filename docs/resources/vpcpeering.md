---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpcpeering"
sidebar_current: "docs-resource-vpcpeering"
description: |-
  VPC Peering connects two VPCs for private communication.
---

# arubacloud_vpcpeering

VPC Peering allows resources in different VPCs to communicate securely.

## Usage example

```hcl
resource "arubacloud_vpcpeering" "example" {
  name     = "example-vpc-peering"
  location = "ITBG-Bergamo"
  tags     = ["peering", "prod"]
  peer_vpc = arubacloud_vpc.peer.id
}
```

## Argument reference

* `id` - (Computed)[string] The ID of the VPC Peering.
* `name` - (Required)[string] The name of the VPC Peering.
* `peer_vpc` - (Required)[string] The peer VPC ID.
* ...other arguments...

## Import

To import a VPC Peering, define an empty resource in your plan:

```
resource "arubacloud_vpcpeering" "example" {
}
```

Import using the VPC Peering ID:

```
terraform import arubacloud_vpcpeering.example <vpcpeering_id>
```
