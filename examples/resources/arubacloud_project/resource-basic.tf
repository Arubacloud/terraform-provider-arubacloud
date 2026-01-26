resource "arubacloud_project" "basic" {
  name        = "basic-project"
  description = "Project for testing Terraform provider"
  tags        = ["terraform", "test"]
}
