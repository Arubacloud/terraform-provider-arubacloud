# KaaS (Kubernetes as a Service) Example
# Note: This example assumes you have already created VPC and Subnet resources

resource "arubacloud_kaas" "basic" {
  name       = "basic-kaas"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = "your-project-id"  # Replace with your project ID
  tags       = ["k8s", "test"]

  # Use URI references for VPC and Subnet
  vpc_uri_ref    = arubacloud_vpc.example.uri
  subnet_uri_ref = arubacloud_subnet.example.uri

  # Node CIDR configuration
  node_cidr = {
    address = "10.0.2.0/24"  # CIDR notation
    name    = "kaas-node-cidr"
  }

  security_group_name = "kaas-security-group"
  kubernetes_version  = "1.33.2"  # Kubernetes version (see https://api.arubacloud.com/docs/metadata#kubernetes-version)

  # Node pools configuration
  # Using KaaS flavor K2A4: 2 CPU, 4GB RAM, 40GB storage
  # See https://api.arubacloud.com/docs/metadata#kaas-flavors for available flavors
  node_pools = [
    {
      name        = "pool-1"
      nodes       = 2
      instance    = "K2A4"  # KaaS flavor: 2 CPU, 4GB RAM, 40GB storage
      zone        = "ITBG-1"
      autoscaling = true
      min_count   = 1
      max_count   = 5
    }
  ]

  # Optional fields
  ha             = true
  billing_period = "Hour"
  pod_cidr       = "10.0.3.0/24"
}
