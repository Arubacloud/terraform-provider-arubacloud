---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasebackup"
sidebar_current: "docs-datasource-databasebackup"
description: |-
  Data source for querying Database Backup resources in ArubaCloud.
---

# arubacloud_databasebackup (Data Source)

Use this data source to retrieve information about a Database Backup resource.

## Usage example

```hcl
data "arubacloud_databasebackup" "example" {
  id = "databasebackup-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the database backup to query.

## Attribute reference

* `name` - (Computed)[string] The name of the database backup.
* `database` - (Computed)[string] The database ID.
* ...other attributes...
