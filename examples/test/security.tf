# Copyright (c) HashiCorp, Inc.

## KMS Example Resource
resource "arubacloud_kms" "example" {
  name           = "example-kms"
  project_id     = "example-project-id"
  location       = "eu-central-1"
  tags           = ["security", "encryption"]
  billing_period = "monthly"
}

# KMIP Example Resource
resource "arubacloud_kmip" "example" {
  name       = "example-kmip"
  project_id = "example-project-id"
  kms_id     = arubacloud_kms.example.id
}
