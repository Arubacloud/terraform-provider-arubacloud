terraform {
  required_providers {
    arubacloud = {
      source = "aruba/arubacloud"
    }
  }
}

# Create a project
resource "arubacloud_project" "example" {
  name        = "security-project"
  description = "Project for KMS and KMIP"
}

# Create a KMS instance
resource "arubacloud_kms" "example" {
  name           = "production-kms"
  project_id     = arubacloud_project.example.id
  location       = "it-mil1"
  billing_period = "monthly"
  tags           = ["production", "kmip"]
}

# Create a KMIP endpoint
resource "arubacloud_kmip" "example" {
  name       = "production-kmip"
  project_id = arubacloud_project.example.id
  kms_id     = arubacloud_kms.example.id
}

# Output the KMIP ID
output "kmip_id" {
  value       = arubacloud_kmip.example.id
  description = "The ID of the KMIP endpoint"
}
