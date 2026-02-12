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
  description = "Project for KMS and encryption keys"
}

# Create a KMS instance
resource "arubacloud_kms" "example" {
  name           = "production-kms"
  project_id     = arubacloud_project.example.id
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  tags           = ["production", "encryption"]
}

# Create an encryption key for AES
resource "arubacloud_key" "aes_key" {
  name       = "aes-encryption-key"
  project_id = arubacloud_project.example.id
  kms_id     = arubacloud_kms.example.id
  algorithm  = "Aes"
}

# Create an RSA key for asymmetric encryption
resource "arubacloud_key" "rsa_key" {
  name       = "rsa-encryption-key"
  project_id = arubacloud_project.example.id
  kms_id     = arubacloud_kms.example.id
  algorithm  = "Rsa"
}

# Output the key IDs
output "aes_key_id" {
  value       = arubacloud_key.aes_key.id
  description = "The ID of the AES encryption key"
}

output "rsa_key_id" {
  value       = arubacloud_key.rsa_key.id
  description = "The ID of the RSA encryption key"
}
