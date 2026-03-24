# Step 1: Create a project as infrastructure container
resource "arubacloud_project" "test" {
  name        = "terraform-wordpress-test-project"
  description = "Project for testing WordPress + MySQL deployment"
  tags        = ["terraform", "wordpress", "test"]
}
