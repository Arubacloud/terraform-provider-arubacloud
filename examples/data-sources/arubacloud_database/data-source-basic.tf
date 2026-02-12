data "arubacloud_database" "basic" {
  id = "your-database-id"
}

output "database_project_id" {
  value = data.arubacloud_database.basic.project_id
}
output "database_dbaas_id" {
  value = data.arubacloud_database.basic.dbaas_id
}
output "database_name" {
  value = data.arubacloud_database.basic.name
}
