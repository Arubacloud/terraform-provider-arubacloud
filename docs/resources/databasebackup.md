page_title: "arubacloud_database_backup Resource - ArubaCloud"
# arubacloud_database_backup (Resource)
```terraform
resource "arubacloud_database_backup" "basic" {
  name = "basic-database-backup"
  database_id = "database-id"
}
```
  Manages an ArubaCloud Database Backup.
---

# arubacloud_databasebackup (Resource)

Manages an ArubaCloud Database Backup.

## Example Usage

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
