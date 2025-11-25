---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_backup"
sidebar_current: "docs-resource-backup"
description: |-
  Backup allows you to create point-in-time copies of your data for disaster recovery.
---

# arubacloud_backup

Backups can be scheduled or created manually for block storage volumes.

## Usage example

```hcl
resource "arubacloud_backup" "example" {
  name           = "example-backup"
  location       = "ITBG-Bergamo"
  tags           = ["backup", "test"]
  project_id     = arubacloud_project.example.id
  type           = "Full"
  volume_id      = arubacloud_blockstorage.example.id
  retention_days = 30
  billing_period = "Hour"
}
```

## Argument reference

* `name` - (Required)[string] The name of the backup.
* `project_id` - (Required)[string] The project ID.
* `volume_id` - (Required)[string] The block storage volume ID.
* ...other arguments...

## Attribute reference

* `id` - (Computed)[string] The ID of the backup.

## Import

To import a backup, define an empty resource in your plan:

```
resource "arubacloud_backup" "example" {
}
```

Import using the backup ID:

```
terraform import arubacloud_backup.example <backup_id>
```
