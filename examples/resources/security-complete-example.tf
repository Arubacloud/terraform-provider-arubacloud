terraform {
  required_providers {
    arubacloud = {
      source = "aruba/arubacloud"
    }
  }
}

provider "arubacloud" {
  # Configure your API credentials
  # api_key    = "your-api-key"
  # api_secret = "your-api-secret"
}

# Create a project for security resources
resource "arubacloud_project" "security" {
  name        = "enterprise-security"
  description = "Project for enterprise security infrastructure"
}

# Create a KMS instance for key management
resource "arubacloud_kms" "main" {
  name           = "enterprise-kms"
  project_id     = arubacloud_project.security.id
  location       = "it-mil1"
  billing_period = "monthly"
  tags           = ["production", "security", "compliance"]
}

# Create multiple encryption keys with different algorithms

# AES-256 key for database encryption
resource "arubacloud_key" "database" {
  name        = "database-encryption-key"
  project_id  = arubacloud_project.security.id
  kms_id      = arubacloud_kms.main.id
  algorithm   = "AES"
  size        = 256
  description = "AES-256 encryption key for database at rest encryption"
}

# AES-128 key for file encryption
resource "arubacloud_key" "files" {
  name        = "file-encryption-key"
  project_id  = arubacloud_project.security.id
  kms_id      = arubacloud_kms.main.id
  algorithm   = "AES"
  size        = 128
  description = "AES-128 encryption key for file system encryption"
}

# RSA-2048 key for secure communications
resource "arubacloud_key" "communications" {
  name        = "rsa-communications-key"
  project_id  = arubacloud_project.security.id
  kms_id      = arubacloud_kms.main.id
  algorithm   = "RSA"
  size        = 2048
  description = "RSA-2048 key for secure inter-service communications"
}

# RSA-4096 key for high-security use cases
resource "arubacloud_key" "high_security" {
  name        = "rsa-high-security-key"
  project_id  = arubacloud_project.security.id
  kms_id      = arubacloud_kms.main.id
  algorithm   = "RSA"
  size        = 4096
  description = "RSA-4096 key for high-security critical operations"
}

# Create KMIP endpoint for third-party integration
resource "arubacloud_kmip" "enterprise" {
  name       = "enterprise-kmip-endpoint"
  project_id = arubacloud_project.security.id
  kms_id     = arubacloud_kms.main.id
}

# Outputs for reference
output "project_id" {
  value       = arubacloud_project.security.id
  description = "Security project ID"
}

output "kms_id" {
  value       = arubacloud_kms.main.id
  description = "KMS instance ID"
}

output "kms_uri" {
  value       = arubacloud_kms.main.uri
  description = "KMS instance URI"
}

output "database_key_id" {
  value       = arubacloud_key.database.id
  description = "Database encryption key ID"
}

output "database_key_status" {
  value       = arubacloud_key.database.status
  description = "Database encryption key status"
}

output "communications_key_id" {
  value       = arubacloud_key.communications.id
  description = "Communications RSA key ID"
}

output "kmip_id" {
  value       = arubacloud_kmip.enterprise.id
  description = "KMIP endpoint ID"
}
