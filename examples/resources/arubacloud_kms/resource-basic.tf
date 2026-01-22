resource "arubacloud_kms" "basic" {
  name           = "basic-kms"
  project_id     = "your-project-id"
  location       = "ITBG-Bergamo"
  billing_period = "Hour"
  tags           = ["encryption", "security"]
}
