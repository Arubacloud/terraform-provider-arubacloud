# Copyright (c) HashiCorp, Inc.

# KeyPair Example Resource
resource "arubacloud_keypair" "example" {
  name     = "example-keypair"
  location = "ITBG-Bergamo"
  tags     = ["keypair", "test"]
  value    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
}

# CloudServer Example Resource
# Note: vpc_uri_ref, subnet_uri_refs, securitygroup_uri_refs, key_pair_uri_ref, and elastic_ip_uri_ref use URI references
resource "arubacloud_cloudserver" "example" {
  name                 = "example-cloudserver"
  location             = "ITBG-Bergamo"
  project_id           = arubacloud_project.example.id
  zone                 = "ITBG-1"
  vpc_uri_ref          = arubacloud_vpc.example.uri                    # URI reference
  flavor_name          = "c2.medium"
  elastic_ip_uri_ref   = arubacloud_elasticip.example.uri               # URI reference
  boot_volume          = arubacloud_blockstorage.example.id             # Boot volume ID or image ID like "LU22-001"
  key_pair_uri_ref     = arubacloud_keypair.example.uri                 # URI reference
  subnet_uri_refs      = [arubacloud_subnet.example.uri, arubacloud_subnet.example2.uri]  # URI references
  securitygroup_uri_refs = [arubacloud_securitygroup.example.uri, arubacloud_securitygroup.example2.uri]  # URI references
}

