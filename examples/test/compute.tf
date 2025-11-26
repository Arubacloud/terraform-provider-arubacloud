# Copyright (c) HashiCorp, Inc.

# KeyPair Example Resource
resource "arubacloud_keypair" "example" {
  name     = "example-keypair"
  location = "ITBG-Bergamo"
  tags     = ["keypair", "test"]
  value    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
}

# CloudServer Example Resource
resource "arubacloud_cloudserver" "example" {
  name            = "example-cloudserver"
  location        = "ITBG-Bergamo"
  project_id      = arubacloud_project.example.id
  zone            = "ITBG-1"
  vpc_id          = arubacloud_vpc.example.id
  flavor_name     = "c2.medium"
  elastic_ip_id   = arubacloud_elasticip.example.id
  boot_volume     = arubacloud_blockstorage.example.id
  key_pair_id     = arubacloud_keypair.example.id
  subnets         = [arubacloud_subnet.example.id, arubacloud_subnet2.example.id]
  securitygroups  = [arubacloud_securitygroup.example.id, arubacloud_securitygroup2.example.id]
}

