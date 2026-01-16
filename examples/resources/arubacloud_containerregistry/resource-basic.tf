# Container Registry Example
# Note: This example assumes you have already created VPC, Subnet, Security Group, Elastic IP, and Block Storage resources

resource "arubacloud_containerregistry" "example" {
  name       = "example-container-registry"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  tags       = ["container", "test"]

  # Network configuration
  network = {
    public_ip_uri_ref      = arubacloud_elasticip.example.uri
    vpc_uri_ref            = arubacloud_vpc.example.uri
    subnet_uri_ref         = arubacloud_subnet.example.uri
    security_group_uri_ref = arubacloud_securitygroup.example.uri
  }

  # Storage configuration
  storage = {
    block_storage_uri_ref = arubacloud_blockstorage.example.uri
  }

  # Settings configuration
  settings = {
    concurrent_users_flavor = "medium"  # Options: small, medium, large
    admin_user       = "adminuser"
  }

  # Optional fields
  billing_period = "Hour"
}
