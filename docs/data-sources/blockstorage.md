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

* `id` - (Required)[string] The ID of the block storage to query.

## Attribute reference

* `name` - (Computed)[string] The name of the block storage.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
