resource "arubacloud_databasebackup" "basic" {
  name       = "example-database-backup"
  database   = "example-database"
  location   = "example-location"
  project_id = "example-project"
}
