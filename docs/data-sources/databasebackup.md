page_title: "arubacloud_database_backup Data Source - ArubaCloud"
# arubacloud_database_backup (Data Source)
```terraform
data "arubacloud_database_backup" "basic" {
  id = "database-backup-id"
}
```
  Reads an existing ArubaCloud database backup.
---

# arubacloud_databasebackup (Data Source)

Reads an existing ArubaCloud database backup.

```terraform
data "arubacloud_database_backup" "example" {
  name       = "example-database-backup"
  project_id = "example-project"
  location   = "eu-1"
  type       = "Full"
}
```


## Schema

<no value>
