terraform {
  required_providers {
    arubacloud = {
      source = "aruba/arubacloud"
    }
  }
}

# Reference existing resources
data "arubacloud_project" "example" {
  id = "project-12345"
}

data "arubacloud_kms" "example" {
  id         = "kms-67890"
  project_id = data.arubacloud_project.example.id
}

# Retrieve information about an existing key
data "arubacloud_key" "existing" {
  id         = "key-abcdef"
  project_id = data.arubacloud_project.example.id
  kms_id     = data.arubacloud_kms.example.id
}

# Output key information
output "key_name" {
  value       = data.arubacloud_key.existing.name
  description = "The name of the key"
}

output "key_algorithm" {
  value       = data.arubacloud_key.existing.algorithm
  description = "The encryption algorithm used by the key"
}

output "key_size" {
  value       = data.arubacloud_key.existing.size
  description = "The size of the key in bits"
}

output "key_status" {
  value       = data.arubacloud_key.existing.status
  description = "The status of the key"
}
