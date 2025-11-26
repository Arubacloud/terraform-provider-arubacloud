resource "arubacloud_databasegrant" "example" {
  name       = "example-database-grant"
  database   = "example-database"
  user       = "example-user"
  privileges = ["SELECT", "INSERT"]
  project_id = "example-project"
}
