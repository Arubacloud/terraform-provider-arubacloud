---
page_title: "arubacloud_databasebackup Resource - ArubaCloud"
subcategory: "Database"
description: |-
  Manages an ArubaCloud Database Backup.
---

# arubacloud_databasebackup

Manages an ArubaCloud Database Backup. 

```terraform
resource "arubacloud_databasebackup" "basic" {
  name       = "example-database-backup"
  database   = "example-database"
  location   = "example-location"
  project_id = "example-project"
}
```


## Argument Reference

<!-- tfplugindocs will inject schema-based arguments here -->

## Attribute Reference

<!-- tfplugindocs will inject schema-based attributes here -->

## Import

```shell
terraform import arubacloud_databasebackup.example <databasebackup_id>
```
