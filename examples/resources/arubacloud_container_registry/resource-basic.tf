# Container Registry Example
# Note: This example assumes you have already created VPC, Subnet, Security Group, Elastic IP, and Block Storage resources

resource "arubacloud_container_registry" "basic" {
  name       = "basic-container-registry"
  location   = "ITBG-Bergamo"
  project_id = "your-project-id"
  tags       = ["container", "test"]

  network = {
    public_ip_uri_ref      = arubacloud_elasticip.example.uri
    vpc_uri_ref            = arubacloud_vpc.example.uri
    subnet_uri_ref         = arubacloud_subnet.example.uri
    security_group_uri_ref = arubacloud_securitygroup.example.uri
  }

  storage = {
    block_storage_uri_ref = arubacloud_blockstorage.example.uri
  }

  billing_period = "Hour"

  settings = {
    admin_user               = "adminuser"
    concurrent_users_flavor  = "Medium"  # Options: Small, Medium, HighPerf
  }
}
