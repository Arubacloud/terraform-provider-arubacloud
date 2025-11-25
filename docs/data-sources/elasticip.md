---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_elasticip"
sidebar_current: "docs-datasource-elasticip"
description: |-
  Data source for querying Elastic IP resources in ArubaCloud.
---

# arubacloud_elasticip (Data Source)

Use this data source to retrieve information about an Elastic IP resource.

## Usage example

```hcl
data "arubacloud_elasticip" "example" {
  id = "elasticip-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Elastic IP to query.

## Attribute reference

* `address` - (Computed)[string] The Elastic IP address.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
