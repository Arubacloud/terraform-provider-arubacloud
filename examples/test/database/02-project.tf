# Step 1: Create a Project (Foundation for all resources)
resource "arubacloud_project" "test" {
  name        = "terraform-database-test-project"
  description = "Project for testing Terraform database resources"
  tags        = ["terraform", "database", "test"]
}
