resource "arubacloud_vpc" "basic" {
  name       = "basic-vpc"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  tags       = ["network", "test"]
}
