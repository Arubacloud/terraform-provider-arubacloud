---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_subnet"
sidebar_current: "docs-datasource-subnet"
description: |-
  Data source for querying Subnet resources in ArubaCloud.
---

# arubacloud_subnet (Data Source)

Use this data source to retrieve information about a Subnet resource.

## Usage example

```hcl
data "arubacloud_subnet" "example" {
  id = "subnet-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Subnet to query.

## Attribute reference

* `name` - [string] The name of the Subnet.
* `location` - [string] The location of the Subnet.
* `tags` - [list(string)] Tags for the Subnet.
* `project_id` - [string] The project ID.
* `vpc_id` - [string] The VPC ID.
* `type` - [string] Subnet type (Basic or Advanced).
* `network` - [object]
  * `address` - [string] Network address in CIDR notation.
* `dhcp` - [object]
  * `enabled` - [bool] DHCP enabled.
  * `range` - [object]
  * `start` - [string] Starting IP address.
  * `count` - [int] Number of available IP addresses.
* `routes` - [list(object)] Routes:
  * `address` - [string] IP address of the route.
  * `gateway` - [string] Gateway.
* `dns` - [list(string)] List of DNS IP addresses.
