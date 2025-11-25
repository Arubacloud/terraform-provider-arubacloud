---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasegrant"
sidebar_current: "docs-datasource-databasegrant"
description: |-
  Data source for querying Database Grant resources in ArubaCloud.
---

# arubacloud_databasegrant (Data Source)

Use this data source to retrieve information about a Database Grant resource.

## Usage example

```hcl
data "arubacloud_databasegrant" "example" {
  id = "databasegrant-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the database grant to query.

## Attribute reference

* `database` - (Computed)[string] The database ID.
* `user_id` - (Computed)[string] The user ID.
* `role` - (Computed)[string] The role assigned.
* ...other attributes...
