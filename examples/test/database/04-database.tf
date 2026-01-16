# Step 3: Create Database Resources
# Note: Adjust location and zone to match your ArubaCloud region

# DBaaS (Database as a Service) - Managed database instance
resource "arubacloud_dbaas" "test" {
  name                  = "test-dbaas"
  location              = "ITBG-Bergamo"  # Change to your region
  zone                  = "ITBG-1"  # Change to your zone
  tags                  = ["dbaas", "database", "test"]
  project_id            = arubacloud_project.test.id
  engine_id             = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor                = "DBO2A8"    # 2 CPU, 8GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
  # Storage configuration
  storage = {
    size_gb = 10  # Storage size in GB
    autoscaling = {
      enabled         = true
      available_space = 2  # Minimum threshold in GB - autoscaling triggers when available space falls below this
      step_size       = 5   # Amount in GB to increase storage by when autoscaling triggers
    }
  }
  
  # Network configuration
  network = {
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_ref         = arubacloud_subnet.test.uri
    security_group_uri_ref = arubacloud_securitygroup.dbaas.uri
    elastic_ip_uri_ref     = arubacloud_elasticip.dbaas.uri
  }
  
  # Billing period
  billing_period = "Hour"
}

# Database - Create a database within the DBaaS instance
resource "arubacloud_database" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  name       = "testdb"
}

# DBaaS User - Create a database user
# Password must be 8-20 characters, using at least one number, one uppercase letter, 
# one lowercase letter, and one special character. Spaces are not allowed.
# The password must be base64 encoded using the base64encode() function.
resource "arubacloud_dbaasuser" "test" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.test.id
  username   = "restapi"
  password   = base64encode("Prova123456789AC!")  # In production, use a secure password or variable
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
