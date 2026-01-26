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
  type       = "Basic"  # Required: "Basic" or "Advanced"
#  network = {
#    address = "10.0.1.0/24"  # CIDR notation
#    dhcp = {
#      enabled = true  # Required for Advanced type subnets
#      range = {
#        start = "10.0.1.10"
#        count = 100
#      }
#      routes = []
#      dns = ["8.8.8.8", "8.8.4.4"]
#    }
#  }
  tags = ["network", "test", "updated"]
}


# Security Group - Depends on VPC
resource "arubacloud_securitygroup" "test" {
  name       = "test-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "test", "updated"]
}

# Elastic IP - Public IP address for the cloud server
resource "arubacloud_elasticip" "test" {
  name           = "test-elastic-ip"
  location       = "ITBG-Bergamo"
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "test","updated"]
}

# Security Rule - Allow HTTP from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "test" {
  name              = "test-http-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.test.id
#  tags              = ["security", "test", "http"]
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "80"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Security Rule - Allow SSH from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "ssh" {
  name              = "test-ssh-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.test.id
#  tags              = ["security", "test", "ssh"]
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

# Security Rule - Allow all outbound traffic (default egress)
resource "arubacloud_securityrule" "default_egress" {
  name              = "test-default-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.test.id
#  tags              = ["security", "test", "egress"]
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
