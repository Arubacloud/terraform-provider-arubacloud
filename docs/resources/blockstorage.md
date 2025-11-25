---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_blockstorage"
sidebar_current: "docs-resource-blockstorage"
description: |-
  Block Storage provides persistent, high-performance storage volumes for your cloud resources.
---

# arubacloud_blockstorage

Block Storage volumes can be attached to cloud servers and used for data storage.

## Usage example

```hcl
resource "arubacloud_blockstorage" "example" {
  name       = "example-block-storage"
  project_id = arubacloud_project.example.id
  properties = {
    size_gb        = 100
    billing_period = "Hour"
    zone           = "ITBG-Bergamo"
    type           = "Standard"
    bootable       = true
    image          = "ubuntu-22.04"
    snapshot_id    = arubacloud_snapshot.example.id
  }
}
```

## Argument reference

* `name` - (Required)[string] The name of the block storage volume.
* `project_id` - (Required)[string] The project ID.
* `properties` - (Required)[object] Block storage properties:
  * `size_gb` - (Required)[int] Size of the block storage in GB.
  * `billing_period` - (Required)[string] Billing period (only "Hour" allowed).
  * `zone` - (Required)[string] Zone where block storage will be created.
  * `type` - (Required)[string] Type of block storage ("Standard", "Performance").
  * `snapshot_id` - (Optional)[string] Snapshot ID to restore from.
  * `bootable` - (Optional)[bool] Whether the block storage is bootable.
  * `image` - (Optional)[string] Image for the block storage.

## Attribute reference

* `id` - (Computed)[string] The ID of the block storage volume.

## Import

To import a block storage volume, define an empty resource in your plan:

```
resource "arubacloud_blockstorage" "example" {
}
```

Import using the block storage ID:

```
terraform import arubacloud_blockstorage.example <blockstorage_id>
```
