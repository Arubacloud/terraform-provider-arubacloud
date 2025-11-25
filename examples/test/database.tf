# Copyright (c) HashiCorp, Inc.

# DBaaS Example
resource "arubacloud_dbaas" "example" {
  name           = "example-dbaas"
  location       = "ITBG-Bergamo"
  tags           = ["dbaas", "test"]
  project_id     = arubacloud_project.example.id
  engine         = "mysql-8.0"
  zone           = "ITBG-1"
  flavor         = "db.t3.medium"
  storage_size   = 50
  billing_period = "Hour"
  network {
    vpc_id            = arubacloud_vpc.example.id
    subnet_id         = arubacloud_subnet.example.id
    security_group_id = arubacloud_securitygroup.example.id
    elastic_ip_id     = arubacloud_elasticip.example.id
  }
  autoscaling {
    enabled         = true
    available_space = 100
    step_size       = 10
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
