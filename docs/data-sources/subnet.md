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

* `id` - (Required)[string] The ID of the subnet to query.

## Attribute reference

* `cidr_block` - (Computed)[string] The subnet CIDR block.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
