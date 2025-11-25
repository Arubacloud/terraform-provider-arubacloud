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

* `id` - (Required)[string] The ID of the Database to query.

## Attribute reference

* `dbaas_id` - (Computed)[string] The DBaaS ID this database belongs to.
* `name` - (Computed)[string] The name of the Database.
