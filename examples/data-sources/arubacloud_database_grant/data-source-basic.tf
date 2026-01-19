data "arubacloud_databasegrant" "example" {
  project_id = "project-123"
  dbaas_id   = "dbaas-456"
  database   = "mydb"
  user_id    = "myuser"
}
