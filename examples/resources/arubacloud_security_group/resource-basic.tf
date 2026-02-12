resource "arubacloud_securitygroup" "basic" {
  name       = "basic-security-group"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  vpc_id     = "your-vpc-id"  # Replace with your VPC ID
  tags       = ["security", "test"]
}
