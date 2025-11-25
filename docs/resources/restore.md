---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_restore"
sidebar_current: "docs-resource-restore"
description: |-
  Restore allows you to recover data from a backup or snapshot to a block storage volume.
---

# arubacloud_restore

Restores allow you to recover block storage volumes to a previous state.

## Usage example

```hcl
resource "arubacloud_restore" "example" {
  name       = "example-restore"
  location   = "ITBG-Bergamo"
  tags       = ["restore", "test"]
  project_id = arubacloud_project.example.id
  volume_id  = arubacloud_blockstorage.example.id
}
```

## Argument reference

* `name` - (Required)[string] The name of the restore operation.
* `location` - (Required)[string] The location for the restore.
* `tags` - (Optional)[list(string)] Tags for the restore resource.
* `project_id` - (Required)[string] The project ID.
* `volume_id` - (Required)[string] The ID of the block storage volume to restore.

## Attribute reference

* `id` - (Computed)[string] The ID of the restore operation.

## Import

To import a restore, define an empty resource in your plan:

```
resource "arubacloud_restore" "example" {
}
```

Import using the restore ID:

```
terraform import arubacloud_restore.example <restore_id>
```
