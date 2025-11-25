---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_restore"
sidebar_current: "docs-datasource-restore"
description: |-
  Data source for querying Restore resources in ArubaCloud.
---

# arubacloud_restore (Data Source)

Use this data source to retrieve information about a Restore resource.

## Usage example

```hcl
data "arubacloud_restore" "example" {
  id = "restore-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the restore to query.

## Attribute reference

* `name` - (Computed)[string] The name of the restore.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
