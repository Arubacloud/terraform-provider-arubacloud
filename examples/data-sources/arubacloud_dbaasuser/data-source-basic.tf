data "arubacloud_dbaasuser" "example" {
  username   = "your-db-username"
  project_id = "your-project-id"
  dbaas_id   = "your-dbaas-id"
}

output "dbaasuser_id" {
  value = data.arubacloud_dbaasuser.example.id
}
