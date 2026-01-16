# Step 1: Create a Project (Foundation for all resources)
resource "arubacloud_project" "test" {
  name        = "terraform-container-test-project"
  description = "Project for testing Terraform container resources"
  tags        = ["terraform", "container", "test"]
}
