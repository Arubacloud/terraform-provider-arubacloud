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

* `id` - (Required)[string] The ID of the Backup to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Backup.
* `location` - (Computed)[string] The location of the Backup.
* `project_id` - (Computed)[string] The project ID.
* `type` - (Computed)[string] Type of backup (Full, Incremental).
* `volume_id` - (Computed)[string] Volume ID for the Backup.
* `billing_period` - (Computed)[string] Billing period.
* `retention_days` - (Computed)[int] Retention days for the Backup.
* `tags` - (Computed)[list(string)] Tags for the Backup.
