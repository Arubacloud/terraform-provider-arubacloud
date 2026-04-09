---
page_title: "arubacloud_snapshot"
subcategory: "Storage"
description: |-
  Reads an existing ArubaCloud snapshot.
---

# arubacloud_snapshot

Reads an existing ArubaCloud snapshot.

```terraform
data "arubacloud_snapshot" "basic" {
  id         = "snapshot-id"
  project_id = "your-project-id"
}

output "snapshot_name" {
  value = data.arubacloud_snapshot.basic.name
}
output "snapshot_location" {
  value = data.arubacloud_snapshot.basic.location
}
output "snapshot_volume_id" {
  value = data.arubacloud_snapshot.basic.volume_id
}
```

## Schema

### Arguments

The following arguments are supported:

#### Required

- `id` (String) Snapshot identifier
- `project_id` (String) ID of the project this Snapshot belongs to

### Attributes Reference

In addition to all arguments above, the following attributes are exported:

#### Read-Only

- `billing_period` (String) Billing period (only 'Hour' allowed)
- `location` (String) Snapshot location
- `name` (String) Snapshot name
- `volume_id` (String) ID of the volume this snapshot is for
