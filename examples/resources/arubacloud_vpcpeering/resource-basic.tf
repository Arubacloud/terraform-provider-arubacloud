resource "arubacloud_vpcpeering" "example" {
  name       = "example-vpc-peering"
  location   = "example-location"
  tags       = ["tag1", "tag2"]
  peer_vpc   = "peer-vpc-id"
}
