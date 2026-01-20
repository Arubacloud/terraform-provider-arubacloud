# Step 3: Create Encryption & Key Management Resources

# KMS - Key Management Service
resource "arubacloud_kms" "test" {
  name           = "test-kms"
  project_id     = arubacloud_project.test.id
  location       = "ITBG-Bergamo"
  tags           = ["security", "encryption", "kms", "test"]
  billing_period = "Hour"
}
