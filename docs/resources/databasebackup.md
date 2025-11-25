---
subcategory: "Database"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_databasebackup"
sidebar_current: "docs-resource-databasebackup"
description: |-
  Database Backup allows you to create backups of your databases for disaster recovery.
---

# arubacloud_databasebackup

Database Backups can be scheduled or created manually for databases in DBaaS clusters.

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
* `dbaas_id` - (Required)[string] The DBaaS instance ID.
* `database` - (Required)[string] The database ID.
* ...other arguments...

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
