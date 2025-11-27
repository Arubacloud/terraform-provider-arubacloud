
---
page_title: "arubacloud_backup Resource - ArubaCloud"
subcategory: "Storage"
description: |-
  Manages an ArubaCloud Storage Backup.
---

# arubacloud_backup

Manages an ArubaCloud Storage Backup.

```terraform
resource "arubacloud_backup" "basic" {
  name          = "example-backup"
  location      = "de-1"
  project_id    = "project-123"
  type          = "full"
  volume_id     = "volume-123"
  billing_period = "monthly"

  # optional
  retention_days = 30
  tags           = ["env:dev", "service:demo"]
}
```

## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_backup.example <backup_id>
```
