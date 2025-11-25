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

* `id` - (Computed)[string] The ID of the Elastic IP.
* `name` - (Computed)[string] The name of the Elastic IP.
* `location` - (Computed)[string] The location of the Elastic IP.
* `tags` - (Computed)[list(string)] The tags associated with the Elastic IP.
* `billing_period` - (Computed)[string] The billing period for the Elastic IP.
* `address` - (Computed)[string] The Elastic IP address.
* `project_id` - (Computed)[string] The project ID associated with the Elastic IP.
