# Step 4: Create managed MySQL and application database/user

resource "arubacloud_dbaas" "wordpress" {
  name       = "wp-dbaas"
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"
  tags       = ["dbaas", "mysql", "wordpress", "test"]
  project_id = arubacloud_project.test.id
  engine_id  = "mysql-8.0"
  flavor     = "DBO2A8"

  storage = {
    size_gb = 10
    autoscaling = {
      enabled         = true
      available_space = 2
      step_size       = 5
    }
  }

  network = {
    vpc_uri_ref            = arubacloud_vpc.test.uri
    subnet_uri_ref         = arubacloud_subnet.test.uri
    security_group_uri_ref = arubacloud_securitygroup.dbaas.uri
    elastic_ip_uri_ref     = arubacloud_elasticip.dbaas.uri
  }

  billing_period = "Hour"
}

resource "arubacloud_database" "wordpress" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.wordpress.id
  name       = "wordpress"
}

resource "arubacloud_dbaasuser" "wordpress" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.wordpress.id
  username   = "wordpress"
  password   = var.database_password
}

# Grant the user access to the wordpress database (required for MySQL auth)
resource "arubacloud_databasegrant" "wordpress" {
  project_id = arubacloud_project.test.id
  dbaas_id   = arubacloud_dbaas.wordpress.id
  database   = arubacloud_database.wordpress.id
  user_id    = arubacloud_dbaasuser.wordpress.id
  role       = "liteadmin"
}
