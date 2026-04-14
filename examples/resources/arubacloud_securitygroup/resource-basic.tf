resource "arubacloud_securitygroup" "example" {
  name       = "example-security-group"
  location   = "ITBG-Bergamo"
  project_id = "your-project-id"
  vpc_id     = "your-vpc-id"
  tags       = ["security", "example"]
}
