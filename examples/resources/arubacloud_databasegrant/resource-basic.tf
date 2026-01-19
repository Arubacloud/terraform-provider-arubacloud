resource "arubacloud_databasegrant" "example" {
  project_id = arubacloud_project.example.id
  dbaas_id   = arubacloud_dbaas.example.id
  database   = arubacloud_database.example.id
  user_id    = arubacloud_dbaasuser.example.id
  role       = "readwrite"  # Options: liteadmin, readwrite, readonly
}
