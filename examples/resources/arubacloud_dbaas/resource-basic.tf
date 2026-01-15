resource "arubacloud_dbaas" "basic" {
  name       = "basic-dbaas"
  location   = "ITBG-Bergamo"
  zone       = "ITBG-1"  # Change to your zone
  project_id = arubacloud_project.example.id
  engine_id      = "mysql-8.0"  # See https://api.arubacloud.com/docs/metadata/#dbaas-engines
  flavor         = "DBO2A4"     # 2 CPU, 4GB RAM (see https://api.arubacloud.com/docs/metadata/#dbaas-flavors)
  
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
  
  tags = ["dbaas", "test"]
}
