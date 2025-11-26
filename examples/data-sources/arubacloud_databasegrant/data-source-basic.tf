data "arubacloud_database_grant" "example" {
  name       = "example-database-grant"
  project_id = "example-project"
  database   = "example-db"
  privileges = ["SELECT", "INSERT"]
}
