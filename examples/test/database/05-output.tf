output "dbaas_id" {
  value       = arubacloud_dbaas.test.id
  description = "DBaaS instance ID"
}

output "dbaas_uri" {
  value       = arubacloud_dbaas.test.uri
  description = "DBaaS instance URI"
}

output "database_id" {
  value       = arubacloud_database.test.id
  description = "Database ID"
}

output "database_name" {
  value       = arubacloud_database.test.name
  description = "Database name"
}

output "dbaas_user_id" {
  value       = arubacloud_dbaasuser.test.id
  description = "DBaaS user ID (username)"
}

output "dbaas_elastic_ip" {
  value       = arubacloud_elasticip.dbaas.address
  description = "Elastic IP address for DBaaS"
}
