resource "arubacloud_databasegrant" "example" {
  project_id = "example-project-id"
  dbaas_id   = "example-dbaas-id"
  database   = "example-database"
  user_id    = "example-user"
  role       = "readwrite"  # Valid roles: readonly, readwrite, liteadmin
}
