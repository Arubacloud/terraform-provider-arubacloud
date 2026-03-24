output "wordpress_url" {
  value       = arubacloud_elasticip.vm.address != null ? format("http://%s", arubacloud_elasticip.vm.address) : "http://<vm-elastic-ip>"
  description = "URL to open WordPress setup/login page"
}

output "wordpress_admin_user" {
  value       = "admin"
  description = "WordPress admin username"
}

output "wordpress_admin_password" {
  value       = var.wordpress_admin_password
  sensitive   = true
  description = "WordPress admin password"
}

output "wordpress_db_host" {
  value       = arubacloud_elasticip.dbaas.address
  description = "MySQL endpoint (DBaaS Elastic IP)"
}

output "wordpress_db_name" {
  value       = arubacloud_database.wordpress.name
  description = "WordPress MySQL database name"
}

output "wordpress_db_user" {
  value       = arubacloud_dbaasuser.wordpress.username
  description = "WordPress MySQL username"
}
