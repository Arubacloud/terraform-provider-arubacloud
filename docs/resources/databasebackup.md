---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasebackup"
sidebar_current: "docs-resource-databasebackup"
description: |-
  Database Backup provides backups for individual databases within DBaaS instances.
---

# arubacloud_databasebackup

Database Backups allow you to protect and restore individual databases.

## Usage example

```hcl
resource "arubacloud_databasebackup" "example" {
  name           = "example-db-backup"
  location       = "ITBG-Bergamo"
  tags           = ["dbbackup", "test"]
  zone           = "ITBG-1"
  dbaas_id       = arubacloud_dbaas.example.id
  database       = arubacloud_database.example.id
  billing_period = "Hour"
}
```

## Argument reference

* `name` - (Required)[string] The name of the database backup.
* `location` - (Required)[string] The location for the database backup.
* `tags` - (Optional)[list(string)] Tags for the database backup resource.
* `zone` - (Required)[string] Zone for the database backup.
* `dbaas_id` - (Required)[string] The ID of the DBaaS instance this backup belongs to.
* `database` - (Required)[string] The ID of the database to back up.
* `billing_period` - (Required)[string] Billing period.

## Attribute reference

* `id` - (Computed)[string] The ID of the database backup.

## Import

To import a database backup, define an empty resource in your plan:

```
resource "arubacloud_databasebackup" "example" {
}
```

Import using the database backup ID:

```
terraform import arubacloud_databasebackup.example <databasebackup_id>
```
