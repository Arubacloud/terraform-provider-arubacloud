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

* `id` - (Required)[string] The ID of the Snapshot to query.

## Attribute reference

* `name` - (Computed)[string] The name of the Snapshot.
* `project_id` - (Computed)[string] The project ID.
* `location` - (Computed)[string] The location of the Snapshot.
* `billing_period` - (Computed)[string] Billing period (only 'Hour' allowed).
* `volume_id` - (Computed)[string] ID of the volume this snapshot is for.
