---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_snapshot"
sidebar_current: "docs-resource-snapshot"
description: |-
  Snapshot allows you to capture the state of a block storage volume at a specific point in time.
---

# arubacloud_snapshot

Snapshots can be used to restore block storage volumes or create new volumes from a saved state.

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
* `volume_id` - (Required)[string] The block storage volume ID.
* ...other arguments...

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
