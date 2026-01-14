# Step 2: Create a Project (Foundation for all resources)
resource "arubacloud_project" "test" {
  name        = "terraform-test-project"
  description = "Project for testing Terraform provider"
  tags        = ["terraform", "test"]
}

