# Step 4: Create Container Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Container Registry - Private container registry for Docker images
resource "arubacloud_containerregistry" "test" {
  name       = "test-container-registry"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["container", "registry", "test"]

  # Network configuration
  network = {
    public_ip_uri_ref      = arubacloud_elasticip.container_registry.uri
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_ref         = arubacloud_subnet.test.uri
    security_group_uri_ref = arubacloud_securitygroup.container_registry.uri
  }

  # Storage configuration
  storage = {
    block_storage_uri_ref = arubacloud_blockstorage.container_registry.uri
  }

  # Settings configuration
  settings = {
    concurrent_users_flavor = "small"  # Options: small, medium, large
    admin_user              = "adminuser"
  }

  # Optional fields
  billing_period = "Hour"
}

# KaaS (Kubernetes as a Service) - Managed Kubernetes cluster
resource "arubacloud_kaas" "test" {
  name       = "test-kaas"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["test-kaas", "container", "kubernetes"]

  # Network configuration
  network = {
    # Use URI references for VPC and Subnet
    vpc_uri_ref    = arubacloud_vpc.test.uri
    subnet_uri_ref = arubacloud_subnet.test.uri

    # Security group name (created afterwards)
    security_group_name = "kaas-sg"

    # Node CIDR configuration
    node_cidr = {
      address = "192.168.0.0 / 24"  # CIDR notation for node network
      name    = "kaas-node-cidr"
    }

    # Pod CIDR configuration
    pod_cidr = "10.0.0.0/16"  # CIDR notation for pod network
  }

  # Settings configuration
  settings = {
    kubernetes_version = "1.33.2"  # Kubernetes version (see https://api.arubacloud.com/docs/metadata#kubernetes-version)

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
    controlplane_ha = true
  }

  # Optional fields
  billing_period = "Hour"
}
