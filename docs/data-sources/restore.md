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

* `id` - (Required)[string] The ID of the Restore to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Restore.
* `location` - (Computed)[string] The location of the Restore.
* `project_id` - (Computed)[string] The project ID.
* `volume_id` - (Computed)[string] Volume ID to restore.
* `tags` - (Computed)[list(string)] Tags for the Restore.
