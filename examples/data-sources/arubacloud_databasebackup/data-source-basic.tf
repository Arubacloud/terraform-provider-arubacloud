data "arubacloud_database_backup" "example" {
  name       = "example-database-backup"
  project_id = "example-project"
  location   = "eu-1"
  type       = "Full"
}
