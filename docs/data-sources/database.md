---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_database"
sidebar_current: "docs-datasource-database"
description: |-
  Data source for querying Database resources in ArubaCloud.
---

# arubacloud_database (Data Source)

Use this data source to retrieve information about a Database resource.

## Usage example

```hcl
data "arubacloud_database" "example" {
  id = "database-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the database to query.

## Attribute reference

* `name` - (Computed)[string] The name of the database.
* `dbaas_id` - (Computed)[string] The DBaaS ID.
* ...other attributes...
