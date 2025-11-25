---
subcategory: "Storage"
layout: "arubacloud"
page_title: "ArubaCloud: arubacloud_backup"
sidebar_current: "docs-datasource-backup"
description: |-
  Data source for querying Backup resources in ArubaCloud.
---

# arubacloud_backup (Data Source)

Use this data source to retrieve information about a Backup resource.

## Usage example

```hcl
data "arubacloud_backup" "example" {
  id = "backup-id"
}
```

## Argument reference

* `id` - (Required)[string] The ID of the backup to query.

## Attribute reference

* `name` - (Computed)[string] The name of the backup.
* `project_id` - (Computed)[string] The project ID.
* ...other attributes...
