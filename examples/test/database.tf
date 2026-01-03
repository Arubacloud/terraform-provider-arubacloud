# Copyright (c) HashiCorp, Inc.

# DBaaS Example
resource "arubacloud_dbaas" "example" {
  name                  = "example-dbaas"
  location              = "ITBG-Bergamo"
  tags                  = ["dbaas", "test"]
  project_id            = arubacloud_project.example.id
  engine_id             = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor                = "DBO2A4"    # 2 CPU, 4GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
  # Required network resources (URI references)
  vpc_uri_ref            = arubacloud_vpc.example.uri
  subnet_uri_ref         = arubacloud_subnet.example.uri
  security_group_uri_ref = arubacloud_securitygroup.example.uri
  
  # Optional Elastic IP (URI reference)
  elastic_ip_uri_ref = arubacloud_elasticip.example.uri
  
  # Optional autoscaling configuration
  autoscaling = {
    enabled         = true
    available_space = 100  # GB
    step_size       = 10   # GB
  }
}

#Database Example
resource "arubacloud_database" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  name     = "exampledb"
}

# DBaaS User Example
resource "arubacloud_dbaasuser" "example" {
  dbaas_id = arubacloud_dbaas.example.id
  username = "dbuser"
  password = "supersecretpassword"
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
