# Step 1: Create a Project (Foundation for all resources)
resource "arubacloud_project" "test" {
  name        = "terraform-schedule-test-project"
  description = "Project for testing Terraform schedule resources"
  tags        = ["terraform", "schedule", "test"]
}
