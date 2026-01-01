# Step 3: Create Network Resources
# Note: Adjust location and zone to match your ArubaCloud region

# VPC - Virtual Private Cloud
resource "arubacloud_vpc" "test" {
  name       = "test-vpc"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["network", "test"]
}

# Subnet - Depends on VPC (will wait for VPC to be active)
resource "arubacloud_subnet" "test" {
  name       = "test-subnet"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  type       = "Advanced"  # Required: "Basic" or "Advanced"
  network = {
    address = "10.0.1.0/24"  # CIDR notation
  }
  tags       = ["network", "test"]
}


# Security Group - Depends on VPC
resource "arubacloud_securitygroup" "test" {
  name       = "test-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "test"]
}

# Elastic IP - Public IP address for the cloud server
resource "arubacloud_elasticip" "test" {
  name           = "test-elastic-ip"
  location       = "ITBG-Bergamo"
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "test"]
}

# Security Rule - Allow SSH from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "test" {
  name              = "test-ssh-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.test.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "22"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

output "vpc_id" {
  value       = arubacloud_vpc.test.id
  description = "The ID of the created VPC"
}

output "subnet_id" {
  value       = arubacloud_subnet.test.id
  description = "The ID of the created subnet"
}

output "security_group_id" {
  value       = arubacloud_securitygroup.test.id
  description = "The ID of the created security group"
}

output "elastic_ip_id" {
  value       = arubacloud_elasticip.test.id
  description = "The ID of the created Elastic IP"
}

