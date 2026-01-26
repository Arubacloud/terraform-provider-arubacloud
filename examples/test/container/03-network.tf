# Step 2: Create Network Resources
# Note: Adjust location and zone to match your ArubaCloud region

# VPC - Virtual Private Cloud
resource "arubacloud_vpc" "test" {
  name       = "container-test-vpc"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["network", "container", "test"]
}

# Subnet - Depends on VPC (will wait for VPC to be active)
resource "arubacloud_subnet" "test" {
  name       = "container-test-subnet"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"  # Required: "Basic" or "Advanced"
  tags       = ["network", "container", "test"]
}

# Security Group - For container registry
resource "arubacloud_securitygroup" "container_registry" {
  name       = "container-registry-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "container", "registry", "test"]
}

# Security Group - For KaaS
resource "arubacloud_securitygroup" "kaas" {
  name       = "container-kaas-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "kaas", "kubernetes", "test"]
}

# Security Rule - Allow HTTPS (port 443) from anywhere for container registry
resource "arubacloud_securityrule" "container_registry_https" {
  name              = "container-registry-https-rule"
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

# Security Rule - Allow all outbound traffic for container registry
resource "arubacloud_securityrule" "container_registry_egress" {
  name              = "container-registry-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.container_registry.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Security Rule - Allow all outbound traffic for KaaS
resource "arubacloud_securityrule" "kaas_egress" {
  name              = "kaas-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.kaas.id
  properties = {
    direction = "Egress"
    protocol  = "ANY"
    port      = "*"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Elastic IP - Public IP address for the container registry
resource "arubacloud_elasticip" "container_registry" {
  name           = "container-registry-elastic-ip"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "container", "registry", "test"]
}
