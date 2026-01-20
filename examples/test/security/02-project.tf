# Step 1: Create a Project (Foundation for all resources)
resource "arubacloud_project" "test" {
  name        = "terraform-security-test-project"
  description = "Project for testing Terraform security resources"
  tags        = ["terraform", "security", "test"]
}
