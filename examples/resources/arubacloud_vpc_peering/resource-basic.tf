resource "arubacloud_vpc_peering" "basic" {
  name = "basic-vpc-peering"
  vpc_id = "vpc-id"
  peer_vpc_id = "peer-vpc-id"
}
