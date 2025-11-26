# This file has been removed as part of the legacy resource cleanup.
page_title: "arubacloud_vpc_peering_route Resource - ArubaCloud"
subcategory: "Network"
description: |-
  Manages an ArubaCloud VPC Peering Route.

# arubacloud_vpc_peering_route (Resource)

Manages an ArubaCloud VPC Peering Route.

## Example Usage

```terraform
resource "arubacloud_vpc_peering_route" "basic" {
  vpc_peering_id = "vpc-peering-id"
  destination_cidr = "10.0.0.0/16"
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_vpc_peering_route.example <route_id>
```
