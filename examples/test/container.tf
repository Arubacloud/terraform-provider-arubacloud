# Copyright (c) HashiCorp, Inc.

# KaaS example
resource "arubacloud_kaas" "example" {
  name       = "example-kaas"
  location   = "ITBG-Bergamo"
  tags       = ["k8s", "test"]
  project_id = arubacloud_project.example.id

  # Use URI references for VPC and Subnet
  vpc_uri_ref    = arubacloud_vpc.example.uri
  subnet_uri_ref = arubacloud_subnet.example.uri

  node_cidr = {
    address = "10.0.2.0/24"
    name    = "kaas-node-cidr"
  }
  security_group_name = arubacloud_securitygroup.example.name
  kubernetes_version  = "1.28.0"
  node_pools = [
    {
      name        = "pool-1"
      nodes       = 2
      instance    = "c2.medium"
      zone        = "ITBG-1"
      autoscaling = true
      min_count   = 1
      max_count   = 5
    },
    {
      name        = "pool-2"
      nodes       = 1
      instance    = "c2.large"
      zone        = "ITBG-2"
      autoscaling = false
    }
  ]
  ha             = true
  billing_period = "Hour"
  pod_cidr       = "10.0.3.0/24"
}

#container registry example
resource "arubacloud_containerregistry" "example" {
  name                  = "example-registry"
  location              = "ITBG-Bergamo"
  tags                  = ["container", "test"]
  project_id            = arubacloud_project.example.id
  public_ip_uri_ref     = arubacloud_elasticip.example.uri
  vpc_uri_ref           = arubacloud_vpc.example.uri
  subnet_uri_ref        = arubacloud_subnet.example.uri
  security_group_uri_ref = arubacloud_securitygroup.example.uri
  block_storage_uri_ref  = arubacloud_blockstorage.example.uri
  billing_period        = "Hour"
  admin_user            = "adminuser"
}
