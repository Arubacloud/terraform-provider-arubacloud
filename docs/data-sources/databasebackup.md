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

* `id` - (Required)[string] The ID of the Database Backup to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Database Backup.
* `location` - (Computed)[string] The location of the Database Backup.
* `zone` - (Computed)[string] Zone for the Database Backup.
* `dbaas_id` - (Computed)[string] The DBaaS ID this backup belongs to.
* `database` - (Computed)[string] Database to backup (ID or name).
* `billing_period` - (Computed)[string] Billing period.
* `tags` - (Computed)[list(string)] Tags for the Database Backup.
