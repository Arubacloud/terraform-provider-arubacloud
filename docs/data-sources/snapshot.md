---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_snapshot"
sidebar_current: "docs-datasource-snapshot"
description: |-
  Data source for querying Snapshot resources in ArubaCloud.
---

# arubacloud_snapshot (Data Source)

Use this data source to retrieve information about a Snapshot resource.

## Usage example

```hcl
data "arubacloud_snapshot" "example" {
  id = "snapshot-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the snapshot to query.

## Attribute reference

* `name` - (Computed)[string] The name of the snapshot.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
