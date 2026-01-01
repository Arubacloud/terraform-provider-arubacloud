resource "arubacloud_subnet" "basic" {
  name       = "basic-subnet"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  vpc_id     = "your-vpc-id"  # Replace with your VPC ID
  type       = "Advanced"  # Required: "Basic" or "Advanced"
  network = {
    address = "10.0.1.0/24"  # CIDR notation
  }
  tags       = ["network", "test"]
}
