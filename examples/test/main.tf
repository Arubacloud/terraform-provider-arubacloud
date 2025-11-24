terraform {
    required_providers {
      arubacloud = {
        source  = "hashicorp/arubacloud"
      }
    }
    
}

resource "arubacloud_securityrule" "example" {
  name              = "example-security-rule"
  location          = "ITBG-Bergamo"
  project_id        = arubacloud_project.example.id
  vpc_id            = arubacloud_vpc.example.id
  security_group_id = arubacloud_securitygroup.example.id
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


