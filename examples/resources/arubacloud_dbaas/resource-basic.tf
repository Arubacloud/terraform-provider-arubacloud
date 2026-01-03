resource "arubacloud_dbaas" "basic" {
  name       = "basic-dbaas"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.example.id
  engine_id  = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor     = "DBO2A4"     # 2 CPU, 4GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
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
  
  tags = ["dbaas", "test"]
}
