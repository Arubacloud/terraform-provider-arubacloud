# Step 9: Create Database Resources
# Note: Adjust location and zone to match your ArubaCloud region

# Elastic IP - Public IP address for the DBaaS
resource "arubacloud_elasticip" "dbaas" {
  name           = "test-dbaas-elastic-ip"
  location       = "ITBG-Bergamo"  # Change to your region
  project_id     = arubacloud_project.test.id
  billing_period = "hourly"
  tags           = ["public", "dbaas", "database", "test"]
}

# Security Group - For DBaaS
resource "arubacloud_securitygroup" "dbaas" {
  name       = "test-dbaas-security-group"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.test.id
  vpc_id     = arubacloud_vpc.test.id
  tags       = ["security", "dbaas", "database", "test"]
}

# Security Rule - Allow MySQL (port 3306) from anywhere (0.0.0.0/0)
resource "arubacloud_securityrule" "dbaas_mysql" {
  name              = "test-dbaas-mysql-rule"
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
  name              = "test-dbaas-egress-rule"
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

# DBaaS (Database as a Service) - Managed database instance
resource "arubacloud_dbaas" "test" {
  name                  = "test-dbaas"
  location              = "ITBG-Bergamo"  # Change to your region
  tags                  = ["dbaas", "database", "test"]
  project_id            = arubacloud_project.test.id
  engine_id             = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor                = "DBO2A4"    # 2 CPU, 4GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
  # Required network resources (URI references)
  vpc_uri_ref            = arubacloud_vpc.test.uri
  subnet_uri_ref         = arubacloud_subnet.test.uri
  security_group_uri_ref = arubacloud_securitygroup.dbaas.uri
  
  # Optional Elastic IP (URI reference)
  elastic_ip_uri_ref = arubacloud_elasticip.dbaas.uri
  
  # Optional autoscaling configuration
  autoscaling = {
    enabled         = true
    available_space = 100  # GB
    step_size       = 10   # GB
  }
}

# Database - Create a database within the DBaaS instance
resource "arubacloud_database" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  name       = "testdb"
}

# DBaaS User - Create a database user
resource "arubacloud_dbaasuser" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  username   = "dbuser"
  password   = "supersecretpassword123!"  # In production, use a secure password or variable
}

# Database Grant - Associate the user with the database and grant permissions
# Note: This resource is currently disabled in the provider due to GrantRole type conversion issues
# Uncomment when the provider issue is resolved
# resource "arubacloud_databasegrant" "test" {
#   project_id = arubacloud_project.test.id
#   dbaas_id   = arubacloud_dbaas.test.id
#   database   = arubacloud_database.test.id
#   user_id    = arubacloud_dbaasuser.test.id
#   role       = "admin"  # Role: read, write, or admin
# }
