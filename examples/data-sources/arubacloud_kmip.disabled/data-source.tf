terraform {
  required_providers {
    arubacloud = {
      source = "aruba/arubacloud"
    }
  }
}

# Retrieve information about an existing KMIP endpoint
data "arubacloud_kmip" "existing" {
  id = "kmip-12345"
}

# Output KMIP information
output "kmip_name" {
  value       = data.arubacloud_kmip.existing.name
  description = "The name of the KMIP endpoint"
}

output "kmip_endpoint" {
  value       = data.arubacloud_kmip.existing.endpoint
  description = "The KMIP endpoint URL"
}

output "kmip_description" {
  value       = data.arubacloud_kmip.existing.description
  description = "The description of the KMIP endpoint"
}
