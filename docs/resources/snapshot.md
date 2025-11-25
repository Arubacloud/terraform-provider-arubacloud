---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_snapshot"
sidebar_current: "docs-resource-snapshot"
description: |-
  Snapshot provides point-in-time copies of block storage volumes for backup and recovery.
---

# arubacloud_snapshot

Snapshots allow you to back up and restore block storage volumes.

## Usage example

```hcl
resource "arubacloud_snapshot" "example" {
  name           = "example-snapshot"
  project_id     = arubacloud_project.example.id
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  volume_id      = arubacloud_blockstorage.example.id
}
```

## Argument reference

* `name` - (Required)[string] The name of the snapshot.
* `project_id` - (Required)[string] The project ID.
* `location` - (Required)[string] The location for the snapshot.
* `billing_period` - (Required)[string] Billing period (only "Hour" allowed).
* `volume_id` - (Required)[string] The ID of the block storage volume to snapshot.

## Attribute reference

* `id` - (Computed)[string] The ID of the snapshot.

## Import

To import a snapshot, define an empty resource in your plan:

```
resource "arubacloud_snapshot" "example" {
}
```

Import using the snapshot ID:

```
terraform import arubacloud_snapshot.example <snapshot_id>
```
