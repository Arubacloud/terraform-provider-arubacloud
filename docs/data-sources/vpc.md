---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_vpc"
sidebar_current: "docs-datasource-vpc"
description: |-
  Data source for querying VPC resources in ArubaCloud.
---

# arubacloud_vpc (Data Source)

Use this data source to retrieve information about a VPC resource.

## Usage example

```hcl
data "arubacloud_vpc" "example" {
  id = "vpc-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the VPC to query.

## Attribute reference

* `name` - (Computed)[string] The name of the VPC.
* ...other attributes...
