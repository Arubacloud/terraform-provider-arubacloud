# Copyright (c) HashiCorp, Inc.

# DBaaS Example
resource "arubacloud_dbaas" "example" {
  name                  = "example-dbaas"
  location              = "ITBG-Bergamo"
  zone                  = "ITBG-1"  # Change to your zone
  tags                  = ["dbaas", "test"]
  project_id            = arubacloud_project.example.id
  engine_id             = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor                = "DBO2A4"    # 2 CPU, 4GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
  # Storage configuration
  storage = {
    size_gb = 100  # Storage size in GB
    autoscaling = {
      enabled         = true
      available_space = 100  # Minimum threshold in GB - autoscaling triggers when available space falls below this
      step_size       = 10   # Amount in GB to increase storage by when autoscaling triggers
    }
  }
  
  # Network configuration
  network = {
    vpc_uri_ref            = arubacloud_vpc.example.uri
    subnet_uri_ref         = arubacloud_subnet.example.uri
    security_group_uri_ref = arubacloud_securitygroup.example.uri
    elastic_ip_uri_ref     = arubacloud_elasticip.example.uri
  }
  
  # Billing period
  billing_period = "Hour"
}

#Database Example
resource "arubacloud_database" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  name     = "exampledb"
}

# DBaaS User Example
# Password must be 8-20 characters, using at least one number, one uppercase letter, 
# one lowercase letter, and one special character. Spaces are not allowed.
# The password must be base64 encoded using the base64encode() function.
resource "arubacloud_dbaasuser" "example" {
  project_id = arubacloud_project.example.id
  dbaas_id   = arubacloud_dbaas.example.id
  username   = "dbuser"
  password   = base64encode("SuperSecret123!")  # In production, use a secure password or variable
}

# Database Grant Example
resource "arubacloud_databasegrant" "example" {
  database = arubacloud_database.example.id
  user_id  = arubacloud_dbaasuser.example.id
  role     = "admin"
}

# Database Backup Example
resource "arubacloud_databasebackup" "example" {
  name           = "example-db-backup"
  location       = "ITBG-Bergamo"
  tags           = ["dbbackup", "test"]
  zone           = "ITBG-1"
  dbaas_id       = arubacloud_dbaas.example.id
  database       = arubacloud_database.example.id
  billing_period = "Hour"
}
