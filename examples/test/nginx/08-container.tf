# Step 8: Create Container Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Elastic IP - Public IP address for the container registry
resource "arubacloud_elasticip" "container_registry" {
  name           = "test-container-registry-elastic-ip"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "container", "registry", "test"]
}

# Security Group - For container registry
resource "arubacloud_securitygroup" "container_registry" {
  name       = "test-container-registry-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "container", "registry", "test"]
}

# Security Rule - Allow HTTPS (port 443) from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "container_registry_https" {
  name              = "test-container-registry-https-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.container_registry.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "443"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Security Rule - Allow all outbound traffic (default egress)
resource "arubacloud_securityrule" "container_registry_egress" {
  name              = "test-container-registry-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.container_registry.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"  # Will be automatically ignored for ANY/ICMP protocols
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Block Storage - Storage disk for container registry
resource "arubacloud_blockstorage" "container_registry" {
  name          = "test-container-registry-storage"
  project_id    = arubacloud_project.test.id
  location      = "ITBG-Bergamo"  # Change to your region
  size_gb       = 100
  billing_period = "Hour"
  zone          = "ITBG-1"  # Change to your zone
  type          = "Standard"
  tags          = ["storage", "container", "registry", "test"]
}

# Container Registry - Private container registry for Docker images
resource "arubacloud_containerregistry" "test" {
  name                  = "test-container-registry"
  location              = "ITBG-Bergamo"  # Change to your region
  project_id            = arubacloud_project.test.id
  tags                  = ["container", "registry", "test"]

  # Use URI references for all required resources
  public_ip_uri_ref     = arubacloud_elasticip.container_registry.uri
  vpc_uri_ref           = arubacloud_vpc.test.uri
  subnet_uri_ref        = arubacloud_subnet.test.uri
  security_group_uri_ref = arubacloud_securitygroup.container_registry.uri
  block_storage_uri_ref  = arubacloud_blockstorage.container_registry.uri

  # Optional fields
  billing_period = "Hour"
  admin_user     = "adminuser"
}

# KaaS (Kubernetes as a Service) - Managed Kubernetes cluster
resource "arubacloud_kaas" "test" {
  name       = "test-kaas"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["k8s", "kubernetes", "test"]

  # Use URI references for VPC and Subnet
  vpc_uri_ref    = arubacloud_vpc.test.uri
  subnet_uri_ref = arubacloud_subnet.test.uri

  # Node CIDR configuration
  node_cidr = {
    address = "10.0.0.0/24"  # CIDR notation for node network
    name    = "kaas-node-cidr"
  }

  # Security group name (must match existing security group)
  security_group_name = "test-kaas-security-group"
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
  pod_cidr       = "10.0.3.0/24"  # CIDR notation for pod network
}
