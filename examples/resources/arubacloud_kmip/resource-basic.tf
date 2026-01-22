# Note: KMIP creation may return "invalid status" error depending on KMS configuration
resource "arubacloud_kmip" "basic" {
  name       = "basic-kmip"
  project_id = "your-project-id"
  kms_id     = "your-kms-id"
}
