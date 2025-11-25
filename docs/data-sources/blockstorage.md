---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_blockstorage"
sidebar_current: "docs-datasource-blockstorage"
description: |-
  Data source for querying Block Storage resources in ArubaCloud.
---

# arubacloud_blockstorage (Data Source)

Use this data source to retrieve information about a Block Storage resource.

## Usage example

```hcl
data "arubacloud_blockstorage" "example" {
  id = "blockstorage-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the Block Storage to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Block Storage.
* `project_id` - (Computed)[string] The project ID.
* `properties` - (Computed)[object] Block Storage properties:
  * `size_gb` - (Computed)[int] Size of the block storage in GB.
  * `billing_period` - (Computed)[string] Billing period (only 'Hour' allowed).
  * `zone` - (Computed)[string] Zone of the block storage.
  * `type` - (Computed)[string] Type of block storage (Standard, Performance).
  * `snapshot_id` - (Computed)[string] Snapshot ID for the block storage.
  * `bootable` - (Computed)[bool] Whether the block storage is bootable.
  * `image` - (Computed)[string] Image for the block storage.
