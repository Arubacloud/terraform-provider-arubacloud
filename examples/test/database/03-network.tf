# Step 2: Create Network Resources
# Note: Adjust location and zone to match your ArubaCloud region

# VPC - Virtual Private Cloud
resource "arubacloud_vpc" "test" {
  name       = "database-test-vpc"
  location   = "ITBG-Bergamo"  # Change to your region
  project_id = arubacloud_project.test.id
  tags       = ["network", "database", "test"]
}

# Subnet - Depends on VPC (will wait for VPC to be active)
resource "arubacloud_subnet" "test" {
  name       = "database-test-subnet"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  type       = "Basic"  # Required: "Basic" or "Advanced"
  tags       = ["network", "database", "test"]
}

# Security Group - For DBaaS
resource "arubacloud_securitygroup" "dbaas" {
  name       = "dbaas-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "dbaas", "database", "test"]
}

# Security Rule - Allow MySQL (port 3306) from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "dbaas_mysql" {
  name              = "dbaas-mysql-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.dbaas.id
  properties = {
    direction = "Ingress"
    protocol  = "TCP"
    port      = "3306"
    target = {
      kind  = "Ip"
      value = "0.0.0.0/0"
    }
  }
}

# Security Rule - Allow all outbound traffic (default egress)
resource "arubacloud_securityrule" "dbaas_egress" {
  name              = "dbaas-egress-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.test.id
  vpc_id            = arubacloud_vpc.test.id
  security_group_id = arubacloud_securitygroup.dbaas.id
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

# Elastic IP - Public IP address for the DBaaS
resource "arubacloud_elasticip" "dbaas" {
  name           = "dbaas-elastic-ip"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "dbaas", "database", "test"]
}
