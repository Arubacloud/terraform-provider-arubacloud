resource "arubacloud_cloudserver" "basic" {
  name                  = "example-cloudserver"
  location              = "ITBG-Bergamo"
  project_id            = arubacloud_project.example.id
  zone                  = "ITBG-1"
  vpc_uri_ref           = arubacloud_vpc.example.uri
  flavor_name           = "CSO8A16"  # 8 CPU, 16GB RAM (see https://api.arubacloud.com/docs/metadata/#cloudserver-flavors)
  elastic_ip_uri_ref    = arubacloud_elasticip.example.uri
  boot_volume           = "LU22-001"  # Image ID or arubacloud_blockstorage.example.id
  key_pair_uri_ref      = arubacloud_keypair.example.uri
  subnet_uri_refs       = [arubacloud_subnet.example.uri]
  securitygroup_uri_refs = [arubacloud_securitygroup.example.uri]
  tags                  = ["compute", "example"]
}
