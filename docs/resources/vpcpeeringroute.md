---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpcpeeringroute"
sidebar_current: "docs-resource-vpcpeeringroute"
description: |-
  VPC Peering Route defines routing between peered VPCs.
---

# arubacloud_vpcpeeringroute

VPC Peering Routes allow you to control network traffic between peered VPCs.

## Usage example

```hcl
resource "arubacloud_vpcpeeringroute" "example" {
  name                   = "example-vpc-peering-route"
  location               = "ITBG-Bergamo"
  tags                   = ["route", "prod"]
  local_network_address  = "10.0.0.0/24"
  remote_network_address = "192.168.1.0/24"
  billing_period         = "Hour"
  vpc_peering_id         = arubacloud_vpcpeering.example.id
}
```

## Argument reference

* `id` - (Computed)[string] The ID of the VPC Peering Route.
* `name` - (Required)[string] The name of the VPC Peering Route.
* `local_network_address` - (Required)[string] Local network address.
* `remote_network_address` - (Required)[string] Remote network address.
* ...other arguments...

## Import

To import a VPC Peering Route, define an empty resource in your plan:

```
resource "arubacloud_vpcpeeringroute" "example" {
}
```

Import using the VPC Peering Route ID:

```
terraform import arubacloud_vpcpeeringroute.example <vpcpeeringroute_id>
```
