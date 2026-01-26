resource "arubacloud_cloudserver" "basic" {
  name       = "example-cloudserver"
  location   = "ITBG-Bergamo"
  project_id = arubacloud_project.example.id
  zone       = "ITBG-1"
  tags       = ["compute", "example"]

  network = {
    vpc_uri_ref            = arubacloud_vpc.example.uri
    elastic_ip_uri_ref     = arubacloud_elasticip.example.uri
    subnet_uri_refs        = [arubacloud_subnet.example.uri]
    securitygroup_uri_refs = [arubacloud_securitygroup.example.uri]
  }

  settings = {
    flavor_name      = "CSO4A8"  # 4 CPU, 8GB RAM (see https://api.arubacloud.com/docs/metadata/#cloudserver-flavors)
    key_pair_uri_ref = arubacloud_keypair.example.uri
    # Optional: cloud-init user data for bootstrapping (raw cloud-init YAML content)
    # user_data      = file("cloud-init.yaml")
  }

  storage = {
    boot_volume_uri_ref = arubacloud_blockstorage.example.uri
  }
}
