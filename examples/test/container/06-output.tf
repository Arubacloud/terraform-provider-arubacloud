output "container_registry_id" {
  value       = arubacloud_containerregistry.test.id
  description = "Container Registry ID"
}

output "container_registry_uri" {
  value       = arubacloud_containerregistry.test.uri
  description = "Container Registry URI"
}

output "kaas_id" {
  value       = arubacloud_kaas.test.id
  description = "KaaS cluster ID"
}

output "kaas_uri" {
  value       = arubacloud_kaas.test.uri
  description = "KaaS cluster URI"
}

output "container_registry_elastic_ip" {
  value       = arubacloud_elasticip.container_registry.address
  description = "Elastic IP address for container registry"
}
