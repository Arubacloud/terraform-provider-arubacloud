## KMS Example Resource
resource "arubacloud_kms" "example" {
  name          = "example-kms"
  project_id    = arubacloud_project.example.id
  location      = "ITBG-Bergamo"
  tags          = ["security", "encryption"]
  billing_period = "monthly"
}
