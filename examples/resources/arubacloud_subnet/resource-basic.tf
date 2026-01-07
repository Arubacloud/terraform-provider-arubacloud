resource "arubacloud_subnet" "basic" {
  name       = "basic-subnet"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  vpc_id     = "your-vpc-id"  # Replace with your VPC ID
  type       = "Advanced"  # Required: "Basic" or "Advanced"
  network = {
    address = "10.0.1.0/24"  # CIDR notation
  }
  dhcp = {
    enabled = true  # Required for Advanced type subnets
    range = {
      start = "10.0.1.10"
      count = 100
    }
    routes = [
      {
        address = "0.0.0.0/0"
        gateway = "10.0.1.1"
      }
    ]
    dns = ["8.8.8.8", "8.8.4.4"]
  }
  tags = ["network", "test"]
}
