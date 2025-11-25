---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_backup"
sidebar_current: "docs-resource-backup"
description: |-
  Backup provides scheduled or manual backups of block storage volumes for data protection.
---

# arubacloud_backup

Backups allow you to protect and restore block storage volumes.

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
* `location` - (Required)[string] The location for the backup.
* `tags` - (Optional)[list(string)] Tags for the backup resource.
* `project_id` - (Required)[string] The project ID.
* `type` - (Required)[string] Type of backup ("Full", "Incremental").
* `volume_id` - (Required)[string] The ID of the block storage volume to back up.
* `retention_days` - (Optional)[int] Number of days to retain the backup.
* `billing_period` - (Required)[string] Billing period.

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
