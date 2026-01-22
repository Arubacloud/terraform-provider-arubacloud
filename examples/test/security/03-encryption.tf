# Step 3: Create Encryption & Key Management Resources

# KMS - Key Management Service
resource "arubacloud_kms" "test" {
  name           = "test-kms"
  project_id     = arubacloud_project.test.id
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  tags           = ["encryption", "security"]
}

# KMIP - Key Management Interoperability Protocol
# Note: KMIP creation returns "invalid status" error - may require special KMS configuration or feature enablement
# resource "arubacloud_kmip" "test" {
#   name       = "test-kmip"
#   project_id = arubacloud_project.test.id
#   kms_id     = arubacloud_kms.test.id
# }

# Key - Encryption Key within KMS
resource "arubacloud_key" "test" {
  name       = "test-encryption-key"
  project_id = arubacloud_project.test.id
  kms_id     = arubacloud_kms.test.id
  algorithm  = "Aes"
}
