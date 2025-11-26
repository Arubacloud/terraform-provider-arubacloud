resource "arubacloud_vpc_peering_route" "basic" {
  vpc_peering_id = "vpc-peering-id"
  destination_cidr = "10.0.0.0/16"
}
