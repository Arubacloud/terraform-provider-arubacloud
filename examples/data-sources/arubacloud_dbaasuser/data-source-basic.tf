data "arubacloud_dbaas_user" "example" {
  name       = "example-dbaas-user"
  project_id = "example-project"
  database   = "example-db"
  role       = "admin"
}
