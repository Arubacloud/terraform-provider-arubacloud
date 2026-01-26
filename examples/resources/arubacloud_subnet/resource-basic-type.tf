resource "arubacloud_subnet" "basic_type" {
  name       = "basic-subnet"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  vpc_id     = "your-vpc-id"  # Replace with your VPC ID
  type       = "Basic"  # Basic type doesn't require network block
  tags       = ["network", "basic"]
}
