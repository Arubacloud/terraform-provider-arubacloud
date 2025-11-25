---
subcategory: "Network"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_securitygroup"
sidebar_current: "docs-datasource-securitygroup"
description: |-
  Data source for querying Security Group resources in ArubaCloud.
---

# arubacloud_securitygroup (Data Source)

Use this data source to retrieve information about a Security Group resource.

## Usage example

```hcl
data "arubacloud_securitygroup" "example" {
  id = "securitygroup-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the security group to query.

## Attribute reference

* `name` - (Computed)[string] The name of the security group.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
